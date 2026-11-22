package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Bundle is several products sold together at one price. Offers never apply
// to bundles: the bundle price is already the deal.
type Bundle struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	NameAr        string          `json:"name_ar"`
	Description   string          `json:"description"`
	DescriptionAr string          `json:"description_ar"`
	ImageURL      *string         `json:"image_url"`
	Price         decimal.Decimal `json:"price"`
	StartsAt      *time.Time      `json:"starts_at"` // optional sale window
	EndsAt        *time.Time      `json:"ends_at"`
	IsActive      bool            `json:"is_active"`
	Items         []BundleItem    `json:"items"`
	// Computed:
	OriginalTotal  decimal.Decimal `json:"original_total"`  // what the items cost separately, today
	Savings        decimal.Decimal `json:"savings"`         // original_total - price (never below 0)
	AvailableStock int             `json:"available_stock"` // how many bundles can be sold right now
	IsAvailable    bool            `json:"is_available"`
	// UnavailableReason explains why IsAvailable is false.
	UnavailableReason *string   `json:"unavailable_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// BundleItem is one product inside a bundle.
type BundleItem struct {
	ProductID       int64           `json:"product_id"`
	ProductName     string          `json:"product_name"`
	ProductNameAr   string          `json:"product_name_ar"`
	PrimaryImageURL *string         `json:"primary_image_url"`
	Quantity        int             `json:"quantity"` // per bundle
	SelectedColor   *string         `json:"selected_color"`
	SelectedColorAr *string         `json:"selected_color_ar"`
	UnitPrice       decimal.Decimal `json:"unit_price"` // the product's own price today (for comparison)
	Stock           int             `json:"stock"`
	IsActive        bool            `json:"is_active"`
}

// BundleInput is the admin body for creating or replacing a bundle.
type BundleInput struct {
	Name          string            `json:"name"`
	NameAr        string            `json:"name_ar"`
	Description   string            `json:"description"`
	DescriptionAr string            `json:"description_ar"`
	ImageURL      *string           `json:"image_url"`
	Price         decimal.Decimal   `json:"price"`
	StartsAt      *time.Time        `json:"starts_at"`
	EndsAt        *time.Time        `json:"ends_at"`
	IsActive      *bool             `json:"is_active"`
	Items         []BundleItemInput `json:"items"`
}

type BundleItemInput struct {
	ProductID     int64   `json:"product_id"`
	Quantity      int     `json:"quantity"`
	SelectedColor *string `json:"selected_color"`
}

type BundleFilter struct {
	IncludeUnavailable bool // admin: also inactive / outside the window
	Limit              int
	Offset             int
}

// CartBundle is a bundle sitting in the cart.
type CartBundle struct {
	ID                int64           `json:"id"`
	BundleID          int64           `json:"bundle_id"`
	Name              string          `json:"name"`
	NameAr            string          `json:"name_ar"`
	ImageURL          *string         `json:"image_url"`
	UnitPrice         decimal.Decimal `json:"unit_price"`
	Quantity          int             `json:"quantity"`
	LineTotal         decimal.Decimal `json:"line_total"`
	Items             []BundleItem    `json:"items"`
	IsAvailable       bool            `json:"is_available"`
	UnavailableReason *string         `json:"unavailable_reason,omitempty"`
	AddedAt           time.Time       `json:"added_at"`
}

type AddCartBundleRequest struct {
	BundleID int64 `json:"bundle_id"`
	Quantity int   `json:"quantity"`
}

type UpdateCartBundleRequest struct {
	Quantity int `json:"quantity"`
}

// OrderBundle is the snapshot of a bundle inside an order.
type OrderBundle struct {
	ID           int64             `json:"id"`
	OrderID      int64             `json:"order_id"`
	BundleID     *int64            `json:"bundle_id"`
	BundleName   string            `json:"bundle_name"`
	BundleNameAr *string           `json:"bundle_name_ar"`
	Quantity     int               `json:"quantity"`
	UnitPrice    decimal.Decimal   `json:"unit_price"`
	Subtotal     decimal.Decimal   `json:"subtotal"`
	Items        []OrderBundleItem `json:"items"`
}

type OrderBundleItem struct {
	ProductID       *int64  `json:"product_id"`
	ProductName     string  `json:"product_name"`
	ProductNameAr   *string `json:"product_name_ar"`
	Quantity        int     `json:"quantity"` // per bundle
	SelectedColor   *string `json:"selected_color"`
	SelectedColorAr *string `json:"selected_color_ar"`
}
