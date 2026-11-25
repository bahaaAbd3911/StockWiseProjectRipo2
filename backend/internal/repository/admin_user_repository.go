package repository

import (
	"context"
	"database/sql"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

// AdminUserRepository is user management for admins.
type AdminUserRepository struct {
	db *sql.DB
}

func NewAdminUserRepository(db *sql.DB) *AdminUserRepository { return &AdminUserRepository{db: db} }

// adminUserColumns = userColumns (prefixed u.) + shopping totals.
const adminUserColumns = `
	u.id, u.first_name, u.last_name, u.email, u.phone, u.profile_image_url, u.role, u.is_active, u.token_version,
	u.email_verified_at IS NOT NULL, COALESCE(u.created_at, NOW()), COALESCE(u.updated_at, NOW()),
	(SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id AND o.status <> 'cancelled'),
	(SELECT COALESCE(SUM(o.total_amount), 0) FROM orders o WHERE o.user_id = u.id AND o.status <> 'cancelled')`

func scanAdminUser(s scanner) (models.AdminUser, error) {
	var a models.AdminUser
	u, err := scanUser(s, &a.OrdersCount, &a.TotalSpent)
	a.User, a.IsActive = u, u.IsActive
	return a, err
}

func (r *AdminUserRepository) List(ctx context.Context, f models.AdminUserFilter) (models.Page[models.AdminUser], error) {
	var w whereBuilder
	if f.Search != "" {
		w.add(`(u.first_name || ' ' || u.last_name ILIKE ? OR u.email ILIKE ? OR COALESCE(u.phone, '') ILIKE ?)`,
			"%"+escapeLike(f.Search)+"%")
	}
	if f.Role != "" {
		w.add(`u.role = ?`, f.Role)
	}
	page := models.Page[models.AdminUser]{Items: []models.AdminUser{}, Limit: f.Limit, Offset: f.Offset}
	where := w.sql()
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users u`+where, w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	limit, offset := w.next(f.Limit), w.next(f.Offset)
	rows, err := r.db.QueryContext(ctx, `SELECT `+adminUserColumns+` FROM users u`+where+
		` ORDER BY u.created_at DESC NULLS LAST, u.id DESC LIMIT `+limit+` OFFSET `+offset, w.args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		a, err := scanAdminUser(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, a)
	}
	return page, rows.Err()
}

func (r *AdminUserRepository) Get(ctx context.Context, id int64) (models.AdminUser, error) {
	a, err := scanAdminUser(r.db.QueryRowContext(ctx, `SELECT `+adminUserColumns+` FROM users u WHERE u.id = $1`, id))
	return a, apperr.FromDB(err, "user not found")
}

var errEmailTaken = apperr.Validation(map[string]string{"email": "an account with this email already exists"})

// Create adds a user. Accounts made by an admin count as email-confirmed.
func (r *AdminUserRepository) Create(ctx context.Context, in models.AdminUserInput, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash, phone, role, is_active, email_verified_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, COALESCE($7, TRUE), NOW())
		RETURNING id`,
		in.FirstName, in.LastName, in.Email, passwordHash, in.Phone, in.Role, in.IsActive).Scan(&id)
	if apperr.IsUniqueViolation(err, "users_email_key") {
		return 0, errEmailTaken
	}
	return id, err
}

// Update edits a user. A new password, a role change or deactivation also
// logs the user out everywhere (token_version + 1).
func (r *AdminUserRepository) Update(ctx context.Context, id int64, in models.AdminUserInput, passwordHash *string) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE users SET
				first_name = $1, last_name = $2, email = $3, phone = NULLIF($4::text, ''), role = $5::text,
				is_active = COALESCE($6::boolean, is_active),
				password_hash = COALESCE($7::text, password_hash),
				token_version = token_version + CASE
					WHEN $7::text IS NOT NULL OR role <> $5::text OR (is_active AND $6::boolean IS FALSE) THEN 1 ELSE 0 END,
				updated_at = NOW()
			WHERE id = $8`,
			in.FirstName, in.LastName, in.Email, in.Phone, in.Role, in.IsActive, passwordHash, id)
		if apperr.IsUniqueViolation(err, "users_email_key") {
			return errEmailTaken
		}
		return expectOneRow(res, err, "user not found")
	})
}

// Delete removes an account that never ordered and has no debt history.
// Accounts with orders must stay (orders reference them), so they are
// deactivated instead. It reports which of the two happened.
func (r *AdminUserRepository) Delete(ctx context.Context, id int64) (deleted bool, err error) {
	err = database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var history bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (SELECT 1 FROM orders WHERE user_id = $1)
			    OR EXISTS (SELECT 1 FROM customer_debt_transactions WHERE user_id = $1)
			FROM users WHERE id = $1 FOR UPDATE`, id).Scan(&history); err != nil {
			return apperr.FromDB(err, "user not found")
		}
		if history {
			_, err := tx.ExecContext(ctx, `
				UPDATE users SET is_active = FALSE, token_version = token_version + 1, updated_at = NOW()
				WHERE id = $1`, id)
			return err
		}
		deleted = true
		_, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
		return err
	})
	return deleted, err
}

// ActiveAdmins counts active admin accounts other than `except`.
func (r *AdminUserRepository) ActiveAdminsExcept(ctx context.Context, except int64) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE role = 'admin' AND is_active AND id <> $1`, except).Scan(&n)
	return n, err
}
