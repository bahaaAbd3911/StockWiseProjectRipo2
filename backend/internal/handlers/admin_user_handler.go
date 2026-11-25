package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type AdminUserHandler struct {
	users *services.AdminUserService
}

func NewAdminUserHandler(u *services.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{users: u}
}

// List: GET /api/admin/users?search=&role=&limit=&offset=
func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.users.List(r.Context(), models.AdminUserFilter{
		Search: r.URL.Query().Get("search"), Role: r.URL.Query().Get("role"), Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// Get: GET /api/admin/users/{id} -> {user, orders, favorites}
func (h *AdminUserHandler) Get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	d, err := h.users.Details(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, d)
}

// Create: POST /api/admin/users
func (h *AdminUserHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.AdminUserInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	u, err := h.users.Create(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, u)
}

// Update: PUT /api/admin/users/{id}
func (h *AdminUserHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.AdminUserInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	u, err := h.users.Update(r.Context(), auth.UserID(r.Context()), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, u)
}

// Delete: DELETE /api/admin/users/{id} (deactivates instead when the user has orders)
func (h *AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	msg, err := h.users.Delete(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, msg)
}
