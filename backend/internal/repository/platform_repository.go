package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/tenant"
)

// PlatformRepository is the data of the platform super admin: platform admins
// and stores. Everything here works across stores, so every method runs in
// platform mode (row-level security lets it see and write every store's
// rows). Only platform handlers, behind platform authentication, use it.
type PlatformRepository struct {
	db *sql.DB
}

func NewPlatformRepository(db *sql.DB) *PlatformRepository { return &PlatformRepository{db: db} }

// ---------- platform admins ----------

const platformAdminColumns = `id, email, first_name, last_name, is_active, token_version, created_at`

func scanPlatformAdmin(s scanner, extra ...any) (models.PlatformAdmin, error) {
	var a models.PlatformAdmin
	err := s.Scan(append([]any{&a.ID, &a.Email, &a.FirstName, &a.LastName, &a.IsActive, &a.TokenVersion, &a.CreatedAt}, extra...)...)
	return a, err
}

// AdminByEmail returns the admin and their password hash (sql.ErrNoRows if none).
func (r *PlatformRepository) AdminByEmail(ctx context.Context, email string) (models.PlatformAdmin, string, error) {
	var hash string
	a, err := scanPlatformAdmin(r.db.QueryRowContext(ctx,
		`SELECT `+platformAdminColumns+`, password_hash FROM platform_admins WHERE lower(email) = lower($1)`, email), &hash)
	return a, hash, err
}

func (r *PlatformRepository) AdminByID(ctx context.Context, id int64) (models.PlatformAdmin, error) {
	return scanPlatformAdmin(r.db.QueryRowContext(ctx,
		`SELECT `+platformAdminColumns+` FROM platform_admins WHERE id = $1`, id))
}

// CreateAdmin adds a platform admin (used by the command line tool).
func (r *PlatformRepository) CreateAdmin(ctx context.Context, email, first, last, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO platform_admins (email, first_name, last_name, password_hash) VALUES ($1, $2, $3, $4)
		RETURNING id`, email, first, last, passwordHash).Scan(&id)
	if apperr.IsUniqueViolation(err, "platform_admins_email_key") {
		return 0, apperr.Conflict("a platform admin with this email already exists")
	}
	return id, err
}

// ---------- stores ----------

const platformStoreColumns = `s.id, s.slug, s.name, s.status, s.currency, s.timezone,
	(SELECT u.email FROM users u WHERE u.store_id = s.id AND u.role = 'admin' ORDER BY u.id LIMIT 1),
	(SELECT COUNT(*) FROM products p WHERE p.store_id = s.id AND p.is_active),
	(SELECT COUNT(*) FROM users u WHERE u.store_id = s.id AND u.role = 'customer' AND u.is_active),
	(SELECT COUNT(*) FROM orders o WHERE o.store_id = s.id AND o.status <> 'cancelled'),
	(SELECT COALESCE(SUM(o.total_amount), 0) FROM orders o WHERE o.store_id = s.id AND o.status <> 'cancelled'),
	COALESCE((SELECT p.domain FROM store_domains p WHERE p.store_id = s.id AND p.is_primary AND p.verified_at IS NOT NULL), ''),
	s.created_at, s.updated_at`

func scanPlatformStore(sc scanner) (models.PlatformStore, error) {
	var s models.PlatformStore
	err := sc.Scan(&s.ID, &s.Slug, &s.Name, &s.Status, &s.Currency, &s.Timezone, &s.AdminEmail,
		&s.ProductsCount, &s.CustomersCount, &s.OrdersCount, &s.Revenue, &s.PrimaryDomain, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *PlatformRepository) ListStores(ctx context.Context, f models.StoreFilter) (models.Page[models.PlatformStore], error) {
	ctx = tenant.AsPlatform(ctx)
	var w whereBuilder
	if f.Search != "" {
		w.add(`(s.slug ILIKE ? OR s.name ILIKE ?)`, "%"+escapeLike(f.Search)+"%")
	}
	if f.Status != "" {
		w.add(`s.status = ?`, f.Status)
	}
	page := models.Page[models.PlatformStore]{Items: []models.PlatformStore{}, Limit: f.Limit, Offset: f.Offset}
	where := w.sql()
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM stores s`+where, w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	limit, offset := w.next(f.Limit), w.next(f.Offset)
	rows, err := r.db.QueryContext(ctx, `SELECT `+platformStoreColumns+` FROM stores s`+where+
		` ORDER BY s.created_at DESC, s.id DESC LIMIT `+limit+` OFFSET `+offset, w.args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		s, err := scanPlatformStore(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, s)
	}
	return page, rows.Err()
}

func (r *PlatformRepository) GetStore(ctx context.Context, id int64) (models.PlatformStore, error) {
	s, err := scanPlatformStore(r.db.QueryRowContext(tenant.AsPlatform(ctx),
		`SELECT `+platformStoreColumns+` FROM stores s WHERE s.id = $1`, id))
	return s, apperr.FromDB(err, "store not found")
}

// NewStore is what CreateStore needs besides the request.
type NewStore struct {
	AdminPasswordHash string    // an unusable random hash: the admin sets a password from the invite
	InviteTokenHash   string    // SHA-256 of the invite link's token
	InviteExpires     time.Time // when the invite link stops working
}

// CreateStore creates, in one transaction: the store, its settings row (with
// the store's name), its admin account (email counts as confirmed) and the
// admin's invite token. Returns the store id.
func (r *PlatformRepository) CreateStore(ctx context.Context, in models.CreateStoreRequest, n NewStore) (int64, error) {
	ctx = tenant.AsPlatform(ctx)
	var storeID int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, `
			INSERT INTO stores (slug, name, currency, timezone) VALUES ($1, $2, $3, $4) RETURNING id`,
			in.Slug, in.Name, in.Currency, in.Timezone).Scan(&storeID)
		if apperr.IsUniqueViolation(err, "stores_slug_key") {
			return apperr.Validation(map[string]string{"slug": "this address is already taken"})
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO store_settings (store_id, data) VALUES ($1, jsonb_build_object('storeName', $2::text))`,
			storeID, in.Name); err != nil {
			return err
		}
		var userID int64
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO users (store_id, first_name, last_name, email, password_hash, role, email_verified_at)
			VALUES ($1, $2, $3, $4, $5, 'admin', NOW()) RETURNING id`,
			storeID, in.Admin.FirstName, in.Admin.LastName, in.Admin.Email, n.AdminPasswordHash).Scan(&userID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO password_reset_tokens (store_id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`,
			storeID, userID, n.InviteTokenHash, n.InviteExpires)
		return err
	})
	return storeID, err
}

// StoreAdmin is who an invite goes to.
type StoreAdmin struct {
	Email     string
	FirstName string
}

// NewInvite replaces the invite link of a store's first admin (for "send
// the invite again") and returns who it goes to.
func (r *PlatformRepository) NewInvite(ctx context.Context, storeID int64, tokenHash string, expires time.Time) (StoreAdmin, error) {
	ctx = tenant.AsPlatform(ctx)
	var a StoreAdmin
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var userID int64
		err := tx.QueryRowContext(ctx, `
			SELECT id, email, first_name FROM users
			WHERE store_id = $1 AND role = 'admin' AND is_active ORDER BY id LIMIT 1`, storeID).Scan(&userID, &a.Email, &a.FirstName)
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.NotFound("this store has no active admin")
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO password_reset_tokens (store_id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`,
			storeID, userID, tokenHash, expires)
		return err
	})
	return a, err
}

func (r *PlatformRepository) UpdateStore(ctx context.Context, id int64, in models.UpdateStoreRequest) error {
	res, err := r.db.ExecContext(tenant.AsPlatform(ctx), `
		UPDATE stores SET name = $1, currency = $2, timezone = $3, updated_at = NOW() WHERE id = $4`,
		in.Name, in.Currency, in.Timezone, id)
	return expectOneRow(res, err, "store not found")
}

// SetStoreStatus activates or suspends a store. Suspended stores answer 403
// on every request; their data stays untouched.
func (r *PlatformRepository) SetStoreStatus(ctx context.Context, id int64, status string) error {
	res, err := r.db.ExecContext(tenant.AsPlatform(ctx),
		`UPDATE stores SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return expectOneRow(res, err, "store not found")
}
