package services

import (
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"fmt"
)

type CartService struct {
	Repository *repository.CartRepository
}

func NewCartService(repo *repository.CartRepository) *CartService {
	return &CartService{
		Repository: repo,
	}
}

func (s *CartService) GetCart(userID int64) (models.Cart, error) {
	return s.Repository.GetCartByUserID(userID)
}

// Add "fmt" to your imports at the top of the file if it's not already there.

func (s *CartService) AddItem(userID int64, request models.AddCartItemRequest) error {
	if request.ProductID <= 0 {
		return fmt.Errorf("invalid product ID")
	}

	if request.Quantity <= 0 {
		return fmt.Errorf("quantity must be at least 1")
	}

	return s.Repository.AddItem(userID, request)
}

func (s *CartService) UpdateItemQuantity(
	userID int64,
	itemID int64,
	quantity int,
) error {
	if itemID <= 0 {
		return fmt.Errorf("invalid cart item ID")
	}

	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}

	return s.Repository.UpdateItemQuantity(userID, itemID, quantity)
}

func (s *CartService) RemoveItem(userID int64, itemID int64) error {
	if itemID <= 0 {
		return fmt.Errorf("invalid cart item ID")
	}

	return s.Repository.RemoveItem(userID, itemID)
}

func (s *CartService) ClearCart(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	return s.Repository.ClearCart(userID)
}
