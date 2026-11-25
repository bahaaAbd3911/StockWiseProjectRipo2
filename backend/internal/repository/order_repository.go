package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
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
	o.building_number, o.apartment_number, o.additional_instructions,
	o.subtotal_amount, o.discount_amount, o.coupon_code, o.total_amount,
	o.status, o.payment_method, o.payment_status, o.created_at, o.updated_at`

// adminOrderColumns adds the customer (requires JOIN users u).
const adminOrderColumns = orderColumns + `, u.first_name, u.last_name, u.email`

func scanOrder(s scanner, withCustomer bool) (models.Order, error) {
	var o models.Order
	dest := []any{&o.ID, &o.UserID, &o.AddressID, &o.RecipientName, &o.Phone, &o.City, &o.Area, &o.Street,
		&o.BuildingNumber, &o.ApartmentNumber, &o.AdditionalInstructions,
		&o.SubtotalAmount, &o.DiscountAmount, &o.CouponCode, &o.TotalAmount,
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
	o.Bundles = []models.OrderBundle{}
	return o, nil
}

// Checkout turns the user's cart (products AND bundles) into an order.
// Everything happens in ONE transaction, so either all of it happens or none:
//
//  1. load the delivery address (must belong to the user; default if none given)
//  2. pay later: lock the user row, so two pay-later checkouts cannot both
//     use the same available credit
//  3. lock the cart row         -> one checkout per user at a time
//  4. lock EVERY product the cart needs - single items and the products inside
//     bundles - in one statement, in id order (no deadlocks)
//  5. check every line with the same rules the cart view uses, plus the
//     combined demand per product (a product can be in the cart alone AND
//     inside a bundle); report ALL problems at once
//  6. compute prices: best of product discount / live offers; bundle price
//  7. pay later: balance + total must fit within the credit limit
//  8. insert the order (address copy), items (name/price/offer copy), bundles
//     (with their products), reduce stock, charge the debt account if pay later
//  9. empty the cart
func (r *OrderRepository) Checkout(ctx context.Context, userID int64, addressID *int64, method string) (int64, error) {
	var orderID int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		addr, err := checkoutAddress(ctx, tx, userID, addressID)
		if err != nil {
			return err
		}
		if method == models.PaymentPayLater {
			if err := lockUser(ctx, tx, userID); err != nil {
				return err
			}
		}

		var cartID int64
		err = tx.QueryRowContext(ctx, `SELECT id FROM carts WHERE user_id = $1 FOR UPDATE`, userID).Scan(&cartID)
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.Conflict("your cart is empty")
		}
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			SELECT id FROM products WHERE id IN (
				SELECT product_id FROM cart_items WHERE cart_id = $1
				UNION
				SELECT bi.product_id FROM cart_bundles cb JOIN bundle_items bi ON bi.bundle_id = cb.bundle_id
				WHERE cb.cart_id = $1)
			ORDER BY id FOR UPDATE`, cartID); err != nil {
			return err
		}

		lines, err := loadCartLines(ctx, tx, cartID, "ci.product_id", "")
		if err != nil {
			return err
		}
		bundles, err := loadCartBundleLines(ctx, tx, cartID)
		if err != nil {
			return err
		}
		if len(lines) == 0 && len(bundles) == 0 {
			return apperr.Conflict("your cart is empty")
		}

		// Per-line problems, then the combined demand for each product.
		var problems []string
		need := map[int64]int{}
		stock := map[int64]int{}
		names := map[int64]string{}
		for _, l := range lines {
			if p := l.problem(); p != "" {
				problems = append(problems, fmt.Sprintf("%s: %s", l.Name, p))
			}
			need[l.ProductID] += l.Quantity
			stock[l.ProductID], names[l.ProductID] = l.Stock, l.Name
		}
		for _, b := range bundles {
			if p := b.problem(); p != "" {
				problems = append(problems, fmt.Sprintf("%s: %s", b.bundle.Name, p))
			}
			for _, c := range b.comps {
				need[c.ProductID] += c.Quantity * b.Quantity
				stock[c.ProductID], names[c.ProductID] = c.Stock, c.ProductName
			}
		}
		if len(problems) == 0 {
			for pid, n := range need {
				if n > stock[pid] {
					problems = append(problems, fmt.Sprintf("%s: only %d left in stock, but your cart needs %d",
						names[pid], stock[pid], n))
				}
			}
		}
		if len(problems) > 0 {
			sort.Strings(problems)
			return apperr.Conflict("some items in your cart cannot be ordered - " + strings.Join(problems, "; "))
		}

		subtotal := decimal.Zero
		for _, l := range lines {
			unit, _ := l.unitPrice()
			subtotal = subtotal.Add(pricing.LineTotal(unit, l.Quantity))
		}
		for _, b := range bundles {
			subtotal = subtotal.Add(pricing.LineTotal(b.bundle.Price, b.Quantity))
		}

		// Coupon code on the cart: checked again with the coupon row locked,
		// so two checkouts cannot both take the last use of a limited code.
		discount := decimal.Zero
		var coupon *couponRow
		couponID, err := cartCouponID(ctx, tx, cartID)
		if err != nil {
			return err
		}
		if couponID != nil {
			c, err := loadCoupon(ctx, tx, `c.id = $1`, *couponID, userID, true)
			if err != nil {
				return err
			}
			d, problem := c.evaluate(lines, bundles)
			if problem != "" {
				return apperr.Conflict(fmt.Sprintf("the code %s cannot be used: %s. Remove it from your cart to continue", c.Code, problem))
			}
			discount, coupon = d, &c
		}
		total := subtotal.Sub(discount)

		paymentStatus := models.PaymentPending
		if method == models.PaymentPayLater {
			if err := checkCredit(ctx, tx, userID, total); err != nil {
				return err
			}
			paymentStatus = models.PaymentDeferred
		}

		var couponRef *int64
		var couponCode *string
		if coupon != nil {
			couponRef, couponCode = &coupon.ID, &coupon.Code
		}
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO orders (user_id, address_id, recipient_name, phone, city, area, street,
			                    building_number, apartment_number, additional_instructions,
			                    subtotal_amount, discount_amount, total_amount, coupon_id, coupon_code,
			                    status, payment_method, payment_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			RETURNING id`,
			userID, addr.ID, addr.RecipientName, addr.Phone, addr.City, addr.Area, addr.Street,
			addr.BuildingNumber, addr.ApartmentNumber, addr.AdditionalInstructions,
			subtotal, discount, total, couponRef, couponCode,
			models.OrderPending, method, paymentStatus,
		).Scan(&orderID); err != nil {
			return err
		}
		if coupon != nil {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO coupon_redemptions (coupon_id, user_id, order_id, discount_amount) VALUES ($1, $2, $3, $4)`,
				coupon.ID, userID, orderID, discount); err != nil {
				return err
			}
		}

		for _, l := range lines {
			unit, offer := l.unitPrice()
			var offerID *int64
			var offerName *string
			if offer != nil {
				offerID, offerName = &offer.ID, &offer.Name
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO order_items (order_id, product_id, product_name, product_name_ar, quantity,
				                         unit_price, selected_color, selected_color_ar, subtotal, offer_id, offer_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
				orderID, l.ProductID, l.Name, l.NameAr, l.Quantity,
				unit, l.Color, l.ColorAr, pricing.LineTotal(unit, l.Quantity), offerID, offerName,
			); err != nil {
				return err
			}
		}

		for _, b := range bundles {
			var obID int64
			if err := tx.QueryRowContext(ctx, `
				INSERT INTO order_bundles (order_id, bundle_id, bundle_name, bundle_name_ar, quantity, unit_price, subtotal)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
				orderID, b.bundle.ID, b.bundle.Name, b.bundle.NameAr, b.Quantity, b.bundle.Price,
				pricing.LineTotal(b.bundle.Price, b.Quantity)).Scan(&obID); err != nil {
				return err
			}
			for _, c := range b.comps {
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO order_bundle_items (order_bundle_id, product_id, product_name, product_name_ar,
					                                quantity, selected_color, selected_color_ar)
					VALUES ($1, $2, $3, $4, $5, $6, $7)`,
					obID, c.ProductID, c.ProductName, c.ProductNameAr, c.Quantity, c.SelectedColor, c.SelectedColorAr,
				); err != nil {
					return err
				}
			}
		}

		// Reduce stock once per product with the combined quantity. The rows are
		// locked and checked above; "stock >= $1" is a second line of defence.
		for pid, n := range need {
			res, err := tx.ExecContext(ctx, `
				UPDATE products SET stock = stock - $1, updated_at = NOW()
				WHERE id = $2 AND stock >= $1`, n, pid)
			if err != nil {
				return err
			}
			if k, _ := res.RowsAffected(); k != 1 {
				return apperr.Conflict(fmt.Sprintf("%s: not enough stock", names[pid]))
			}
		}

		if method == models.PaymentPayLater {
			note := fmt.Sprintf("order #%d (pay later)", orderID)
			if err := insertDebt(ctx, tx, userID, &orderID, models.DebtCharge, total, &note, nil); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM cart_bundles WHERE cart_id = $1`, cartID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE carts SET coupon_id = NULL WHERE id = $1`, cartID); err != nil {
			return err
		}
		return touchCart(ctx, tx, cartID)
	})
	return orderID, err
}

// checkCredit fails unless balance + amount fits within the credit limit.
func checkCredit(ctx context.Context, tx *sql.Tx, userID int64, amount decimal.Decimal) error {
	var limit decimal.Decimal
	if err := tx.QueryRowContext(ctx, `SELECT credit_limit FROM users WHERE id = $1`, userID).Scan(&limit); err != nil {
		return err
	}
	if !limit.IsPositive() {
		return apperr.Conflict("pay later is not available for your account; please contact the shop")
	}
	balance, err := debtBalance(ctx, tx, userID)
	if err != nil {
		return err
	}
	available := limit.Sub(balance)
	if amount.GreaterThan(available) {
		if available.IsNegative() {
			available = decimal.Zero
		}
		return apperr.Conflict(fmt.Sprintf("this order (%s) is more than your available credit (%s)",
			amount.StringFixed(2), available.StringFixed(2)))
	}
	return nil
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

// loadItems fills Items and Bundles for every order with a fixed number of
// queries (not one per order).
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
		SELECT oi.id, oi.order_id, oi.product_id, oi.product_name, oi.product_name_ar, oi.quantity,
		       oi.unit_price, oi.selected_color, oi.selected_color_ar, oi.subtotal, oi.offer_id, oi.offer_name,
		       (SELECT pi.image_url FROM product_images pi WHERE pi.product_id = oi.product_id
		         ORDER BY pi.is_primary DESC, pi.display_order, pi.id LIMIT 1)
		FROM order_items oi
		WHERE oi.order_id = ANY($1)
		ORDER BY oi.order_id, oi.id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.ProductName, &it.ProductNameAr,
			&it.Quantity, &it.UnitPrice, &it.SelectedColor, &it.SelectedColorAr, &it.Subtotal,
			&it.OfferID, &it.OfferName, &it.ProductImageURL); err != nil {
			return err
		}
		i := index[it.OrderID]
		orders[i].Items = append(orders[i].Items, it)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return loadOrderBundles(ctx, q, orders, ids, index)
}

func loadOrderBundles(ctx context.Context, q database.Querier, orders []models.Order, ids []int64, index map[int64]int) error {
	rows, err := q.QueryContext(ctx, `
		SELECT ob.id, ob.order_id, ob.bundle_id, ob.bundle_name, ob.bundle_name_ar, ob.quantity, ob.unit_price, ob.subtotal,
		       obi.product_id, obi.product_name, obi.product_name_ar, obi.quantity, obi.selected_color, obi.selected_color_ar
		FROM order_bundles ob
		JOIN order_bundle_items obi ON obi.order_bundle_id = ob.id
		WHERE ob.order_id = ANY($1)
		ORDER BY ob.order_id, ob.id, obi.id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var b models.OrderBundle
		var it models.OrderBundleItem
		if err := rows.Scan(&b.ID, &b.OrderID, &b.BundleID, &b.BundleName, &b.BundleNameAr, &b.Quantity,
			&b.UnitPrice, &b.Subtotal, &it.ProductID, &it.ProductName, &it.ProductNameAr, &it.Quantity,
			&it.SelectedColor, &it.SelectedColorAr); err != nil {
			return err
		}
		o := &orders[index[b.OrderID]]
		if n := len(o.Bundles); n == 0 || o.Bundles[n-1].ID != b.ID {
			b.Items = []models.OrderBundleItem{}
			o.Bundles = append(o.Bundles, b)
		}
		last := &o.Bundles[len(o.Bundles)-1]
		last.Items = append(last.Items, it)
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

// List returns a page of all orders, optionally filtered by status / payment method (admin).
func (r *OrderRepository) List(ctx context.Context, f models.OrderFilter) (models.Page[models.Order], error) {
	var w whereBuilder
	if f.Status != "" {
		w.add("o.status = ?::text", f.Status)
	}
	if f.PaymentMethod != "" {
		w.add("o.payment_method = ?::text", f.PaymentMethod)
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
		return setStatus(ctx, tx, orderID, userID, models.OrderCancelled, "")
	})
}

// UpdateStatus moves an order along its life cycle (admin). Side effects:
//   - cancelled -> stock is put back (single items and bundle contents), and
//     anything charged to the customer's debt for this order is reversed
//   - delivered -> the customer's reviews of these products become verified,
//     and the payment is settled:
//   - pay_later orders stay "deferred" (the debt was charged at checkout)
//   - cash orders become "paid", unless paymentCollected is false: then the
//     amount is charged to the customer's debt and the payment is "deferred"
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, to string, paymentCollected *bool, adminID int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var from, method string
		var userID int64
		var total decimal.Decimal
		if err := tx.QueryRowContext(ctx,
			`SELECT status, payment_method, user_id, total_amount FROM orders WHERE id = $1 FOR UPDATE`, orderID,
		).Scan(&from, &method, &userID, &total); err != nil {
			return apperr.FromDB(err, "order not found")
		}
		if !models.CanTransitionOrder(from, to) {
			return apperr.Conflict(fmt.Sprintf("cannot change order status from %s to %s", from, to))
		}
		if paymentCollected != nil && (to != models.OrderDelivered || method != models.PaymentCashOnDelivery) {
			return apperr.Invalid("payment_collected is only used when a cash_on_delivery order is delivered")
		}

		payment := ""
		if to == models.OrderDelivered {
			if err := verifyReviews(ctx, tx, orderID); err != nil {
				return err
			}
			switch {
			case method == models.PaymentPayLater:
				// already on the debt account since checkout
			case paymentCollected != nil && !*paymentCollected:
				// The admin decides to let the customer owe it; the credit limit
				// only restricts customers' own pay-later checkouts.
				if err := lockUser(ctx, tx, userID); err != nil {
					return err
				}
				note := fmt.Sprintf("order #%d delivered, cash not collected", orderID)
				if err := insertDebt(ctx, tx, userID, &orderID, models.DebtCharge, total, &note, &adminID); err != nil {
					return err
				}
				payment = models.PaymentDeferred
			default:
				payment = models.PaymentPaid
			}
		}
		return setStatus(ctx, tx, orderID, userID, to, payment)
	})
}

// verifyReviews marks the customer's reviews of every product in the order -
// single items and products inside bundles - as verified purchases.
func verifyReviews(ctx context.Context, tx *sql.Tx, orderID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE product_reviews r SET verified_purchase = TRUE
		FROM orders o
		WHERE o.id = $1 AND r.user_id = o.user_id AND NOT r.verified_purchase
		  AND r.product_id IN (
			SELECT oi.product_id FROM order_items oi WHERE oi.order_id = o.id
			UNION
			SELECT obi.product_id FROM order_bundles ob
			JOIN order_bundle_items obi ON obi.order_bundle_id = ob.id
			WHERE ob.order_id = o.id)`, orderID)
	return err
}

// setStatus writes the new status. Cancelling restores stock and reverses any
// debt charged for the order. payment == "" keeps the current payment status.
func setStatus(ctx context.Context, tx *sql.Tx, orderID, userID int64, status, payment string) error {
	if status == models.OrderCancelled {
		// One row per product: single items + (bundle qty x qty inside bundle).
		// Products that were hard-deleted have product_id NULL and are skipped.
		if _, err := tx.ExecContext(ctx, `
			UPDATE products p SET stock = p.stock + x.qty, updated_at = NOW()
			FROM (
				SELECT product_id, SUM(qty) AS qty FROM (
					SELECT oi.product_id, oi.quantity AS qty
					FROM order_items oi WHERE oi.order_id = $1
					UNION ALL
					SELECT obi.product_id, obi.quantity * ob.quantity
					FROM order_bundles ob JOIN order_bundle_items obi ON obi.order_bundle_id = ob.id
					WHERE ob.order_id = $1
				) parts
				WHERE product_id IS NOT NULL
				GROUP BY product_id
			) x
			WHERE p.id = x.product_id`, orderID); err != nil {
			return err
		}
		if err := reverseOrderDebt(ctx, tx, userID, orderID); err != nil {
			return err
		}
		// Give the coupon use back (the order keeps the code and discount it had).
		if _, err := tx.ExecContext(ctx, `DELETE FROM coupon_redemptions WHERE order_id = $1`, orderID); err != nil {
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
