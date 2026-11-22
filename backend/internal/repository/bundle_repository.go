package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"

	"github.com/shopspring/decimal"
)

type BundleRepository struct {
	db *sql.DB
}

func NewBundleRepository(db *sql.DB) *BundleRepository { return &BundleRepository{db: db} }

// bundleLiveSQL: the bundle is switched on and today is inside its optional window.
const bundleLiveSQL = `(b.is_active AND (b.starts_at IS NULL OR b.starts_at <= NOW()) AND (b.ends_at IS NULL OR b.ends_at > NOW()))`

const bundleColumns = `
	b.id, b.name, b.name_ar, COALESCE(b.description, ''), COALESCE(b.description_ar, ''), b.image_url,
	b.price, b.starts_at, b.ends_at, b.is_active, b.created_at, b.updated_at, ` + bundleLiveSQL

// bundleRow is a bundle plus whether it is live right now.
type bundleRow struct {
	models.Bundle
	live bool
}

func scanBundle(s scanner, prefix ...any) (bundleRow, error) {
	var b bundleRow
	dest := append(prefix, &b.ID, &b.Name, &b.NameAr, &b.Description, &b.DescriptionAr, &b.ImageURL,
		&b.Price, &b.StartsAt, &b.EndsAt, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.live)
	err := s.Scan(dest...)
	return b, err
}

// bundleComponent is one product inside a bundle with its LIVE product data.
type bundleComponent struct {
	models.BundleItem
	price      decimal.Decimal
	discount   decimal.Decimal
	colorCount int
}

// loadBundleComponents returns the products of each bundle, with current
// stock, status and best price (offers included).
func loadBundleComponents(ctx context.Context, q database.Querier, bundleIDs []int64) (map[int64][]bundleComponent, error) {
	out := map[int64][]bundleComponent{}
	if len(bundleIDs) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `
		SELECT bi.bundle_id, p.id, p.name, p.name_ar,
		       (SELECT pi.image_url FROM product_images pi WHERE pi.product_id = p.id
		         ORDER BY pi.is_primary DESC, pi.display_order, pi.id LIMIT 1),
		       bi.quantity, bi.selected_color, pc.name_ar, p.price, p.discount_value, p.stock, p.is_active,
		       (SELECT COUNT(*) FROM product_colors x WHERE x.product_id = p.id)
		FROM bundle_items bi
		JOIN products p ON p.id = bi.product_id
		LEFT JOIN product_colors pc ON pc.product_id = p.id AND pc.name = bi.selected_color
		WHERE bi.bundle_id = ANY($1)
		ORDER BY bi.bundle_id, p.id`, bundleIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var productIDs []int64
	for rows.Next() {
		var bid int64
		var c bundleComponent
		if err := rows.Scan(&bid, &c.ProductID, &c.ProductName, &c.ProductNameAr, &c.PrimaryImageURL,
			&c.Quantity, &c.SelectedColor, &c.SelectedColorAr, &c.price, &c.discount, &c.Stock, &c.IsActive,
			&c.colorCount); err != nil {
			return nil, err
		}
		out[bid] = append(out[bid], c)
		productIDs = append(productIDs, c.ProductID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	offers, err := loadLiveOffers(ctx, q, productIDs)
	if err != nil {
		return nil, err
	}
	for bid := range out {
		for i := range out[bid] {
			c := &out[bid][i]
			c.UnitPrice, _ = pricing.BestPrice(c.price, c.discount, offers[c.ProductID])
		}
	}
	return out, nil
}

// bundleProblem returns "" when `qty` bundles can be bought right now,
// otherwise the reason. Shared by the bundle pages, the cart and checkout.
func bundleProblem(live bool, comps []bundleComponent, qty int) string {
	if !live {
		return "this bundle is no longer available"
	}
	if len(comps) == 0 {
		return "this bundle has no products"
	}
	for _, c := range comps {
		switch {
		case !c.IsActive:
			return fmt.Sprintf("%s is no longer available", c.ProductName)
		case c.SelectedColor != nil && c.SelectedColorAr == nil:
			return fmt.Sprintf("the color of %s in this bundle is no longer available", c.ProductName)
		case c.SelectedColor == nil && c.colorCount > 0:
			return fmt.Sprintf("%s in this bundle has no color set; please contact the shop", c.ProductName)
		case c.Stock < c.Quantity*qty:
			return fmt.Sprintf("not enough stock of %s", c.ProductName)
		}
	}
	return ""
}

// bundleAvailableStock is how many complete bundles the stock allows.
func bundleAvailableStock(comps []bundleComponent) int {
	if len(comps) == 0 {
		return 0
	}
	n := -1
	for _, c := range comps {
		if !c.IsActive {
			return 0
		}
		if k := c.Stock / c.Quantity; n < 0 || k < n {
			n = k
		}
	}
	return n
}

// fillBundle copies the components into the bundle and computes totals.
func fillBundle(b *bundleRow, comps []bundleComponent) {
	b.Items = []models.BundleItem{}
	b.OriginalTotal = decimal.Zero
	for _, c := range comps {
		b.Items = append(b.Items, c.BundleItem)
		b.OriginalTotal = b.OriginalTotal.Add(pricing.LineTotal(c.UnitPrice, c.Quantity))
	}
	b.Savings = b.OriginalTotal.Sub(b.Price)
	if b.Savings.IsNegative() {
		b.Savings = decimal.Zero
	}
	b.AvailableStock = bundleAvailableStock(comps)
	if p := bundleProblem(b.live, comps, 1); p != "" {
		b.UnavailableReason = &p
	} else {
		b.IsAvailable = true
	}
}

func (r *BundleRepository) withComponents(ctx context.Context, rows []bundleRow) ([]models.Bundle, error) {
	ids := make([]int64, len(rows))
	for i, b := range rows {
		ids[i] = b.ID
	}
	comps, err := loadBundleComponents(ctx, r.db, ids)
	if err != nil {
		return nil, err
	}
	out := make([]models.Bundle, len(rows))
	for i := range rows {
		fillBundle(&rows[i], comps[rows[i].ID])
		out[i] = rows[i].Bundle
	}
	return out, nil
}

// List returns bundles. Customers only see live ones.
func (r *BundleRepository) List(ctx context.Context, f models.BundleFilter) (models.Page[models.Bundle], error) {
	var w whereBuilder
	if !f.IncludeUnavailable {
		w.addRaw(bundleLiveSQL)
	}
	page := models.Page[models.Bundle]{Items: []models.Bundle{}, Limit: f.Limit, Offset: f.Offset}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM bundles b`+w.sql(), w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+bundleColumns+` FROM bundles b`+w.sql()+
		` ORDER BY b.created_at DESC, b.id DESC LIMIT `+w.next(f.Limit)+` OFFSET `+w.next(f.Offset), w.args...)
	if err != nil {
		return page, err
	}
	var list []bundleRow
	for rows.Next() {
		b, err := scanBundle(rows)
		if err != nil {
			rows.Close()
			return page, err
		}
		list = append(list, b)
	}
	if err := rows.Close(); err != nil {
		return page, err
	}
	if err := rows.Err(); err != nil {
		return page, err
	}
	items, err := r.withComponents(ctx, list)
	if err != nil {
		return page, err
	}
	page.Items = items
	return page, nil
}

// Get returns one bundle; customers (includeUnavailable=false) only live ones.
func (r *BundleRepository) Get(ctx context.Context, id int64, includeUnavailable bool) (models.Bundle, error) {
	b, err := scanBundle(r.db.QueryRowContext(ctx,
		`SELECT `+bundleColumns+` FROM bundles b WHERE b.id = $1 AND (`+bundleLiveSQL+` OR $2)`, id, includeUnavailable))
	if err != nil {
		return models.Bundle{}, apperr.FromDB(err, "bundle not found")
	}
	items, err := r.withComponents(ctx, []bundleRow{b})
	if err != nil {
		return models.Bundle{}, err
	}
	return items[0], nil
}

// writeItems validates every item's color against the product and stores the
// canonical color name. Products with colors must have one chosen.
func writeItems(ctx context.Context, tx *sql.Tx, bundleID int64, items []models.BundleItemInput) error {
	ids := make([]int64, len(items))
	for i, it := range items {
		ids[i] = it.ProductID
	}
	if err := requireProducts(ctx, tx, ids); err != nil {
		return err
	}
	for _, it := range items {
		color, err := resolveColor(ctx, tx, it.ProductID, it.SelectedColor)
		if err != nil {
			if e, ok := apperr.As(err); ok && e.Fields != nil {
				return apperr.Validation(map[string]string{
					"items": fmt.Sprintf("product %d: %s", it.ProductID, e.Fields["selected_color"])})
			}
			return err
		}
		var name *string
		if color != nil {
			name = &color.Name
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO bundle_items (bundle_id, product_id, quantity, selected_color)
			VALUES ($1, $2, $3, $4)`, bundleID, it.ProductID, it.Quantity, name); err != nil {
			return err
		}
	}
	return nil
}

func (r *BundleRepository) Create(ctx context.Context, in models.BundleInput) (int64, error) {
	var id int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO bundles (name, name_ar, description, description_ar, image_url, price, starts_at, ends_at, is_active)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7, $8, COALESCE($9, TRUE))
			RETURNING id`,
			in.Name, in.NameAr, in.Description, in.DescriptionAr, in.ImageURL, in.Price,
			nullTime(in.StartsAt), nullTime(in.EndsAt), in.IsActive,
		).Scan(&id); err != nil {
			return err
		}
		return writeItems(ctx, tx, id, in.Items)
	})
	return id, apperr.FromDB(err, "bundle not found")
}

// Update replaces the bundle and its items. Bundles already in carts follow
// the new definition; past orders keep their own snapshot.
func (r *BundleRepository) Update(ctx context.Context, id int64, in models.BundleInput) error {
	return apperr.FromDB(database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE bundles SET name = $1, name_ar = $2, description = NULLIF($3, ''), description_ar = NULLIF($4, ''),
			       image_url = $5, price = $6, starts_at = $7, ends_at = $8,
			       is_active = COALESCE($9, is_active), updated_at = NOW()
			WHERE id = $10`,
			in.Name, in.NameAr, in.Description, in.DescriptionAr, in.ImageURL, in.Price,
			nullTime(in.StartsAt), nullTime(in.EndsAt), in.IsActive, id)
		if err := expectOneRow(res, err, "bundle not found"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM bundle_items WHERE bundle_id = $1`, id); err != nil {
			return err
		}
		return writeItems(ctx, tx, id, in.Items)
	}), "bundle not found")
}

// Deactivate is a soft delete: hidden from customers, kept for order history.
func (r *BundleRepository) Deactivate(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND is_active`, id)
	return expectOneRow(res, err, "bundle not found or already inactive")
}

// nullTime passes a *time.Time as NULL when nil.
func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}
