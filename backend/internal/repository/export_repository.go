package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// ExportRepository reads whole tables for the admin's CSV downloads. Row-level
// security limits every query to the current store.
type ExportRepository struct {
	db *sql.DB
}

func NewExportRepository(db *sql.DB) *ExportRepository { return &ExportRepository{db: db} }

// OrderExportRow is one order in the orders CSV.
type OrderExportRow struct {
	ID            int64
	CreatedAt     time.Time
	Status        string
	PaymentMethod string
	PaymentStatus string
	CustomerName  string
	CustomerEmail string
	Recipient     string
	Phone         string
	City          string
	Area          string
	Street        string
	Items         string // "Kettle x2; Desk set x1"
	Units         int    // products + bundles, by quantity
	Subtotal      decimal.Decimal
	Discount      decimal.Decimal
	CouponCode    string
	Shipping      decimal.Decimal
	Total         decimal.Decimal
}

// MaxExportRows keeps one download bounded (a shop's whole history fits easily).
const MaxExportRows = 100000

// Orders returns the orders placed in [from, to) (zero times = no limit),
// optionally with one status and one payment method, oldest first.
func (r *ExportRepository) Orders(ctx context.Context, from, to time.Time, status, payment string) ([]OrderExportRow, error) {
	var w whereBuilder
	if !from.IsZero() {
		w.add("o.created_at >= ?", from)
	}
	if !to.IsZero() {
		w.add("o.created_at < ?", to)
	}
	if status != "" {
		w.add("o.status = ?::text", status)
	}
	if payment != "" {
		w.add("o.payment_method = ?::text", payment)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id, o.created_at, o.status, o.payment_method, o.payment_status,
		       trim(u.first_name || ' ' || u.last_name), u.email,
		       o.recipient_name, o.phone, o.city, o.area, o.street,
		       COALESCE(concat_ws('; ',
		           (SELECT string_agg(oi.product_name || ' x' || oi.quantity, '; ' ORDER BY oi.id) FROM order_items oi WHERE oi.order_id = o.id),
		           (SELECT string_agg(ob.bundle_name || ' x' || ob.quantity, '; ' ORDER BY ob.id) FROM order_bundles ob WHERE ob.order_id = o.id)), ''),
		       COALESCE((SELECT SUM(oi.quantity) FROM order_items oi WHERE oi.order_id = o.id), 0)
		     + COALESCE((SELECT SUM(ob.quantity) FROM order_bundles ob WHERE ob.order_id = o.id), 0),
		       o.subtotal_amount, o.discount_amount, COALESCE(o.coupon_code, ''), o.shipping_amount, o.total_amount
		FROM orders o JOIN users u ON u.id = o.user_id`+w.sql()+`
		ORDER BY o.created_at, o.id LIMIT `+w.next(MaxExportRows), w.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OrderExportRow{}
	for rows.Next() {
		var o OrderExportRow
		if err := rows.Scan(&o.ID, &o.CreatedAt, &o.Status, &o.PaymentMethod, &o.PaymentStatus,
			&o.CustomerName, &o.CustomerEmail, &o.Recipient, &o.Phone, &o.City, &o.Area, &o.Street,
			&o.Items, &o.Units, &o.Subtotal, &o.Discount, &o.CouponCode, &o.Shipping, &o.Total); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// CustomerExportRow is one account in the customers CSV.
type CustomerExportRow struct {
	ID            int64
	FirstName     string
	LastName      string
	Email         string
	Phone         string
	Role          string
	Active        bool
	EmailVerified bool
	JoinedAt      *time.Time
	Orders        int             // not cancelled
	Spent         decimal.Decimal // total of those orders
	CreditLimit   decimal.Decimal
	Balance       decimal.Decimal // what the customer owes (pay later)
}

// Customers returns every account of the store, oldest first.
func (r *ExportRepository) Customers(ctx context.Context) ([]CustomerExportRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.email, COALESCE(u.phone, ''), u.role, u.is_active,
		       u.email_verified_at IS NOT NULL, u.created_at,
		       (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id AND o.status <> 'cancelled'),
		       COALESCE((SELECT SUM(o.total_amount) FROM orders o WHERE o.user_id = u.id AND o.status <> 'cancelled'), 0),
		       u.credit_limit,
		       COALESCE((SELECT SUM(t.amount) FROM customer_debt_transactions t WHERE t.user_id = u.id), 0)
		FROM users u
		ORDER BY u.created_at NULLS FIRST, u.id
		LIMIT $1`, MaxExportRows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CustomerExportRow{}
	for rows.Next() {
		var c CustomerExportRow
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Role, &c.Active,
			&c.EmailVerified, &c.JoinedAt, &c.Orders, &c.Spent, &c.CreditLimit, &c.Balance); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
