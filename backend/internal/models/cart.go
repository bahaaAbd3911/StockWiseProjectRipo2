package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Cart is the user's cart with live prices.
// Subtotal only counts items that can actually be bought right now.
type Cart struct {
	ID        int64           `json:"id"`
	Items     []CartItem      `json:"items"`
	Bundles   []CartBundle    `json:"bundles"`
	ItemCount int             `json:"item_count"` // sum of quantities (products + bundles)
	Subtotal  decimal.Decimal `json:"subtotal"`
	Coupon    *CartCoupon     `json:"coupon"`   // the code typed by the customer, if any
	Discount  decimal.Decimal `json:"discount"` // what the coupon takes off (0 without one)
	// Shipping for ShippingCity: the city the page asked about (?city=), else
	// the default address's city, else "" (the standard fee). Checkout
	// charges the fee of the order's real delivery city.
	Shipping            decimal.Decimal  `json:"shipping"`
	ShippingCity        string           `json:"shipping_city"`
	FreeShippingOver    *decimal.Decimal `json:"free_shipping_over"` // items total (after the coupon) that ships free; null = never
	Total               decimal.Decimal  `json:"total"`              // subtotal - discount + shipping: what checkout will charge
	HasUnavailableItems bool             `json:"has_unavailable_items"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type CartItem struct {
	ID               int64           `json:"id"`
	ProductID        int64           `json:"product_id"`
	ProductName      string          `json:"product_name"`
	ProductNameAr    string          `json:"product_name_ar"`
	PrimaryImageURL  *string         `json:"primary_image_url"`
	OriginalPrice    decimal.Decimal `json:"original_price"`
	DiscountValue    decimal.Decimal `json:"discount_value"`
	UnitPrice        decimal.Decimal `json:"unit_price"` // best price: discount or live offer
	Offer            *AppliedOffer   `json:"offer"`
	Quantity         int             `json:"quantity"`
	LineTotal        decimal.Decimal `json:"line_total"`
	SelectedColor    *string         `json:"selected_color"`
	SelectedColorAr  *string         `json:"selected_color_ar"`
	SelectedColorHex *string         `json:"selected_color_hex"`
	Stock            int             `json:"stock"`
	IsAvailable      bool            `json:"is_available"`
	// UnavailableReason explains why IsAvailable is false.
	UnavailableReason *string   `json:"unavailable_reason,omitempty"`
	AddedAt           time.Time `json:"added_at"`
}

type AddCartItemRequest struct {
	ProductID     int64   `json:"product_id"`
	Quantity      int     `json:"quantity"`
	SelectedColor *string `json:"selected_color"`
}

// UpdateCartItemRequest changes the quantity and, optionally, the color.
type UpdateCartItemRequest struct {
	Quantity      int     `json:"quantity"`
	SelectedColor *string `json:"selected_color"`
}
