package repository

import (
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/models"
)

type FavoriteRepository struct {
	DB *sql.DB
}

func NewFavoriteRepository(db *sql.DB) *FavoriteRepository {
	return &FavoriteRepository{DB: db}
}

func (r *FavoriteRepository) AddFavorite(userID int64, productID int64) error {
	query := `
		INSERT INTO favorites (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
	`

	_, err := r.DB.Exec(query, userID, productID)
	if err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

func (r *FavoriteRepository) GetFavorites(userID int64) ([]models.Favorite, error) {
	query := `
		SELECT
			f.id,
			f.user_id,
			f.product_id,
			p.name,
			p.name_ar,
			p.price,
			f.added_at
		FROM favorites f
		JOIN products p ON p.id = f.product_id
		WHERE f.user_id = $1
		ORDER BY f.added_at DESC
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()

	favorites := []models.Favorite{}

	for rows.Next() {
		var favorite models.Favorite

		err := rows.Scan(
			&favorite.ID,
			&favorite.UserID,
			&favorite.ProductID,
			&favorite.ProductName,
			&favorite.ProductNameAr,
			&favorite.Price,
			&favorite.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}

		favorites = append(favorites, favorite)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read favorites: %w", err)
	}

	return favorites, nil
}

func (r *FavoriteRepository) RemoveFavorite(userID int64, productID int64) error {
	query := `
		DELETE FROM favorites
		WHERE user_id = $1 AND product_id = $2
	`

	result, err := r.DB.Exec(query, userID, productID)
	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check removed favorite: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
