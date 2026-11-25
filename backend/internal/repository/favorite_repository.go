package repository

import (
	"context"
	"database/sql"
	"errors"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
)

type FavoriteRepository struct {
	db *sql.DB
}

func NewFavoriteRepository(db *sql.DB) *FavoriteRepository { return &FavoriteRepository{db: db} }

// Add saves a product as favorite. Adding it twice is harmless: created is
// false the second time.
func (r *FavoriteRepository) Add(ctx context.Context, userID, productID int64) (created bool, err error) {
	var active bool
	err = r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM products WHERE id = $1 AND is_active)`, productID).Scan(&active)
	if err != nil {
		return false, err
	}
	if !active {
		return false, apperr.NotFound("product not found")
	}

	var id int64
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO favorites (user_id, product_id) VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
		RETURNING id`, userID, productID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // already a favorite
	}
	return err == nil, apperr.FromDB(err, "product not found")
}

// List returns favorites with a full product summary. Products that were
// deactivated stay in the list with is_active=false so the user sees why.
func (r *FavoriteRepository) List(ctx context.Context, userID int64) ([]models.Favorite, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.id, f.added_at, `+productColumns+`
		FROM favorites f
		JOIN products p ON p.id = f.product_id
		WHERE f.user_id = $1
		ORDER BY f.added_at DESC, f.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Favorite{}
	for rows.Next() {
		var fav models.Favorite
		p, err := scanProduct(rows, &fav.ID, &fav.AddedAt)
		if err != nil {
			return nil, err
		}
		fav.Product = p
		out = append(out, fav)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ptrs := make([]*models.Product, len(out))
	for i := range out {
		ptrs[i] = &out[i].Product
	}
	if err := attachOffers(ctx, r.db, ptrs...); err != nil {
		return out, err
	}
	return out, attachMedia(ctx, r.db, ptrs...)
}

func (r *FavoriteRepository) Remove(ctx context.Context, userID, productID int64) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND product_id = $2`, userID, productID)
	return expectOneRow(res, err, "favorite not found")
}
