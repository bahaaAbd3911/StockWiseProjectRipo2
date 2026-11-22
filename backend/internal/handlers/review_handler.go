package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type ReviewHandler struct {
	reviews *services.ReviewService
}

func NewReviewHandler(reviews *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviews: reviews}
}

// ListForProduct: GET /api/products/{id}/reviews?sort=newest|helpful|rating_high|rating_low
func (h *ReviewHandler) ListForProduct(w http.ResponseWriter, r *http.Request) error {
	productID, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	list, err := h.reviews.ListForProduct(r.Context(), productID, r.URL.Query().Get("sort"))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, list)
}

// Create: POST /api/products/{id}/reviews {comment, comment_ar, rating}
func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) error {
	productID, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.ReviewInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	review, err := h.reviews.Create(r.Context(), auth.UserID(r.Context()), productID, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, review)
}

// Update: PUT /api/reviews/{id} (author only)
func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.ReviewInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	review, err := h.reviews.Update(r.Context(), auth.UserID(r.Context()), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, review)
}

// Delete: DELETE /api/reviews/{id} (author only)
func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.reviews.Delete(r.Context(), auth.UserID(r.Context()), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "review deleted")
}

// AddHelpfulVote: POST /api/reviews/{id}/helpful
func (h *ReviewHandler) AddHelpfulVote(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	review, err := h.reviews.AddHelpfulVote(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, review)
}

// RemoveHelpfulVote: DELETE /api/reviews/{id}/helpful
func (h *ReviewHandler) RemoveHelpfulVote(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	review, err := h.reviews.RemoveHelpfulVote(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, review)
}
