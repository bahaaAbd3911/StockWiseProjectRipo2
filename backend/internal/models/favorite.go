package models

import "time"

// Favorite is a saved product. The full product summary is embedded so the
// frontend can show price, image and stock without another request.
type Favorite struct {
	ID      int64     `json:"id"`
	AddedAt time.Time `json:"added_at"`
	Product Product   `json:"product"`
}

type AddFavoriteRequest struct {
	ProductID int64 `json:"product_id"`
}
