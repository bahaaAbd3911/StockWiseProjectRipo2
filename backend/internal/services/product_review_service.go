package services

import (
	"errors"
	"strings"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
)

type ProductReviewService struct {
	Repository *repository.ProductReviewRepository
}

func NewProductReviewService(repo *repository.ProductReviewRepository) *ProductReviewService {
	return &ProductReviewService{
		Repository: repo,
	}
}

func (s *ProductReviewService) CreateReview(
	review models.ProductReview,
) (models.ProductReview, error) {

	if review.ProductID <= 0 || review.UserID <= 0 {
		return models.ProductReview{}, errors.New("invalid product ID or user ID")
	}

	review.Comment = strings.TrimSpace(review.Comment)

	if review.Comment == "" {
		return models.ProductReview{}, errors.New("review comment is required")
	}

	if review.CommentAr != nil {
		arabicComment := strings.TrimSpace(*review.CommentAr)
		if arabicComment == "" {
			review.CommentAr = nil
		} else {
			review.CommentAr = &arabicComment
		}
	}

	if review.Rating < 1 || review.Rating > 5 {
		return models.ProductReview{}, errors.New("rating must be between 1 and 5")
	}

	verified, err := s.Repository.HasDeliveredPurchase(review.UserID, review.ProductID)
	if err != nil {
		return models.ProductReview{}, err
	}

	review.VerifiedPurchase = verified

	return s.Repository.Create(review)
}

func (s *ProductReviewService) GetReviewsByProductID(
	productID int64,
) ([]models.ProductReview, error) {

	if productID <= 0 {
		return nil, errors.New("invalid product ID")
	}

	return s.Repository.GetByProductID(productID)
}

func (s *ProductReviewService) UpdateReview(
	reviewID int64,
	userID int64,
	comment string,
	commentAr *string,
	rating int,
) (models.ProductReview, error) {

	if reviewID <= 0 || userID <= 0 {
		return models.ProductReview{}, errors.New("invalid review ID or user ID")
	}

	comment = strings.TrimSpace(comment)

	if comment == "" {
		return models.ProductReview{}, errors.New("review comment is required")
	}

	if commentAr != nil {
		arabicComment := strings.TrimSpace(*commentAr)
		if arabicComment == "" {
			commentAr = nil
		} else {
			commentAr = &arabicComment
		}
	}

	if rating < 1 || rating > 5 {
		return models.ProductReview{}, errors.New("rating must be between 1 and 5")
	}

	return s.Repository.Update(reviewID, userID, comment, commentAr, rating)
}

func (s *ProductReviewService) DeleteReview(
	reviewID int64,
	userID int64,
) error {

	if reviewID <= 0 || userID <= 0 {
		return errors.New("invalid review ID or user ID")
	}

	return s.Repository.Delete(reviewID, userID)
}

func (s *ProductReviewService) AddHelpfulVote(reviewID, userID int64) error {
	if reviewID <= 0 || userID <= 0 {
		return errors.New("invalid review ID or user ID")
	}

	return s.Repository.AddHelpfulVote(reviewID, userID)
}

func (s *ProductReviewService) RemoveHelpfulVote(reviewID, userID int64) error {
	if reviewID <= 0 || userID <= 0 {
		return errors.New("invalid review ID or user ID")
	}

	return s.Repository.RemoveHelpfulVote(reviewID, userID)
}
