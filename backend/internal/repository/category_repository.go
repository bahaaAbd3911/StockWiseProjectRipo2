package repository

import (
	"context"
	"database/sql"
	"errors"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

// CategoryRepository manages the categories table. Products still store their
// category as text; this table holds what admins manage per category (Arabic
// name, icon, on/off) and keeps the product texts in step when renaming.
type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository { return &CategoryRepository{db: db} }

const categoryColumns = `c.id, c.name, c.name_ar, c.icon, c.is_active,
	(SELECT COUNT(*) FROM products p WHERE p.is_active AND lower(p.category) = lower(c.name)),
	c.created_at, c.updated_at`

func scanCategory(s scanner) (models.Category, error) {
	var c models.Category
	err := s.Scan(&c.ID, &c.Name, &c.NameAr, &c.Icon, &c.IsActive, &c.ProductCount, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

var errCategoryExists = apperr.Conflict("a category with this name already exists")

// List returns categories by name. Public lists only show active ones.
func (r *CategoryRepository) List(ctx context.Context, includeInactive bool) ([]models.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+categoryColumns+` FROM categories c
		WHERE c.is_active OR $1 ORDER BY lower(c.name)`, includeInactive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepository) Get(ctx context.Context, id int64) (models.Category, error) {
	c, err := scanCategory(r.db.QueryRowContext(ctx, `SELECT `+categoryColumns+` FROM categories c WHERE c.id = $1`, id))
	return c, apperr.FromDB(err, "category not found")
}

func (r *CategoryRepository) Create(ctx context.Context, in models.CategoryInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO categories (name, name_ar, icon, is_active) VALUES ($1, $2, $3, COALESCE($4, TRUE))
		RETURNING id`, in.Name, in.NameAr, in.Icon, in.IsActive).Scan(&id)
	if apperr.IsUniqueViolation(err, "categories_name_key") {
		return 0, errCategoryExists
	}
	return id, err
}

// Update edits a category. A new name or Arabic name is copied to all of its
// products in the same transaction, so filters and product pages stay right.
func (r *CategoryRepository) Update(ctx context.Context, id int64, in models.CategoryInput) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var oldName string
		if err := tx.QueryRowContext(ctx, `SELECT name FROM categories WHERE id = $1 FOR UPDATE`, id).Scan(&oldName); err != nil {
			return apperr.FromDB(err, "category not found")
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE categories SET name = $1, name_ar = $2, icon = $3, is_active = COALESCE($4, is_active), updated_at = NOW()
			WHERE id = $5`, in.Name, in.NameAr, in.Icon, in.IsActive, id)
		if apperr.IsUniqueViolation(err, "categories_name_key") {
			return errCategoryExists
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE products SET category = $1, category_ar = NULLIF($2, ''), updated_at = NOW()
			WHERE lower(category) = lower($3)`, in.Name, in.NameAr, oldName)
		return err
	})
}

// Delete removes a category and hides (deactivates) its products: they stay
// in the database for order history and can be moved to another category and
// switched on again from the admin product list.
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var name string
		if err := tx.QueryRowContext(ctx, `DELETE FROM categories WHERE id = $1 RETURNING name`, id).Scan(&name); err != nil {
			return apperr.FromDB(err, "category not found")
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE products SET is_active = FALSE, updated_at = NOW()
			WHERE lower(category) = lower($1) AND is_active`, name)
		return err
	})
}

// ensureCategory makes the product use an existing category's exact spelling
// (and its Arabic name when none was sent), or creates the category.
func ensureCategory(ctx context.Context, tx *sql.Tx, in *models.ProductInput) error {
	if in.Category == "" {
		return nil
	}
	var name, nameAr string
	err := tx.QueryRowContext(ctx,
		`SELECT name, name_ar FROM categories WHERE lower(name) = lower($1)`, in.Category).Scan(&name, &nameAr)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO categories (name, name_ar) VALUES ($1, $2)
			ON CONFLICT (store_id, lower(name)) DO NOTHING`, in.Category, in.CategoryAr)
		return err
	}
	if err != nil {
		return err
	}
	in.Category = name
	if in.CategoryAr == "" {
		in.CategoryAr = nameAr
	}
	return nil
}
