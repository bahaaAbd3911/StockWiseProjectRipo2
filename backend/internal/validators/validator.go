// Package validators checks request data BEFORE it reaches the database.
//
// Each Validate... function first normalises the input (trims spaces,
// lower-cases emails, turns "" into nil for optional fields) and then checks it
// against the limits in the schema (varchar lengths, numeric(10,2), CHECK
// constraints). All problems are collected and returned together as
// {"error":"validation failed","fields":{...}} so a form can highlight every
// bad field at once.
package validators

import (
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"ecommerce-backend/internal/apperr"

	"github.com/shopspring/decimal"
)

// Checker collects field errors. The first message per field wins.
type Checker struct {
	fields map[string]string
}

func New() *Checker { return &Checker{fields: map[string]string{}} }

// Add records msg for field when ok is false.
func (c *Checker) Add(ok bool, field, msg string) {
	if ok {
		return
	}
	if _, exists := c.fields[field]; !exists {
		c.fields[field] = msg
	}
}

// Required checks a trimmed string is not empty.
func (c *Checker) Required(field, value string) {
	c.Add(value != "", field, "is required")
}

// MaxLen counts characters (not bytes), like PostgreSQL varchar(n). Arabic
// letters are 2 bytes each in UTF-8 but count as one character.
func (c *Checker) MaxLen(field, value string, n int) {
	c.Add(utf8.RuneCountInString(value) <= n, field, "is too long (maximum "+itoa(n)+" characters)")
}

// OptionalMaxLen is MaxLen for nullable fields.
func (c *Checker) OptionalMaxLen(field string, value *string, n int) {
	if value != nil {
		c.MaxLen(field, *value, n)
	}
}

// Err returns nil when there were no problems.
func (c *Checker) Err() error {
	if len(c.fields) == 0 {
		return nil
	}
	return apperr.Validation(c.fields)
}

// ---- normalisation helpers ----

// Trim trims a required string in place.
func Trim(s *string) { *s = strings.TrimSpace(*s) }

// TrimOptional trims an optional string and turns "" into nil (NULL).
func TrimOptional(s **string) {
	if *s == nil {
		return
	}
	v := strings.TrimSpace(**s)
	if v == "" {
		*s = nil
		return
	}
	*s = &v
}

// NormalizeEmail trims and lower-cases an email so "Ali@X.com" and
// "ali@x.com" are the same account.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ---- format checks ----

var phonePattern = regexp.MustCompile(`^\+?[0-9][0-9 ()-]{5,18}$`)

func IsPhone(s string) bool { return phonePattern.MatchString(s) }

func IsEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	// ParseAddress accepts "Name <a@b.c>"; we want only the bare address.
	return err == nil && addr.Address == s && strings.Contains(s, ".")
}

func IsHTTPURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func IsHexColor(s string) bool { return hexColorPattern.MatchString(s) }

// maxMoney is the largest value numeric(10,2) can hold.
var maxMoney = decimal.RequireFromString("99999999.99")

// Money checks 0 <= v <= 99,999,999.99 with at most 2 decimal places.
func (c *Checker) Money(field string, v decimal.Decimal) {
	c.Add(!v.IsNegative(), field, "cannot be negative")
	c.Add(v.LessThanOrEqual(maxMoney), field, "is too large")
	c.Add(v.Equal(v.Round(2)), field, "can have at most 2 decimal places")
}

// Percent checks 0 <= v <= 100 with at most 2 decimal places.
func (c *Checker) Percent(field string, v decimal.Decimal) {
	c.Add(!v.IsNegative() && v.LessThanOrEqual(decimal.NewFromInt(100)), field, "must be between 0 and 100")
	c.Add(v.Equal(v.Round(2)), field, "can have at most 2 decimal places")
}

func itoa(n int) string { return strconv.Itoa(n) }

// FieldError is a shortcut for a validation error on a single field.
func FieldError(field, msg string) error {
	return apperr.Validation(map[string]string{field: msg})
}
