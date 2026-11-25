package repository

import (
	"context"
	"database/sql"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

// attachMedia fills ImageURLs (primary first) and Colors for a set of
// products with two queries in total, not two per product.
func attachMedia(ctx context.Context, q database.Querier, products ...*models.Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]int64, len(products))
	byID := make(map[int64][]*models.Product, len(products))
	for i, p := range products {
		ids[i] = p.ID
		p.ImageURLs, p.Colors = []string{}, []models.ProductColor{}
		byID[p.ID] = append(byID[p.ID], p)
	}

	rows, err := q.QueryContext(ctx, `
		SELECT product_id, image_url FROM product_images
		WHERE product_id = ANY($1)
		ORDER BY product_id, is_primary DESC, display_order, id`, ids)
	if err != nil {
		return err
	}
	for rows.Next() {
		var pid int64
		var url string
		if err := rows.Scan(&pid, &url); err != nil {
			rows.Close()
			return err
		}
		for _, p := range byID[pid] {
			p.ImageURLs = append(p.ImageURLs, url)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = q.QueryContext(ctx, `
		SELECT id, product_id, name, name_ar, hex_code FROM product_colors
		WHERE product_id = ANY($1)
		ORDER BY product_id, id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.ProductColor
		if err := rows.Scan(&c.ID, &c.ProductID, &c.Name, &c.NameAr, &c.HexCode); err != nil {
			return err
		}
		for _, p := range byID[c.ProductID] {
			p.Colors = append(p.Colors, c)
		}
	}
	return rows.Err()
}

// replaceMedia sets a product's images and/or colors to exactly the given
// lists (nil = leave unchanged). The first image becomes the primary one.
// Colors are matched by name, so carts that picked a color keep it when only
// its Arabic name or hex code changed.
func replaceMedia(ctx context.Context, tx *sql.Tx, productID int64, images *[]string, colors *[]models.ProductColorInput) error {
	if images != nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM product_images WHERE product_id = $1`, productID); err != nil {
			return err
		}
		for i, url := range *images {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO product_images (product_id, image_url, is_primary, display_order)
				VALUES ($1, $2, $3, $4)`, productID, url, i == 0, i); err != nil {
				return err
			}
		}
	}
	if colors != nil {
		names := make([]string, len(*colors))
		for i, c := range *colors {
			names[i] = c.Name
			res, err := tx.ExecContext(ctx, `
				UPDATE product_colors SET name_ar = $3, hex_code = $4
				WHERE product_id = $1 AND name = $2`, productID, c.Name, c.NameAr, c.HexCode)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO product_colors (product_id, name, name_ar, hex_code) VALUES ($1, $2, $3, $4)`,
					productID, c.Name, c.NameAr, c.HexCode); err != nil {
					return err
				}
			}
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM product_colors WHERE product_id = $1 AND NOT (name = ANY($2))`, productID, names); err != nil {
			return err
		}
	}
	return nil
}
