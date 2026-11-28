package services

import (
	"context"
	"strings"
	"unicode/utf8"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// CartService: every change returns the updated cart, so the frontend can
// re-render from the response without another request.
type CartService struct {
	carts *repository.CartRepository
}

func NewCartService(carts *repository.CartRepository) *CartService { return &CartService{carts: carts} }

// Get returns the cart; city ("" = the default address) decides the shipping estimate.
func (s *CartService) Get(ctx context.Context, userID int64, city string) (models.Cart, error) {
	city = strings.TrimSpace(city)
	if utf8.RuneCountInString(city) > 100 {
		return models.Cart{}, validators.FieldError("city", "is too long (maximum 100 characters)")
	}
	return s.carts.GetForCity(ctx, userID, city)
}

func (s *CartService) AddItem(ctx context.Context, userID int64, req models.AddCartItemRequest) (models.Cart, error) {
	if req.Quantity == 0 {
		req.Quantity = 1 // "add to cart" without a quantity means one
	}
	if err := validators.ValidateAddCartItem(&req); err != nil {
		return models.Cart{}, err
	}
	if err := s.carts.AddItem(ctx, userID, req); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) UpdateItem(ctx context.Context, userID, itemID int64, req models.UpdateCartItemRequest) (models.Cart, error) {
	if err := validators.ValidateUpdateCartItem(&req); err != nil {
		return models.Cart{}, err
	}
	if err := s.carts.UpdateItem(ctx, userID, itemID, req); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) RemoveItem(ctx context.Context, userID, itemID int64) (models.Cart, error) {
	if err := s.carts.RemoveItem(ctx, userID, itemID); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) Clear(ctx context.Context, userID int64) (models.Cart, error) {
	if err := s.carts.Clear(ctx, userID); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

// ---------- bundles in the cart ----------

func (s *CartService) AddBundle(ctx context.Context, userID int64, req models.AddCartBundleRequest) (models.Cart, error) {
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	if err := validators.ValidateAddCartBundle(&req); err != nil {
		return models.Cart{}, err
	}
	if err := s.carts.AddBundle(ctx, userID, req); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) UpdateBundle(ctx context.Context, userID, lineID int64, req models.UpdateCartBundleRequest) (models.Cart, error) {
	c := validators.New()
	validators.ValidateQuantity(c, req.Quantity)
	if err := c.Err(); err != nil {
		return models.Cart{}, err
	}
	if err := s.carts.UpdateBundle(ctx, userID, lineID, req.Quantity); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) RemoveBundle(ctx context.Context, userID, lineID int64) (models.Cart, error) {
	if err := s.carts.RemoveBundle(ctx, userID, lineID); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}
