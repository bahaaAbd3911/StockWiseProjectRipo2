package models

import "time"

type Cart struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Items     []CartItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
type CartItem struct {
	ID            int64     `json:"id"`
	CartID        int64     `json:"cart_id"`
	ProductID     int64     `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductNameAr string    `json:"product_name_ar"`
	Price         float64   `json:"price"`
	Quantity      int       `json:"quantity"`
	SelectedColor *string   `json:"selected_color"`
	AddedAt       time.Time `json:"added_at"`
}
type AddCartItemRequest struct {
	ProductID     int64   `json:"product_id"`
	Quantity      int     `json:"quantity"`
	SelectedColor *string `json:"selected_color"`
}
