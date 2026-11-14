package services

import (
	"errors"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
)

type OrderService struct {
	Repository *repository.OrderRepository
}

func NewOrderService(
	repo *repository.OrderRepository,
) *OrderService {
	return &OrderService{
		Repository: repo,
	}
}

func (s *OrderService) CreateOrder(
	userID int64,
	request models.CreateOrderRequest,
) (models.Order, error) {

	if userID <= 0 {
		return models.Order{}, errors.New("invalid user ID")
	}

	if request.AddressID <= 0 {
		return models.Order{}, errors.New("valid address ID is required")
	}

	return s.Repository.CreateOrder(
		userID,
		request.AddressID,
	)
}
func (s *OrderService) GetOrdersByUserID(
	userID int64,
) ([]models.Order, error) {

	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.Repository.GetOrdersByUserID(userID)
}
func (s *OrderService) GetOrderByID(
	userID int64,
	orderID int64,
) (models.Order, error) {

	if userID <= 0 || orderID <= 0 {
		return models.Order{}, errors.New("invalid user ID or order ID")
	}

	return s.Repository.GetOrderByID(userID, orderID)
}
func (s *OrderService) CancelOrder(
	userID int64,
	orderID int64,
) error {

	if userID <= 0 || orderID <= 0 {
		return errors.New("invalid user ID or order ID")
	}

	return s.Repository.CancelOrder(userID, orderID)
}
