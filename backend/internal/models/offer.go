package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Offer discount types.
const (
	OfferPercentage = "percentage" // discount_value is a percent (0 < v <= 100)
	OfferFixed      = "fixed"      // discount_value is an amount taken off the price
)

// Offer scopes.
const (
	OfferScopeProducts = "products" // applies to the products listed in product_ids
	OfferScopeCategory = "category" // applies to every product of one category
)

// Offer is a timed promotion. It is "live" when is_active is true and now is
// inside [starts_at, ends_at).
type Offer struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	NameAr        string          `json:"name_ar"`
	Description   string          `json:"description"`
	DescriptionAr string          `json:"description_ar"`
	DiscountType  string          `json:"discount_type"`
	DiscountValue decimal.Decimal `json:"discount_value"`
	Scope         string          `json:"scope"`
	Category      *string         `json:"category"`
	ProductIDs    []int64         `json:"product_ids"`
	StartsAt      time.Time       `json:"starts_at"`
	EndsAt        time.Time       `json:"ends_at"`
	IsActive      bool            `json:"is_active"`
	Status        string          `json:"status"` // computed: live | scheduled | expired | disabled
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Offer statuses (computed by the database query, not stored).
const (
	OfferLive      = "live"
	OfferScheduled = "scheduled"
	OfferExpired   = "expired"
	OfferDisabled  = "disabled"
)

// OfferInput is the admin body for creating or replacing an offer.
type OfferInput struct {
	Name          string          `json:"name"`
	NameAr        string          `json:"name_ar"`
	Description   string          `json:"description"`
	DescriptionAr string          `json:"description_ar"`
	DiscountType  string          `json:"discount_type"`
	DiscountValue decimal.Decimal `json:"discount_value"`
	Scope         string          `json:"scope"`
	Category      *string         `json:"category"`
	ProductIDs    []int64         `json:"product_ids"`
	StartsAt      time.Time       `json:"starts_at"`
	EndsAt        time.Time       `json:"ends_at"`
	IsActive      *bool           `json:"is_active"`
}

// AppliedOffer is the short form of an offer shown on a product, cart line or
// order item: "this price comes from offer X, which ends at Y".
type AppliedOffer struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	NameAr        string          `json:"name_ar"`
	DiscountType  string          `json:"discount_type"`
	DiscountValue decimal.Decimal `json:"discount_value"`
	EndsAt        time.Time       `json:"ends_at"`
}

type OfferFilter struct {
	Status string // admin: live | scheduled | expired | disabled
	Limit  int
	Offset int
}
