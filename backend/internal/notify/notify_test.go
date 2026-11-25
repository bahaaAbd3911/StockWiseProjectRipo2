package notify

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

func TestBackoffGrowsAndIsCapped(t *testing.T) {
	want := []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute}
	for i, w := range want {
		if got := Backoff(i + 1); got != w {
			t.Errorf("attempt %d: got %s, want %s", i+1, got, w)
		}
	}
	if Backoff(50) != 6*time.Hour {
		t.Fatalf("backoff must stop at 6h, got %s", Backoff(50))
	}
}

func TestOrderEmails(t *testing.T) {
	tpl := Templates{Store: "Shop", FrontendURL: "https://shop.example.com/", Currency: "JOD"}
	to := Recipient{Email: "a@example.com", FirstName: "Alia"}
	color := "Black"
	o := models.Order{
		ID: 42, Status: models.OrderPending, PaymentMethod: models.PaymentPayLater,
		TotalAmount: decimal.RequireFromString("99.5"), City: "Amman", Area: "Abdoun", Street: "Main St",
		Items: []models.OrderItem{{ProductName: "Laptop", Quantity: 1, Subtotal: decimal.RequireFromString("99.5"), SelectedColor: &color}},
	}
	m := tpl.OrderPlaced(to, o)
	for _, want := range []string{"#42", "Laptop (Black)", "99.50 JOD", "pay later", "https://shop.example.com/orders/42", "مرحباً Alia"} {
		if !strings.Contains(m.Subject+m.Text, want) {
			t.Errorf("order email should contain %q:\n%s\n%s", want, m.Subject, m.Text)
		}
	}

	o.Status = models.OrderPending
	if _, ok := tpl.OrderStatus(to, o); ok {
		t.Fatal("no email for 'pending' (the order-received email covers it)")
	}
	o.Status = models.OrderCancelled
	m, ok := tpl.OrderStatus(to, o)
	if !ok || !strings.Contains(m.Text, "cancelled") || !strings.Contains(m.Text, "balance has been removed") {
		t.Fatalf("cancelled pay-later order should mention the balance: %s", m.Text)
	}
}

func TestAddQuery(t *testing.T) {
	got, err := AddQuery("https://shop.example.com/verify-email?lang=ar", "token", "a+b/c")
	if err != nil || got != "https://shop.example.com/verify-email?lang=ar&token=a%2Bb%2Fc" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestMaxAttemptsConstantsAgree(t *testing.T) {
	if maxAttemptsSQL != strconv.Itoa(maxAttempts) {
		t.Fatalf("maxAttemptsSQL (%s) must equal maxAttempts (%d)", maxAttemptsSQL, maxAttempts)
	}
}
