package services

import (
	"context"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

type ReviewService struct {
	reviews *repository.ReviewRepository
	catalog *repository.CatalogRepository
}

func NewReviewService(reviews *repository.ReviewRepository, catalog *repository.CatalogRepository) *ReviewService {
	return &ReviewService{reviews: reviews, catalog: catalog}
}

func (s *ReviewService) requireActiveProduct(ctx context.Context, productID int64) error {
	ok, err := s.catalog.ProductExists(ctx, productID, true)
	if err != nil {
		return err
	}
	if !ok {
		return apperr.NotFound("product not found")
	}
	return nil
}

func (s *ReviewService) ListForProduct(ctx context.Context, productID int64, sort string) ([]models.Review, error) {
	if sort == "" {
		sort = "newest"
	}
	if !repository.IsValidReviewSort(sort) {
		return nil, validators.FieldError("sort", "must be one of: newest, helpful, rating_high, rating_low")
	}
	if err := s.requireActiveProduct(ctx, productID); err != nil {
		return nil, err
	}
	return s.reviews.ListForProduct(ctx, productID, sort)
}

// Create adds the user's review. "verified_purchase" is decided here by the
// server (does the user have a delivered order with this product?), never by
// the client.
func (s *ReviewService) Create(ctx context.Context, userID, productID int64, in models.ReviewInput) (models.Review, error) {
	if err := validators.ValidateReview(&in); err != nil {
		return models.Review{}, err
	}
	if err := s.requireActiveProduct(ctx, productID); err != nil {
		return models.Review{}, err
	}
	verified, err := s.reviews.HasDeliveredPurchase(ctx, userID, productID)
	if err != nil {
		return models.Review{}, err
	}
	id, err := s.reviews.Create(ctx, userID, productID, in, verified)
	if err != nil {
		return models.Review{}, err
	}
	return s.reviews.Get(ctx, id)
}

func (s *ReviewService) Update(ctx context.Context, userID, reviewID int64, in models.ReviewInput) (models.Review, error) {
	if err := validators.ValidateReview(&in); err != nil {
		return models.Review{}, err
	}
	if err := s.reviews.Update(ctx, userID, reviewID, in); err != nil {
		return models.Review{}, err
	}
	return s.reviews.Get(ctx, reviewID)
}

func (s *ReviewService) Delete(ctx context.Context, userID, reviewID int64) error {
	return s.reviews.Delete(ctx, userID, reviewID)
}

func (s *ReviewService) AddHelpfulVote(ctx context.Context, userID, reviewID int64) (models.Review, error) {
	if err := s.reviews.AddHelpfulVote(ctx, userID, reviewID); err != nil {
		return models.Review{}, err
	}
	return s.reviews.Get(ctx, reviewID)
}

func (s *ReviewService) RemoveHelpfulVote(ctx context.Context, userID, reviewID int64) (models.Review, error) {
	if err := s.reviews.RemoveHelpfulVote(ctx, userID, reviewID); err != nil {
		return models.Review{}, err
	}
	return s.reviews.Get(ctx, reviewID)
}
