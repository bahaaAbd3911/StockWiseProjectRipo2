package models

import (
	"time"

	"github.com/shopspring/decimal"
)

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
	TotalAmount            decimal.Decimal `json:"total_amount"`
	Status                 string          `json:"status"`
	PaymentMethod          string          `json:"payment_method"`
	PaymentStatus          string          `json:"payment_status"`
	Items                  []OrderItem     `json:"items"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

type OrderItem struct {
	ID              int64           `json:"id"`
	OrderID         int64           `json:"order_id"`
	ProductID       *int64          `json:"product_id"`
	ProductName     string          `json:"product_name"`
	ProductNameAr   *string         `json:"product_name_ar"`
	Quantity        int             `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	SelectedColor   *string         `json:"selected_color"`
	SelectedColorAr *string         `json:"selected_color_ar"`
	Subtotal        decimal.Decimal `json:"subtotal"`
}

type CreateOrderRequest struct {
	AddressID int64 `json:"address_id"`
}
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}
