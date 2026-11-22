// Package repository is the only layer that talks SQL.
//
// Rules every repository follows:
//   - every query takes a context (so a cancelled request stops its query)
//   - user-owned rows are ALWAYS filtered by user_id, so one user can never read
//     or change another user's data
//   - multi-step writes run inside database.WithTx (all-or-nothing)
//   - rows that may be changed concurrently are locked with SELECT ... FOR UPDATE
//   - database errors are translated with apperr.FromDB into not-found /
//     conflict / invalid errors the HTTP layer understands
package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"
)

// scanner is satisfied by *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// productColumns selects a product (alias p) plus computed values: its primary
// image, average rating and review count. Used by products and favorites.
const productColumns = `
	p.id, p.name, p.name_ar,
	COALESCE(p.description, ''), COALESCE(p.description_ar, ''),
	COALESCE(p.category, ''), COALESCE(p.category_ar, ''),
	p.price, p.discount_value, p.stock, p.is_active, p.created_at, p.updated_at,
	(SELECT pi.image_url FROM product_images pi
	  WHERE pi.product_id = p.id
	  ORDER BY pi.is_primary DESC, pi.display_order, pi.id LIMIT 1),
	COALESCE((SELECT ROUND(AVG(r.rating), 1) FROM product_reviews r WHERE r.product_id = p.id), 0)::float8,
	(SELECT COUNT(*) FROM product_reviews r WHERE r.product_id = p.id)`

// scanProduct reads productColumns (after any prefix columns) and fills in the
// computed fields FinalPrice and InStock.
func scanProduct(s scanner, prefix ...any) (models.Product, error) {
	var p models.Product
	dest := append(prefix,
		&p.ID, &p.Name, &p.NameAr,
		&p.Description, &p.DescriptionAr,
		&p.Category, &p.CategoryAr,
		&p.Price, &p.DiscountValue, &p.Stock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&p.PrimaryImageURL, &p.RatingAverage, &p.ReviewCount,
	)
	if err := s.Scan(dest...); err != nil {
		return models.Product{}, err
	}
	p.FinalPrice = pricing.DiscountedPrice(p.Price, p.DiscountValue)
	p.InStock = p.Stock > 0
	return p, nil
}

// resolveColor checks a requested color against the product's color options.
//
//   - product has no colors  -> color must be empty
//   - product has colors     -> a color is required and must exist (matched
//     case-insensitively); the stored spelling is returned
func resolveColor(ctx context.Context, q database.Querier, productID int64, color *string) (*models.ProductColor, error) {
	var count int
	if err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM product_colors WHERE product_id = $1`, productID,
	).Scan(&count); err != nil {
		return nil, err
	}
	if count == 0 {
		if color != nil {
			return nil, apperr.Validation(map[string]string{"selected_color": "this product has no color options"})
		}
		return nil, nil
	}
	if color == nil {
		return nil, apperr.Validation(map[string]string{"selected_color": "please choose a color for this product"})
	}

	var c models.ProductColor
	err := q.QueryRowContext(ctx, `
		SELECT id, product_id, name, name_ar, hex_code
		FROM product_colors
		WHERE product_id = $1 AND lower(name) = lower($2)`,
		productID, *color,
	).Scan(&c.ID, &c.ProductID, &c.Name, &c.NameAr, &c.HexCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.Validation(map[string]string{"selected_color": "this color is not available for this product"})
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// escapeLike escapes %, _ and \ so user search text is matched literally.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// whereBuilder assembles a WHERE clause with numbered $n placeholders.
type whereBuilder struct {
	conds []string
	args  []any
}

// add appends a condition; every "?" in cond becomes the same new $n.
func (w *whereBuilder) add(cond string, arg any) {
	w.args = append(w.args, arg)
	w.conds = append(w.conds, strings.ReplaceAll(cond, "?", placeholder(len(w.args))))
}

func (w *whereBuilder) addRaw(cond string) { w.conds = append(w.conds, cond) }

func (w *whereBuilder) sql() string {
	if len(w.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.conds, " AND ")
}

// next returns the placeholder for one more argument (for LIMIT/OFFSET).
func (w *whereBuilder) next(arg any) string {
	w.args = append(w.args, arg)
	return placeholder(len(w.args))
}

func placeholder(n int) string {
	return "$" + strconv.Itoa(n)
}
