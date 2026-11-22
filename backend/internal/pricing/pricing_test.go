package pricing

import (
	"testing"

	"ecommerce-backend/internal/models"

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

func TestOfferPrice(t *testing.T) {
	if got := OfferPrice(d("100"), models.OfferFixed, d("15.50")); !got.Equal(d("84.50")) {
		t.Errorf("fixed offer: got %s", got)
	}
	if got := OfferPrice(d("10"), models.OfferFixed, d("25")); !got.Equal(d("0")) {
		t.Errorf("fixed offer must not go below zero: got %s", got)
	}
	if got := OfferPrice(d("80"), models.OfferPercentage, d("25")); !got.Equal(d("60")) {
		t.Errorf("percentage offer: got %s", got)
	}
}

func TestBestPricePicksTheSingleBestDiscount(t *testing.T) {
	offers := []models.AppliedOffer{
		{ID: 1, DiscountType: models.OfferPercentage, DiscountValue: d("20")}, // 80
		{ID: 2, DiscountType: models.OfferFixed, DiscountValue: d("30")},      // 70  <- best
		{ID: 3, DiscountType: models.OfferPercentage, DiscountValue: d("5")},  // 95
	}
	price, applied := BestPrice(d("100"), d("10"), offers) // own discount -> 90
	if !price.Equal(d("70")) || applied == nil || applied.ID != 2 {
		t.Fatalf("got %s via %+v, want 70 via offer 2", price, applied)
	}

	price, applied = BestPrice(d("100"), d("40"), offers) // own discount 60 beats every offer
	if !price.Equal(d("60")) || applied != nil {
		t.Fatalf("own discount should win: got %s via %+v", price, applied)
	}

	price, applied = BestPrice(d("100"), d("30"), offers[1:2]) // tie 70 vs 70: no offer label
	if !price.Equal(d("70")) || applied != nil {
		t.Fatalf("tie should keep the product discount: got %s via %+v", price, applied)
	}
}
