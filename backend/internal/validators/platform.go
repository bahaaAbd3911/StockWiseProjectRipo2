package validators

import (
	"regexp"
	"strings"
	"time"

	"ecommerce-backend/internal/models"
)

// A store's slug is its subdomain: a DNS label (same rule as the database check).
var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ReservedSlugs cannot be stores: they are used by the platform itself or
// would confuse customers.
var ReservedSlugs = map[string]bool{
	"admin": true, "www": true, "api": true, "app": true, "platform": true, "mail": true, "smtp": true,
	"static": true, "assets": true, "cdn": true, "uploads": true, "help": true, "support": true,
	"status": true, "blog": true, "docs": true, "dashboard": true, "billing": true, "login": true,
}

func checkStoreDetails(c *Checker, name, currency, timezone *string) {
	Trim(name)
	*currency = strings.ToUpper(strings.TrimSpace(*currency))
	*timezone = strings.TrimSpace(*timezone)
	if *timezone == "" {
		*timezone = "Asia/Amman"
	}
	c.Required("name", *name)
	c.MaxLen("name", *name, 150)
	c.MaxLen("currency", *currency, 10)
	_, err := time.LoadLocation(*timezone)
	c.Add(err == nil && len(*timezone) <= 64, "timezone", "must be a time zone such as Asia/Amman or Asia/Riyadh")
}

// ValidateCreateStore normalises and checks a new store with its admin.
func ValidateCreateStore(in *models.CreateStoreRequest) error {
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	Trim(&in.Admin.FirstName)
	Trim(&in.Admin.LastName)
	in.Admin.Email = NormalizeEmail(in.Admin.Email)

	c := New()
	c.Add(slugPattern.MatchString(in.Slug), "slug",
		"must be 1-63 lower-case letters, digits or hyphens (not at the start or end), e.g. ahmad-store")
	c.Add(!ReservedSlugs[in.Slug], "slug", "this address is reserved; choose another")
	checkStoreDetails(c, &in.Name, &in.Currency, &in.Timezone)
	c.Required("admin.first_name", in.Admin.FirstName)
	c.MaxLen("admin.first_name", in.Admin.FirstName, 100)
	c.Required("admin.last_name", in.Admin.LastName)
	c.MaxLen("admin.last_name", in.Admin.LastName, 100)
	c.Required("admin.email", in.Admin.Email)
	c.Add(in.Admin.Email == "" || IsEmail(in.Admin.Email), "admin.email", "must be a valid email address")
	c.MaxLen("admin.email", in.Admin.Email, 255)
	return c.Err()
}

func ValidateUpdateStore(in *models.UpdateStoreRequest) error {
	c := New()
	checkStoreDetails(c, &in.Name, &in.Currency, &in.Timezone)
	return c.Err()
}
