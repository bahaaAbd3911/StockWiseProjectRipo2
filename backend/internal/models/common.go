// Package models defines the data shapes passed between layers and sent as
// JSON: database records (Product, Order, ...), request bodies (...Request /
// ...Input) and a few pure domain rules (allowed status transitions).
package models

import "github.com/shopspring/decimal"

func init() {
	// Money is decimal.Decimal everywhere. By default it would be written to JSON
	// as a string ("19.99"); this makes it a JSON number (19.99) instead, which
	// is friendlier for the frontend. The value is still exact.
	decimal.MarshalJSONWithoutQuotes = true
}

// Page is a paginated list response.
type Page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
