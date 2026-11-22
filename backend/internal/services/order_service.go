package services

import (
	"context"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

type OrderService struct {
	orders *repository.OrderRepository
}

func NewOrderService(orders *repository.OrderRepository) *OrderService {
	return &OrderService{orders: orders}
}

// Checkout places an order from the cart and returns the created order.
func (s *OrderService) Checkout(ctx context.Context, userID int64, req models.CreateOrderRequest) (models.Order, error) {
	if req.AddressID != nil && *req.AddressID <= 0 {
		return models.Order{}, validators.FieldError("address_id", "must be a positive number")
	}
	id, err := s.orders.Checkout(ctx, userID, req.AddressID)
	if err != nil {
		return models.Order{}, err
	}
	return s.orders.GetForUser(ctx, userID, id)
}

func (s *OrderService) ListMine(ctx context.Context, userID int64) ([]models.Order, error) {
	return s.orders.ListForUser(ctx, userID)
}

func (s *OrderService) GetMine(ctx context.Context, userID, orderID int64) (models.Order, error) {
	return s.orders.GetForUser(ctx, userID, orderID)
}

func (s *OrderService) CancelMine(ctx context.Context, userID, orderID int64) (models.Order, error) {
	if err := s.orders.CancelByUser(ctx, userID, orderID); err != nil {
		return models.Order{}, err
	}
	return s.orders.GetForUser(ctx, userID, orderID)
}

// ---------- admin ----------

func (s *OrderService) List(ctx context.Context, f models.OrderFilter) (models.Page[models.Order], error) {
	f.Status = strings.TrimSpace(f.Status)
	if f.Status != "" && !models.IsOrderStatus(f.Status) {
		return models.Page[models.Order]{}, validators.FieldError("status", "is not a valid order status")
	}
	return s.orders.List(ctx, f)
}

func (s *OrderService) Get(ctx context.Context, orderID int64) (models.Order, error) {
	return s.orders.Get(ctx, orderID)
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID int64, status string) (models.Order, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if !models.IsOrderStatus(status) {
		return models.Order{}, validators.FieldError("status",
			"must be one of: confirmed, processing, shipped, delivered, cancelled")
	}
	if status == models.OrderPending {
		return models.Order{}, apperr.Conflict("an order cannot be moved back to pending")
	}
	if err := s.orders.UpdateStatus(ctx, orderID, status); err != nil {
		return models.Order{}, err
	}
	return s.orders.Get(ctx, orderID)
}
