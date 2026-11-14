package repository

import (
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) CreateUser(
	user models.RegisterUserRequest,
	passwordHash string,
) (models.User, error) {

	query := `
        INSERT INTO users (
            first_name,
            last_name,
            email,
            password_hash,
            phone,
            profile_image_url
        )
        VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
        RETURNING
            id,
            first_name,
            last_name,
            email,
            phone,
            profile_image_url,
			role,
            created_at,
            updated_at;
    `

	var newUser models.User
	var profileImageURL sql.NullString

	err := r.DB.QueryRow(
		query,
		user.FirstName,
		user.LastName,
		user.Email,
		passwordHash,
		user.Phone,
		user.ProfileImageURL,
	).Scan(
		&newUser.ID,
		&newUser.FirstName,
		&newUser.LastName,
		&newUser.Email,
		&newUser.Phone,
		&profileImageURL,
		&newUser.Role,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)

	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	if profileImageURL.Valid {
		newUser.ProfileImageURL = &profileImageURL.String
	}

	return newUser, nil
}
func (r *UserRepository) GetUserByEmail(
	email string,
) (models.User, string, error) {

	query := `
        SELECT
            id,
            first_name,
            last_name,
            email,
            phone,
            profile_image_url,
			role,
            is_active,
            created_at,
            updated_at,
            password_hash
        FROM users
        WHERE email = $1;
    `

	var user models.User
	var passwordHash string
	var profileImageURL sql.NullString

	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&profileImageURL,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&passwordHash,
	)

	if err != nil {
		return models.User{}, "", fmt.Errorf(
			"failed to find user: %w", err,
		)
	}
	if profileImageURL.Valid {
		user.ProfileImageURL = &profileImageURL.String
	}

	return user, passwordHash, nil
}
func (r *UserRepository) GetUserByID(id int64) (models.User, error) {
	query := `
        SELECT
            id,
            first_name,
            last_name,
            email,
            phone,
            profile_image_url,
            role,
            is_active,
            created_at,
            updated_at
        FROM users
        WHERE id = $1;
    `

	var user models.User
	var profileImageURL sql.NullString

	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&profileImageURL,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return models.User{}, err
	}

	if profileImageURL.Valid {
		user.ProfileImageURL = &profileImageURL.String
	}

	return user, nil
}
func (r *UserRepository) UpdateUserProfile(
	id int64,
	request models.UpdateProfileRequest,
) (models.User, error) {
	query := `
        UPDATE users
        SET
            first_name = $1,
            last_name = $2,
            phone = $3,
            profile_image_url = NULLIF($4, ''),
            updated_at = NOW()
        WHERE id = $5
        RETURNING
            id,
            first_name,
            last_name,
            email,
            phone,
            profile_image_url,
			role,
            is_active,
            created_at,
            updated_at;
    `

	var user models.User
	var profileImageURL sql.NullString

	err := r.DB.QueryRow(
		query,
		request.FirstName,
		request.LastName,
		request.Phone,
		request.ProfileImageURL,
		id,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&profileImageURL,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return models.User{}, err
	}

	if profileImageURL.Valid {
		user.ProfileImageURL = &profileImageURL.String
	}

	return user, nil
}
func (r *UserRepository) UpdatePassword(userID int64, passwordHash string) error {
	query := `
        UPDATE users
        SET password_hash = $1,
            updated_at = NOW()
        WHERE id = $2;
    `

	result, err := r.DB.Exec(query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
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
func (r *UserRepository) GetPasswordHashByID(userID int64) (string, error) {
	query := `
        SELECT password_hash
        FROM users
        WHERE id = $1;
    `

	var passwordHash string

	err := r.DB.QueryRow(query, userID).Scan(&passwordHash)
	if err != nil {
		return "", err
	}

	return passwordHash, nil
}
func (r *UserRepository) DeactivateUser(userID int64) error {
	query := `
        UPDATE users
        SET is_active = FALSE,
            updated_at = NOW()
        WHERE id = $1 AND is_active = TRUE;
    `

	result, err := r.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deactivated user: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
