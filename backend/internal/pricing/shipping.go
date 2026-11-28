package pricing

import (
	"strings"

	"github.com/shopspring/decimal"

	"ecommerce-backend/internal/models"
)

// ShippingFee is the only place the delivery fee is decided; the cart and
// checkout both use it, so they always agree.
//
//   - no shipping settings: free (stores that never set a fee keep the
//     behaviour from before shipping fees existed)
//   - itemsTotal (after the coupon) reaches FreeOver: free
//   - the city has its own fee: that fee
//   - otherwise: the standard fee
//
// An empty cart costs nothing to ship.
func ShippingFee(s *models.ShippingSettings, city string, itemsTotal decimal.Decimal) decimal.Decimal {
	if s == nil || !itemsTotal.IsPositive() {
		return decimal.Zero
	}
	if s.FreeOver != nil && itemsTotal.GreaterThanOrEqual(*s.FreeOver) {
		return decimal.Zero
	}
	city = strings.TrimSpace(city)
	for _, c := range s.Cities {
		if city != "" && strings.EqualFold(strings.TrimSpace(c.City), city) {
			return c.Fee
		}
	}
	return s.Fee
}
