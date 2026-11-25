package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Product is a catalog item as shown in lists.
// FinalPrice, InStock, PrimaryImageURL, RatingAverage and ReviewCount are
// computed (not stored in the products table).
type Product struct {
	ID              int64           `json:"id"`
	Name            string          `json:"name"`
	NameAr          string          `json:"name_ar"`
	Description     string          `json:"description"`
	DescriptionAr   string          `json:"description_ar"`
	Category        string          `json:"category"`
	CategoryAr      string          `json:"category_ar"`
	Price           decimal.Decimal `json:"price"`
	DiscountValue   decimal.Decimal `json:"discount_value"` // percent, 0..100
	FinalPrice      decimal.Decimal `json:"final_price"`    // best price: product discount or live offer
	Offer           *AppliedOffer   `json:"offer"`          // the offer that gave final_price, if any
	Stock           int             `json:"stock"`
	InStock         bool            `json:"in_stock"`
	IsActive        bool            `json:"is_active"`
	PrimaryImageURL *string         `json:"primary_image_url"`
	RatingAverage   float64         `json:"rating_average"`
	ReviewCount     int             `json:"review_count"`
	// ImageURLs are all images, primary first, and Colors the colors a customer
	// can pick. They are included in lists so a product card can show a gallery
	// and add the product to the cart (a product with colors needs one).
	ImageURLs []string       `json:"image_urls"`
	Colors    []ProductColor `json:"colors"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// ProductDetail is a single product with its full image records (ids,
// primary flag, order - for the admin) on top of everything in Product.
type ProductDetail struct {
	Product
	Images []ProductImage `json:"images"`
}

// ProductInput is the body for creating or replacing a product (admin).
// IsActive is a pointer so "not sent" can be told apart from false:
// on create it defaults to true, on update it keeps the current value.
type ProductInput struct {
	Name          string          `json:"name"`
	NameAr        string          `json:"name_ar"`
	Description   string          `json:"description"`
	DescriptionAr string          `json:"description_ar"`
	Category      string          `json:"category"`
	CategoryAr    string          `json:"category_ar"`
	Price         decimal.Decimal `json:"price"`
	DiscountValue decimal.Decimal `json:"discount_value"`
	Stock         int             `json:"stock"`
	IsActive      *bool           `json:"is_active"`
	// Optional: when sent (even as []), the product's images / colors are
	// replaced by these in the same transaction, so a whole product form is
	// saved in one request. Leave them out to keep the current ones (the
	// separate image and color endpoints still work).
	Images *[]string            `json:"images"` // URLs; the first becomes the primary image
	Colors *[]ProductColorInput `json:"colors"`
}

// ProductFilter controls GET /api/products.
type ProductFilter struct {
	Category        string
	Search          string
	Sort            string // newest | price_asc | price_desc | name (price = final price incl. offers)
	InStockOnly     bool
	OnOfferOnly     bool  // only products with a live offer
	OfferID         int64 // only products covered by this live offer
	IncludeInactive bool  // admin listing only
	Limit           int
	Offset          int
}

// Category is one entry of GET /api/categories.
type Category struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	NameAr       string    `json:"name_ar"`
	Icon         string    `json:"icon"` // e.g. "fa-solid fa-laptop"
	IsActive     bool      `json:"is_active"`
	ProductCount int       `json:"product_count"` // active products in it
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CategoryInput is the admin body for creating or editing a category.
// IsActive omitted: true on create, unchanged on edit.
type CategoryInput struct {
	Name     string `json:"name"`
	NameAr   string `json:"name_ar"`
	Icon     string `json:"icon"`
	IsActive *bool  `json:"is_active"`
}

type ProductImage struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	ImageURL     string `json:"image_url"`
	IsPrimary    bool   `json:"is_primary"`
	DisplayOrder int    `json:"display_order"`
}

type ProductImageInput struct {
	ImageURL     string `json:"image_url"`
	IsPrimary    bool   `json:"is_primary"`
	DisplayOrder int    `json:"display_order"`
}

type ProductColor struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	NameAr    string `json:"name_ar"`
	HexCode   string `json:"hex_code"`
}

type ProductColorInput struct {
	Name    string `json:"name"`
	NameAr  string `json:"name_ar"`
	HexCode string `json:"hex_code"`
}
