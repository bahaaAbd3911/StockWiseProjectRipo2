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

// CouponRepository manages coupon codes. The rule that decides what a coupon
// is worth for a cart (couponRow.evaluate) is shared by the cart view and by
// checkout, like cartLine.problem, so the cart never promises a discount that
// checkout then refuses.
type CouponRepository struct {
	db *sql.DB
}

func NewCouponRepository(db *sql.DB) *CouponRepository { return &CouponRepository{db: db} }

// usedSQL counts the orders that used coupon c (cancelled orders do not count:
// cancelling deletes the redemption).
const couponUsedSQL = `(SELECT COUNT(*) FROM coupon_redemptions r WHERE r.coupon_id = c.id)`

const couponStatusSQL = `CASE
	WHEN NOT c.is_active THEN 'disabled'
	WHEN c.starts_at IS NOT NULL AND c.starts_at > NOW() THEN 'scheduled'
	WHEN c.ends_at IS NOT NULL AND c.ends_at <= NOW() THEN 'expired'
	WHEN c.usage_limit IS NOT NULL AND ` + couponUsedSQL + ` >= c.usage_limit THEN 'used_up'
	ELSE 'live' END`

const couponColumns = `c.id, c.code, c.description, c.description_ar, c.discount_type, c.discount_value,
	c.scope, c.category, c.min_order_amount, c.max_discount, c.starts_at, c.ends_at,
	c.usage_limit, c.per_customer_limit, c.is_active, ` + couponStatusSQL + `, ` + couponUsedSQL + `,
	COALESCE((SELECT SUM(r.discount_amount) FROM coupon_redemptions r WHERE r.coupon_id = c.id), 0),
	c.created_at, c.updated_at`

func scanCoupon(s scanner) (models.Coupon, error) {
	var c models.Coupon
	err := s.Scan(&c.ID, &c.Code, &c.Description, &c.DescriptionAr, &c.DiscountType, &c.DiscountValue,
		&c.Scope, &c.Category, &c.MinOrderAmount, &c.MaxDiscount, &c.StartsAt, &c.EndsAt,
		&c.UsageLimit, &c.PerCustomerLimit, &c.IsActive, &c.Status, &c.TimesUsed, &c.TotalDiscount,
		&c.CreatedAt, &c.UpdatedAt)
	c.ProductIDs = []int64{}
	return c, err
}

// couponStatusFilter maps a status filter to a fixed condition.
var couponStatusFilter = map[string]string{
	models.CouponLive:      couponStatusSQL + ` = 'live'`,
	models.CouponScheduled: couponStatusSQL + ` = 'scheduled'`,
	models.CouponExpired:   couponStatusSQL + ` = 'expired'`,
	models.CouponDisabled:  couponStatusSQL + ` = 'disabled'`,
	models.CouponUsedUp:    couponStatusSQL + ` = 'used_up'`,
}

func IsValidCouponStatus(s string) bool { _, ok := couponStatusFilter[s]; return ok }

// ---------- admin ----------

func (r *CouponRepository) List(ctx context.Context, f models.CouponFilter) (models.Page[models.Coupon], error) {
	w := &whereBuilder{}
	if cond, ok := couponStatusFilter[f.Status]; ok {
		w.addRaw(cond)
	}
	if f.Search != "" {
		w.add(`c.code ILIKE '%' || ? || '%'`, escapeLike(f.Search))
	}
	page := models.Page[models.Coupon]{Items: []models.Coupon{}, Limit: f.Limit, Offset: f.Offset}
	where := w.sql()
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM coupons c`+where, w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	limit, offset := w.next(f.Limit), w.next(f.Offset)
	rows, err := r.db.QueryContext(ctx, `SELECT `+couponColumns+` FROM coupons c`+where+
		` ORDER BY c.created_at DESC, c.id DESC LIMIT `+limit+` OFFSET `+offset, w.args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanCoupon(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, c)
	}
	if err := rows.Err(); err != nil {
		return page, err
	}
	return page, r.attachProductIDs(ctx, r.db, page.Items)
}

func (r *CouponRepository) Get(ctx context.Context, id int64) (models.Coupon, error) {
	c, err := scanCoupon(r.db.QueryRowContext(ctx, `SELECT `+couponColumns+` FROM coupons c WHERE c.id = $1`, id))
	if err != nil {
		return c, apperr.FromDB(err, "coupon not found")
	}
	list := []models.Coupon{c}
	err = r.attachProductIDs(ctx, r.db, list)
	return list[0], err
}

func (r *CouponRepository) attachProductIDs(ctx context.Context, q database.Querier, coupons []models.Coupon) error {
	if len(coupons) == 0 {
		return nil
	}
	ids := make([]int64, len(coupons))
	index := map[int64]int{}
	for i, c := range coupons {
		ids[i], index[c.ID] = c.ID, i
	}
	rows, err := q.QueryContext(ctx,
		`SELECT coupon_id, product_id FROM coupon_products WHERE coupon_id = ANY($1) ORDER BY product_id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, pid int64
		if err := rows.Scan(&cid, &pid); err != nil {
			return err
		}
		coupons[index[cid]].ProductIDs = append(coupons[index[cid]].ProductIDs, pid)
	}
	return rows.Err()
}

func couponWriteArgs(in models.CouponInput) []any {
	return []any{strings.ToUpper(in.Code), in.Description, in.DescriptionAr, in.DiscountType, in.DiscountValue,
		in.Scope, in.Category, in.MinOrderAmount, in.MaxDiscount, in.StartsAt, in.EndsAt,
		in.UsageLimit, in.PerCustomerLimit}
}

func duplicateCode(err error) error {
	if apperr.IsUniqueViolation(err, "coupons_code_key") {
		return apperr.Conflict("a coupon with this code already exists")
	}
	return err
}

func (r *CouponRepository) Create(ctx context.Context, in models.CouponInput) (int64, error) {
	var id int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := requireProducts(ctx, tx, in.ProductIDs); err != nil {
			return err
		}
		active := in.IsActive == nil || *in.IsActive
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO coupons (code, description, description_ar, discount_type, discount_value, scope, category,
			                     min_order_amount, max_discount, starts_at, ends_at, usage_limit, per_customer_limit, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`,
			append(couponWriteArgs(in), active)...).Scan(&id); err != nil {
			return duplicateCode(err)
		}
		return setCouponProducts(ctx, tx, id, in.ProductIDs)
	})
	return id, err
}

// Update replaces the coupon. An omitted is_active keeps the current value.
func (r *CouponRepository) Update(ctx context.Context, id int64, in models.CouponInput) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := requireProducts(ctx, tx, in.ProductIDs); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE coupons SET code = $1, description = $2, description_ar = $3, discount_type = $4,
			       discount_value = $5, scope = $6, category = $7, min_order_amount = $8, max_discount = $9,
			       starts_at = $10, ends_at = $11, usage_limit = $12, per_customer_limit = $13,
			       is_active = COALESCE($14, is_active), updated_at = NOW()
			WHERE id = $15`, append(couponWriteArgs(in), in.IsActive, id)...)
		if err := expectOneRow(res, duplicateCode(err), "coupon not found"); err != nil {
			return err
		}
		return setCouponProducts(ctx, tx, id, in.ProductIDs)
	})
}

func setCouponProducts(ctx context.Context, tx *sql.Tx, couponID int64, productIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM coupon_products WHERE coupon_id = $1`, couponID); err != nil {
		return err
	}
	if len(productIDs) == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO coupon_products (coupon_id, product_id) SELECT $1, unnest($2::bigint[])`, couponID, productIDs)
	return err
}

// Delete removes a coupon. Orders keep the code and the discount they got.
func (r *CouponRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM coupons WHERE id = $1`, id)
	return expectOneRow(res, err, "coupon not found")
}

// ---------- the coupon rule (cart view + checkout) ----------

// couponRow is a coupon as seen by one customer: the coupon, how often that
// customer already used it, and its product list.
type couponRow struct {
	models.Coupon
	usedByUser int
	products   map[int64]bool
}

// loadCoupon reads a coupon for a customer. lock = true (checkout) locks the
// coupon row, so two checkouts cannot both take the last use.
func loadCoupon(ctx context.Context, q database.Querier, where string, arg any, userID int64, lock bool) (couponRow, error) {
	query := `SELECT ` + couponColumns + ` FROM coupons c WHERE ` + where
	if lock {
		// Lock first (FOR UPDATE cannot be combined with the aggregates in
		// couponColumns), then read the coupon with its counts.
		var id int64
		if err := q.QueryRowContext(ctx, `SELECT c.id FROM coupons c WHERE `+where+` FOR UPDATE`, arg).Scan(&id); err != nil {
			return couponRow{}, err
		}
	}
	c, err := scanCoupon(q.QueryRowContext(ctx, query, arg))
	if err != nil {
		return couponRow{}, err
	}
	row := couponRow{Coupon: c, products: map[int64]bool{}}
	if err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`, c.ID, userID,
	).Scan(&row.usedByUser); err != nil {
		return couponRow{}, err
	}
	if c.Scope == models.CouponScopeProducts {
		rows, err := q.QueryContext(ctx, `SELECT product_id FROM coupon_products WHERE coupon_id = $1`, c.ID)
		if err != nil {
			return couponRow{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var pid int64
			if err := rows.Scan(&pid); err != nil {
				return couponRow{}, err
			}
			row.products[pid] = true
		}
		if err := rows.Err(); err != nil {
			return couponRow{}, err
		}
	}
	return row, nil
}

// errCouponNotValid is the answer for unknown and switched-off codes alike.
var errCouponNotValid = validationCode("this code is not valid")

func validationCode(msg string) error { return apperr.Validation(map[string]string{"code": msg}) }

// evaluate returns the discount the coupon gives on these cart lines, or the
// reason it cannot be used. Only lines that can be bought count.
func (c couponRow) evaluate(lines []cartLine, bundles []cartBundleLine) (decimal.Decimal, string) {
	switch c.Status {
	case models.CouponDisabled:
		return decimal.Zero, "this code is not valid"
	case models.CouponScheduled:
		return decimal.Zero, "this code cannot be used yet"
	case models.CouponExpired:
		return decimal.Zero, "this code has expired"
	case models.CouponUsedUp:
		return decimal.Zero, "this code has reached its usage limit"
	}
	if c.PerCustomerLimit != nil && c.usedByUser >= *c.PerCustomerLimit {
		if *c.PerCustomerLimit == 1 {
			return decimal.Zero, "you have already used this code"
		}
		return decimal.Zero, fmt.Sprintf("you have already used this code %d times (the maximum)", c.usedByUser)
	}

	subtotal, eligible := decimal.Zero, decimal.Zero
	for _, l := range lines {
		if l.problem() != "" {
			continue
		}
		unit, _ := l.unitPrice()
		line := pricing.LineTotal(unit, l.Quantity)
		subtotal = subtotal.Add(line)
		if c.applies(l) {
			eligible = eligible.Add(line)
		}
	}
	for _, b := range bundles {
		if b.problem() != "" {
			continue
		}
		line := pricing.LineTotal(b.bundle.Price, b.Quantity)
		subtotal = subtotal.Add(line)
		if c.Scope == models.CouponScopeAll { // bundles already have their own special price
			eligible = eligible.Add(line)
		}
	}
	if !eligible.IsPositive() {
		return decimal.Zero, "this code does not apply to the items in your cart"
	}
	if subtotal.LessThan(c.MinOrderAmount) {
		return decimal.Zero, fmt.Sprintf("add %s more to use this code (minimum order %s)",
			c.MinOrderAmount.Sub(subtotal).StringFixed(2), c.MinOrderAmount.StringFixed(2))
	}
	return pricing.CouponDiscount(c.DiscountType, c.DiscountValue, c.MaxDiscount, eligible), ""
}

// applies reports whether a product line is covered by the coupon's scope.
func (c couponRow) applies(l cartLine) bool {
	switch c.Scope {
	case models.CouponScopeProducts:
		return c.products[l.ProductID]
	case models.CouponScopeCategory:
		return c.Category != nil && l.Category != nil && strings.EqualFold(*c.Category, *l.Category)
	}
	return true
}

func (c couponRow) cartModel(discount decimal.Decimal, problem string) *models.CartCoupon {
	cc := &models.CartCoupon{
		Code: c.Code, Description: c.Description, DescriptionAr: c.DescriptionAr,
		DiscountType: c.DiscountType, DiscountValue: c.DiscountValue, DiscountAmount: discount,
	}
	if problem != "" {
		cc.Problem = &problem
	}
	return cc
}

// cartCouponID returns the coupon attached to a cart (nil = none).
func cartCouponID(ctx context.Context, q database.Querier, cartID int64) (*int64, error) {
	var id *int64
	err := q.QueryRowContext(ctx, `SELECT coupon_id FROM carts WHERE id = $1`, cartID).Scan(&id)
	return id, err
}

// ApplyToCart attaches a code to the user's cart after checking it against
// the current cart. A code that cannot be used is refused with the reason.
func (r *CouponRepository) ApplyToCart(ctx context.Context, userID int64, code string) error {
	cartID, _, err := ensureCart(ctx, r.db, userID)
	if err != nil {
		return err
	}
	c, err := loadCoupon(ctx, r.db, `upper(c.code) = upper($1)`, code, userID, false)
	if errors.Is(err, sql.ErrNoRows) {
		return errCouponNotValid
	}
	if err != nil {
		return err
	}
	lines, err := loadCartLines(ctx, r.db, cartID, "ci.id", "")
	if err != nil {
		return err
	}
	bundles, err := loadCartBundleLines(ctx, r.db, cartID)
	if err != nil {
		return err
	}
	if _, problem := c.evaluate(lines, bundles); problem != "" {
		return validationCode(problem)
	}
	_, err = r.db.ExecContext(ctx, `UPDATE carts SET coupon_id = $1, updated_at = NOW() WHERE id = $2`, c.ID, cartID)
	return err
}

// RemoveFromCart detaches the code (no error if there was none).
func (r *CouponRepository) RemoveFromCart(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE carts SET coupon_id = NULL, updated_at = NOW() WHERE user_id = $1`, userID)
	return err
}
