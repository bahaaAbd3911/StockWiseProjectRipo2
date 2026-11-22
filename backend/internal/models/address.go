package models

import "time"

type Address struct {
	ID                     int64     `json:"id"`
	UserID                 int64     `json:"user_id"`
	AddressLabel           *string   `json:"address_label"`
	RecipientName          string    `json:"recipient_name"`
	Phone                  string    `json:"phone"`
	City                   string    `json:"city"`
	Area                   string    `json:"area"`
	Street                 string    `json:"street"`
	BuildingNumber         *string   `json:"building_number"`
	ApartmentNumber        *string   `json:"apartment_number"`
	AdditionalInstructions *string   `json:"additional_instructions"`
	IsDefault              bool      `json:"is_default"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// AddressInput is used for both create and update. IsDefault is only read on
// create; use PUT /api/addresses/{id}/default to change the default later.
type AddressInput struct {
	AddressLabel           *string `json:"address_label"`
	RecipientName          string  `json:"recipient_name"`
	Phone                  string  `json:"phone"`
	City                   string  `json:"city"`
	Area                   string  `json:"area"`
	Street                 string  `json:"street"`
	BuildingNumber         *string `json:"building_number"`
	ApartmentNumber        *string `json:"apartment_number"`
	AdditionalInstructions *string `json:"additional_instructions"`
	IsDefault              bool    `json:"is_default"`
}
