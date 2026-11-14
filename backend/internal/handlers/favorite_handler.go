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

type FavoriteHandler struct {
	Service *services.FavoriteService
}

func NewFavoriteHandler(service *services.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{Service: service}
}

func (h *FavoriteHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var request models.AddFavoriteRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.Service.AddFavorite(userID, request.ProductID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Product added to favorites successfully",
	})
}

func (h *FavoriteHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	favorites, err := h.Service.GetFavorites(userID)
	if err != nil {
		http.Error(w, "Failed to get favorites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (h *FavoriteHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	productID, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.Service.RemoveFavorite(userID, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Favorite not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to remove favorite", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Favorite removed successfully",
	})
}
