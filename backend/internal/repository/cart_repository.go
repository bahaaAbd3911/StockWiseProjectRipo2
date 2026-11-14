package repository

import (
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/models"
)

type CartRepository struct {
	DB *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{
		DB: db,
	}
}

func (r *CartRepository) GetCartByUserID(userID int64) (models.Cart, error) {
	var cart models.Cart

	// Create a cart if the user doesn't already have one.
	_, err := r.DB.Exec(`
		INSERT INTO carts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING;
	`, userID)

	if err != nil {
		return cart, fmt.Errorf("failed to create cart: %w", err)
	}

	// Retrieve the cart.
	err = r.DB.QueryRow(`
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1;
	`, userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		return cart, fmt.Errorf("failed to retrieve cart: %w", err)
	}

	// Initialize an empty items list.
	cart.Items = []models.CartItem{}

	// Retrieve the cart items and their product information.
	rows, err := r.DB.Query(`
		SELECT
			ci.id,
			ci.cart_id,
			ci.product_id,
			p.name,
			p.name_ar,
			p.price,
			ci.quantity,
			ci.selected_color,
			ci.added_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.added_at DESC;
	`, cart.ID)

	if err != nil {
		return cart, fmt.Errorf("failed to retrieve cart items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.CartItem
		var selectedColor sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductNameAr,
			&item.Price,
			&item.Quantity,
			&selectedColor,
			&item.AddedAt,
		)

		if err != nil {
			return cart, fmt.Errorf("failed to scan cart item: %w", err)
		}

		if selectedColor.Valid {
			item.SelectedColor = &selectedColor.String
		}

		cart.Items = append(cart.Items, item)
	}

	if err := rows.Err(); err != nil {
		return cart, fmt.Errorf("failed to read cart items: %w", err)
	}

	return cart, nil
}

func (r *CartRepository) AddItem(
	userID int64,
	request models.AddCartItemRequest,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check that the product exists, is active, and has sufficient stock.
	var stock int

	err = tx.QueryRow(`
		SELECT stock
		FROM products
		WHERE id = $1 AND is_active = TRUE
		FOR UPDATE;
	`, request.ProductID).Scan(&stock)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product not found or inactive")
		}
		return fmt.Errorf("failed to check product: %w", err)
	}

	// Get or create the user's cart.
	_, err = tx.Exec(`
		INSERT INTO carts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING;
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to create cart: %w", err)
	}

	var cartID int64

	err = tx.QueryRow(`
		SELECT id FROM carts WHERE user_id = $1;
	`, userID).Scan(&cartID)

	if err != nil {
		return fmt.Errorf("failed to retrieve cart: %w", err)
	}

	// Check the quantity already in the cart.
	var existingQuantity int

	err = tx.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0)
		FROM cart_items
		WHERE cart_id = $1 AND product_id = $2;
	`, cartID, request.ProductID).Scan(&existingQuantity)

	if err != nil {
		return fmt.Errorf("failed to check cart quantity: %w", err)
	}

	if existingQuantity+request.Quantity > stock {
		return fmt.Errorf("requested quantity exceeds available stock")
	}

	// Add the item or increase its existing quantity.
	_, err = tx.Exec(`
		INSERT INTO cart_items (
			cart_id, product_id, quantity, selected_color
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET
			quantity = cart_items.quantity + EXCLUDED.quantity,
			selected_color = EXCLUDED.selected_color;
	`, cartID, request.ProductID, request.Quantity, request.SelectedColor)

	if err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	return tx.Commit()
}

func (r *CartRepository) UpdateItemQuantity(
	userID int64,
	itemID int64,
	quantity int,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Find the cart item and lock the product row.
	var productID int64
	var stock int

	err = tx.QueryRow(`
		SELECT ci.product_id, p.stock
		FROM cart_items ci
		JOIN carts c ON c.id = ci.cart_id
		JOIN products p ON p.id = ci.product_id
		WHERE ci.id = $1
		  AND c.user_id = $2
		  AND p.is_active = TRUE
		FOR UPDATE OF p;
	`, itemID, userID).Scan(&productID, &stock)

	if err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return fmt.Errorf("failed to find cart item: %w", err)
	}

	if quantity > stock {
		return fmt.Errorf("requested quantity exceeds available stock")
	}

	// Update the quantity.
	_, err = tx.Exec(`
		UPDATE cart_items
		SET quantity = $1
		WHERE id = $2;
	`, quantity, itemID)

	if err != nil {
		return fmt.Errorf("failed to update cart item: %w", err)
	}

	return tx.Commit()
}

func (r *CartRepository) RemoveItem(userID int64, itemID int64) error {
	result, err := r.DB.Exec(`
		DELETE FROM cart_items ci
		USING carts c
		WHERE ci.cart_id = c.id
		  AND ci.id = $1
		  AND c.user_id = $2;
	`, itemID, userID)

	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check removed cart item: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *CartRepository) ClearCart(userID int64) error {
	_, err := r.DB.Exec(`
		DELETE FROM cart_items ci
		USING carts c
		WHERE ci.cart_id = c.id
		  AND c.user_id = $1;
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}
