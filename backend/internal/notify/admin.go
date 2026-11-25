package notify

import (
	"context"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/tenant"
)

// Admin view of the email queue: what is waiting, what failed, and buttons to
// retry or cancel.

// statusSQL computes the status shown to admins from the stored columns.
const statusSQL = `CASE WHEN sent_at IS NOT NULL THEN 'sent'
	WHEN attempts >= ` + maxAttemptsSQL + ` THEN 'failed' ELSE 'pending' END`

const emailColumns = `id, kind, to_address, subject, ` + statusSQL + `, attempts, last_error,
	CASE WHEN sent_at IS NULL AND attempts < ` + maxAttemptsSQL + ` THEN next_attempt_at END,
	sent_at, created_at`

func scanEmail(s interface{ Scan(...any) error }, extra ...any) (models.Email, error) {
	var e models.Email
	err := s.Scan(append([]any{&e.ID, &e.Kind, &e.To, &e.Subject, &e.Status, &e.Attempts, &e.LastError,
		&e.NextAttemptAt, &e.SentAt, &e.CreatedAt}, extra...)...)
	return e, err
}

// List returns queued emails, newest first.
func (o *Outbox) List(ctx context.Context, f models.EmailFilter) (models.Page[models.Email], error) {
	where := ` WHERE ($1 = '' OR ` + statusSQL + ` = $1)
		AND ($2 = '' OR kind = $2)
		AND ($3 = '' OR to_address ILIKE '%' || $3 || '%')`
	args := []any{f.Status, f.Kind, f.To}
	page := models.Page[models.Email]{Items: []models.Email{}, Limit: f.Limit, Offset: f.Offset}
	if err := o.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM email_outbox`+where, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := o.db.QueryContext(ctx, `SELECT `+emailColumns+` FROM email_outbox`+where+
		` ORDER BY id DESC LIMIT $4 OFFSET $5`, append(args, f.Limit, f.Offset)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		e, err := scanEmail(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, e)
	}
	return page, rows.Err()
}

// Get returns one email with its full text.
func (o *Outbox) Get(ctx context.Context, id int64) (models.Email, error) {
	var body string
	e, err := scanEmail(o.db.QueryRowContext(ctx,
		`SELECT `+emailColumns+`, body FROM email_outbox WHERE id = $1`, id), &body)
	if err != nil {
		return models.Email{}, apperr.FromDB(err, "email not found")
	}
	e.Body = body
	return e, nil
}

// Summary counts the queue.
func (o *Outbox) Summary(ctx context.Context) (models.EmailSummary, error) {
	var s models.EmailSummary
	err := o.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE sent_at IS NULL AND attempts < `+maxAttemptsSQL+`),
			COUNT(*) FILTER (WHERE sent_at IS NULL AND attempts >= `+maxAttemptsSQL+`),
			COUNT(*) FILTER (WHERE sent_at > NOW() - INTERVAL '24 hours'),
			MIN(created_at) FILTER (WHERE sent_at IS NULL AND attempts < `+maxAttemptsSQL+`)
		FROM email_outbox`).Scan(&s.Pending, &s.Failed, &s.SentLast24h, &s.OldestPendingAt)
	return s, err
}

// Retry sends an unsent email again as soon as possible, with a fresh set of
// attempts (use it after fixing the mail settings or the address).
func (o *Outbox) Retry(ctx context.Context, id int64) (models.Email, error) {
	res, err := o.db.ExecContext(ctx, `
		UPDATE email_outbox SET attempts = 0, next_attempt_at = NOW(), last_error = NULL
		WHERE id = $1 AND sent_at IS NULL`, id)
	if err := o.expectUnsent(ctx, id, res, err); err != nil {
		return models.Email{}, err
	}
	return o.Get(ctx, id)
}

// RetryFailed gives every failed email a fresh set of attempts and returns
// how many there were.
func (o *Outbox) RetryFailed(ctx context.Context) (int64, error) {
	res, err := o.db.ExecContext(ctx, `
		UPDATE email_outbox SET attempts = 0, next_attempt_at = NOW(), last_error = NULL
		WHERE sent_at IS NULL AND attempts >= `+maxAttemptsSQL)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Cancel deletes an email that has not been sent (e.g. a wrong address).
func (o *Outbox) Cancel(ctx context.Context, id int64) error {
	res, err := o.db.ExecContext(ctx, `DELETE FROM email_outbox WHERE id = $1 AND sent_at IS NULL`, id)
	return o.expectUnsent(ctx, id, res, err)
}

// expectUnsent turns "0 rows changed" into 404 (no such email) or 409 (already sent).
func (o *Outbox) expectUnsent(ctx context.Context, id int64, res interface{ RowsAffected() (int64, error) }, err error) error {
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n > 0 {
		return err
	}
	var exists bool
	if err := o.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM email_outbox WHERE id = $1)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return apperr.NotFound("email not found")
	}
	return apperr.Conflict("this email has already been sent")
}

// DeleteOldSent removes delivered emails older than the retention period
// (the queue is not an archive; it would otherwise grow forever).
func (o *Outbox) DeleteOldSent(ctx context.Context, keepDays int) (int64, error) {
	ctx = tenant.AsPlatform(ctx) // retention applies to every store
	res, err := o.db.ExecContext(ctx,
		`DELETE FROM email_outbox WHERE sent_at < NOW() - make_interval(days => $1)`, keepDays)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
