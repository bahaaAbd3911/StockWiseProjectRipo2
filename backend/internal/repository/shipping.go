package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/shopspring/decimal"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/pricing"
)

// shippingRule reads the current store's shipping settings (nil = free
// shipping). Works on the pool or inside a transaction (checkout).
func shippingRule(ctx context.Context, q database.Querier) (*models.ShippingSettings, error) {
	var raw []byte
	err := q.QueryRowContext(ctx, `SELECT data->'shipping' FROM store_settings`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) || len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rule models.ShippingSettings
	if err := json.Unmarshal(raw, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// defaultAddressCity is the city of the customer's default address ("" if none):
// the cart's shipping estimate when the page does not say which city.
func defaultAddressCity(ctx context.Context, q database.Querier, userID int64) (string, error) {
	var city string
	err := q.QueryRowContext(ctx,
		`SELECT city FROM user_addresses WHERE user_id = $1 AND is_default`, userID).Scan(&city)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return city, err
}

// shippingFor works out the fee for an amount (items after the coupon) and a city.
func shippingFor(ctx context.Context, q database.Querier, city string, itemsTotal decimal.Decimal) (decimal.Decimal, *models.ShippingSettings, error) {
	rule, err := shippingRule(ctx, q)
	if err != nil {
		return decimal.Zero, nil, err
	}
	return pricing.ShippingFee(rule, city, itemsTotal), rule, nil
}
