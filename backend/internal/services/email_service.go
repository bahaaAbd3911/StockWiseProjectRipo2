package services

import (
	"context"
	"strings"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/notify"
	"ecommerce-backend/internal/validators"
)

// EmailService is the admin view of the notification email queue: see what
// is waiting or failed, retry, or cancel.
type EmailService struct {
	outbox *notify.Outbox
}

func NewEmailService(outbox *notify.Outbox) *EmailService { return &EmailService{outbox: outbox} }

func (s *EmailService) List(ctx context.Context, f models.EmailFilter) (models.Page[models.Email], error) {
	f.Status = strings.ToLower(strings.TrimSpace(f.Status))
	if f.Status != "" && !models.IsEmailStatus(f.Status) {
		return models.Page[models.Email]{}, validators.FieldError("status", "must be pending, failed or sent")
	}
	f.Kind = strings.TrimSpace(f.Kind)
	f.To = strings.TrimSpace(f.To)
	return s.outbox.List(ctx, f)
}

func (s *EmailService) Get(ctx context.Context, id int64) (models.Email, error) {
	return s.outbox.Get(ctx, id)
}

func (s *EmailService) Summary(ctx context.Context) (models.EmailSummary, error) {
	return s.outbox.Summary(ctx)
}

func (s *EmailService) Retry(ctx context.Context, id int64) (models.Email, error) {
	return s.outbox.Retry(ctx, id)
}

func (s *EmailService) RetryFailed(ctx context.Context) (int64, error) {
	return s.outbox.RetryFailed(ctx)
}

func (s *EmailService) Cancel(ctx context.Context, id int64) error {
	return s.outbox.Cancel(ctx, id)
}
