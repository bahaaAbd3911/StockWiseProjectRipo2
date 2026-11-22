package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"
)

// offerAppliesSQL is true when offer `o` is live right now and covers product
// `p`. It is the single SQL definition of "this offer applies to this product"
// and is reused by the offer loader, the on_offer filter and price sorting.
const offerAppliesSQL = `o.is_active AND o.starts_at <= NOW() AND o.ends_at > NOW() AND (
	(o.scope = 'products' AND EXISTS (
		SELECT 1 FROM offer_products op WHERE op.offer_id = o.id AND op.product_id = p.id))
	OR (o.scope = 'category' AND p.category IS NOT NULL AND lower(o.category) = lower(p.category)))`

// effectivePriceSQL mirrors pricing.BestPrice in SQL. It is ONLY used to sort
// product lists by price; the price shown to the customer is always computed
// in Go by pricing.BestPrice.
const effectivePriceSQL = `LEAST(
	ROUND(p.price * (100 - p.discount_value) / 100, 2),
	COALESCE((SELECT MIN(CASE WHEN o.discount_type = 'percentage'
	                          THEN ROUND(p.price * (100 - o.discount_value) / 100, 2)
	                          ELSE GREATEST(p.price - o.discount_value, 0) END)
	          FROM offers o WHERE ` + offerAppliesSQL + `), p.price))`

// loadLiveOffers returns, for each product ID, the offers that apply to it now.
func loadLiveOffers(ctx context.Context, q database.Querier, productIDs []int64) (map[int64][]models.AppliedOffer, error) {
	out := map[int64][]models.AppliedOffer{}
	if len(productIDs) == 0 {
		return out, nil
	}
	rows, err := q.QueryContext(ctx, `
		SELECT p.id, o.id, o.name, o.name_ar, o.discount_type, o.discount_value, o.ends_at
		FROM products p
		JOIN offers o ON `+offerAppliesSQL+`
		WHERE p.id = ANY($1)
		ORDER BY o.id`, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var o models.AppliedOffer
		if err := rows.Scan(&pid, &o.ID, &o.Name, &o.NameAr, &o.DiscountType, &o.DiscountValue, &o.EndsAt); err != nil {
			return nil, err
		}
		out[pid] = append(out[pid], o)
	}
	return out, rows.Err()
}

// attachOffers sets FinalPrice and Offer on products using their live offers.
func attachOffers(ctx context.Context, q database.Querier, products ...*models.Product) error {
	ids := make([]int64, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}
	offers, err := loadLiveOffers(ctx, q, ids)
	if err != nil {
		return err
	}
	for _, p := range products {
		p.FinalPrice, p.Offer = pricing.BestPrice(p.Price, p.DiscountValue, offers[p.ID])
	}
	return nil
}

// ---------- admin CRUD ----------

type OfferRepository struct {
	db *sql.DB
}

func NewOfferRepository(db *sql.DB) *OfferRepository { return &OfferRepository{db: db} }

const offerColumns = `
	o.id, o.name, o.name_ar, COALESCE(o.description, ''), COALESCE(o.description_ar, ''),
	o.discount_type, o.discount_value, o.scope, o.category, o.starts_at, o.ends_at, o.is_active,
	CASE WHEN NOT o.is_active THEN 'disabled'
	     WHEN o.starts_at > NOW() THEN 'scheduled'
	     WHEN o.ends_at <= NOW() THEN 'expired'
	     ELSE 'live' END,
	o.created_at, o.updated_at`

// offerStatusSQL maps a status filter to a fixed condition.
var offerStatusSQL = map[string]string{
	models.OfferLive:      "o.is_active AND o.starts_at <= NOW() AND o.ends_at > NOW()",
	models.OfferScheduled: "o.is_active AND o.starts_at > NOW()",
	models.OfferExpired:   "o.is_active AND o.ends_at <= NOW()",
	models.OfferDisabled:  "NOT o.is_active",
}

func IsValidOfferStatus(s string) bool {
	_, ok := offerStatusSQL[s]
	return ok
}

func scanOffer(s scanner) (models.Offer, error) {
	var o models.Offer
	err := s.Scan(&o.ID, &o.Name, &o.NameAr, &o.Description, &o.DescriptionAr,
		&o.DiscountType, &o.DiscountValue, &o.Scope, &o.Category, &o.StartsAt, &o.EndsAt, &o.IsActive,
		&o.Status, &o.CreatedAt, &o.UpdatedAt)
	o.ProductIDs = []int64{}
	return o, err
}

// List returns a page of offers. Status "" = all.
func (r *OfferRepository) List(ctx context.Context, f models.OfferFilter) (models.Page[models.Offer], error) {
	var w whereBuilder
	if cond, ok := offerStatusSQL[f.Status]; ok {
		w.addRaw(cond)
	}
	page := models.Page[models.Offer]{Items: []models.Offer{}, Limit: f.Limit, Offset: f.Offset}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM offers o`+w.sql(), w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+offerColumns+` FROM offers o`+w.sql()+
		` ORDER BY o.ends_at, o.id LIMIT `+w.next(f.Limit)+` OFFSET `+w.next(f.Offset), w.args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		o, err := scanOffer(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, o)
	}
	if err := rows.Err(); err != nil {
		return page, err
	}
	return page, r.loadProductIDs(ctx, page.Items)
}

func (r *OfferRepository) Get(ctx context.Context, id int64) (models.Offer, error) {
	o, err := scanOffer(r.db.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers o WHERE o.id = $1`, id))
	if err != nil {
		return o, apperr.FromDB(err, "offer not found")
	}
	list := []models.Offer{o}
	if err := r.loadProductIDs(ctx, list); err != nil {
		return o, err
	}
	return list[0], nil
}

func (r *OfferRepository) loadProductIDs(ctx context.Context, offers []models.Offer) error {
	if len(offers) == 0 {
		return nil
	}
	ids := make([]int64, len(offers))
	index := map[int64]int{}
	for i, o := range offers {
		ids[i] = o.ID
		index[o.ID] = i
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT offer_id, product_id FROM offer_products WHERE offer_id = ANY($1) ORDER BY product_id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var oid, pid int64
		if err := rows.Scan(&oid, &pid); err != nil {
			return err
		}
		offers[index[oid]].ProductIDs = append(offers[index[oid]].ProductIDs, pid)
	}
	return rows.Err()
}

// requireProducts fails with 404 naming the missing IDs.
func requireProducts(ctx context.Context, q database.Querier, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	rows, err := q.QueryContext(ctx, `SELECT id FROM products WHERE id = ANY($1)`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	found := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		found[id] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	var missing []int64
	for _, id := range ids {
		if !found[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return apperr.NotFound(fmt.Sprintf("products not found: %v", missing))
	}
	return nil
}

func (r *OfferRepository) Create(ctx context.Context, in models.OfferInput) (int64, error) {
	var id int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := requireProducts(ctx, tx, in.ProductIDs); err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO offers (name, name_ar, description, description_ar, discount_type, discount_value,
			                    scope, category, starts_at, ends_at, is_active)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7, $8, $9, $10, COALESCE($11, TRUE))
			RETURNING id`,
			in.Name, in.NameAr, in.Description, in.DescriptionAr, in.DiscountType, in.DiscountValue,
			in.Scope, in.Category, in.StartsAt, in.EndsAt, in.IsActive,
		).Scan(&id); err != nil {
			return err
		}
		return setOfferProducts(ctx, tx, id, in.ProductIDs)
	})
	return id, apperr.FromDB(err, "offer not found")
}

func (r *OfferRepository) Update(ctx context.Context, id int64, in models.OfferInput) error {
	return apperr.FromDB(database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := requireProducts(ctx, tx, in.ProductIDs); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE offers SET name = $1, name_ar = $2, description = NULLIF($3, ''), description_ar = NULLIF($4, ''),
			       discount_type = $5, discount_value = $6, scope = $7, category = $8,
			       starts_at = $9, ends_at = $10, is_active = COALESCE($11, is_active), updated_at = NOW()
			WHERE id = $12`,
			in.Name, in.NameAr, in.Description, in.DescriptionAr, in.DiscountType, in.DiscountValue,
			in.Scope, in.Category, in.StartsAt, in.EndsAt, in.IsActive, id)
		if err := expectOneRow(res, err, "offer not found"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM offer_products WHERE offer_id = $1`, id); err != nil {
			return err
		}
		return setOfferProducts(ctx, tx, id, in.ProductIDs)
	}), "offer not found")
}

func setOfferProducts(ctx context.Context, tx *sql.Tx, offerID int64, productIDs []int64) error {
	if len(productIDs) == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO offer_products (offer_id, product_id) SELECT $1, unnest($2::bigint[])`, offerID, productIDs)
	return err
}

// Delete removes an offer. Orders keep the offer's name (order_items.offer_name);
// their offer_id simply becomes NULL. To pause an offer instead, set is_active=false.
func (r *OfferRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM offers WHERE id = $1`, id)
	return expectOneRow(res, err, "offer not found")
}
