package repository

import (
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/models"
)

type AddressRepository struct {
	DB *sql.DB
}

func NewAddressRepository(db *sql.DB) *AddressRepository {
	return &AddressRepository{DB: db}
}

func (r *AddressRepository) CreateAddress(
	userID int64,
	request models.CreateAddressRequest,
) (models.Address, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return models.Address{}, err
	}
	defer tx.Rollback()

	// Lock the user row to safely manage default addresses.
	var id int64
	err = tx.QueryRow(
		"SELECT id FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&id)

	if err != nil {
		return models.Address{}, err
	}

	if request.IsDefault {
		_, err = tx.Exec(
			"UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1",
			userID,
		)
		if err != nil {
			return models.Address{}, err
		}
	}

	query := `
		INSERT INTO user_addresses (
			user_id, address_label, recipient_name, phone,
			city, area, street, building_number,
			apartment_number, additional_instructions, is_default
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, address_label, recipient_name, phone,
			city, area, street, building_number, apartment_number,
			additional_instructions, is_default, created_at, updated_at
	`

	var address models.Address

	err = tx.QueryRow(
		query,
		userID,
		request.AddressLabel,
		request.RecipientName,
		request.Phone,
		request.City,
		request.Area,
		request.Street,
		request.BuildingNumber,
		request.ApartmentNumber,
		request.AdditionalInstructions,
		request.IsDefault,
	).Scan(
		&address.ID,
		&address.UserID,
		&address.AddressLabel,
		&address.RecipientName,
		&address.Phone,
		&address.City,
		&address.Area,
		&address.Street,
		&address.BuildingNumber,
		&address.ApartmentNumber,
		&address.AdditionalInstructions,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		return models.Address{}, fmt.Errorf("failed to create address: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return models.Address{}, err
	}

	return address, nil
}

func (r *AddressRepository) GetAddresses(userID int64) ([]models.Address, error) {
	query := `
		SELECT id, user_id, address_label, recipient_name, phone,
			city, area, street, building_number, apartment_number,
			additional_instructions, is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := []models.Address{}

	for rows.Next() {
		var address models.Address

		err := rows.Scan(
			&address.ID,
			&address.UserID,
			&address.AddressLabel,
			&address.RecipientName,
			&address.Phone,
			&address.City,
			&address.Area,
			&address.Street,
			&address.BuildingNumber,
			&address.ApartmentNumber,
			&address.AdditionalInstructions,
			&address.IsDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (r *AddressRepository) UpdateAddress(
	userID int64,
	addressID int64,
	request models.UpdateAddressRequest,
) (models.Address, error) {
	query := `
		UPDATE user_addresses
		SET address_label = $1,
			recipient_name = $2,
			phone = $3,
			city = $4,
			area = $5,
			street = $6,
			building_number = $7,
			apartment_number = $8,
			additional_instructions = $9,
			updated_at = NOW()
		WHERE id = $10 AND user_id = $11
		RETURNING id, user_id, address_label, recipient_name, phone,
			city, area, street, building_number, apartment_number,
			additional_instructions, is_default, created_at, updated_at
	`

	var address models.Address

	err := r.DB.QueryRow(
		query,
		request.AddressLabel,
		request.RecipientName,
		request.Phone,
		request.City,
		request.Area,
		request.Street,
		request.BuildingNumber,
		request.ApartmentNumber,
		request.AdditionalInstructions,
		addressID,
		userID,
	).Scan(
		&address.ID,
		&address.UserID,
		&address.AddressLabel,
		&address.RecipientName,
		&address.Phone,
		&address.City,
		&address.Area,
		&address.Street,
		&address.BuildingNumber,
		&address.ApartmentNumber,
		&address.AdditionalInstructions,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		return models.Address{}, err
	}

	return address, nil
}

func (r *AddressRepository) DeleteAddress(userID int64, addressID int64) error {
	result, err := r.DB.Exec(
		"DELETE FROM user_addresses WHERE id = $1 AND user_id = $2",
		addressID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *AddressRepository) SetDefaultAddress(
	userID int64,
	addressID int64,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Serialize default-address changes for this user.
	var id int64
	err = tx.QueryRow(
		"SELECT id FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&id)

	if err != nil {
		return err
	}

	// Confirm that the address belongs to this user.
	err = tx.QueryRow(
		"SELECT id FROM user_addresses WHERE id = $1 AND user_id = $2",
		addressID,
		userID,
	).Scan(&id)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE user_addresses SET is_default = TRUE, updated_at = NOW() WHERE id = $1 AND user_id = $2",
		addressID,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
