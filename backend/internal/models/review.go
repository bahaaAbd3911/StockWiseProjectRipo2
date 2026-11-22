package models

import "time"

type Review struct {
	ID               int64     `json:"id"`
	ProductID        int64     `json:"product_id"`
	UserID           int64     `json:"user_id"`
	ReviewerName     string    `json:"reviewer_name"` // "Sara K." - never the full surname or email
	Comment          string    `json:"comment"`
	CommentAr        *string   `json:"comment_ar"`
	Rating           int       `json:"rating"`
	VerifiedPurchase bool      `json:"verified_purchase"`
	HelpfulCount     int       `json:"helpful_count"`
	IsEdited         bool      `json:"is_edited"`
	CreatedAt        time.Time `json:"created_at"`
}

type ReviewInput struct {
	Comment   string  `json:"comment"`
	CommentAr *string `json:"comment_ar"`
	Rating    int     `json:"rating"`
}
