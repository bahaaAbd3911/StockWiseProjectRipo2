package services

import (
	"context"
	"strings"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// WaitingListService: customers ask to be told when an out-of-stock product
// comes back; admins work through the list. Marking a request "notified" only
// records the time - sending the actual email/SMS needs a provider that is not
// part of this project yet.
type WaitingListService struct {
	items *repository.WaitingListRepository
}

func NewWaitingListService(items *repository.WaitingListRepository) *WaitingListService {
	return &WaitingListService{items: items}
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
	return s.items.UpdateStatus(ctx, id, status)
}
