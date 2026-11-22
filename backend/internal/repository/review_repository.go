package repository

import (
	"context"
	"database/sql"
	"errors"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository { return &ReviewRepository{db: db} }

// The reviewer is shown as "First L." - the full surname and email stay private.
const reviewColumns = `
	r.id, r.product_id, r.user_id,
	u.first_name || ' ' || LEFT(u.last_name, 1) || '.',
	r.comment, r.comment_ar, r.rating, r.verified_purchase, r.helpful_count, r.is_edited, r.created_at`

const reviewFrom = ` FROM product_reviews r JOIN users u ON u.id = r.user_id`

func scanReview(s scanner) (models.Review, error) {
	var v models.Review
	err := s.Scan(&v.ID, &v.ProductID, &v.UserID, &v.ReviewerName,
		&v.Comment, &v.CommentAr, &v.Rating, &v.VerifiedPurchase, &v.HelpfulCount, &v.IsEdited, &v.CreatedAt)
	return v, err
}

var reviewSorts = map[string]string{
	"newest":      "r.created_at DESC, r.id DESC",
	"helpful":     "r.helpful_count DESC, r.created_at DESC",
	"rating_high": "r.rating DESC, r.created_at DESC",
	"rating_low":  "r.rating ASC, r.created_at DESC",
}

func IsValidReviewSort(s string) bool {
	_, ok := reviewSorts[s]
	return ok
}

func (r *ReviewRepository) ListForProduct(ctx context.Context, productID int64, sort string) ([]models.Review, error) {
	order, ok := reviewSorts[sort]
	if !ok {
		order = reviewSorts["newest"]
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+reviewColumns+reviewFrom+` WHERE r.product_id = $1 ORDER BY `+order, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Review{}
	for rows.Next() {
		v, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *ReviewRepository) Get(ctx context.Context, reviewID int64) (models.Review, error) {
	v, err := scanReview(r.db.QueryRowContext(ctx, `SELECT `+reviewColumns+reviewFrom+` WHERE r.id = $1`, reviewID))
	return v, apperr.FromDB(err, "review not found")
}

// HasDeliveredPurchase is true when the user has a delivered order containing
// the product, on its own or inside a bundle. It decides the "verified purchase" badge.
func (r *ReviewRepository) HasDeliveredPurchase(ctx context.Context, userID, productID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM orders o JOIN order_items oi ON oi.order_id = o.id
			WHERE o.user_id = $1 AND oi.product_id = $2 AND o.status = 'delivered'
		) OR EXISTS (
			SELECT 1 FROM orders o
			JOIN order_bundles ob ON ob.order_id = o.id
			JOIN order_bundle_items obi ON obi.order_bundle_id = ob.id
			WHERE o.user_id = $1 AND obi.product_id = $2 AND o.status = 'delivered')`,
		userID, productID).Scan(&ok)
	return ok, err
}

func (r *ReviewRepository) Create(ctx context.Context, userID, productID int64, in models.ReviewInput, verified bool) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO product_reviews (product_id, user_id, comment, comment_ar, rating, verified_purchase)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`, productID, userID, in.Comment, in.CommentAr, in.Rating, verified).Scan(&id)
	if apperr.IsUniqueViolation(err, "unique_user_product_review") {
		return 0, apperr.Conflict("you have already reviewed this product; edit your existing review instead")
	}
	return id, apperr.FromDB(err, "product not found")
}

// Update only succeeds for the review's author.
func (r *ReviewRepository) Update(ctx context.Context, userID, reviewID int64, in models.ReviewInput) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE product_reviews SET comment = $1, comment_ar = $2, rating = $3, is_edited = TRUE
		WHERE id = $4 AND user_id = $5`, in.Comment, in.CommentAr, in.Rating, reviewID, userID)
	return expectOneRow(res, err, "review not found")
}

// Delete only succeeds for the author (helpful votes are removed by CASCADE).
func (r *ReviewRepository) Delete(ctx context.Context, userID, reviewID int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM product_reviews WHERE id = $1 AND user_id = $2`, reviewID, userID)
	return expectOneRow(res, err, "review not found")
}

// AddHelpfulVote records one vote and increments helpful_count in the same
// transaction, so the counter always equals the number of vote rows.
func (r *ReviewRepository) AddHelpfulVote(ctx context.Context, userID, reviewID int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var authorID int64
		if err := tx.QueryRowContext(ctx,
			`SELECT user_id FROM product_reviews WHERE id = $1`, reviewID).Scan(&authorID); err != nil {
			return apperr.FromDB(err, "review not found")
		}
		if authorID == userID {
			return apperr.Invalid("you cannot vote on your own review")
		}
		var voteID int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO review_helpful_votes (review_id, user_id) VALUES ($1, $2)
			ON CONFLICT (review_id, user_id) DO NOTHING
			RETURNING id`, reviewID, userID).Scan(&voteID)
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.Conflict("you have already marked this review as helpful")
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE product_reviews SET helpful_count = helpful_count + 1 WHERE id = $1`, reviewID)
		return err
	})
}

func (r *ReviewRepository) RemoveHelpfulVote(ctx context.Context, userID, reviewID int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var voteID int64
		err := tx.QueryRowContext(ctx,
			`DELETE FROM review_helpful_votes WHERE review_id = $1 AND user_id = $2 RETURNING id`,
			reviewID, userID).Scan(&voteID)
		if err != nil {
			return apperr.FromDB(err, "you have not marked this review as helpful")
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE product_reviews SET helpful_count = GREATEST(helpful_count - 1, 0) WHERE id = $1`, reviewID)
		return err
	})
}
