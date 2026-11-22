package repository

import (
	"context"
	"database/sql"
	"time"

	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

// DashboardRepository computes the admin overview. Every query runs in one
// read-only REPEATABLE READ transaction: all numbers describe the same moment,
// even while orders keep coming in.
type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository { return &DashboardRepository{db: db} }

// DashboardRange is the time window, already converted to the store's days.
type DashboardRange struct {
	TodayFrom, TodayTo time.Time
	From, To           time.Time // the period
	PrevFrom, PrevTo   time.Time // the period before it
	Timezone           string    // IANA name, used to group orders by local date
	LowStockThreshold  int
}

// orders counted as sales: everything except cancelled.
const soldOrder = `o.status <> 'cancelled'`

func (r *DashboardRepository) Build(ctx context.Context, rg DashboardRange) (models.Dashboard, error) {
	d := models.Dashboard{
		Timezone:       rg.Timezone,
		PaymentMethods: map[string]models.MethodSales{},
		OrdersByStatus: map[string]int{},
		SalesByDay:     []models.DaySales{},
		BestSellers:    []models.BestSeller{},
		WaitingList:    []models.WaitingDemand{},
		TopCoupons:     []models.CouponUsage{},
		LowStock:       models.LowStockReport{Threshold: rg.LowStockThreshold, Products: []models.LowStockProduct{}},
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return d, err
	}
	defer tx.Rollback() //nolint:errcheck // read-only

	var nowDB time.Time
	if err := tx.QueryRowContext(ctx, `SELECT NOW()`).Scan(&nowDB); err != nil {
		return d, err
	}
	d.GeneratedAt = nowDB

	steps := []func() error{
		func() (err error) { d.Today, err = figures(ctx, tx, rg.TodayFrom, rg.TodayTo); return },
		func() (err error) { d.Period, err = figures(ctx, tx, rg.From, rg.To); return },
		func() (err error) { d.PreviousPeriod, err = figures(ctx, tx, rg.PrevFrom, rg.PrevTo); return },
		func() (err error) { d.SalesByDay, err = salesByDay(ctx, tx, rg); return },
		func() error { return paymentMethods(ctx, tx, rg, d.PaymentMethods) },
		func() error { return ordersByStatus(ctx, tx, d.OrdersByStatus) },
		func() (err error) { d.BestSellers, err = bestSellers(ctx, tx, rg); return },
		func() error { return lowStock(ctx, tx, &d.LowStock) },
		func() (err error) { d.Debts, err = debtTotals(ctx, tx, rg); return },
		func() (err error) { d.WaitingList, err = waitingDemand(ctx, tx); return },
		func() (err error) { d.TopCoupons, err = topCoupons(ctx, tx, rg); return },
		func() (err error) { d.Customers, err = customerFigures(ctx, tx, rg); return },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return d, err
		}
	}
	return d, nil
}

func figures(ctx context.Context, tx *sql.Tx, from, to time.Time) (models.SalesFigures, error) {
	f := models.SalesFigures{From: from, To: to}
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FILTER (WHERE `+soldOrder+`),
		       COALESCE(SUM(o.total_amount) FILTER (WHERE `+soldOrder+`), 0),
		       COALESCE(SUM(o.discount_amount) FILTER (WHERE `+soldOrder+`), 0),
		       COUNT(*) FILTER (WHERE o.status = 'cancelled')
		FROM orders o WHERE o.created_at >= $1 AND o.created_at < $2`, from, to,
	).Scan(&f.Orders, &f.Revenue, &f.DiscountsGiven, &f.CancelledOrders); err != nil {
		return f, err
	}
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(units), 0) FROM (
			SELECT oi.quantity AS units FROM order_items oi JOIN orders o ON o.id = oi.order_id
			WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
			UNION ALL
			SELECT obi.quantity * ob.quantity FROM order_bundles ob
			JOIN order_bundle_items obi ON obi.order_bundle_id = ob.id
			JOIN orders o ON o.id = ob.order_id
			WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
		) x`, from, to).Scan(&f.ItemsSold); err != nil {
		return f, err
	}
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE role = 'customer' AND created_at >= $1 AND created_at < $2`, from, to,
	).Scan(&f.NewCustomers); err != nil {
		return f, err
	}
	f.AverageOrderValue = decimal.Zero
	if f.Orders > 0 {
		f.AverageOrderValue = f.Revenue.Div(decimal.NewFromInt(int64(f.Orders))).Round(2)
	}
	return f, nil
}

// salesByDay returns one entry per local day of the period, zeros included,
// so a chart has no gaps.
func salesByDay(ctx context.Context, tx *sql.Tx, rg DashboardRange) ([]models.DaySales, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT to_char((o.created_at AT TIME ZONE $3)::date, 'YYYY-MM-DD'), COUNT(*), SUM(o.total_amount)
		FROM orders o
		WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
		GROUP BY 1`, rg.From, rg.To, rg.Timezone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byDate := map[string]models.DaySales{}
	for rows.Next() {
		var s models.DaySales
		if err := rows.Scan(&s.Date, &s.Orders, &s.Revenue); err != nil {
			return nil, err
		}
		byDate[s.Date] = s
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(rg.Timezone)
	if err != nil {
		return nil, err
	}
	var out []models.DaySales
	for day := rg.From.In(loc); day.Before(rg.To); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		s, ok := byDate[key]
		if !ok {
			s = models.DaySales{Date: key, Revenue: decimal.Zero}
		}
		out = append(out, s)
	}
	return out, nil
}

func paymentMethods(ctx context.Context, tx *sql.Tx, rg DashboardRange, out map[string]models.MethodSales) error {
	for _, m := range []string{models.PaymentCashOnDelivery, models.PaymentPayLater} {
		out[m] = models.MethodSales{Revenue: decimal.Zero}
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT o.payment_method, COUNT(*), SUM(o.total_amount) FROM orders o
		WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
		GROUP BY 1`, rg.From, rg.To)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m string
		var s models.MethodSales
		if err := rows.Scan(&m, &s.Orders, &s.Revenue); err != nil {
			return err
		}
		out[m] = s
	}
	return rows.Err()
}

func ordersByStatus(ctx context.Context, tx *sql.Tx, out map[string]int) error {
	for _, s := range []string{models.OrderPending, models.OrderConfirmed, models.OrderProcessing,
		models.OrderShipped, models.OrderDelivered, models.OrderCancelled} {
		out[s] = 0
	}
	rows, err := tx.QueryContext(ctx, `SELECT status, COUNT(*) FROM orders GROUP BY status`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		var n int
		if err := rows.Scan(&s, &n); err != nil {
			return err
		}
		out[s] = n
	}
	return rows.Err()
}

func bestSellers(ctx context.Context, tx *sql.Tx, rg DashboardRange) ([]models.BestSeller, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT product_id, MAX(name), MAX(name_ar), SUM(units), SUM(single) FROM (
			SELECT oi.product_id, oi.product_name AS name, oi.product_name_ar AS name_ar,
			       oi.quantity AS units, oi.subtotal AS single
			FROM order_items oi JOIN orders o ON o.id = oi.order_id
			WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
			UNION ALL
			SELECT obi.product_id, obi.product_name, obi.product_name_ar, obi.quantity * ob.quantity, 0
			FROM order_bundles ob
			JOIN order_bundle_items obi ON obi.order_bundle_id = ob.id
			JOIN orders o ON o.id = ob.order_id
			WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
		) x
		GROUP BY product_id, CASE WHEN product_id IS NULL THEN name END
		ORDER BY SUM(units) DESC, SUM(single) DESC
		LIMIT 10`, rg.From, rg.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.BestSeller{}
	for rows.Next() {
		var b models.BestSeller
		if err := rows.Scan(&b.ProductID, &b.Name, &b.NameAr, &b.Units, &b.SingleSales); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func lowStock(ctx context.Context, tx *sql.Tx, out *models.LowStockReport) error {
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM products WHERE is_active AND stock = 0`).Scan(&out.OutOfStock); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT p.id, p.name, p.name_ar, p.stock,
		       (SELECT COUNT(*) FROM waiting_list_items w WHERE w.product_id = p.id AND w.status = 'waiting') AS waiting
		FROM products p
		WHERE p.is_active AND p.stock <= $1
		ORDER BY p.stock, waiting DESC, p.name
		LIMIT 20`, out.Threshold)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.LowStockProduct
		if err := rows.Scan(&p.ProductID, &p.Name, &p.NameAr, &p.Stock, &p.Waiting); err != nil {
			return err
		}
		out.Products = append(out.Products, p)
	}
	return rows.Err()
}

func debtTotals(ctx context.Context, tx *sql.Tx, rg DashboardRange) (models.DebtTotals, error) {
	var t models.DebtTotals
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(balance), 0), COUNT(*), COUNT(*) FILTER (WHERE balance > credit_limit)
		FROM (
			SELECT u.credit_limit, COALESCE(SUM(t.amount), 0) AS balance
			FROM users u JOIN customer_debt_transactions t ON t.user_id = u.id
			GROUP BY u.id, u.credit_limit
		) b
		WHERE balance > 0`).Scan(&t.Outstanding, &t.CustomersOwing, &t.OverCreditLimit); err != nil {
		return t, err
	}
	// Payments are stored as negative amounts in the ledger.
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(-SUM(amount), 0) FROM customer_debt_transactions
		WHERE type = 'payment' AND created_at >= $1 AND created_at < $2`, rg.From, rg.To).Scan(&t.PaymentsInPeriod)
	return t, err
}

func waitingDemand(ctx context.Context, tx *sql.Tx) ([]models.WaitingDemand, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT p.id, p.name, p.name_ar, COUNT(*), SUM(w.quantity)
		FROM waiting_list_items w JOIN products p ON p.id = w.product_id
		WHERE w.status = 'waiting'
		GROUP BY p.id, p.name, p.name_ar
		ORDER BY COUNT(*) DESC, SUM(w.quantity) DESC, p.id
		LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WaitingDemand{}
	for rows.Next() {
		var w models.WaitingDemand
		if err := rows.Scan(&w.ProductID, &w.Name, &w.NameAr, &w.Requests, &w.Units); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func topCoupons(ctx context.Context, tx *sql.Tx, rg DashboardRange) ([]models.CouponUsage, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT o.coupon_code, COUNT(*), SUM(o.discount_amount) FROM orders o
		WHERE o.coupon_code IS NOT NULL AND `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2
		GROUP BY o.coupon_code
		ORDER BY COUNT(*) DESC, SUM(o.discount_amount) DESC
		LIMIT 10`, rg.From, rg.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CouponUsage{}
	for rows.Next() {
		var c models.CouponUsage
		if err := rows.Scan(&c.Code, &c.Uses, &c.Discount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func customerFigures(ctx context.Context, tx *sql.Tx, rg DashboardRange) (models.CustomerFigures, error) {
	var c models.CustomerFigures
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FILTER (WHERE is_active),
		       COUNT(*) FILTER (WHERE is_active AND email_verified_at IS NULL),
		       COUNT(*) FILTER (WHERE created_at >= $1 AND created_at < $2)
		FROM users WHERE role = 'customer'`, rg.From, rg.To).Scan(&c.Total, &c.Unverified, &c.NewInPeriod); err != nil {
		return c, err
	}
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT o.user_id) FROM orders o
		WHERE `+soldOrder+` AND o.created_at >= $1 AND o.created_at < $2`, rg.From, rg.To).Scan(&c.OrderedInPeriod)
	return c, err
}
