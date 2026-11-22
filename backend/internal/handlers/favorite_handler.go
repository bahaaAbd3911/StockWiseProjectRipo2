package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type FavoriteHandler struct {
	favorites *services.FavoriteService
}

func NewFavoriteHandler(favorites *services.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

// List: GET /api/favorites
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) error {
	favs, err := h.favorites.List(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, favs)
}

// Add: POST /api/favorites {product_id} -> 201 if new, 200 if it already was a favorite
func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) error {
	var req models.AddFavoriteRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	created, err := h.favorites.Add(r.Context(), auth.UserID(r.Context()), req.ProductID)
	if err != nil {
		return err
	}
	if !created {
		return httpx.Message(w, http.StatusOK, "product is already in your favorites")
	}
	return httpx.Message(w, http.StatusCreated, "product added to favorites")
}

// Remove: DELETE /api/favorites/{product_id}
func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) error {
	productID, err := httpx.PathID(r, "product_id")
	if err != nil {
		return err
	}
	if err := h.favorites.Remove(r.Context(), auth.UserID(r.Context()), productID); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "product removed from favorites")
}
