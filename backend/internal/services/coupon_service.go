package services

import (
	"context"
	"strings"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// CouponService: admins manage codes; customers apply one to their cart.
//
// The discount comes off the order AFTER the automatic offers (it stacks on
// top of them), on the whole order or only on listed products / one category.
// Limits: minimum order, maximum discount (percentage coupons), validity
// window, total uses, uses per customer (default once).
type CouponService struct {
	coupons *repository.CouponRepository
	carts   *repository.CartRepository
}

func NewCouponService(coupons *repository.CouponRepository, carts *repository.CartRepository) *CouponService {
	return &CouponService{coupons: coupons, carts: carts}
}

// ApplyToCart checks the code against the current cart and attaches it.
// Returns the updated cart (with the discount and the new total).
func (s *CouponService) ApplyToCart(ctx context.Context, userID int64, req models.ApplyCouponRequest) (models.Cart, error) {
	code := validators.NormalizeCouponCode(req.Code)
	if code == "" {
		return models.Cart{}, validators.FieldError("code", "is required")
	}
	if err := s.coupons.ApplyToCart(ctx, userID, code); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CouponService) RemoveFromCart(ctx context.Context, userID int64) (models.Cart, error) {
	if err := s.coupons.RemoveFromCart(ctx, userID); err != nil {
		return models.Cart{}, err
	}
	return s.carts.Get(ctx, userID)
}

// ---------- admin ----------

func (s *CouponService) List(ctx context.Context, f models.CouponFilter) (models.Page[models.Coupon], error) {
	f.Status = strings.ToLower(strings.TrimSpace(f.Status))
	if f.Status != "" && !repository.IsValidCouponStatus(f.Status) {
		return models.Page[models.Coupon]{}, validators.FieldError("status", "must be live, scheduled, expired, disabled or used_up")
	}
	f.Search = strings.TrimSpace(f.Search)
	return s.coupons.List(ctx, f)
}

func (s *CouponService) Get(ctx context.Context, id int64) (models.Coupon, error) {
	return s.coupons.Get(ctx, id)
}

func (s *CouponService) Create(ctx context.Context, in models.CouponInput) (models.Coupon, error) {
	if err := validators.ValidateCoupon(&in); err != nil {
		return models.Coupon{}, err
	}
	id, err := s.coupons.Create(ctx, in)
	if err != nil {
		return models.Coupon{}, err
	}
	return s.coupons.Get(ctx, id)
}

func (s *CouponService) Update(ctx context.Context, id int64, in models.CouponInput) (models.Coupon, error) {
	if err := validators.ValidateCoupon(&in); err != nil {
		return models.Coupon{}, err
	}
	if err := s.coupons.Update(ctx, id, in); err != nil {
		return models.Coupon{}, err
	}
	return s.coupons.Get(ctx, id)
}

func (s *CouponService) Delete(ctx context.Context, id int64) error {
	return s.coupons.Delete(ctx, id)
}
