package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Coupon scopes: what part of the cart the code discounts.
const (
	CouponScopeAll      = "all"      // the whole order (products and bundles)
	CouponScopeProducts = "products" // only the products listed in product_ids
	CouponScopeCategory = "category" // only products of one category
)

// Coupon statuses (computed, not stored).
const (
	CouponLive      = "live"
	CouponScheduled = "scheduled"
	CouponExpired   = "expired"
	CouponDisabled  = "disabled"
	CouponUsedUp    = "used_up" // usage_limit reached
)

// Coupon is a code customers type at checkout. Its discount comes off the
// order total AFTER the automatic offers (it stacks on top of them).
type Coupon struct {
	ID               int64            `json:"id"`
	Code             string           `json:"code"`
	Description      string           `json:"description"`
	DescriptionAr    string           `json:"description_ar"`
	DiscountType     string           `json:"discount_type"` // percentage | fixed (same as offers)
	DiscountValue    decimal.Decimal  `json:"discount_value"`
	Scope            string           `json:"scope"`
	Category         *string          `json:"category"`
	ProductIDs       []int64          `json:"product_ids"`
	MinOrderAmount   decimal.Decimal  `json:"min_order_amount"`
	MaxDiscount      *decimal.Decimal `json:"max_discount"`
	StartsAt         *time.Time       `json:"starts_at"`
	EndsAt           *time.Time       `json:"ends_at"`
	UsageLimit       *int             `json:"usage_limit"`
	PerCustomerLimit *int             `json:"per_customer_limit"`
	IsActive         bool             `json:"is_active"`
	Status           string           `json:"status"`
	TimesUsed        int              `json:"times_used"`     // orders that used it (cancelled orders give the use back)
	TotalDiscount    decimal.Decimal  `json:"total_discount"` // money given away through it
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// CouponInput is the admin body for creating or replacing a coupon.
type CouponInput struct {
	Code             string           `json:"code"`
	Description      string           `json:"description"`
	DescriptionAr    string           `json:"description_ar"`
	DiscountType     string           `json:"discount_type"`
	DiscountValue    decimal.Decimal  `json:"discount_value"`
	Scope            string           `json:"scope"` // default all
	Category         *string          `json:"category"`
	ProductIDs       []int64          `json:"product_ids"`
	MinOrderAmount   decimal.Decimal  `json:"min_order_amount"`
	MaxDiscount      *decimal.Decimal `json:"max_discount"`
	StartsAt         *time.Time       `json:"starts_at"`
	EndsAt           *time.Time       `json:"ends_at"`
	UsageLimit       *int             `json:"usage_limit"`
	PerCustomerLimit *int             `json:"per_customer_limit"`
	// PerCustomerLimitSet is true when the JSON contained per_customer_limit
	// (even as null = unlimited); otherwise the default of 1 is used.
	PerCustomerLimitSet bool  `json:"-"`
	IsActive            *bool `json:"is_active"`
}

type CouponFilter struct {
	Status string // live | scheduled | expired | disabled | used_up
	Search string // part of the code
	Limit  int
	Offset int
}

// ApplyCouponRequest: POST /api/cart/coupon
type ApplyCouponRequest struct {
	Code string `json:"code"`
}

// CartCoupon is the code attached to a cart and what it is worth right now.
// When it cannot be used (expired, minimum not reached...), Problem says why
// and the discount is 0; checkout is refused until it is removed or fixed.
type CartCoupon struct {
	Code           string          `json:"code"`
	Description    string          `json:"description"`
	DescriptionAr  string          `json:"description_ar"`
	DiscountType   string          `json:"discount_type"`
	DiscountValue  decimal.Decimal `json:"discount_value"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	Problem        *string         `json:"problem,omitempty"`
}

// UnmarshalJSON notes whether per_customer_limit was sent, so that leaving it
// out means "once per customer" while an explicit null means "unlimited".
func (in *CouponInput) UnmarshalJSON(b []byte) error {
	type plain CouponInput
	if err := json.Unmarshal(b, (*plain)(in)); err != nil {
		return err
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(b, &keys); err != nil {
		return err
	}
	_, in.PerCustomerLimitSet = keys["per_customer_limit"]
	return nil
}