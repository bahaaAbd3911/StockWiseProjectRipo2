package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"
	"ecommerce-backend/internal/validators"

	"github.com/shopspring/decimal"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository { return &CartRepository{db: db} }

// cartLine is one cart row joined with its live product and color data. It is
// shared by the cart view and by checkout so both apply the same rules.
type cartLine struct {
	ItemID          int64
	ProductID       int64
	Name            string
	NameAr          string
	PrimaryImageURL *string
	Price           decimal.Decimal
	Discount        decimal.Decimal
	Stock           int
	Active          bool
	Quantity        int
	Color           *string // what the user picked
	ColorAr         *string // NULL if that color no longer exists
	ColorHex        *string
	ColorCount      int
	AddedAt         time.Time
}

// cartLinesSQL selects the lines of one cart ($1). Checkout passes a lock clause.
func cartLinesSQL(orderBy, lock string) string {
	return `
		SELECT ci.id, ci.product_id, p.name, p.name_ar,
		       (SELECT pi.image_url FROM product_images pi WHERE pi.product_id = p.id
		         ORDER BY pi.is_primary DESC, pi.display_order, pi.id LIMIT 1),
		       p.price, p.discount_value, p.stock, p.is_active,
		       ci.quantity, ci.selected_color, pc.name_ar, pc.hex_code,
		       (SELECT COUNT(*) FROM product_colors x WHERE x.product_id = p.id),
		       ci.added_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		LEFT JOIN product_colors pc ON pc.product_id = p.id AND pc.name = ci.selected_color
		WHERE ci.cart_id = $1
		ORDER BY ` + orderBy + ` ` + lock
}

func loadCartLines(ctx context.Context, q database.Querier, cartID int64, orderBy, lock string) ([]cartLine, error) {
	rows, err := q.QueryContext(ctx, cartLinesSQL(orderBy, lock), cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []cartLine
	for rows.Next() {
		var l cartLine
		if err := rows.Scan(&l.ItemID, &l.ProductID, &l.Name, &l.NameAr, &l.PrimaryImageURL,
			&l.Price, &l.Discount, &l.Stock, &l.Active,
			&l.Quantity, &l.Color, &l.ColorAr, &l.ColorHex, &l.ColorCount, &l.AddedAt); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

// problem returns "" when the line can be bought, otherwise the reason.
// This single rule is used both to display the cart and to accept checkout.
func (l cartLine) problem() string {
	switch {
	case !l.Active:
		return "this product is no longer available"
	case l.Stock == 0:
		return "this product is out of stock"
	case l.Quantity > l.Stock:
		return fmt.Sprintf("only %d left in stock", l.Stock)
	case l.Color != nil && l.ColorAr == nil:
		return "the selected color is no longer available"
	case l.Color == nil && l.ColorCount > 0:
		return "please choose a color"
	}
	return ""
}

func (l cartLine) unitPrice() decimal.Decimal { return pricing.DiscountedPrice(l.Price, l.Discount) }

// ensureCart returns the user's cart, creating it on first use.
func ensureCart(ctx context.Context, q database.Querier, userID int64) (int64, time.Time, error) {
	var id int64
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT id, updated_at FROM carts WHERE user_id = $1`, userID).Scan(&id, &updated)
	if !errors.Is(err, sql.ErrNoRows) {
		return id, updated, err
	}
	// ON CONFLICT handles two first requests racing each other.
	err = q.QueryRowContext(ctx, `
		INSERT INTO carts (user_id) VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = carts.updated_at
		RETURNING id, updated_at`, userID).Scan(&id, &updated)
	return id, updated, err
}

func touchCart(ctx context.Context, q database.Querier, cartID int64) error {
	_, err := q.ExecContext(ctx, `UPDATE carts SET updated_at = NOW() WHERE id = $1`, cartID)
	return err
}

// Get returns the cart with live prices, line totals and availability.
func (r *CartRepository) Get(ctx context.Context, userID int64) (models.Cart, error) {
	cartID, updated, err := ensureCart(ctx, r.db, userID)
	if err != nil {
		return models.Cart{}, err
	}
	lines, err := loadCartLines(ctx, r.db, cartID, "ci.added_at DESC, ci.id DESC", "")
	if err != nil {
		return models.Cart{}, err
	}

	cart := models.Cart{ID: cartID, Items: []models.CartItem{}, Subtotal: decimal.Zero, UpdatedAt: updated}
	for _, l := range lines {
		unit := l.unitPrice()
		item := models.CartItem{
			ID: l.ItemID, ProductID: l.ProductID, ProductName: l.Name, ProductNameAr: l.NameAr,
			PrimaryImageURL: l.PrimaryImageURL,
			OriginalPrice:   l.Price, DiscountValue: l.Discount, UnitPrice: unit,
			Quantity: l.Quantity, LineTotal: pricing.LineTotal(unit, l.Quantity),
			SelectedColor: l.Color, SelectedColorAr: l.ColorAr, SelectedColorHex: l.ColorHex,
			Stock: l.Stock, AddedAt: l.AddedAt,
		}
		cart.ItemCount += l.Quantity
		if p := l.problem(); p != "" {
			item.UnavailableReason = &p
			cart.HasUnavailableItems = true
		} else {
			item.IsAvailable = true
			cart.Subtotal = cart.Subtotal.Add(item.LineTotal)
		}
		cart.Items = append(cart.Items, item)
	}
	return cart, nil
}

// AddItem adds a product, or increases the quantity if it is already in the
// cart. The schema allows one row per product per cart, so the same product in
// a different color is refused with a clear message.
func (r *CartRepository) AddItem(ctx context.Context, userID int64, req models.AddCartItemRequest) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var stock int
		err := tx.QueryRowContext(ctx,
			`SELECT stock FROM products WHERE id = $1 AND is_active FOR SHARE`, req.ProductID).Scan(&stock)
		if err != nil {
			return apperr.FromDB(err, "product not found")
		}

		color, err := resolveColor(ctx, tx, req.ProductID, req.SelectedColor)
		if err != nil {
			return err
		}
		var colorName *string
		if color != nil {
			colorName = &color.Name
		}

		cartID, _, err := ensureCart(ctx, tx, userID)
		if err != nil {
			return err
		}

		var itemID int64
		var existingQty int
		var existingColor *string
		err = tx.QueryRowContext(ctx, `
			SELECT id, quantity, selected_color FROM cart_items
			WHERE cart_id = $1 AND product_id = $2
			FOR UPDATE`, cartID, req.ProductID).Scan(&itemID, &existingQty, &existingColor)
		exists := err == nil
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if exists && !sameString(existingColor, colorName) {
			return apperr.Conflict(fmt.Sprintf(
				"this product is already in your cart in %s; change the color on that cart item instead",
				describeColor(existingColor)))
		}

		newQty := existingQty + req.Quantity
		switch {
		case stock == 0:
			return apperr.Conflict("this product is out of stock; you can join the waiting list instead")
		case newQty > stock:
			if exists {
				return apperr.Conflict(fmt.Sprintf("only %d left in stock (you already have %d in your cart)", stock, existingQty))
			}
			return apperr.Conflict(fmt.Sprintf("only %d left in stock", stock))
		case newQty > validators.MaxQuantity:
			return apperr.Invalid("quantity per product cannot exceed 1000")
		}

		if exists {
			_, err = tx.ExecContext(ctx, `UPDATE cart_items SET quantity = $1 WHERE id = $2`, newQty, itemID)
		} else {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO cart_items (cart_id, product_id, quantity, selected_color)
				VALUES ($1, $2, $3, $4)`, cartID, req.ProductID, newQty, colorName)
		}
		if err != nil {
			return apperr.FromDB(err, "cart item not found")
		}
		return touchCart(ctx, tx, cartID)
	})
}

// UpdateItem sets a new quantity and, optionally, a new color.
func (r *CartRepository) UpdateItem(ctx context.Context, userID, itemID int64, req models.UpdateCartItemRequest) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var cartID, productID int64
		var color *string
		var stock int
		var active bool
		err := tx.QueryRowContext(ctx, `
			SELECT ci.cart_id, ci.product_id, ci.selected_color, p.stock, p.is_active
			FROM cart_items ci
			JOIN carts c ON c.id = ci.cart_id
			JOIN products p ON p.id = ci.product_id
			WHERE ci.id = $1 AND c.user_id = $2
			FOR UPDATE OF ci`, itemID, userID).Scan(&cartID, &productID, &color, &stock, &active)
		if err != nil {
			return apperr.FromDB(err, "cart item not found")
		}
		if !active {
			return apperr.Conflict("this product is no longer available; please remove it from your cart")
		}
		if req.SelectedColor != nil {
			c, err := resolveColor(ctx, tx, productID, req.SelectedColor)
			if err != nil {
				return err
			}
			color = &c.Name
		}
		if req.Quantity > stock {
			return apperr.Conflict(fmt.Sprintf("only %d left in stock", stock))
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE cart_items SET quantity = $1, selected_color = $2 WHERE id = $3`,
			req.Quantity, color, itemID); err != nil {
			return err
		}
		return touchCart(ctx, tx, cartID)
	})
}

func (r *CartRepository) RemoveItem(ctx context.Context, userID, itemID int64) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items ci USING carts c
		WHERE ci.cart_id = c.id AND ci.id = $1 AND c.user_id = $2`, itemID, userID)
	return expectOneRow(res, err, "cart item not found")
}

func (r *CartRepository) Clear(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items ci USING carts c
		WHERE ci.cart_id = c.id AND c.user_id = $1`, userID)
	return err
}

func sameString(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func describeColor(c *string) string {
	if c == nil {
		return "no color"
	}
	return fmt.Sprintf("color %q", *c)
}
