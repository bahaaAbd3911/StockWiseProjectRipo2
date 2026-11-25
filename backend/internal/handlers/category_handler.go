package handlers

import (
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type CategoryHandler struct {
	categories *services.CategoryService
}

func NewCategoryHandler(c *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: c}
}

// List: GET /api/categories (active only)
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) error {
	cats, err := h.categories.List(r.Context(), false)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cats)
}

// AdminList: GET /api/admin/categories (inactive ones too)
func (h *CategoryHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	cats, err := h.categories.List(r.Context(), true)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cats)
}

// Create: POST /api/admin/categories {name, name_ar, icon, is_active?}
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.CategoryInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	c, err := h.categories.Create(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, c)
}

// Update: PUT /api/admin/categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.CategoryInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	c, err := h.categories.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, c)
}

// Delete: DELETE /api/admin/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.categories.Delete(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "category deleted; its products are now hidden")
}
