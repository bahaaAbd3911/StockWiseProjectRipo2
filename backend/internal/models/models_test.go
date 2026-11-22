package models

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestOrderTransitions(t *testing.T) {
	allowed := [][2]string{
		{OrderPending, OrderConfirmed}, {OrderConfirmed, OrderProcessing},
		{OrderProcessing, OrderShipped}, {OrderShipped, OrderDelivered},
		{OrderPending, OrderCancelled}, {OrderConfirmed, OrderCancelled}, {OrderProcessing, OrderCancelled},
	}
	for _, tr := range allowed {
		if !CanTransitionOrder(tr[0], tr[1]) {
			t.Errorf("%s -> %s should be allowed", tr[0], tr[1])
		}
	}
	forbidden := [][2]string{
		{OrderPending, OrderShipped},     // skipping steps
		{OrderDelivered, OrderShipped},   // going backwards
		{OrderShipped, OrderCancelled},   // already on the way
		{OrderCancelled, OrderConfirmed}, // final
		{OrderDelivered, OrderCancelled}, // final
	}
	for _, tr := range forbidden {
		if CanTransitionOrder(tr[0], tr[1]) {
			t.Errorf("%s -> %s should be forbidden", tr[0], tr[1])
		}
	}
}

func TestWaitingTransitions(t *testing.T) {
	if !CanTransitionWaiting(WaitingWaiting, WaitingNotified) || !CanTransitionWaiting(WaitingNotified, WaitingFulfilled) {
		t.Error("normal waiting-list flow should be allowed")
	}
	if CanTransitionWaiting(WaitingFulfilled, WaitingWaiting) || CanTransitionWaiting(WaitingCancelled, WaitingNotified) {
		t.Error("fulfilled and cancelled must be final")
	}
}

func TestMoneyIsAJSONNumber(t *testing.T) {
	b, _ := json.Marshal(struct {
		Price decimal.Decimal `json:"price"`
	}{decimal.RequireFromString("19.90")})
	if !strings.Contains(string(b), `"price":19.9`) {
		t.Fatalf("money should be encoded as a JSON number, got %s", b)
	}
}
