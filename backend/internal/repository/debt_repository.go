package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

// DebtRepository manages customer credit limits and the debt ledger.
//
// The ledger (customer_debt_transactions) is append-only: rows are never
// edited or deleted. A mistake is fixed with an "adjustment" row, so the full
// history of what happened stays visible.
type DebtRepository struct {
	db *sql.DB
}

func NewDebtRepository(db *sql.DB) *DebtRepository { return &DebtRepository{db: db} }

// ---------- helpers shared with the order repository ----------

// debtBalance is what the user owes right now (SUM of the ledger).
func debtBalance(ctx context.Context, q database.Querier, userID int64) (decimal.Decimal, error) {
	var b decimal.Decimal
	err := q.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM customer_debt_transactions WHERE user_id = $1`, userID).Scan(&b)
	return b, err
}

// insertDebt appends one ledger row. createdBy nil = automatic (checkout/cancel).
func insertDebt(ctx context.Context, tx *sql.Tx, userID int64, orderID *int64, typ string,
	amount decimal.Decimal, note *string, createdBy *int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO customer_debt_transactions (user_id, order_id, type, amount, note, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)`, userID, orderID, typ, amount, note, createdBy)
	return err
}

// reverseOrderDebt cancels whatever is still charged for an order (charges
// minus earlier reversals). Payments linked to the order are NOT reversed:
// that money was really received, so it stays as credit on the account.
func reverseOrderDebt(ctx context.Context, tx *sql.Tx, userID, orderID int64) error {
	var net decimal.Decimal
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM customer_debt_transactions
		WHERE order_id = $1 AND type IN ('charge', 'reversal')`, orderID).Scan(&net); err != nil {
		return err
	}
	if !net.IsPositive() {
		return nil
	}
	note := fmt.Sprintf("order #%d cancelled", orderID)
	return insertDebt(ctx, tx, userID, &orderID, models.DebtReversal, net.Neg(), &note, nil)
}

// ---------- accounts ----------

const accountColumns = `
	u.id, u.first_name, u.last_name, u.email, u.phone, u.credit_limit,
	COALESCE((SELECT SUM(t.amount) FROM customer_debt_transactions t WHERE t.user_id = u.id), 0) AS balance,
	(SELECT MAX(t.created_at) FROM customer_debt_transactions t WHERE t.user_id = u.id) AS last_transaction_at`

func scanAccount(s scanner) (models.DebtAccount, error) {
	var a models.DebtAccount
	if err := s.Scan(&a.UserID, &a.FirstName, &a.LastName, &a.Email, &a.Phone, &a.CreditLimit,
		&a.Balance, &a.LastTransactionAt); err != nil {
		return a, err
	}
	a.AvailableCredit = a.CreditLimit.Sub(a.Balance)
	if a.AvailableCredit.IsNegative() {
		a.AvailableCredit = decimal.Zero
	}
	a.PayLaterAllowed = a.CreditLimit.IsPositive() && a.AvailableCredit.IsPositive()
	return a, nil
}

func (r *DebtRepository) Account(ctx context.Context, userID int64) (models.DebtAccount, error) {
	a, err := scanAccount(r.db.QueryRowContext(ctx, `SELECT `+accountColumns+` FROM users u WHERE u.id = $1`, userID))
	return a, apperr.FromDB(err, "customer not found")
}

// ListAccounts lists customers with a credit limit or any ledger history,
// biggest debt first.
func (r *DebtRepository) ListAccounts(ctx context.Context, f models.DebtAccountFilter) (models.Page[models.DebtAccount], error) {
	base := `FROM (SELECT ` + accountColumns + ` FROM users u
		WHERE u.credit_limit > 0 OR EXISTS (SELECT 1 FROM customer_debt_transactions t WHERE t.user_id = u.id)) a`
	filter := ""
	if f.OutstandingOnly {
		filter = " WHERE a.balance > 0"
	}
	page := models.Page[models.DebtAccount]{Items: []models.DebtAccount{}, Limit: f.Limit, Offset: f.Offset}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) `+base+filter).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT a.* `+base+filter+
		` ORDER BY a.balance DESC, a.id LIMIT $1 OFFSET $2`, f.Limit, f.Offset)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, a)
	}
	return page, rows.Err()
}

// Transactions returns the ledger newest first, each row with the running
// balance right after it.
func (r *DebtRepository) Transactions(ctx context.Context, userID int64, limit, offset int) (models.Page[models.DebtTransaction], error) {
	page := models.Page[models.DebtTransaction]{Items: []models.DebtTransaction{}, Limit: limit, Offset: offset}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM customer_debt_transactions WHERE user_id = $1`, userID).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, order_id, type, amount, balance_after, note, created_by, created_at FROM (
			SELECT t.*, SUM(t.amount) OVER (ORDER BY t.created_at, t.id) AS balance_after
			FROM customer_debt_transactions t WHERE t.user_id = $1
		) x
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var t models.DebtTransaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.OrderID, &t.Type, &t.Amount, &t.BalanceAfter,
			&t.Note, &t.CreatedBy, &t.CreatedAt); err != nil {
			return page, err
		}
		page.Items = append(page.Items, t)
	}
	return page, rows.Err()
}

// ---------- admin actions ----------

// SetCreditLimit changes how much a customer may owe. Lowering it below the
// current balance is allowed: it only blocks new pay-later orders.
func (r *DebtRepository) SetCreditLimit(ctx context.Context, userID int64, limit decimal.Decimal) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET credit_limit = $1, updated_at = NOW() WHERE id = $2`, limit, userID)
	return expectOneRow(res, err, "customer not found")
}

// requireOrderOf checks that an order belongs to the customer.
func requireOrderOf(ctx context.Context, q database.Querier, userID int64, orderID *int64) error {
	if orderID == nil {
		return nil
	}
	var ok bool
	if err := q.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM orders WHERE id = $1 AND user_id = $2)`, *orderID, userID).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return apperr.NotFound("order not found for this customer")
	}
	return nil
}

// RecordPayment stores money received. A payment cannot be larger than the
// debt (the shop does not hold customer credit).
func (r *DebtRepository) RecordPayment(ctx context.Context, adminID, userID int64, req models.RecordPaymentRequest) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := lockUser(ctx, tx, userID); err != nil {
			return apperr.NotFound("customer not found")
		}
		if err := requireOrderOf(ctx, tx, userID, req.OrderID); err != nil {
			return err
		}
		balance, err := debtBalance(ctx, tx, userID)
		if err != nil {
			return err
		}
		if !balance.IsPositive() {
			return apperr.Conflict("this customer has no outstanding debt")
		}
		if req.Amount.GreaterThan(balance) {
			return apperr.Conflict(fmt.Sprintf("payment of %s is more than the outstanding balance of %s",
				req.Amount.StringFixed(2), balance.StringFixed(2)))
		}
		return insertDebt(ctx, tx, userID, req.OrderID, models.DebtPayment, req.Amount.Neg(), req.Note, &adminID)
	})
}

// RecordAdjustment adds a manual correction (positive or negative). The
// balance may not go below zero.
func (r *DebtRepository) RecordAdjustment(ctx context.Context, adminID, userID int64, req models.RecordAdjustmentRequest) error {
	return database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := lockUser(ctx, tx, userID); err != nil {
			return apperr.NotFound("customer not found")
		}
		if err := requireOrderOf(ctx, tx, userID, req.OrderID); err != nil {
			return err
		}
		balance, err := debtBalance(ctx, tx, userID)
		if err != nil {
			return err
		}
		if balance.Add(req.Amount).IsNegative() {
			return apperr.Conflict(fmt.Sprintf("this adjustment would make the balance negative (current balance %s)",
				balance.StringFixed(2)))
		}
		note := req.Note
		return insertDebt(ctx, tx, userID, req.OrderID, models.DebtAdjustment, req.Amount, &note, &adminID)
	})
}
