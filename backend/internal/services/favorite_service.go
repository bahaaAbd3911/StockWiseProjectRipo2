package services

import (
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"errors"
)

type FavoriteService struct {
	Repository *repository.FavoriteRepository
}

func NewFavoriteService(repo *repository.FavoriteRepository) *FavoriteService {
	return &FavoriteService{Repository: repo}
}

func (s *FavoriteService) AddFavorite(userID int64, productID int64) error {
	if userID <= 0 || productID <= 0 {
		return errors.New("invalid user ID or product ID")
	}

	return s.Repository.AddFavorite(userID, productID)
}

func (s *FavoriteService) GetFavorites(userID int64) ([]models.Favorite, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.Repository.GetFavorites(userID)
}

func (s *FavoriteService) RemoveFavorite(userID int64, productID int64) error {
	if userID <= 0 || productID <= 0 {
		return errors.New("invalid user ID or product ID")
	}

	return s.Repository.RemoveFavorite(userID, productID)
}
