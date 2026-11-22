package repository

import (
	"context"
	"database/sql"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

type AddressRepository struct {
	db *sql.DB
}

func NewAddressRepository(db *sql.DB) *AddressRepository { return &AddressRepository{db: db} }

const addressColumns = `
	id, user_id, address_label, recipient_name, phone, city, area, street,
	building_number, apartment_number, additional_instructions, is_default, created_at, updated_at`

func scanAddress(s scanner) (models.Address, error) {
	var a models.Address
	err := s.Scan(&a.ID, &a.UserID, &a.AddressLabel, &a.RecipientName, &a.Phone, &a.City, &a.Area, &a.Street,
		&a.BuildingNumber, &a.ApartmentNumber, &a.AdditionalInstructions, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

// lockUser serialises default-address changes for one user. Without it, two
// simultaneous requests could both try to set a default and one would fail on
// the unique_default_address_per_user index.
func lockUser(ctx context.Context, tx *sql.Tx, userID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&id)
	return apperr.FromDB(err, "user not found")
}

func (r *AddressRepository) List(ctx context.Context, userID int64) ([]models.Address, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+addressColumns+` FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Address{}
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AddressRepository) Get(ctx context.Context, userID, id int64) (models.Address, error) {
	a, err := scanAddress(r.db.QueryRowContext(ctx,
		`SELECT `+addressColumns+` FROM user_addresses WHERE id = $1 AND user_id = $2`, id, userID))
	return a, apperr.FromDB(err, "address not found")
}

// Create inserts an address. The user's FIRST address automatically becomes the
// default; otherwise it becomes default only when is_default is true.
func (r *AddressRepository) Create(ctx context.Context, userID int64, in models.AddressInput) (models.Address, error) {
	var a models.Address
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := lockUser(ctx, tx, userID); err != nil {
			return err
		}
		var hasDefault bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM user_addresses WHERE user_id = $1 AND is_default)`, userID,
		).Scan(&hasDefault); err != nil {
			return err
		}
		makeDefault := in.IsDefault || !hasDefault
		if makeDefault && hasDefault {
			if _, err := tx.ExecContext(ctx,
				`UPDATE user_addresses SET is_default = FALSE, updated_at = NOW() WHERE user_id = $1 AND is_default`, userID); err != nil {
				return err
			}
		}
		var err error
		a, err = scanAddress(tx.QueryRowContext(ctx, `
			INSERT INTO user_addresses (user_id, address_label, recipient_name, phone, city, area, street,
			                            building_number, apartment_number, additional_instructions, is_default)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING `+addressColumns,
			userID, in.AddressLabel, in.RecipientName, in.Phone, in.City, in.Area, in.Street,
			in.BuildingNumber, in.ApartmentNumber, in.AdditionalInstructions, makeDefault))
		return err
	})
	return a, apperr.FromDB(err, "address not found")
}

// Update changes the address fields (not the default flag).
func (r *AddressRepository) Update(ctx context.Context, userID, id int64, in models.AddressInput) (models.Address, error) {
	a, err := scanAddress(r.db.QueryRowContext(ctx, `
		UPDATE user_addresses SET
			address_label = $1, recipient_name = $2, phone = $3, city = $4, area = $5, street = $6,
			building_number = $7, apartment_number = $8, additional_instructions = $9, updated_at = NOW()
		WHERE id = $10 AND user_id = $11
		RETURNING `+addressColumns,
		in.AddressLabel, in.RecipientName, in.Phone, in.City, in.Area, in.Street,
		in.BuildingNumber, in.ApartmentNumber, in.AdditionalInstructions, id, userID))
	return a, apperr.FromDB(err, "address not found")
}

// Delete removes an address. Past orders are unaffected: they keep their own
// copy of the address (orders.address_id simply becomes NULL). If the default
// address is deleted, the most recent remaining address becomes the default.
func (r *AddressRepository) Delete(ctx context.Context, userID, id int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := lockUser(ctx, tx, userID); err != nil {
			return err
		}
		var wasDefault bool
		if err := tx.QueryRowContext(ctx,
			`DELETE FROM user_addresses WHERE id = $1 AND user_id = $2 RETURNING is_default`, id, userID,
		).Scan(&wasDefault); err != nil {
			return apperr.FromDB(err, "address not found")
		}
		if !wasDefault {
			return nil
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE user_addresses SET is_default = TRUE, updated_at = NOW()
			WHERE id = (SELECT id FROM user_addresses WHERE user_id = $1
			            ORDER BY created_at DESC, id DESC LIMIT 1)`, userID)
		return err
	})
}

func (r *AddressRepository) SetDefault(ctx context.Context, userID, id int64) (models.Address, error) {
	var a models.Address
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := lockUser(ctx, tx, userID); err != nil {
			return err
		}
		var exists bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM user_addresses WHERE id = $1 AND user_id = $2)`, id, userID,
		).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return apperr.NotFound("address not found")
		}
		// Unset first, then set: the partial unique index allows only one TRUE.
		if _, err := tx.ExecContext(ctx, `
			UPDATE user_addresses SET is_default = FALSE, updated_at = NOW()
			WHERE user_id = $1 AND is_default AND id <> $2`, userID, id); err != nil {
			return err
		}
		var err error
		a, err = scanAddress(tx.QueryRowContext(ctx, `
			UPDATE user_addresses SET is_default = TRUE, updated_at = NOW()
			WHERE id = $1 AND user_id = $2
			RETURNING `+addressColumns, id, userID))
		return err
	})
	return a, apperr.FromDB(err, "address not found")
}
