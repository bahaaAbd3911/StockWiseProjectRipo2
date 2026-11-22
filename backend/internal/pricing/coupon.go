package pricing

import (
	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

// CouponDiscount is what a coupon takes off the eligible part of an order
// (the lines the coupon applies to, at their prices AFTER offers):
//
//	percentage: eligible * value / 100, rounded to cents, at most maxDiscount
//	fixed:      value, but never more than the eligible amount
//
// The discount never exceeds the eligible amount, so a total can never go
// negative.
//
//	CouponDiscount(percentage, 10, nil, 89.97)  = 9.00   (8.997 -> 9.00)
//	CouponDiscount(percentage, 50, 20, 100.00)  = 20.00  (capped)
//	CouponDiscount(fixed, 15, nil, 12.00)       = 12.00  (not more than the items)
func CouponDiscount(discountType string, value decimal.Decimal, maxDiscount *decimal.Decimal, eligible decimal.Decimal) decimal.Decimal {
	if !eligible.IsPositive() || !value.IsPositive() {
		return decimal.Zero
	}
	var d decimal.Decimal
	if discountType == models.OfferPercentage {
		d = eligible.Mul(value).Div(hundred).Round(2)
		if maxDiscount != nil && d.GreaterThan(*maxDiscount) {
			d = *maxDiscount
		}
	} else {
		d = value
	}
	if d.GreaterThan(eligible) {
		d = eligible
	}
	return d.Round(2)
}
