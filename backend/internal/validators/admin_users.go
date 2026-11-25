package validators

import (
	"strings"

	"ecommerce-backend/internal/models"
)

// ValidateAdminUser checks an admin's create/edit-user body. creating = true
// requires a password; on edit an empty password keeps the current one. The
// phone is optional here (an admin may not know it).
func ValidateAdminUser(in *models.AdminUserInput, creating bool) error {
	Trim(&in.FirstName)
	Trim(&in.LastName)
	Trim(&in.Phone)
	in.Email = NormalizeEmail(in.Email)
	in.Role = strings.ToLower(strings.TrimSpace(in.Role))
	if in.Role == "" {
		in.Role = "customer"
	}

	c := New()
	checkName(c, in.FirstName, in.LastName)
	c.Required("email", in.Email)
	c.Add(in.Email == "" || IsEmail(in.Email), "email", "must be a valid email address")
	c.MaxLen("email", in.Email, 255)
	if creating || in.Password != "" {
		checkPassword(c, "password", in.Password)
	}
	c.Add(in.Phone == "" || IsPhone(in.Phone), "phone", "must be a valid phone number")
	c.MaxLen("phone", in.Phone, 20)
	c.Add(in.Role == "customer" || in.Role == "admin", "role", "must be customer or admin")
	return c.Err()
}
