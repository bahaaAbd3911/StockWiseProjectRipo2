package services

import (
	"context"
	"log/slog"
	"strings"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/notify"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// WaitingListService: customers ask to be told when an out-of-stock product
// comes back. They are emailed ("back in stock") automatically when an admin
// gives the product stock again, or when an admin marks a request "notified".
type WaitingListService struct {
	items  *repository.WaitingListRepository
	users  *repository.UserRepository
	outbox *notify.Outbox
	tpl    notify.Templates
}

func NewWaitingListService(items *repository.WaitingListRepository, users *repository.UserRepository,
	outbox *notify.Outbox, tpl notify.Templates) *WaitingListService {
	return &WaitingListService{items: items, users: users, outbox: outbox, tpl: tpl}
}

// NotifyRestock is called after a product gets stock: every waiting customer
// is marked notified and emailed. Errors are logged; they never fail the
// product update that triggered them.
func (s *WaitingListService) NotifyRestock(ctx context.Context, productID int64) {
	items, err := s.items.MarkProductNotified(context.WithoutCancel(ctx), productID)
	if err != nil {
		slog.Error("waiting list: mark notified", "product_id", productID, "error", err)
	}
	for _, it := range items {
		s.backInStockEmail(ctx, it)
	}
}

func (s *WaitingListService) backInStockEmail(ctx context.Context, it models.WaitingListItem) {
	if to, ok := lookupRecipient(ctx, s.users, it.UserID); ok {
		s.outbox.EnqueueOrLog(ctx, notify.KindBackInStock, s.tpl.BackInStock(to, it.ProductID, it.ProductName, it.ProductNameAr))
	}
}

func (s *WaitingListService) Join(ctx context.Context, userID int64, req models.CreateWaitingListRequest) (models.WaitingListItem, error) {
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	if err := validators.ValidateWaitingList(&req); err != nil {
		return models.WaitingListItem{}, err
	}
	return s.items.Create(ctx, userID, req)
}

func (s *WaitingListService) ListMine(ctx context.Context, userID int64) ([]models.WaitingListItem, error) {
	return s.items.ListForUser(ctx, userID)
}

func (s *WaitingListService) CancelMine(ctx context.Context, userID, id int64) error {
	return s.items.CancelByUser(ctx, userID, id)
}

func (s *WaitingListService) List(ctx context.Context, f models.WaitingListFilter) (models.Page[models.WaitingListItem], error) {
	f.Status = strings.TrimSpace(f.Status)
	if f.Status != "" && !models.IsWaitingStatus(f.Status) {
		return models.Page[models.WaitingListItem]{}, validators.FieldError("status", "is not a valid waiting-list status")
	}
	return s.items.List(ctx, f)
}

func (s *WaitingListService) UpdateStatus(ctx context.Context, id int64, status string) (models.WaitingListItem, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if !models.IsWaitingStatus(status) {
		return models.WaitingListItem{}, validators.FieldError("status",
			"must be one of: waiting, notified, fulfilled, cancelled")
	}
	item, err := s.items.UpdateStatus(ctx, id, status)
	if err != nil {
		return item, err
	}
	if status == models.WaitingNotified {
		s.backInStockEmail(ctx, item)
	}
	return item, nil
}
