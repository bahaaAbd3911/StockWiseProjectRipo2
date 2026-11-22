package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Dashboard is the admin overview. "Revenue" is the value of orders placed
// (total_amount, after coupons) that were not cancelled - booked sales, whether
// already paid, owed (pay later) or still to be collected on delivery.
type Dashboard struct {
	Timezone    string    `json:"timezone"`
	GeneratedAt time.Time `json:"generated_at"`

	Today          SalesFigures `json:"today"`
	Period         SalesFigures `json:"period"`          // the last `days` days, today included
	PreviousPeriod SalesFigures `json:"previous_period"` // the same number of days just before, for comparison

	SalesByDay     []DaySales             `json:"sales_by_day"`     // every day of the period, also days without orders
	PaymentMethods map[string]MethodSales `json:"payment_methods"`  // in the period
	OrdersByStatus map[string]int         `json:"orders_by_status"` // all orders right now, e.g. how many wait to be confirmed
	BestSellers    []BestSeller           `json:"best_sellers"`     // top 10 in the period
	LowStock       LowStockReport         `json:"low_stock"`
	Debts          DebtTotals             `json:"debts"`
	WaitingList    []WaitingDemand        `json:"waiting_list"` // most-requested out-of-stock products
	TopCoupons     []CouponUsage          `json:"top_coupons"`  // in the period
	Customers      CustomerFigures        `json:"customers"`
}

type SalesFigures struct {
	From              time.Time       `json:"from"`
	To                time.Time       `json:"to"` // exclusive
	Orders            int             `json:"orders"`
	Revenue           decimal.Decimal `json:"revenue"`
	AverageOrderValue decimal.Decimal `json:"average_order_value"`
	ItemsSold         int             `json:"items_sold"`      // units, bundle contents included
	DiscountsGiven    decimal.Decimal `json:"discounts_given"` // coupon discounts
	CancelledOrders   int             `json:"cancelled_orders"`
	NewCustomers      int             `json:"new_customers"`
}

type DaySales struct {
	Date    string          `json:"date"` // YYYY-MM-DD in the store's time zone
	Orders  int             `json:"orders"`
	Revenue decimal.Decimal `json:"revenue"`
}

type MethodSales struct {
	Orders  int             `json:"orders"`
	Revenue decimal.Decimal `json:"revenue"`
}

type BestSeller struct {
	ProductID   *int64          `json:"product_id"` // null if the product was deleted
	Name        string          `json:"name"`
	NameAr      *string         `json:"name_ar"`
	Units       int             `json:"units"`        // sold alone + inside bundles
	SingleSales decimal.Decimal `json:"single_sales"` // money from units sold alone (bundles have one price for all their products)
}

type LowStockReport struct {
	Threshold  int               `json:"threshold"`
	OutOfStock int               `json:"out_of_stock"` // active products with stock 0
	Products   []LowStockProduct `json:"products"`     // stock <= threshold, lowest first (max 20)
}

type LowStockProduct struct {
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	NameAr    string `json:"name_ar"`
	Stock     int    `json:"stock"`
	Waiting   int    `json:"waiting"` // customers on its waiting list
}

type DebtTotals struct {
	Outstanding      decimal.Decimal `json:"outstanding"` // total owed by all customers
	CustomersOwing   int             `json:"customers_owing"`
	OverCreditLimit  int             `json:"over_credit_limit"` // customers owing more than their limit
	PaymentsInPeriod decimal.Decimal `json:"payments_in_period"`
}

type WaitingDemand struct {
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	NameAr    string `json:"name_ar"`
	Requests  int    `json:"requests"`
	Units     int    `json:"units"` // total quantity asked for
}

type CouponUsage struct {
	Code     string          `json:"code"`
	Uses     int             `json:"uses"`
	Discount decimal.Decimal `json:"discount"`
}

type CustomerFigures struct {
	Total           int `json:"total"`      // active customer accounts
	Unverified      int `json:"unverified"` // email not confirmed yet
	NewInPeriod     int `json:"new_in_period"`
	OrderedInPeriod int `json:"ordered_in_period"` // distinct customers with an order in the period
}
