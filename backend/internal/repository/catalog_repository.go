package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

// CatalogRepository manages product images and colors.
type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository { return &CatalogRepository{db: db} }

// ProductExists reports whether a product exists (optionally: and is active).
func (r *CatalogRepository) ProductExists(ctx context.Context, id int64, activeOnly bool) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1 AND (is_active OR NOT $2))`, id, activeOnly,
	).Scan(&ok)
	return ok, err
}

// ---------- images ----------

func (r *CatalogRepository) ListImages(ctx context.Context, productID int64) ([]models.ProductImage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, image_url, is_primary, display_order
		FROM product_images
		WHERE product_id = $1
		ORDER BY is_primary DESC, display_order, id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ProductImage{}
	for rows.Next() {
		var img models.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.IsPrimary, &img.DisplayOrder); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

// lockProduct locks the product row so concurrent image changes for the same
// product run one after another (needed for the "one primary image" rule).
func lockProduct(ctx context.Context, tx *sql.Tx, productID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&id)
	return apperr.FromDB(err, "product not found")
}

// AddImage inserts an image. The first image of a product always becomes the
// primary one; adding a new primary image un-marks the old one.
func (r *CatalogRepository) AddImage(ctx context.Context, productID int64, in models.ProductImageInput) (models.ProductImage, error) {
	var img models.ProductImage
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := lockProduct(ctx, tx, productID); err != nil {
			return err
		}
		var hasPrimary bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM product_images WHERE product_id = $1 AND is_primary)`, productID,
		).Scan(&hasPrimary); err != nil {
			return err
		}
		primary := in.IsPrimary || !hasPrimary
		if primary && hasPrimary {
			if _, err := tx.ExecContext(ctx,
				`UPDATE product_images SET is_primary = FALSE WHERE product_id = $1 AND is_primary`, productID); err != nil {
				return err
			}
		}
		return tx.QueryRowContext(ctx, `
			INSERT INTO product_images (product_id, image_url, is_primary, display_order)
			VALUES ($1, $2, $3, $4)
			RETURNING id, product_id, image_url, is_primary, display_order`,
			productID, in.ImageURL, primary, in.DisplayOrder,
		).Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.IsPrimary, &img.DisplayOrder)
	})
	return img, apperr.FromDB(err, "product not found")
}

// SetPrimaryImage makes one image the primary image of its product.
func (r *CatalogRepository) SetPrimaryImage(ctx context.Context, imageID int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var productID int64
		if err := tx.QueryRowContext(ctx,
			`SELECT product_id FROM product_images WHERE id = $1`, imageID,
		).Scan(&productID); err != nil {
			return apperr.FromDB(err, "image not found")
		}
		if err := lockProduct(ctx, tx, productID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE product_images SET is_primary = FALSE WHERE product_id = $1 AND is_primary`, productID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE product_images SET is_primary = TRUE WHERE id = $1`, imageID)
		return err
	})
}

// DeleteImage removes an image. If it was the primary one, the next image (by
// display_order) is promoted so the product keeps a cover picture. It returns
// the deleted image's URL so an uploaded file can be removed too.
func (r *CatalogRepository) DeleteImage(ctx context.Context, imageID int64) (imageURL string, err error) {
	err = database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var productID int64
		var wasPrimary bool
		if err := tx.QueryRowContext(ctx,
			`DELETE FROM product_images WHERE id = $1 RETURNING product_id, is_primary, image_url`, imageID,
		).Scan(&productID, &wasPrimary, &imageURL); err != nil {
			return apperr.FromDB(err, "image not found")
		}
		if !wasPrimary {
			return nil
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE product_images SET is_primary = TRUE
			WHERE id = (SELECT id FROM product_images WHERE product_id = $1
			            ORDER BY display_order, id LIMIT 1)`, productID)
		return err
	})
	return imageURL, err
}

// ---------- colors ----------

func (r *CatalogRepository) ListColors(ctx context.Context, productID int64) ([]models.ProductColor, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, name, name_ar, hex_code
		FROM product_colors WHERE product_id = $1 ORDER BY id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ProductColor{}
	for rows.Next() {
		var c models.ProductColor
		if err := rows.Scan(&c.ID, &c.ProductID, &c.Name, &c.NameAr, &c.HexCode); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CatalogRepository) AddColor(ctx context.Context, productID int64, in models.ProductColorInput) (models.ProductColor, error) {
	var c models.ProductColor
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO product_colors (product_id, name, name_ar, hex_code)
		VALUES ($1, $2, $3, $4)
		RETURNING id, product_id, name, name_ar, hex_code`,
		productID, in.Name, in.NameAr, in.HexCode,
	).Scan(&c.ID, &c.ProductID, &c.Name, &c.NameAr, &c.HexCode)
	if apperr.IsUniqueViolation(err, "unique_product_color") {
		return c, apperr.Conflict(fmt.Sprintf("this product already has a color named %q", in.Name))
	}
	return c, apperr.FromDB(err, "product not found")
}

func (r *CatalogRepository) DeleteColor(ctx context.Context, colorID int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM product_colors WHERE id = $1`, colorID)
	return expectOneRow(res, err, "color not found")
}
