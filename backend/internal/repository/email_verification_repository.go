package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
)

// ErrVerifyTokenInvalid covers every reason a confirmation link cannot be used.
var ErrVerifyTokenInvalid = apperr.Invalid("this confirmation link is invalid or has expired; please ask for a new one from your profile")

// CreateVerifyToken stores the hash of a new confirmation token. Older links
// of the same user stop working, so only the newest email is valid.
func (r *UserRepository) CreateVerifyToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
			userID, tokenHash, expiresAt)
		return err
	})
}

// VerifyEmail uses a confirmation token: it marks the email as verified and
// deletes the user's links. Using a link twice is harmless but reports
// "invalid", because the link is gone.
func (r *UserRepository) VerifyEmail(ctx context.Context, tokenHash string) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var userID int64
		var expiresAt time.Time
		var active bool
		err := tx.QueryRowContext(ctx, `
			SELECT t.user_id, t.expires_at, u.is_active
			FROM email_verification_tokens t JOIN users u ON u.id = t.user_id
			WHERE t.token_hash = $1
			FOR UPDATE OF t`, tokenHash).Scan(&userID, &expiresAt, &active)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVerifyTokenInvalid
		}
		if err != nil {
			return err
		}
		if !time.Now().Before(expiresAt) || !active {
			return ErrVerifyTokenInvalid
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()), updated_at = NOW()
			WHERE id = $1`, userID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID)
		return err
	})
}

// SetProfileImage stores a new profile image URL (nil removes it) and returns
// the previous one, so the caller can delete the old file.
func (r *UserRepository) SetProfileImage(ctx context.Context, userID int64, imageURL *string) (old *string, err error) {
	err = database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`SELECT profile_image_url FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&old); err != nil {
			return apperr.FromDB(err, "user not found")
		}
		_, err := tx.ExecContext(ctx,
			`UPDATE users SET profile_image_url = $1, updated_at = NOW() WHERE id = $2`, imageURL, userID)
		return err
	})
	return old, err
}
