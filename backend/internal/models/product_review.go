package models

import "time"

type ProductReview struct {
	ID               int64     `json:"id"`
	ProductID        int64     `json:"product_id"`
	UserID           int64     `json:"user_id"`
	Comment          string    `json:"comment"`
	CommentAr        *string   `json:"comment_ar"`
	Rating           int       `json:"rating"`
	VerifiedPurchase bool      `json:"verified_purchase"`
	HelpfulCount     int       `json:"helpful_count"`
	IsEdited         bool      `json:"is_edited"`
	CreatedAt        time.Time `json:"created_at"`
}
