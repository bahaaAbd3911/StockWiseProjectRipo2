package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
	"ecommerce-backend/internal/tenant"
)

// PlatformHandler: the platform super admin's API (only on the platform host).
type PlatformHandler struct {
	platform *services.PlatformService
}

func NewPlatformHandler(p *services.PlatformService) *PlatformHandler {
	return &PlatformHandler{platform: p}
}

// Login: POST /api/platform/login {email, password}
func (h *PlatformHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req models.LoginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	res, err := h.platform.Login(r.Context(), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, res)
}

// Me: GET /api/platform/me
func (h *PlatformHandler) Me(w http.ResponseWriter, r *http.Request) error {
	a, err := h.platform.Me(r.Context(), auth.PlatformAdminID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, a)
}

// ListStores: GET /api/platform/stores?search=&status=&limit=&offset=
func (h *PlatformHandler) ListStores(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	page, err := h.platform.ListStores(r.Context(), models.StoreFilter{
		Search: q.Get("search"), Status: q.Get("status"), Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// GetStore: GET /api/platform/stores/{id}
func (h *PlatformHandler) GetStore(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	s, err := h.platform.GetStore(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, s)
}

// CreateStore: POST /api/platform/stores {slug, name, currency?, timezone?, admin: {first_name, last_name, email}}
func (h *PlatformHandler) CreateStore(w http.ResponseWriter, r *http.Request) error {
	var in models.CreateStoreRequest
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	res, err := h.platform.CreateStore(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, res)
}

// UpdateStore: PUT /api/platform/stores/{id} {name, currency, timezone}
func (h *PlatformHandler) UpdateStore(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.UpdateStoreRequest
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	s, err := h.platform.UpdateStore(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, s)
}

func (h *PlatformHandler) setStatus(status string) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := httpx.PathID(r, "id")
		if err != nil {
			return err
		}
		s, err := h.platform.SetStatus(r.Context(), id, status)
		if err != nil {
			return err
		}
		return httpx.JSON(w, http.StatusOK, s)
	}
}

// Suspend: POST /api/platform/stores/{id}/suspend
func (h *PlatformHandler) Suspend(w http.ResponseWriter, r *http.Request) error {
	return h.setStatus(tenant.StatusSuspended)(w, r)
}

// Activate: POST /api/platform/stores/{id}/activate
func (h *PlatformHandler) Activate(w http.ResponseWriter, r *http.Request) error {
	return h.setStatus(tenant.StatusActive)(w, r)
}

// ResendInvite: POST /api/platform/stores/{id}/invite
func (h *PlatformHandler) ResendInvite(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	msg, err := h.platform.ResendInvite(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, msg)
}
