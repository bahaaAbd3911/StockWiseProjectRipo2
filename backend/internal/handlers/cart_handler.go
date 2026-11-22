package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

// CartHandler: every endpoint responds with the full, updated cart.
type CartHandler struct {
	carts *services.CartService
}

func NewCartHandler(carts *services.CartService) *CartHandler { return &CartHandler{carts: carts} }

// Get: GET /api/cart
func (h *CartHandler) Get(w http.ResponseWriter, r *http.Request) error {
	cart, err := h.carts.Get(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// AddItem: POST /api/cart/items {product_id, quantity, selected_color}
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) error {
	var req models.AddCartItemRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	cart, err := h.carts.AddItem(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// UpdateItem: PUT /api/cart/items/{id} {quantity, selected_color}
func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.UpdateCartItemRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	cart, err := h.carts.UpdateItem(r.Context(), auth.UserID(r.Context()), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// RemoveItem: DELETE /api/cart/items/{id}
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	cart, err := h.carts.RemoveItem(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// Clear: DELETE /api/cart
func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) error {
	cart, err := h.carts.Clear(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// ---------- bundles ----------

// AddBundle: POST /api/cart/bundles {bundle_id, quantity}
func (h *CartHandler) AddBundle(w http.ResponseWriter, r *http.Request) error {
	var req models.AddCartBundleRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	cart, err := h.carts.AddBundle(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// UpdateBundle: PUT /api/cart/bundles/{id} {quantity}
func (h *CartHandler) UpdateBundle(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.UpdateCartBundleRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	cart, err := h.carts.UpdateBundle(r.Context(), auth.UserID(r.Context()), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// RemoveBundle: DELETE /api/cart/bundles/{id}
func (h *CartHandler) RemoveBundle(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	cart, err := h.carts.RemoveBundle(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}
