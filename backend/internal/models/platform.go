package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PlatformAdmin runs the platform (not a store user).
type PlatformAdmin struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsActive     bool      `json:"is_active"`
	TokenVersion int       `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type PlatformLoginResponse struct {
	Admin     PlatformAdmin `json:"admin"`
	Token     string        `json:"token"`
	ExpiresAt time.Time     `json:"expires_at"`
}

// PlatformStore is a store as the platform admin sees it.
type PlatformStore struct {
	ID             int64           `json:"id"`
	Slug           string          `json:"slug"`
	Name           string          `json:"name"`
	Status         string          `json:"status"` // active | suspended
	Currency       string          `json:"currency"`
	Timezone       string          `json:"timezone"`
	URL            string          `json:"url"`            // the store's public address
	PrimaryDomain  string          `json:"primary_domain"` // verified main custom domain, if any
	AdminEmail     *string         `json:"admin_email"`    // its first admin
	ProductsCount  int             `json:"products_count"`
	CustomersCount int             `json:"customers_count"`
	OrdersCount    int             `json:"orders_count"` // not cancelled
	Revenue        decimal.Decimal `json:"revenue"`      // value of those orders
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// CreateStoreRequest: POST /api/platform/stores. Creates the store, its
// settings and its admin account, and emails the admin a link to choose a
// password.
type CreateStoreRequest struct {
	Slug     string            `json:"slug"` // the subdomain; cannot be changed later
	Name     string            `json:"name"`
	Currency string            `json:"currency"`
	Timezone string            `json:"timezone"`
	Admin    StoreAdminRequest `json:"admin"`
}

type StoreAdminRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// UpdateStoreRequest: PUT /api/platform/stores/{id} (the slug stays).
type UpdateStoreRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
}

type StoreFilter struct {
	Search string // slug or name
	Status string
	Limit  int
	Offset int
}

// CreatedStore is the reply to creating a store.
type CreatedStore struct {
	Store   PlatformStore `json:"store"`
	Message string        `json:"message"`
}
