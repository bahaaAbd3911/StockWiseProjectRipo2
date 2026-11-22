package handlers

import (
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type BundleHandler struct {
	bundles *services.BundleService
}

func NewBundleHandler(bundles *services.BundleService) *BundleHandler {
	return &BundleHandler{bundles: bundles}
}

func (h *BundleHandler) list(w http.ResponseWriter, r *http.Request, admin bool) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.bundles.List(r.Context(), models.BundleFilter{IncludeUnavailable: admin, Limit: limit, Offset: offset})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

func (h *BundleHandler) get(w http.ResponseWriter, r *http.Request, admin bool) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	b, err := h.bundles.Get(r.Context(), id, admin)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, b)
}

// List: GET /api/bundles (live bundles)
func (h *BundleHandler) List(w http.ResponseWriter, r *http.Request) error {
	return h.list(w, r, false)
}

// Get: GET /api/bundles/{id}
func (h *BundleHandler) Get(w http.ResponseWriter, r *http.Request) error { return h.get(w, r, false) }

// AdminList: GET /api/admin/bundles (also inactive / outside their window)
func (h *BundleHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	return h.list(w, r, true)
}

// AdminGet: GET /api/admin/bundles/{id}
func (h *BundleHandler) AdminGet(w http.ResponseWriter, r *http.Request) error {
	return h.get(w, r, true)
}

// Create: POST /api/admin/bundles
func (h *BundleHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.BundleInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	b, err := h.bundles.Create(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, b)
}

// Update: PUT /api/admin/bundles/{id}
func (h *BundleHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.BundleInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	b, err := h.bundles.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, b)
}

// Delete: DELETE /api/admin/bundles/{id} (soft delete)
func (h *BundleHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.bundles.Deactivate(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "bundle deactivated")
}
