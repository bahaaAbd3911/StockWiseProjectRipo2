package models

import "time"

type Product struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	NameAr        string    `json:"name_ar"`
	Description   string    `json:"description"`
	DescriptionAr string    `json:"description_ar"`
	Category      string    `json:"category"`
	CategoryAr    string    `json:"category_ar"`
	Price         float64   `json:"price"`
	DiscountValue float64   `json:"discount_value"`
	Stock         int       `json:"stock"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	IsActive      bool      `json:"is_active"`
}
type CreateProductRequest struct {
	Name          string  `json:"name"`
	NameAr        string  `json:"name_ar"`
	Description   string  `json:"description"`
	DescriptionAr string  `json:"description_ar"`
	Category      string  `json:"category"`
	CategoryAr    string  `json:"category_ar"`
	Price         float64 `json:"price"`
	DiscountValue float64 `json:"discount_value"`
	Stock         int     `json:"stock"`
	IsActive      bool    `json:"is_active"`
}
type UpdateProductRequest struct {
	Name          string  `json:"name"`
	NameAr        string  `json:"name_ar"`
	Description   string  `json:"description"`
	DescriptionAr string  `json:"description_ar"`
	Category      string  `json:"category"`
	CategoryAr    string  `json:"category_ar"`
	Price         float64 `json:"price"`
	DiscountValue float64 `json:"discount_value"`
	Stock         int     `json:"stock"`
	IsActive      bool    `json:"is_active"`
}
