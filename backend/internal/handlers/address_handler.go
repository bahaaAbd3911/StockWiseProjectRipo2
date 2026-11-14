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

type AddressHandler struct {
	Service *services.AddressService
}

func NewAddressHandler(service *services.AddressService) *AddressHandler {
	return &AddressHandler{Service: service}
}

func (h *AddressHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var request models.CreateAddressRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	address, err := h.Service.CreateAddress(userID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(address)
}

func (h *AddressHandler) GetAddresses(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	addresses, err := h.Service.GetAddresses(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve addresses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addresses)
}

func (h *AddressHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	addressID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || addressID <= 0 {
		http.Error(w, "Invalid address ID", http.StatusBadRequest)
		return
	}

	var request models.UpdateAddressRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	address, err := h.Service.UpdateAddress(userID, addressID, request)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Address not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(address)
}

func (h *AddressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	addressID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || addressID <= 0 {
		http.Error(w, "Invalid address ID", http.StatusBadRequest)
		return
	}

	err = h.Service.DeleteAddress(userID, addressID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Address not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to delete address", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Address deleted successfully",
	})
}

func (h *AddressHandler) SetDefaultAddress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	addressID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || addressID <= 0 {
		http.Error(w, "Invalid address ID", http.StatusBadRequest)
		return
	}

	err = h.Service.SetDefaultAddress(userID, addressID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Address not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to set default address", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Default address updated successfully",
	})
}
