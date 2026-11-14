package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type CartHandler struct {
	Service *services.CartService
}

func NewCartHandler(service *services.CartService) *CartHandler {
	return &CartHandler{
		Service: service,
	}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cart, err := h.Service.GetCart(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to retrieve cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cart)
}
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	// 1. Get the authenticated user ID
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse the JSON body
	var req models.AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 3. Call the service
	err := h.Service.AddItem(userID, req)
	if err != nil {
		// If the product wasn't found or is inactive, return a 400 or 404.
		if err.Error() == "product not found or is currently inactive" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Item added to cart successfully",
	})
}

func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || itemID <= 0 {
		http.Error(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Quantity int `json:"quantity"`
	}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Service.UpdateItemQuantity(userID, itemID, request.Quantity)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cart item quantity updated successfully",
	})
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || itemID <= 0 {
		http.Error(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	err = h.Service.RemoveItem(userID, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to remove cart item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cart item removed successfully",
	})
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.Service.ClearCart(userID)
	if err != nil {
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cart cleared successfully",
	})
}
