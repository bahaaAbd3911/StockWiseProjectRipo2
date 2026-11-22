// Package pricing holds the money math. It is the ONLY place a discount is
// applied, so the product page, the cart and checkout always agree.
//
// All amounts are shopspring/decimal values (exact base-10), never float64,
// matching PostgreSQL's numeric(10,2) columns.
package pricing

import "github.com/shopspring/decimal"

var hundred = decimal.NewFromInt(100)

// DiscountedPrice applies a percentage discount (0..100) and rounds to cents.
//
//	DiscountedPrice(200.00, 15) = 170.00
//	DiscountedPrice(19.99, 33)  = 13.39   (19.99 * 0.67 = 13.3933 -> 13.39)
func DiscountedPrice(price, discountPercent decimal.Decimal) decimal.Decimal {
	if discountPercent.LessThanOrEqual(decimal.Zero) {
		return price.Round(2)
	}
	if discountPercent.GreaterThanOrEqual(hundred) {
		return decimal.Zero.Round(2)
	}
	return price.Mul(hundred.Sub(discountPercent)).Div(hundred).Round(2)
}

// LineTotal is unitPrice * quantity. Because unitPrice already has two
// decimals the result is exact, which satisfies the database check
// order_items.subtotal = quantity * unit_price.
func LineTotal(unitPrice decimal.Decimal, quantity int) decimal.Decimal {
	return unitPrice.Mul(decimal.NewFromInt(int64(quantity)))
}
