package models

import "github.com/shopspring/decimal"

// AdminUser is a user as admins see them: the profile plus account status and
// shopping totals.
type AdminUser struct {
	User
	IsActive    bool            `json:"is_active"`
	OrdersCount int             `json:"orders_count"` // orders that were not cancelled
	TotalSpent  decimal.Decimal `json:"total_spent"`  // value of those orders
}

// AdminUserDetails is the admin's user page: the user, their orders and favorites.
type AdminUserDetails struct {
	User      AdminUser  `json:"user"`
	Orders    []Order    `json:"orders"`
	Favorites []Favorite `json:"favorites"`
}

// AdminUserInput is the admin body for creating or editing a user. On edit an
// empty password keeps the current one. IsActive omitted: true on create,
// unchanged on edit.
type AdminUserInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	Role      string `json:"role"` // customer (default) | admin
	IsActive  *bool  `json:"is_active"`
}

type AdminUserFilter struct {
	Search string // name, email or phone
	Role   string
	Limit  int
	Offset int
}
