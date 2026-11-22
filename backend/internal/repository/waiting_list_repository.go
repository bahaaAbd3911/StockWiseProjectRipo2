package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
)

type WaitingListRepository struct {
	db *sql.DB
}

func NewWaitingListRepository(db *sql.DB) *WaitingListRepository {
	return &WaitingListRepository{db: db}
}

const waitingColumns = `
	w.id, w.user_id, w.product_id, p.name, p.name_ar, p.stock, w.quantity, w.selected_color,
	w.status, w.requested_at, w.updated_at, w.notified_at`

const waitingFrom = ` FROM waiting_list_items w JOIN products p ON p.id = w.product_id`

func scanWaiting(s scanner, withCustomer bool) (models.WaitingListItem, error) {
	var x models.WaitingListItem
	dest := []any{&x.ID, &x.UserID, &x.ProductID, &x.ProductName, &x.ProductNameAr, &x.ProductStock,
		&x.Quantity, &x.SelectedColor, &x.Status, &x.RequestedAt, &x.UpdatedAt, &x.NotifiedAt}
	var c models.OrderCustomer
	if withCustomer {
		dest = append(dest, &c.FirstName, &c.LastName, &c.Email)
	}
	if err := s.Scan(dest...); err != nil {
		return x, err
	}
	if withCustomer {
		x.Customer = &c
	}
	return x, nil
}

// Create adds a request for an OUT-OF-STOCK product. Only one active
// (waiting/notified) request per user + product + color is allowed.
func (r *WaitingListRepository) Create(ctx context.Context, userID int64, req models.CreateWaitingListRequest) (models.WaitingListItem, error) {
	var id int64
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		// Two identical requests at the same moment would both pass the
		// "already exists?" check below. This lock (held until the transaction
		// ends) makes them run one after the other. The schema has no unique
		// index for this rule, so the lock is what enforces it.
		key := fmt.Sprintf("waiting-list:%d:%d", userID, req.ProductID)
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, key); err != nil {
			return err
		}

		var stock int
		if err := tx.QueryRowContext(ctx,
			`SELECT stock FROM products WHERE id = $1 AND is_active`, req.ProductID).Scan(&stock); err != nil {
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
		if stock > 0 {
			return apperr.Conflict("this product is in stock; add it to your cart instead")
		}

		var exists bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM waiting_list_items
				WHERE user_id = $1 AND product_id = $2
				  AND selected_color IS NOT DISTINCT FROM $3
				  AND status IN ('waiting', 'notified'))`,
			userID, req.ProductID, colorName).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return apperr.Conflict("you are already on the waiting list for this product")
		}

		return tx.QueryRowContext(ctx, `
			INSERT INTO waiting_list_items (user_id, product_id, quantity, selected_color)
			VALUES ($1, $2, $3, $4) RETURNING id`,
			userID, req.ProductID, req.Quantity, colorName).Scan(&id)
	})
	if err != nil {
		return models.WaitingListItem{}, err
	}
	return r.get(ctx, id)
}

func (r *WaitingListRepository) get(ctx context.Context, id int64) (models.WaitingListItem, error) {
	x, err := scanWaiting(r.db.QueryRowContext(ctx, `SELECT `+waitingColumns+waitingFrom+` WHERE w.id = $1`, id), false)
	return x, apperr.FromDB(err, "waiting-list item not found")
}

func (r *WaitingListRepository) ListForUser(ctx context.Context, userID int64) ([]models.WaitingListItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+waitingColumns+waitingFrom+` WHERE w.user_id = $1 ORDER BY w.requested_at DESC, w.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WaitingListItem{}
	for rows.Next() {
		x, err := scanWaiting(rows, false)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// CancelByUser cancels the user's own active request.
func (r *WaitingListRepository) CancelByUser(ctx context.Context, userID, id int64) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx,
			`SELECT status FROM waiting_list_items WHERE id = $1 AND user_id = $2 FOR UPDATE`, id, userID,
		).Scan(&status); err != nil {
			return apperr.FromDB(err, "waiting-list item not found")
		}
		if !models.CanTransitionWaiting(status, models.WaitingCancelled) {
			return apperr.Conflict(fmt.Sprintf("this request is already %s", status))
		}
		return setWaitingStatus(ctx, tx, id, models.WaitingCancelled)
	})
}

// List returns a page of all requests, oldest first (first come, first served).
func (r *WaitingListRepository) List(ctx context.Context, f models.WaitingListFilter) (models.Page[models.WaitingListItem], error) {
	var w whereBuilder
	if f.Status != "" {
		w.add("w.status = ?::text", f.Status)
	}
	if f.ProductID > 0 {
		w.add("w.product_id = ?", f.ProductID)
	}
	page := models.Page[models.WaitingListItem]{Items: []models.WaitingListItem{}, Limit: f.Limit, Offset: f.Offset}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM waiting_list_items w`+w.sql(), w.args...).Scan(&page.Total); err != nil {
		return page, err
	}
	query := `SELECT ` + waitingColumns + `, u.first_name, u.last_name, u.email` + waitingFrom +
		` JOIN users u ON u.id = w.user_id` + w.sql() +
		` ORDER BY w.requested_at, w.id LIMIT ` + w.next(f.Limit) + ` OFFSET ` + w.next(f.Offset)
	rows, err := r.db.QueryContext(ctx, query, w.args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		x, err := scanWaiting(rows, true)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, x)
	}
	return page, rows.Err()
}

// UpdateStatus changes a request's status (admin), following the allowed
// transitions in models.CanTransitionWaiting.
func (r *WaitingListRepository) UpdateStatus(ctx context.Context, id int64, to string) (models.WaitingListItem, error) {
	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var from string
		if err := tx.QueryRowContext(ctx,
			`SELECT status FROM waiting_list_items WHERE id = $1 FOR UPDATE`, id).Scan(&from); err != nil {
			return apperr.FromDB(err, "waiting-list item not found")
		}
		if !models.CanTransitionWaiting(from, to) {
			return apperr.Conflict(fmt.Sprintf("cannot change waiting-list status from %s to %s", from, to))
		}
		return setWaitingStatus(ctx, tx, id, to)
	})
	if err != nil {
		return models.WaitingListItem{}, err
	}
	return r.get(ctx, id)
}

// setWaitingStatus writes the status. notified_at is set the first time a
// request is marked "notified" and cleared if it goes back to "waiting".
func setWaitingStatus(ctx context.Context, tx *sql.Tx, id int64, status string) error {
	notifiedAt := "notified_at"
	switch status {
	case models.WaitingNotified:
		notifiedAt = "COALESCE(notified_at, NOW())"
	case models.WaitingWaiting:
		notifiedAt = "NULL"
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE waiting_list_items SET status = $1, notified_at = `+notifiedAt+`, updated_at = NOW() WHERE id = $2`,
		status, id)
	return err
}

// MarkProductNotified moves every "waiting" request for a product to
// "notified" (the product is back in stock) and returns the updated requests,
// so the customers can be emailed. Each request is notified only once.
func (r *WaitingListRepository) MarkProductNotified(ctx context.Context, productID int64) ([]models.WaitingListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		UPDATE waiting_list_items
		SET status = 'notified', notified_at = COALESCE(notified_at, NOW()), updated_at = NOW()
		WHERE product_id = $1 AND status = 'waiting'
		RETURNING id`, productID)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	items := make([]models.WaitingListItem, 0, len(ids))
	for _, id := range ids {
		item, err := r.get(ctx, id)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}