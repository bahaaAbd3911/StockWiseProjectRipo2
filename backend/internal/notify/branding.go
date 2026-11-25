package notify

import (
	"context"
	"net/url"
	"strings"

	"ecommerce-backend/internal/tenant"
)

// Branding gives every email the details of the store it is sent for: the
// store's name (subject, signature, sender name), links to that store's own
// address, and its currency. The store comes from the context (requests) or
// is passed in (background jobs that work across stores).
type Branding struct {
	Base Templates // defaults: STORE_NAME, FRONTEND_URL, CURRENCY

	// StoreURL returns a store's frontend address, e.g. https://ahmad.myplatform.com.
	StoreURL func(tenant.Store) string

	// Paths of the pages that emailed links open (taken from
	// EMAIL_VERIFY_URL / PASSWORD_RESET_URL), added to the store's address.
	VerifyPath string // e.g. /verify-email
	ResetPath  string // e.g. /reset-password
}

// NewBranding builds a Branding from the configured links.
func NewBranding(base Templates, storeURL func(tenant.Store) string, verifyURL, resetURL string) *Branding {
	return &Branding{Base: base, StoreURL: storeURL, VerifyPath: pathOf(verifyURL, "/verify-email"), ResetPath: pathOf(resetURL, "/reset-password")}
}

func pathOf(raw, fallback string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return fallback
	}
	if u.RawQuery != "" {
		return u.Path + "?" + u.RawQuery
	}
	return u.Path
}

// For returns the templates of the context's store (the defaults if none).
func (b *Branding) For(ctx context.Context) Templates {
	if b == nil {
		return Templates{}
	}
	s, ok := tenant.FromContext(ctx)
	if !ok {
		return b.withLinks(b.Base)
	}
	return b.ForStore(s)
}

// ForStore returns the templates for a given store.
func (b *Branding) ForStore(s tenant.Store) Templates {
	t := b.Base
	if s.Name != "" {
		t.Store = s.Name
	}
	if s.Currency != "" {
		t.Currency = s.Currency
	}
	if b.StoreURL != nil {
		if u := b.StoreURL(s); u != "" {
			t.FrontendURL = u
		}
	}
	return b.withLinks(t)
}

func (b *Branding) withLinks(t Templates) Templates {
	base := strings.TrimRight(t.FrontendURL, "/")
	t.VerifyURL = base + b.VerifyPath
	t.ResetURL = base + b.ResetPath
	return t
}
