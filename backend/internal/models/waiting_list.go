package models

import (
	"slices"
	"time"
)

// Waiting-list statuses (waiting_list_items_status_check).
const (
	WaitingWaiting   = "waiting"
	WaitingNotified  = "notified"
	WaitingFulfilled = "fulfilled"
	WaitingCancelled = "cancelled"
)

// waitingTransitions: fulfilled and cancelled are final.
var waitingTransitions = map[string][]string{
	WaitingWaiting:  {WaitingNotified, WaitingFulfilled, WaitingCancelled},
	WaitingNotified: {WaitingWaiting, WaitingFulfilled, WaitingCancelled},
}

func IsWaitingStatus(s string) bool {
	switch s {
	case WaitingWaiting, WaitingNotified, WaitingFulfilled, WaitingCancelled:
		return true
	}
	return false
}

func CanTransitionWaiting(from, to string) bool {
	return slices.Contains(waitingTransitions[from], to)
}

type WaitingListItem struct {
	ID            int64          `json:"id"`
	UserID        int64          `json:"user_id"`
	ProductID     int64          `json:"product_id"`
	ProductName   string         `json:"product_name"`
	ProductNameAr string         `json:"product_name_ar"`
	ProductStock  int            `json:"product_stock"`
	Quantity      int            `json:"quantity"`
	SelectedColor *string        `json:"selected_color"`
	Status        string         `json:"status"`
	RequestedAt   time.Time      `json:"requested_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	NotifiedAt    *time.Time     `json:"notified_at"`
	Customer      *OrderCustomer `json:"customer,omitempty"` // admin views only
}

type CreateWaitingListRequest struct {
	ProductID     int64   `json:"product_id"`
	Quantity      int     `json:"quantity"`
	SelectedColor *string `json:"selected_color"`
}

type UpdateWaitingListStatusRequest struct {
	Status string `json:"status"`
}

type WaitingListFilter struct {
	Status    string
	ProductID int64
	Limit     int
	Offset    int
}
