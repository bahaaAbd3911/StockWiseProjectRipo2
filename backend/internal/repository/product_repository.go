package repository

import (
	"context"
	"database/sql"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository { return &ProductRepository{db: db} }

// productSorts maps the ?sort= values to ORDER BY clauses. Only these fixed
// strings are ever put into the SQL, so sorting cannot be used for injection.
var productSorts = map[string]string{
	"newest":     "p.created_at DESC, p.id DESC",
	"price_asc":  "p.price * (100 - p.discount_value) ASC, p.id",
	"price_desc": "p.price * (100 - p.discount_value) DESC, p.id",
	"name":       "p.name ASC, p.id",
}

// IsValidSort reports whether s is an accepted ?sort= value.
func IsValidSort(s string) bool {
	_, ok := productSorts[s]
	return ok
}

// List returns one page of products and the total number of matches.
func (r *ProductRepository) List(ctx context.Context, f models.ProductFilter) (models.Page[models.Product], error) {
	var w whereBuilder
	if !f.IncludeInactive {
		w.addRaw("p.is_active")
	}
	if f.InStockOnly {
		w.addRaw("p.stock > 0")
	}
	if f.Category != "" {
		// The same $n is used twice; ::text pins its type for PostgreSQL.
		w.add("(lower(p.category) = lower(?::text) OR p.category_ar = ?::text)", f.Category)
	}
	if f.Search != "" {
		w.add("(p.name ILIKE ?::text OR p.name_ar ILIKE ?::text OR p.description ILIKE ?::text OR p.description_ar ILIKE ?::text)",
			"%"+escapeLike(f.Search)+"%")
	}

	page := models.Page[models.Product]{Items: []models.Product{}, Limit: f.Limit, Offset: f.Offset}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products p`+w.sql(), w.args...).Scan(&page.Total); err != nil {
		return page, err
	}

	order, ok := productSorts[f.Sort]
	if !ok {
		order = productSorts["newest"]
	}
	query := `SELECT ` + productColumns + ` FROM products p` + w.sql() +
		` ORDER BY ` + order + ` LIMIT ` + w.next(f.Limit) + ` OFFSET ` + w.next(f.Offset)

	rows, err := r.db.QueryContext(ctx, query, w.args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, p)
	}
	return page, rows.Err()
}

// Get returns one product. Inactive (soft-deleted) products are only visible
// when includeInactive is true (admin).
func (r *ProductRepository) Get(ctx context.Context, id int64, includeInactive bool) (models.Product, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+productColumns+` FROM products p WHERE p.id = $1 AND (p.is_active OR $2)`,
		id, includeInactive)
	p, err := scanProduct(row)
	return p, apperr.FromDB(err, "product not found")
}

// Create inserts a product; is_active defaults to true when not given.
// Empty optional texts are stored as NULL.
func (r *ProductRepository) Create(ctx context.Context, in models.ProductInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO products (name, name_ar, description, description_ar, category, category_ar,
		                      price, discount_value, stock, is_active)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''),
		        $7, $8, $9, COALESCE($10, TRUE))
		RETURNING id`,
		in.Name, in.NameAr, in.Description, in.DescriptionAr, in.Category, in.CategoryAr,
		in.Price, in.DiscountValue, in.Stock, in.IsActive,
	).Scan(&id)
	return id, apperr.FromDB(err, "product not found")
}

// Update replaces all editable fields; is_active is kept when not given.
func (r *ProductRepository) Update(ctx context.Context, id int64, in models.ProductInput) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE products SET
			name = $1, name_ar = $2,
			description = NULLIF($3, ''), description_ar = NULLIF($4, ''),
			category = NULLIF($5, ''), category_ar = NULLIF($6, ''),
			price = $7, discount_value = $8, stock = $9,
			is_active = COALESCE($10, is_active),
			updated_at = NOW()
		WHERE id = $11`,
		in.Name, in.NameAr, in.Description, in.DescriptionAr, in.Category, in.CategoryAr,
		in.Price, in.DiscountValue, in.Stock, in.IsActive, id,
	)
	return expectOneRow(res, err, "product not found")
}

// Deactivate soft-deletes: the row stays so order history and reviews keep
// working, but it disappears from the public catalog and cannot be bought.
func (r *ProductRepository) Deactivate(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND is_active`, id)
	return expectOneRow(res, err, "product not found or already inactive")
}

// Categories lists the distinct categories of active products.
func (r *ProductRepository) Categories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT category, COALESCE(MAX(category_ar), ''), COUNT(*)
		FROM products
		WHERE is_active AND category IS NOT NULL AND category <> ''
		GROUP BY category
		ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.Name, &c.NameAr, &c.ProductCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
