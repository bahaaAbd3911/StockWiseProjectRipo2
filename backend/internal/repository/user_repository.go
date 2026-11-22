package repository

import (
	"context"
	"database/sql"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

// users.created_at / updated_at are nullable in the schema, hence COALESCE.
const userColumns = `
	id, first_name, last_name, email, phone, profile_image_url, role, is_active,
	COALESCE(created_at, NOW()), COALESCE(updated_at, NOW())`

func scanUser(s scanner, extra ...any) (models.User, error) {
	var u models.User
	dest := append([]any{
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.ProfileImageURL,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	}, extra...)
	err := s.Scan(dest...)
	return u, err
}

// EmailTaken checks case-insensitively, so existing mixed-case rows also count.
func (r *UserRepository) EmailTaken(ctx context.Context, email string) (bool, error) {
	var taken bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE lower(email) = lower($1))`, email,
	).Scan(&taken)
	return taken, err
}

// Create inserts a customer. role, is_active and timestamps use the column
// defaults, so a client can never register itself as an admin.
func (r *UserRepository) Create(ctx context.Context, req models.RegisterRequest, passwordHash string) (models.User, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash, phone, profile_image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+userColumns,
		req.FirstName, req.LastName, req.Email, passwordHash, req.Phone, req.ProfileImageURL,
	)
	u, err := scanUser(row)
	if apperr.IsUniqueViolation(err, "users_email_key") {
		return models.User{}, apperr.Conflict("an account with this email already exists")
	}
	return u, apperr.FromDB(err, "user not found")
}

// GetByEmail returns the user and their password hash (for login only).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, string, error) {
	var hash string
	row := r.db.QueryRowContext(ctx, `
		SELECT `+userColumns+`, password_hash
		FROM users
		WHERE lower(email) = lower($1)
		ORDER BY id
		LIMIT 1`, email)
	u, err := scanUser(row, &hash)
	return u, hash, err // raw error: the service decides how to report it
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (models.User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	return u, apperr.FromDB(err, "user not found")
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, req models.UpdateProfileRequest) (models.User, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET first_name = $1, last_name = $2, phone = $3, profile_image_url = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING `+userColumns,
		req.FirstName, req.LastName, req.Phone, req.ProfileImageURL, id,
	)
	u, err := scanUser(row)
	return u, apperr.FromDB(err, "user not found")
}

func (r *UserRepository) PasswordHash(ctx context.Context, id int64) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&hash)
	return hash, apperr.FromDB(err, "user not found")
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, hash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, hash, id)
	return expectOneRow(res, err, "user not found")
}

// Deactivate is a soft delete: the row must stay because orders reference it
// (orders.user_id is ON DELETE RESTRICT).
func (r *UserRepository) Deactivate(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND is_active`, id)
	return expectOneRow(res, err, "account not found or already deactivated")
}

// AuthInfo is used by the authentication middleware on every protected request.
// It returns sql.ErrNoRows unchanged when the user does not exist.
func (r *UserRepository) AuthInfo(ctx context.Context, id int64) (string, bool, error) {
	var role string
	var active bool
	err := r.db.QueryRowContext(ctx, `SELECT role, is_active FROM users WHERE id = $1`, id).Scan(&role, &active)
	return role, active, err
}

// expectOneRow turns "0 rows affected" into a NotFound error.
func expectOneRow(res sql.Result, err error, notFound string) error {
	if err != nil {
		return apperr.FromDB(err, notFound)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.NotFound(notFound)
	}
	return nil
}
