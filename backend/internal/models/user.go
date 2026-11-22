package models

import "time"

// User is a row of the users table. password_hash is never part of this struct,
// so it can never be sent to a client by accident.
type User struct {
	ID              int64     `json:"id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           string    `json:"email"`
	Phone           *string   `json:"phone"`
	ProfileImageURL *string   `json:"profile_image_url"`
	Role            string    `json:"role"`
	IsActive        bool      `json:"-"`
	TokenVersion    int       `json:"-"` // see auth.Claims.Version
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	Email           string  `json:"email"`
	Password        string  `json:"password"`
	Phone           string  `json:"phone"`
	ProfileImageURL *string `json:"profile_image_url"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User      User      `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UpdateProfileRequest struct {
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	Phone           string  `json:"phone"`
	ProfileImageURL *string `json:"profile_image_url"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// TokenResponse is returned when an action gives the caller a fresh token
// (e.g. after changing the password, which revokes the old ones).
type TokenResponse struct {
	Message   string    `json:"message"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest: Token comes from the ?token= of the emailed link.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
