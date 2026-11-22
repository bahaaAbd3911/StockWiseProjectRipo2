package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestDiscountedPrice(t *testing.T) {
	cases := []struct{ price, discount, want string }{
		{"200.00", "0", "200.00"},
		{"200.00", "15", "170.00"},
		{"19.99", "33", "13.39"},   // 13.3933 rounds down
		{"10.01", "50", "5.01"},    // 5.005 rounds half away from zero
		{"99.99", "100", "0.00"},   // free
		{"49.95", "12.5", "43.71"}, // 43.70625
	}
	for _, c := range cases {
		got := DiscountedPrice(d(c.price), d(c.discount))
		if !got.Equal(d(c.want)) {
			t.Errorf("DiscountedPrice(%s, %s%%) = %s, want %s", c.price, c.discount, got, c.want)
		}
	}
}

func TestLineTotalIsExact(t *testing.T) {
	unit := DiscountedPrice(d("19.99"), d("33"))
	if got := LineTotal(unit, 3); !got.Equal(d("40.17")) {
		t.Fatalf("LineTotal = %s, want 40.17", got)
	}
}
