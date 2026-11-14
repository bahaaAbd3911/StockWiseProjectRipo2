package repository

import (
	"database/sql"
	"ecommerce-backend/internal/models"
	"errors"
	"fmt"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		DB: db,
	}
}

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	query := `
        SELECT
            id,
            name,
            name_ar,
            description,
            description_ar,
            category,
            category_ar,
            price,
            discount_value,
            stock,
            created_at,
            updated_at,
            is_active
        FROM products
        WHERE is_active = TRUE
        ORDER BY id;
    `

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		var product models.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.NameAr,
			&product.Description,
			&product.DescriptionAr,
			&product.Category,
			&product.CategoryAr,
			&product.Price,
			&product.DiscountValue,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}
func (r *ProductRepository) GetByID(id int64) (models.Product, error) {
	query := `
        SELECT
            id,
            name,
            name_ar,
            description,
            description_ar,
            category,
            category_ar,
            price,
            discount_value,
            stock,
            created_at,
            updated_at,
            is_active
        FROM products
        WHERE id = $1 AND is_active = TRUE;
    `

	var product models.Product

	err := r.DB.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.NameAr,
		&product.Description,
		&product.DescriptionAr,
		&product.Category,
		&product.CategoryAr,
		&product.Price,
		&product.DiscountValue,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.IsActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Product{}, sql.ErrNoRows
	}

	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}
func (r *ProductRepository) Create(product models.CreateProductRequest) (models.Product, error) {
	query := `
        INSERT INTO products (
            name,
            name_ar,
            description,
            description_ar,
            category,
            category_ar,
            price,
            discount_value,
            stock,
            is_active
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING
            id,
            name,
            name_ar,
            description,
            description_ar,
            category,
            category_ar,
            price,
            discount_value,
            stock,
            created_at,
            updated_at,
            is_active;
    `

	var createdProduct models.Product

	err := r.DB.QueryRow(
		query,
		product.Name,
		product.NameAr,
		product.Description,
		product.DescriptionAr,
		product.Category,
		product.CategoryAr,
		product.Price,
		product.DiscountValue,
		product.Stock,
		product.IsActive,
	).Scan(
		&createdProduct.ID,
		&createdProduct.Name,
		&createdProduct.NameAr,
		&createdProduct.Description,
		&createdProduct.DescriptionAr,
		&createdProduct.Category,
		&createdProduct.CategoryAr,
		&createdProduct.Price,
		&createdProduct.DiscountValue,
		&createdProduct.Stock,
		&createdProduct.CreatedAt,
		&createdProduct.UpdatedAt,
		&createdProduct.IsActive,
	)

	if err != nil {
		return models.Product{}, fmt.Errorf("failed to create product: %w", err)
	}

	return createdProduct, nil
}
func (r *ProductRepository) Update(
	id int64,
	product models.UpdateProductRequest,
) (models.Product, error) {

	query := `
        UPDATE products
        SET
            name = $1,
            name_ar = $2,
            description = $3,
            description_ar = $4,
            category = $5,
            category_ar = $6,
            price = $7,
            discount_value = $8,
            stock = $9,
            is_active = $10,
            updated_at = NOW()
        WHERE id = $11
        RETURNING
            id,
            name,
            name_ar,
            description,
            description_ar,
            category,
            category_ar,
            price,
            discount_value,
            stock,
            created_at,
            updated_at,
            is_active;
    `

	var updatedProduct models.Product

	err := r.DB.QueryRow(
		query,
		product.Name,
		product.NameAr,
		product.Description,
		product.DescriptionAr,
		product.Category,
		product.CategoryAr,
		product.Price,
		product.DiscountValue,
		product.Stock,
		product.IsActive,
		id,
	).Scan(
		&updatedProduct.ID,
		&updatedProduct.Name,
		&updatedProduct.NameAr,
		&updatedProduct.Description,
		&updatedProduct.DescriptionAr,
		&updatedProduct.Category,
		&updatedProduct.CategoryAr,
		&updatedProduct.Price,
		&updatedProduct.DiscountValue,
		&updatedProduct.Stock,
		&updatedProduct.CreatedAt,
		&updatedProduct.UpdatedAt,
		&updatedProduct.IsActive,
	)

	if err != nil {
		return models.Product{}, fmt.Errorf("failed to update product: %w", err)
	}

	return updatedProduct, nil
}
func (r *ProductRepository) Delete(id int64) error {
	query := `
        UPDATE products
        SET
            is_active = FALSE,
            updated_at = NOW()
        WHERE id = $1
        AND is_active = TRUE;
    `

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deleted product: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
