package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type ProductReviewHandler struct {
	Service *services.ProductReviewService
}

func NewProductReviewHandler(service *services.ProductReviewService) *ProductReviewHandler {
	return &ProductReviewHandler{
		Service: service,
	}
}

// Create a product review
func (h *ProductReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Comment   string  `json:"comment"`
		CommentAr *string `json:"comment_ar"`
		Rating    int     `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	review := models.ProductReview{
		ProductID: productID,
		UserID:    userID,
		Comment:   request.Comment,
		CommentAr: request.CommentAr,
		Rating:    request.Rating,
	}

	createdReview, err := h.Service.CreateReview(review)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdReview)
}

// Get all reviews for a product
func (h *ProductReviewHandler) GetReviewsByProductID(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	reviews, err := h.Service.GetReviewsByProductID(productID)
	if err != nil {
		http.Error(w, "Failed to retrieve reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reviews)
}

// Update your own review
func (h *ProductReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	reviewID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || reviewID <= 0 {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Comment   string  `json:"comment"`
		CommentAr *string `json:"comment_ar"`
		Rating    int     `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	updatedReview, err := h.Service.UpdateReview(
		reviewID,
		userID,
		request.Comment,
		request.CommentAr,
		request.Rating,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedReview)
}

// Delete your own review
func (h *ProductReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	reviewID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || reviewID <= 0 {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	err = h.Service.DeleteReview(reviewID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Review deleted successfully",
	})
}

func (h *ProductReviewHandler) AddHelpfulVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	reviewID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || reviewID <= 0 {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	err = h.Service.AddHelpfulVote(reviewID, userID)
	if err != nil {
		http.Error(w, "Unable to add helpful vote. You may have already voted or the review may not exist.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Helpful vote added successfully",
	})
}

func (h *ProductReviewHandler) RemoveHelpfulVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	reviewID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || reviewID <= 0 {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	err = h.Service.RemoveHelpfulVote(reviewID, userID)
	if err != nil {
		http.Error(w, "Unable to remove helpful vote or vote does not exist.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Helpful vote removed successfully",
	})
}
