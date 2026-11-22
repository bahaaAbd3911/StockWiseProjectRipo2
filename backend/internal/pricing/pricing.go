// Package pricing holds the money math. It is the ONLY place a discount is
// applied, so the product page, the cart and checkout always agree.
//
// All amounts are shopspring/decimal values (exact base-10), never float64,
// matching PostgreSQL's numeric(10,2) columns.
package pricing

import (
	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

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

// OfferPrice is the price after one offer:
//
//	percentage: price * (100 - value) / 100   (same rounding as DiscountedPrice)
//	fixed:      price - value, never below 0
func OfferPrice(price decimal.Decimal, discountType string, value decimal.Decimal) decimal.Decimal {
	if discountType == models.OfferPercentage {
		return DiscountedPrice(price, value)
	}
	p := price.Sub(value)
	if p.IsNegative() {
		p = decimal.Zero
	}
	return p.Round(2)
}

// BestPrice returns the lowest price a customer can get for a product right
// now: the product's own discount, or the best of its live offers. Discounts
// never stack - only the single best one is used. When an offer wins, it is
// returned so it can be shown ("Eid sale, ends Friday") and saved on the order.
// On a tie the product's own discount wins (no offer label).
func BestPrice(price, productDiscount decimal.Decimal, offers []models.AppliedOffer) (decimal.Decimal, *models.AppliedOffer) {
	best := DiscountedPrice(price, productDiscount)
	var applied *models.AppliedOffer
	for i := range offers {
		p := OfferPrice(price, offers[i].DiscountType, offers[i].DiscountValue)
		if p.LessThan(best) {
			best = p
			o := offers[i]
			applied = &o
		}
	}
	return best, applied
}
