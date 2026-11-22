package pricing

import (
	"testing"

	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

func TestCouponDiscount(t *testing.T) {
	d := decimal.RequireFromString
	cap20 := d("20")
	cases := []struct {
		name, typ, value string
		max              *decimal.Decimal
		eligible, want   string
	}{
		{"10% rounds to cents", models.OfferPercentage, "10", nil, "89.97", "9"},
		{"percentage is capped", models.OfferPercentage, "50", &cap20, "100", "20"},
		{"cap above the discount does nothing", models.OfferPercentage, "10", &cap20, "100", "10"},
		{"100% takes everything", models.OfferPercentage, "100", nil, "45.50", "45.5"},
		{"fixed amount", models.OfferFixed, "15", nil, "60", "15"},
		{"fixed never exceeds the items", models.OfferFixed, "15", nil, "12", "12"},
		{"nothing eligible", models.OfferFixed, "15", nil, "0", "0"},
	}
	for _, c := range cases {
		got := CouponDiscount(c.typ, d(c.value), c.max, d(c.eligible))
		if !got.Equal(d(c.want)) {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}
