package models

import (
	"slices"
	"time"

	"github.com/shopspring/decimal"
)

// Order statuses (must match the orders_status_check constraint).
const (
	OrderPending    = "pending"
	OrderConfirmed  = "confirmed"
	OrderProcessing = "processing"
	OrderShipped    = "shipped"
	OrderDelivered  = "delivered"
	OrderCancelled  = "cancelled"
)

// Payment statuses (orders_payment_status_check).
const (
	PaymentPending = "pending"
	PaymentPaid    = "paid"
	PaymentFailed  = "failed"
	// PaymentDeferred: the amount was moved to the customer's debt account
	// (pay-later orders, or cash orders delivered without collecting payment).
	PaymentDeferred = "deferred"
)

// Payment methods (check_payment_method).
const (
	PaymentCashOnDelivery = "cash_on_delivery"
	PaymentPayLater       = "pay_later" // charged to the customer's debt account
)

func IsPaymentMethod(s string) bool {
	return s == PaymentCashOnDelivery || s == PaymentPayLater
}

// orderTransitions is the order life cycle. An order can only move along these
// arrows; delivered and cancelled are final.
//
//	pending ──► confirmed ──► processing ──► shipped ──► delivered
//	   │            │              │
//	   └────────────┴──────────────┴──► cancelled
var orderTransitions = map[string][]string{
	OrderPending:    {OrderConfirmed, OrderCancelled},
	OrderConfirmed:  {OrderProcessing, OrderCancelled},
	OrderProcessing: {OrderShipped, OrderCancelled},
	OrderShipped:    {OrderDelivered},
}

// IsOrderStatus reports whether s is a known order status.
func IsOrderStatus(s string) bool {
	switch s {
	case OrderPending, OrderConfirmed, OrderProcessing, OrderShipped, OrderDelivered, OrderCancelled:
		return true
	}
	return false
}

// CanTransitionOrder reports whether an order may move from one status to another.
func CanTransitionOrder(from, to string) bool {
	return slices.Contains(orderTransitions[from], to)
}

// Order stores a SNAPSHOT of the delivery address and of every item's name and
// price at the moment of checkout, so history never changes when the address
// or the product is edited later.
type Order struct {
	ID                     int64           `json:"id"`
	UserID                 int64           `json:"user_id"`
	AddressID              *int64          `json:"address_id"`
	RecipientName          string          `json:"recipient_name"`
	Phone                  string          `json:"phone"`
	City                   string          `json:"city"`
	Area                   string          `json:"area"`
	Street                 string          `json:"street"`
	BuildingNumber         *string         `json:"building_number"`
	ApartmentNumber        *string         `json:"apartment_number"`
	AdditionalInstructions *string         `json:"additional_instructions"`
	SubtotalAmount         decimal.Decimal `json:"subtotal_amount"` // items after offers, before the coupon
	DiscountAmount         decimal.Decimal `json:"discount_amount"` // what the coupon took off
	CouponCode             *string         `json:"coupon_code"`     // copy of the code used, if any
	TotalAmount            decimal.Decimal `json:"total_amount"`    // subtotal - discount: what the customer pays
	Status                 string          `json:"status"`
	PaymentMethod          string          `json:"payment_method"`
	PaymentStatus          string          `json:"payment_status"`
	Items                  []OrderItem     `json:"items"`
	Bundles                []OrderBundle   `json:"bundles"`
	Customer               *OrderCustomer  `json:"customer,omitempty"` // admin views only
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

type OrderItem struct {
	ID              int64           `json:"id"`
	OrderID         int64           `json:"order_id"`
	ProductID       *int64          `json:"product_id"` // NULL if the product was hard-deleted
	ProductName     string          `json:"product_name"`
	ProductNameAr   *string         `json:"product_name_ar"`
	Quantity        int             `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	SelectedColor   *string         `json:"selected_color"`
	SelectedColorAr *string         `json:"selected_color_ar"`
	Subtotal        decimal.Decimal `json:"subtotal"`
	OfferID         *int64          `json:"offer_id"`   // offer that set unit_price, if any
	OfferName       *string         `json:"offer_name"` // snapshot of the offer name
	// ProductImageURL is the product's current main image (not a snapshot;
	// null if the product or its images were removed).
	ProductImageURL *string `json:"product_image_url"`
}

type OrderCustomer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// CreateOrderRequest: if AddressID is omitted, the default address is used.
// PaymentMethod defaults to cash_on_delivery; pay_later charges the order to
// the customer's debt account (needs a credit limit set by an admin).
type CreateOrderRequest struct {
	AddressID     *int64 `json:"address_id"`
	PaymentMethod string `json:"payment_method"`
}

// UpdateOrderStatusRequest: PaymentCollected only matters when a cash order is
// marked delivered. true/omitted = the cash was collected (payment "paid");
// false = the customer did not pay, so the amount is moved to their debt.
type UpdateOrderStatusRequest struct {
	Status           string `json:"status"`
	PaymentCollected *bool  `json:"payment_collected"`
}

type OrderFilter struct {
	Status        string
	PaymentMethod string
	Limit         int
	Offset        int
}
