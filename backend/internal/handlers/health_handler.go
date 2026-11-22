// Package handlers is the HTTP boundary. A handler:
//  1. reads path values, query parameters and the JSON body
//  2. takes the user ID from the request context (put there by the auth
//     middleware) - never from the body
//  3. calls ONE service method
//  4. writes the JSON response with the right status code
//
// Handlers return errors instead of writing them; httpx.Handle turns them into
// responses. Handlers contain no business rules and no SQL.
package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"ecommerce-backend/internal/httpx"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler { return &HealthHandler{db: db} }

// Root: GET / - a quick "is the process up" check.
func (h *HealthHandler) Root(w http.ResponseWriter, r *http.Request) error {
	return httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "E-commerce API is running"})
}

// Health: GET /api/health - also checks the database connection.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		return httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "unreachable"})
	}
	return httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}
