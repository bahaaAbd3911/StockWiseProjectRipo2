package pricing

import (
	"testing"

	"github.com/shopspring/decimal"

	"ecommerce-backend/internal/models"
)

func TestShippingFee(t *testing.T) {
	d := decimal.RequireFromString
	freeOver := d("50")
	rule := &models.ShippingSettings{
		Fee:      d("4"),
		FreeOver: &freeOver,
		Cities:   []models.CityShipping{{City: "Amman", Fee: d("2")}, {City: "Aqaba", Fee: d("6.50")}},
	}
	cases := []struct {
		name  string
		rule  *models.ShippingSettings
		city  string
		total string
		want  string
	}{
		{"no settings = free", nil, "Amman", "10", "0"},
		{"empty cart ships nothing", rule, "Amman", "0", "0"},
		{"city fee", rule, "Amman", "10", "2"},
		{"city fee ignores case and spaces", rule, "  aqaba ", "10", "6.5"},
		{"other city = standard fee", rule, "Irbid", "10", "4"},
		{"unknown city = standard fee", rule, "", "10", "4"},
		{"free at the threshold", rule, "Aqaba", "50", "0"},
		{"just under the threshold", rule, "Aqaba", "49.99", "6.5"},
		{"no threshold", &models.ShippingSettings{Fee: d("3")}, "Amman", "1000", "3"},
	}
	for _, c := range cases {
		got := ShippingFee(c.rule, c.city, d(c.total))
		if !got.Equal(d(c.want)) {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}
