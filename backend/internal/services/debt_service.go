package services

import (
	"context"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/notify"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// DebtService: customer credit limits and the debt ledger.
//
// How debt comes into existence:
//   - a customer checks out with payment_method=pay_later   -> "charge"
//   - an admin delivers a cash order with payment_collected=false -> "charge"
//
// How it goes away:
//   - an admin records money received   -> "payment"
//   - a charged order is cancelled      -> "reversal" (automatic)
//   - an admin corrects a mistake       -> "adjustment" (+ or -, note required)
//
// The customer is emailed when a payment or adjustment is recorded (and the
// notify worker sends periodic reminders while a balance is open).
type DebtService struct {
	debts  *repository.DebtRepository
	users  *repository.UserRepository
	outbox *notify.Outbox
	tpl    *notify.Branding // each store's name, links and currency
}

func NewDebtService(debts *repository.DebtRepository, users *repository.UserRepository,
	outbox *notify.Outbox, tpl *notify.Branding) *DebtService {
	return &DebtService{debts: debts, users: users, outbox: outbox, tpl: tpl}
}

// Statement returns the account summary and a page of its ledger.
func (s *DebtService) Statement(ctx context.Context, userID int64, limit, offset int) (models.DebtStatement, error) {
	acc, err := s.debts.Account(ctx, userID)
	if err != nil {
		return models.DebtStatement{}, err
	}
	txs, err := s.debts.Transactions(ctx, userID, limit, offset)
	if err != nil {
		return models.DebtStatement{}, err
	}
	return models.DebtStatement{Account: acc, Transactions: txs}, nil
}

func (s *DebtService) ListAccounts(ctx context.Context, f models.DebtAccountFilter) (models.Page[models.DebtAccount], error) {
	return s.debts.ListAccounts(ctx, f)
}

func (s *DebtService) SetCreditLimit(ctx context.Context, userID int64, req models.SetCreditLimitRequest) (models.DebtAccount, error) {
	if err := validators.ValidateCreditLimit(&req); err != nil {
		return models.DebtAccount{}, err
	}
	if err := s.debts.SetCreditLimit(ctx, userID, req.CreditLimit); err != nil {
		return models.DebtAccount{}, err
	}
	return s.debts.Account(ctx, userID)
}

func (s *DebtService) RecordPayment(ctx context.Context, adminID, userID int64, req models.RecordPaymentRequest) (models.DebtStatement, error) {
	if err := validators.ValidatePayment(&req); err != nil {
		return models.DebtStatement{}, err
	}
	if err := s.debts.RecordPayment(ctx, adminID, userID, req); err != nil {
		return models.DebtStatement{}, err
	}
	st, err := s.Statement(ctx, userID, 20, 0)
	if err != nil {
		return st, err
	}
	if to, ok := lookupRecipient(ctx, s.users, userID); ok {
		s.outbox.EnqueueOrLog(ctx, notify.KindDebtPayment, s.tpl.For(ctx).DebtPayment(to, req.Amount, st.Account.Balance))
	}
	return st, nil
}

func (s *DebtService) RecordAdjustment(ctx context.Context, adminID, userID int64, req models.RecordAdjustmentRequest) (models.DebtStatement, error) {
	if err := validators.ValidateAdjustment(&req); err != nil {
		return models.DebtStatement{}, err
	}
	if err := s.debts.RecordAdjustment(ctx, adminID, userID, req); err != nil {
		return models.DebtStatement{}, err
	}
	st, err := s.Statement(ctx, userID, 20, 0)
	if err != nil {
		return st, err
	}
	if to, ok := lookupRecipient(ctx, s.users, userID); ok {
		s.outbox.EnqueueOrLog(ctx, notify.KindDebtAdjustment, s.tpl.For(ctx).DebtAdjustment(to, req.Amount, req.Note, st.Account.Balance))
	}
	return st, nil
}
