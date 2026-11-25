package repository

import (
	"context"
	"database/sql"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/tenant"
)

// DomainRepository: a store's custom domains.
//
// store_domains is NOT under row-level security, because finding the store of
// a host name must work before any store is known. So every query here names
// the current store explicitly (store_id = the request's store), and a store
// can never see or change another store's domains.
type DomainRepository struct {
	db *sql.DB
}

func NewDomainRepository(db *sql.DB) *DomainRepository { return &DomainRepository{db: db} }

const domainColumns = `id, domain, is_primary, verified_at IS NOT NULL, verified_at, verify_token, created_at`

func scanDomain(s scanner) (models.StoreDomain, error) {
	var d models.StoreDomain
	err := s.Scan(&d.ID, &d.Domain, &d.IsPrimary, &d.Verified, &d.VerifiedAt, &d.VerifyToken, &d.CreatedAt)
	return d, err
}

func (r *DomainRepository) List(ctx context.Context) ([]models.StoreDomain, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+domainColumns+` FROM store_domains WHERE store_id = $1 ORDER BY is_primary DESC, id`, tenant.ID(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.StoreDomain{}
	for rows.Next() {
		d, err := scanDomain(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DomainRepository) Get(ctx context.Context, id int64) (models.StoreDomain, error) {
	d, err := scanDomain(r.db.QueryRowContext(ctx,
		`SELECT `+domainColumns+` FROM store_domains WHERE id = $1 AND store_id = $2`, id, tenant.ID(ctx)))
	return d, apperr.FromDB(err, "domain not found")
}

func (r *DomainRepository) Add(ctx context.Context, domain, token string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO store_domains (store_id, domain, verify_token) VALUES ($1, $2, $3) RETURNING id`,
		tenant.ID(ctx), domain, token).Scan(&id)
	if apperr.IsUniqueViolation(err, "store_domains_domain_key") {
		return 0, apperr.Conflict("this domain is already connected to a store")
	}
	return id, err
}

func (r *DomainRepository) MarkVerified(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE store_domains SET verified_at = COALESCE(verified_at, NOW()) WHERE id = $1 AND store_id = $2`,
		id, tenant.ID(ctx))
	return expectOneRow(res, err, "domain not found")
}

// SetPrimary makes a verified domain the store's main address (used in emails).
func (r *DomainRepository) SetPrimary(ctx context.Context, id int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var verified bool
		if err := tx.QueryRowContext(ctx,
			`SELECT verified_at IS NOT NULL FROM store_domains WHERE id = $1 AND store_id = $2 FOR UPDATE`,
			id, tenant.ID(ctx)).Scan(&verified); err != nil {
			return apperr.FromDB(err, "domain not found")
		}
		if !verified {
			return apperr.Conflict("verify the domain before making it the main address")
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE store_domains SET is_primary = FALSE WHERE store_id = $1 AND is_primary`, tenant.ID(ctx)); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE store_domains SET is_primary = TRUE WHERE id = $1`, id)
		return err
	})
}

func (r *DomainRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM store_domains WHERE id = $1 AND store_id = $2`, id, tenant.ID(ctx))
	return expectOneRow(res, err, "domain not found")
}
