package handlers

import (
	"fmt"
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

// EmailHandler: admin endpoints for the notification email queue.
type EmailHandler struct {
	emails *services.EmailService
}

func NewEmailHandler(emails *services.EmailService) *EmailHandler {
	return &EmailHandler{emails: emails}
}

// List: GET /api/admin/emails?status=pending|failed|sent&kind=&to=&limit=&offset=
func (h *EmailHandler) List(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	page, err := h.emails.List(r.Context(), models.EmailFilter{
		Status: q.Get("status"), Kind: q.Get("kind"), To: q.Get("to"), Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// Summary: GET /api/admin/emails/summary
func (h *EmailHandler) Summary(w http.ResponseWriter, r *http.Request) error {
	s, err := h.emails.Summary(r.Context())
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, s)
}

// Get: GET /api/admin/emails/{id} (includes the full text)
func (h *EmailHandler) Get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	e, err := h.emails.Get(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, e)
}

// Retry: POST /api/admin/emails/{id}/retry
func (h *EmailHandler) Retry(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	e, err := h.emails.Retry(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, e)
}

// RetryFailed: POST /api/admin/emails/retry-failed
func (h *EmailHandler) RetryFailed(w http.ResponseWriter, r *http.Request) error {
	n, err := h.emails.RetryFailed(r.Context())
	if err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, fmt.Sprintf("%d failed email(s) will be sent again", n))
}

// Cancel: DELETE /api/admin/emails/{id} (only emails not sent yet)
func (h *EmailHandler) Cancel(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.emails.Cancel(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "email cancelled")
}
