package repository

import (
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/models"
)

type ProductReviewRepository struct {
	DB *sql.DB
}

func NewProductReviewRepository(db *sql.DB) *ProductReviewRepository {
	return &ProductReviewRepository{
		DB: db,
	}
}

func (r *ProductReviewRepository) Create(
	review models.ProductReview,
) (models.ProductReview, error) {

	query := `
    INSERT INTO product_reviews (
        product_id, user_id, comment, comment_ar, rating, verified_purchase
    )
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING
        id, product_id, user_id, comment, comment_ar, rating,
        verified_purchase, helpful_count, is_edited, created_at;
    `

	var createdReview models.ProductReview

	err := r.DB.QueryRow(
		query,
		review.ProductID,
		review.UserID,
		review.Comment,
		review.CommentAr,
		review.Rating,
		review.VerifiedPurchase,
	).Scan(
		&createdReview.ID,
		&createdReview.ProductID,
		&createdReview.UserID,
		&createdReview.Comment,
		&createdReview.CommentAr,
		&createdReview.Rating,
		&createdReview.VerifiedPurchase,
		&createdReview.HelpfulCount,
		&createdReview.IsEdited,
		&createdReview.CreatedAt,
	)

	if err != nil {
		return models.ProductReview{}, fmt.Errorf(
			"failed to create product review: %w", err,
		)
	}

	return createdReview, nil
}

func (r *ProductReviewRepository) GetByProductID(
	productID int64,
) ([]models.ProductReview, error) {

	query := `
		SELECT
			id,
			product_id,
			user_id,
			comment,
			comment_ar,
			rating,
			verified_purchase,
			helpful_count,
			is_edited,
			created_at
		FROM product_reviews
		WHERE product_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.DB.Query(query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve product reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]models.ProductReview, 0)

	for rows.Next() {
		var review models.ProductReview

		err := rows.Scan(
			&review.ID,
			&review.ProductID,
			&review.UserID,
			&review.Comment,
			&review.CommentAr,
			&review.Rating,
			&review.VerifiedPurchase,
			&review.HelpfulCount,
			&review.IsEdited,
			&review.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product review: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product reviews: %w", err)
	}

	return reviews, nil
}

func (r *ProductReviewRepository) Update(
	reviewID int64,
	userID int64,
	comment string,
	commentAr *string,
	rating int,
) (models.ProductReview, error) {

	query := `
		UPDATE product_reviews
		SET
			comment = $1,
			comment_ar = $2,
			rating = $3,
			is_edited = TRUE
		WHERE id = $4 AND user_id = $5
		RETURNING
			id,
			product_id,
			user_id,
			comment,
			comment_ar,
			rating,
			verified_purchase,
			helpful_count,
			is_edited,
			created_at;
	`

	var updatedReview models.ProductReview

	err := r.DB.QueryRow(
		query,
		comment,
		commentAr,
		rating,
		reviewID,
		userID,
	).Scan(
		&updatedReview.ID,
		&updatedReview.ProductID,
		&updatedReview.UserID,
		&updatedReview.Comment,
		&updatedReview.CommentAr,
		&updatedReview.Rating,
		&updatedReview.VerifiedPurchase,
		&updatedReview.HelpfulCount,
		&updatedReview.IsEdited,
		&updatedReview.CreatedAt,
	)

	if err != nil {
		return models.ProductReview{}, fmt.Errorf(
			"failed to update product review: %w", err,
		)
	}

	return updatedReview, nil
}

func (r *ProductReviewRepository) Delete(
	reviewID int64,
	userID int64,
) error {

	query := `
		DELETE FROM product_reviews
		WHERE id = $1 AND user_id = $2;
	`

	result, err := r.DB.Exec(query, reviewID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete product review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deleted review: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (r *ProductReviewRepository) HasDeliveredPurchase(userID int64, productID int64) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM orders o
            JOIN order_items oi ON oi.order_id = o.id
            WHERE o.user_id = $1
              AND oi.product_id = $2
              AND o.status = 'delivered'
        );
    `

	var exists bool
	err := r.DB.QueryRow(query, userID, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check delivered purchase: %w", err)
	}

	return exists, nil
}
func (r *ProductReviewRepository) AddHelpfulVote(reviewID, userID int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var insertedReviewID int64

	query := `
        INSERT INTO review_helpful_votes (review_id, user_id)
        SELECT id, $2
        FROM product_reviews
        WHERE id = $1 AND user_id <> $2
        RETURNING review_id;
    `

	err = tx.QueryRow(query, reviewID, userID).Scan(&insertedReviewID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        UPDATE product_reviews
        SET helpful_count = helpful_count + 1
        WHERE id = $1;
    `, insertedReviewID)

	if err != nil {
		return err
	}

	return tx.Commit()
}
func (r *ProductReviewRepository) RemoveHelpfulVote(reviewID, userID int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var deletedReviewID int64

	err = tx.QueryRow(`
        DELETE FROM review_helpful_votes
        WHERE review_id = $1 AND user_id = $2
        RETURNING review_id;
    `, reviewID, userID).Scan(&deletedReviewID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        UPDATE product_reviews
        SET helpful_count = GREATEST(helpful_count - 1, 0)
        WHERE id = $1;
    `, deletedReviewID)

	if err != nil {
		return err
	}

	return tx.Commit()
}
