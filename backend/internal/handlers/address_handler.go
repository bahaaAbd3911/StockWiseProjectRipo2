package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type AddressHandler struct {
	addresses *services.AddressService
}

func NewAddressHandler(addresses *services.AddressService) *AddressHandler {
	return &AddressHandler{addresses: addresses}
}

// List: GET /api/addresses (default first)
func (h *AddressHandler) List(w http.ResponseWriter, r *http.Request) error {
	list, err := h.addresses.List(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, list)
}

// Get: GET /api/addresses/{id}
func (h *AddressHandler) Get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	a, err := h.addresses.Get(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, a)
}

// Create: POST /api/addresses
func (h *AddressHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.AddressInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	a, err := h.addresses.Create(r.Context(), auth.UserID(r.Context()), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, a)
}

// Update: PUT /api/addresses/{id}
func (h *AddressHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.AddressInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	a, err := h.addresses.Update(r.Context(), auth.UserID(r.Context()), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, a)
}

// Delete: DELETE /api/addresses/{id}
func (h *AddressHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.addresses.Delete(r.Context(), auth.UserID(r.Context()), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "address deleted")
}

// SetDefault: PUT /api/addresses/{id}/default
func (h *AddressHandler) SetDefault(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	a, err := h.addresses.SetDefault(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, a)
}
