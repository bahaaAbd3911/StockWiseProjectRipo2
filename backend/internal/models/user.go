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
