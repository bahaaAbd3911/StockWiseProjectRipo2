package services

import (
	"context"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

type FavoriteService struct {
	favorites *repository.FavoriteRepository
}

func NewFavoriteService(favorites *repository.FavoriteRepository) *FavoriteService {
	return &FavoriteService{favorites: favorites}
}

func (s *FavoriteService) List(ctx context.Context, userID int64) ([]models.Favorite, error) {
	return s.favorites.List(ctx, userID)
}

// Add returns created=false when the product was already a favorite.
func (s *FavoriteService) Add(ctx context.Context, userID, productID int64) (bool, error) {
	if productID <= 0 {
		return false, validators.FieldError("product_id", "is required")
	}
	return s.favorites.Add(ctx, userID, productID)
}

func (s *FavoriteService) Remove(ctx context.Context, userID, productID int64) error {
	return s.favorites.Remove(ctx, userID, productID)
}
