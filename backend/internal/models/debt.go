package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Debt transaction types. The sign of Amount is enforced by the database:
//
//	charge     > 0  an order was put on the customer's account
//	payment    < 0  the customer paid money back
//	reversal   < 0  a charged order was cancelled
//	adjustment ≠ 0  manual correction by an admin (note required)
//
// The customer's balance (what they owe) is simply the SUM of all amounts.
const (
	DebtCharge     = "charge"
	DebtPayment    = "payment"
	DebtReversal   = "reversal"
	DebtAdjustment = "adjustment"
)

// DebtAccount summarises one customer's credit.
type DebtAccount struct {
	UserID            int64           `json:"user_id"`
	FirstName         string          `json:"first_name"`
	LastName          string          `json:"last_name"`
	Email             string          `json:"email"`
	Phone             *string         `json:"phone"`
	CreditLimit       decimal.Decimal `json:"credit_limit"`
	Balance           decimal.Decimal `json:"balance"`          // amount owed right now
	AvailableCredit   decimal.Decimal `json:"available_credit"` // credit_limit - balance (never below 0)
	PayLaterAllowed   bool            `json:"pay_later_allowed"`
	LastTransactionAt *time.Time      `json:"last_transaction_at"`
}

// DebtTransaction is one ledger row. BalanceAfter is the running balance.
type DebtTransaction struct {
	ID           int64           `json:"id"`
	UserID       int64           `json:"user_id"`
	OrderID      *int64          `json:"order_id"`
	Type         string          `json:"type"`
	Amount       decimal.Decimal `json:"amount"`
	BalanceAfter decimal.Decimal `json:"balance_after"`
	Note         *string         `json:"note"`
	CreatedBy    *int64          `json:"created_by"` // admin who recorded it; null = automatic
	CreatedAt    time.Time       `json:"created_at"`
}

// DebtStatement is an account plus a page of its transactions (newest first).
type DebtStatement struct {
	Account      DebtAccount           `json:"account"`
	Transactions Page[DebtTransaction] `json:"transactions"`
}

type SetCreditLimitRequest struct {
	CreditLimit decimal.Decimal `json:"credit_limit"`
}

// RecordPaymentRequest: Amount is positive (what the customer handed over).
type RecordPaymentRequest struct {
	Amount  decimal.Decimal `json:"amount"`
	Note    *string         `json:"note"`
	OrderID *int64          `json:"order_id"`
}

// RecordAdjustmentRequest: Amount is signed (+ increases the debt, - reduces it).
type RecordAdjustmentRequest struct {
	Amount  decimal.Decimal `json:"amount"`
	Note    string          `json:"note"`
	OrderID *int64          `json:"order_id"`
}

type DebtAccountFilter struct {
	OutstandingOnly bool // only customers who owe something
	Limit           int
	Offset          int
}
