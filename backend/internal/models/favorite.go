package models

import "time"

type Favorite struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	ProductID     int64     `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductNameAr string    `json:"product_name_ar"`
	Price         float64   `json:"price"`
	AddedAt       time.Time `json:"added_at"`
}

type AddFavoriteRequest struct {
	ProductID int64 `json:"product_id"`
}
