package validators

import (
	"errors"
	"net/mail"
	"strings"

	"ecommerce-backend/internal/models"
)

func ValidateUserRegistration(user models.RegisterUserRequest) error {

	if strings.TrimSpace(user.FirstName) == "" {
		return errors.New("first name is required")
	}

	if strings.TrimSpace(user.LastName) == "" {
		return errors.New("last name is required")
	}

	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}

	_, err := mail.ParseAddress(user.Email)
	if err != nil {
		return errors.New("invalid email address")
	}

	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if strings.TrimSpace(user.Phone) == "" {
		return errors.New("phone number is required")
	}

	return nil
}
func ValidateLogin(login models.LoginRequest) error {

	if strings.TrimSpace(login.Email) == "" {
		return errors.New("email is required")
	}

	if strings.TrimSpace(login.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}
