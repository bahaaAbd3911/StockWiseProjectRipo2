package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
)

// ErrResetTokenInvalid covers every reason a reset link cannot be used
// (unknown, already used, expired, account deactivated). One message for all,
// so the response does not reveal which.
var ErrResetTokenInvalid = apperr.Invalid("this password reset link is invalid or has expired; please request a new one")

// CreateResetToken stores the hash of a new reset token. Older unused links
// of the same user stop working, so only the latest email is valid.
func (r *UserRepository) CreateResetToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ip string) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, requested_ip)
			VALUES ($1, $2, $3, NULLIF($4, ''))`, userID, tokenHash, expiresAt, ip)
		return err
	})
}

// ResetPassword uses a reset token: it must exist, be unused, not expired, and
// belong to an active account. In one transaction it sets the new password,
// raises token_version (logging out every device), marks the token used and
// deletes the user's other reset links.
func (r *UserRepository) ResetPassword(ctx context.Context, tokenHash, newPasswordHash string) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var tokenID, userID int64
		var expiresAt time.Time
		var usedAt *time.Time
		var active bool
		err := tx.QueryRowContext(ctx, `
			SELECT t.id, t.user_id, t.expires_at, t.used_at, u.is_active
			FROM password_reset_tokens t JOIN users u ON u.id = t.user_id
			WHERE t.token_hash = $1
			FOR UPDATE OF t`, tokenHash).Scan(&tokenID, &userID, &expiresAt, &usedAt, &active)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrResetTokenInvalid
		}
		if err != nil {
			return err
		}
		if usedAt != nil || !time.Now().Before(expiresAt) || !active {
			return ErrResetTokenInvalid
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE users SET password_hash = $1, token_version = token_version + 1, updated_at = NOW()
			WHERE id = $2`, newPasswordHash, userID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`, tokenID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx,
			`DELETE FROM password_reset_tokens WHERE user_id = $1 AND id <> $2`, userID, tokenID)
		return err
	})
}
