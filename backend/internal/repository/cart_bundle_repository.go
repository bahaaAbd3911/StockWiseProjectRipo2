package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"
	"ecommerce-backend/internal/validators"
)

// cartBundleLine is one bundle in a cart with its live components.
type cartBundleLine struct {
	CartBundleID int64
	Quantity     int
	AddedAt      time.Time
	bundle       bundleRow
	comps        []bundleComponent
}

func (b cartBundleLine) problem() string { return bundleProblem(b.bundle.live, b.comps, b.Quantity) }

func (b cartBundleLine) toModel() models.CartBundle {
	items := make([]models.BundleItem, len(b.comps))
	for i, c := range b.comps {
		items[i] = c.BundleItem
	}
	cb := models.CartBundle{
		ID: b.CartBundleID, BundleID: b.bundle.ID, Name: b.bundle.Name, NameAr: b.bundle.NameAr,
		ImageURL: b.bundle.ImageURL, UnitPrice: b.bundle.Price, Quantity: b.Quantity,
		LineTotal: pricing.LineTotal(b.bundle.Price, b.Quantity), Items: items, AddedAt: b.AddedAt,
	}
	if p := b.problem(); p != "" {
		cb.UnavailableReason = &p
	} else {
		cb.IsAvailable = true
	}
	return cb
}

func loadCartBundleLines(ctx context.Context, q database.Querier, cartID int64) ([]cartBundleLine, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT cb.id, cb.quantity, cb.added_at, `+bundleColumns+`
		FROM cart_bundles cb JOIN bundles b ON b.id = cb.bundle_id
		WHERE cb.cart_id = $1
		ORDER BY cb.added_at DESC, cb.id DESC`, cartID)
	if err != nil {
		return nil, err
	}
	var lines []cartBundleLine
	for rows.Next() {
		var l cartBundleLine
		b, err := scanBundle(rows, &l.CartBundleID, &l.Quantity, &l.AddedAt)
		if err != nil {
			rows.Close()
			return nil, err
		}
		l.bundle = b
		lines = append(lines, l)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ids := make([]int64, len(lines))
	for i, l := range lines {
		ids[i] = l.bundle.ID
	}
	comps, err := loadBundleComponents(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	for i := range lines {
		lines[i].comps = comps[lines[i].bundle.ID]
	}
	return lines, nil
}

// AddBundle puts a bundle in the cart (or adds to its quantity).
func (r *CartRepository) AddBundle(ctx context.Context, userID int64, req models.AddCartBundleRequest) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		b, err := scanBundle(tx.QueryRowContext(ctx,
			`SELECT `+bundleColumns+` FROM bundles b WHERE b.id = $1 AND `+bundleLiveSQL, req.BundleID))
		if err != nil {
			return apperr.FromDB(err, "bundle not found")
		}
		cartID, _, err := ensureCart(ctx, tx, userID)
		if err != nil {
			return err
		}
		var lineID int64
		var existing int
		err = tx.QueryRowContext(ctx, `
			SELECT id, quantity FROM cart_bundles WHERE cart_id = $1 AND bundle_id = $2 FOR UPDATE`,
			cartID, req.BundleID).Scan(&lineID, &existing)
		exists := err == nil
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		newQty := existing + req.Quantity
		if newQty > validators.MaxQuantity {
			return apperr.Invalid("quantity per bundle cannot exceed 1000")
		}
		comps, err := loadBundleComponents(ctx, tx, []int64{b.ID})
		if err != nil {
			return err
		}
		if p := bundleProblem(b.live, comps[b.ID], newQty); p != "" {
			return apperr.Conflict(p)
		}
		if exists {
			_, err = tx.ExecContext(ctx, `UPDATE cart_bundles SET quantity = $1 WHERE id = $2`, newQty, lineID)
		} else {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO cart_bundles (cart_id, bundle_id, quantity) VALUES ($1, $2, $3)`, cartID, b.ID, newQty)
		}
		if err != nil {
			return apperr.FromDB(err, "bundle not found")
		}
		return touchCart(ctx, tx, cartID)
	})
}

// UpdateBundle sets the quantity of a bundle line in the user's cart.
func (r *CartRepository) UpdateBundle(ctx context.Context, userID, lineID int64, qty int) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var cartID int64
		b, err := scanBundle(tx.QueryRowContext(ctx, `
			SELECT cb.cart_id, `+bundleColumns+`
			FROM cart_bundles cb
			JOIN carts c ON c.id = cb.cart_id
			JOIN bundles b ON b.id = cb.bundle_id
			WHERE cb.id = $1 AND c.user_id = $2
			FOR UPDATE OF cb`, lineID, userID), &cartID)
		if err != nil {
			return apperr.FromDB(err, "cart bundle not found")
		}
		comps, err := loadBundleComponents(ctx, tx, []int64{b.ID})
		if err != nil {
			return err
		}
		if p := bundleProblem(b.live, comps[b.ID], qty); p != "" {
			return apperr.Conflict(p)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE cart_bundles SET quantity = $1 WHERE id = $2`, qty, lineID); err != nil {
			return err
		}
		return touchCart(ctx, tx, cartID)
	})
}

func (r *CartRepository) RemoveBundle(ctx context.Context, userID, lineID int64) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_bundles cb USING carts c
		WHERE cb.cart_id = c.id AND cb.id = $1 AND c.user_id = $2`, lineID, userID)
	return expectOneRow(res, err, "cart bundle not found")
}
