package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"

	"github.com/shopspring/decimal"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository { return &OrderRepository{db: db} }

const orderColumns = `
	o.id, o.user_id, o.address_id, o.recipient_name, o.phone, o.city, o.area, o.street,
	o.building_number, o.apartment_number, o.additional_instructions, o.total_amount,
	o.status, o.payment_method, o.payment_status, o.created_at, o.updated_at`

// adminOrderColumns adds the customer (requires JOIN users u).
const adminOrderColumns = orderColumns + `, u.first_name, u.last_name, u.email`

func scanOrder(s scanner, withCustomer bool) (models.Order, error) {
	var o models.Order
	dest := []any{&o.ID, &o.UserID, &o.AddressID, &o.RecipientName, &o.Phone, &o.City, &o.Area, &o.Street,
		&o.BuildingNumber, &o.ApartmentNumber, &o.AdditionalInstructions, &o.TotalAmount,
		&o.Status, &o.PaymentMethod, &o.PaymentStatus, &o.CreatedAt, &o.UpdatedAt}
	var c models.OrderCustomer
	if withCustomer {
		dest = append(dest, &c.FirstName, &c.LastName, &c.Email)
	}
	if err := s.Scan(dest...); err != nil {
		return o, err
	}
	if withCustomer {
		o.Customer = &c
	}
	o.Items = []models.OrderItem{}
	return o, nil
}

// Checkout turns the user's cart into an order. Everything happens in ONE
// transaction, so either all of it happens or none of it does:
//
//  1. load the delivery address (must belong to the user; default if none given)
//  2. lock the cart row       -> two checkouts by the same user run one at a time
//  3. load the cart lines and lock their product rows (in product-id order to
//     avoid deadlocks) -> two customers cannot both buy the last unit
//  4. check every line with the same rule the cart view uses (active, stock,
//     color); report ALL problems at once
//  5. compute prices with the discount (exact decimals)
//  6. insert the order with a COPY of the address
//  7. insert order items with a COPY of name/price/color, and reduce stock
//  8. empty the cart
func (r *OrderRepository) Checkout(ctx context.Context, userID int64, addressID *int64) (int64, error) {
	var orderID int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		addr, err := checkoutAddress(ctx, tx, userID, addressID)
		if err != nil {
			return err
		}

		var cartID int64
		err = tx.QueryRowContext(ctx, `SELECT id FROM carts WHERE user_id = $1 FOR UPDATE`, userID).Scan(&cartID)
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.Conflict("your cart is empty")
		}
		if err != nil {
			return err
		}

		lines, err := loadCartLines(ctx, tx, cartID, "ci.product_id", "FOR UPDATE OF p")
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return apperr.Conflict("your cart is empty")
		}
		var problems []string
		for _, l := range lines {
			if p := l.problem(); p != "" {
				problems = append(problems, fmt.Sprintf("%s: %s", l.Name, p))
			}
		}
		if len(problems) > 0 {
			return apperr.Conflict("some items in your cart cannot be ordered - " + strings.Join(problems, "; "))
		}

		total := decimal.Zero
		for _, l := range lines {
			total = total.Add(pricing.LineTotal(l.unitPrice(), l.Quantity))
		}

		if err := tx.QueryRowContext(ctx, `
			INSERT INTO orders (user_id, address_id, recipient_name, phone, city, area, street,
			                    building_number, apartment_number, additional_instructions,
			                    total_amount, status, payment_method, payment_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING id`,
			userID, addr.ID, addr.RecipientName, addr.Phone, addr.City, addr.Area, addr.Street,
			addr.BuildingNumber, addr.ApartmentNumber, addr.AdditionalInstructions,
			total, models.OrderPending, models.PaymentCashOnDelivery, models.PaymentPending,
		).Scan(&orderID); err != nil {
			return err
		}

		for _, l := range lines {
			unit := l.unitPrice()
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO order_items (order_id, product_id, product_name, product_name_ar, quantity,
				                         unit_price, selected_color, selected_color_ar, subtotal)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
				orderID, l.ProductID, l.Name, l.NameAr, l.Quantity,
				unit, l.Color, l.ColorAr, pricing.LineTotal(unit, l.Quantity),
			); err != nil {
				return err
			}
			// The row is locked and the stock was checked, so this always succeeds;
			// "stock >= $1" is a second line of defence.
			res, err := tx.ExecContext(ctx, `
				UPDATE products SET stock = stock - $1, updated_at = NOW()
				WHERE id = $2 AND stock >= $1`, l.Quantity, l.ProductID)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n != 1 {
				return apperr.Conflict(fmt.Sprintf("%s: not enough stock", l.Name))
			}
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
			return err
		}
		return touchCart(ctx, tx, cartID)
	})
	return orderID, err
}

func checkoutAddress(ctx context.Context, tx *sql.Tx, userID int64, addressID *int64) (models.Address, error) {
	var row *sql.Row
	if addressID != nil {
		row = tx.QueryRowContext(ctx,
			`SELECT `+addressColumns+` FROM user_addresses WHERE user_id = $1 AND id = $2`, userID, *addressID)
	} else {
		row = tx.QueryRowContext(ctx,
			`SELECT `+addressColumns+` FROM user_addresses WHERE user_id = $1 AND is_default`, userID)
	}
	a, err := scanAddress(row)
	if errors.Is(err, sql.ErrNoRows) {
		if addressID == nil {
			return a, apperr.Invalid("no address_id given and you have no default address; add an address first")
		}
		return a, apperr.NotFound("address not found")
	}
	return a, err
}

// loadItems fills Items for every order with ONE query (not one per order).
func loadItems(ctx context.Context, q database.Querier, orders []models.Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]int64, len(orders))
	index := make(map[int64]int, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
		index[o.ID] = i
	}
	rows, err := q.QueryContext(ctx, `
		SELECT id, order_id, product_id, product_name, product_name_ar, quantity,
		       unit_price, selected_color, selected_color_ar, subtotal
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY order_id, id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.ProductName, &it.ProductNameAr,
			&it.Quantity, &it.UnitPrice, &it.SelectedColor, &it.SelectedColorAr, &it.Subtotal); err != nil {
			return err
		}
		i := index[it.OrderID]
		orders[i].Items = append(orders[i].Items, it)
	}
	return rows.Err()
}

func (r *OrderRepository) queryOrders(ctx context.Context, withCustomer bool, query string, args ...any) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	orders := []models.Order{}
	for rows.Next() {
		o, err := scanOrder(rows, withCustomer)
		if err != nil {
			rows.Close()
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, loadItems(ctx, r.db, orders)
}

// ListForUser returns the user's orders, newest first.
func (r *OrderRepository) ListForUser(ctx context.Context, userID int64) ([]models.Order, error) {
	return r.queryOrders(ctx, false,
		`SELECT `+orderColumns+` FROM orders o WHERE o.user_id = $1 ORDER BY o.created_at DESC, o.id DESC`, userID)
}

// GetForUser returns one order only if it belongs to the user.
func (r *OrderRepository) GetForUser(ctx context.Context, userID, orderID int64) (models.Order, error) {
	orders, err := r.queryOrders(ctx, false,
		`SELECT `+orderColumns+` FROM orders o WHERE o.id = $1 AND o.user_id = $2`, orderID, userID)
	if err != nil {
		return models.Order{}, err
	}
	if len(orders) == 0 {
		return models.Order{}, apperr.NotFound("order not found")
	}
	return orders[0], nil
}

// Get returns any order with its customer (admin).
func (r *OrderRepository) Get(ctx context.Context, orderID int64) (models.Order, error) {
	orders, err := r.queryOrders(ctx, true,
		`SELECT `+adminOrderColumns+` FROM orders o JOIN users u ON u.id = o.user_id WHERE o.id = $1`, orderID)
	if err != nil {
		return models.Order{}, err
	}
	if len(orders) == 0 {
		return models.Order{}, apperr.NotFound("order not found")
	}
	return orders[0], nil
}

// List returns a page of all orders, optionally filtered by status (admin).
func (r *OrderRepository) List(ctx context.Context, f models.OrderFilter) (models.Page[models.Order], error) {
	var w whereBuilder
	if f.Status != "" {
		w.add("o.status = ?::text", f.Status)
	}
	page := models.Page[models.Order]{Items: []models.Order{}, Limit: f.Limit, Offset: f.Offset}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders o`+w.sql(), w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	query := `SELECT ` + adminOrderColumns + ` FROM orders o JOIN users u ON u.id = o.user_id` + w.sql() +
		` ORDER BY o.created_at DESC, o.id DESC LIMIT ` + w.next(f.Limit) + ` OFFSET ` + w.next(f.Offset)
	orders, err := r.queryOrders(ctx, true, query, w.args...)
	if err != nil {
		return page, err
	}
	page.Items = orders
	return page, nil
}

// CancelByUser lets a customer cancel their own order while it is still pending.
func (r *OrderRepository) CancelByUser(ctx context.Context, userID, orderID int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx,
			`SELECT status FROM orders WHERE id = $1 AND user_id = $2 FOR UPDATE`, orderID, userID,
		).Scan(&status); err != nil {
			return apperr.FromDB(err, "order not found")
		}
		if status != models.OrderPending {
			return apperr.Conflict(fmt.Sprintf("only pending orders can be cancelled (this order is %s)", status))
		}
		return setStatus(ctx, tx, orderID, models.OrderCancelled, "")
	})
}

// UpdateStatus moves an order along its life cycle (admin). Side effects:
//   - cancelled -> stock is put back
//   - delivered -> payment_status becomes "paid" (cash was collected) and the
//     customer's existing reviews of these products become verified purchases
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, to string) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var from string
		if err := tx.QueryRowContext(ctx,
			`SELECT status FROM orders WHERE id = $1 FOR UPDATE`, orderID,
		).Scan(&from); err != nil {
			return apperr.FromDB(err, "order not found")
		}
		if !models.CanTransitionOrder(from, to) {
			return apperr.Conflict(fmt.Sprintf("cannot change order status from %s to %s", from, to))
		}
		payment := ""
		if to == models.OrderDelivered {
			payment = models.PaymentPaid
			if _, err := tx.ExecContext(ctx, `
				UPDATE product_reviews r SET verified_purchase = TRUE
				FROM order_items oi JOIN orders o ON o.id = oi.order_id
				WHERE o.id = $1 AND r.user_id = o.user_id AND r.product_id = oi.product_id
				  AND NOT r.verified_purchase`, orderID); err != nil {
				return err
			}
		}
		return setStatus(ctx, tx, orderID, to, payment)
	})
}

// setStatus writes the new status (restoring stock for cancellations).
// payment == "" keeps the current payment status.
func setStatus(ctx context.Context, tx *sql.Tx, orderID int64, status, payment string) error {
	if status == models.OrderCancelled {
		// Products that were hard-deleted have product_id NULL and are skipped.
		if _, err := tx.ExecContext(ctx, `
			UPDATE products p SET stock = p.stock + oi.quantity, updated_at = NOW()
			FROM order_items oi
			WHERE oi.order_id = $1 AND oi.product_id = p.id`, orderID); err != nil {
			return err
		}
	}
	var err error
	if payment == "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, orderID)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE orders SET status = $1, payment_status = $2, updated_at = NOW() WHERE id = $3`,
			status, payment, orderID)
	}
	return err
}
