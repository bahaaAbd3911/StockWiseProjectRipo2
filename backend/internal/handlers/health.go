package handlers

import (
	"encoding/json"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"message": "E-commerce API is running!",
		"status":  "success",
	}

	json.NewEncoder(w).Encode(response)
}
