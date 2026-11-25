// Package tenant knows which store a request belongs to.
//
// The store is decided by the server from the host name the visitor used
// (ahmad.myplatform.com -> store "ahmad"; later also custom domains), never
// from a value the client sends. It travels in the request context; the
// database layer reads it from there and sets app.store_id on the connection,
// and PostgreSQL's row-level security then only shows that store's rows.
package tenant

import "context"

// Store is the tenant a request runs for.
type Store struct {
	ID       int64  `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Status   string `json:"status"` // active | suspended
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
	// PrimaryDomain is the store's verified main custom domain, if any (its
	// address in emails), e.g. shop.example.com.
	PrimaryDomain string `json:"primary_domain"`
}

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

type storeKey struct{}
type platformKey struct{}

// WithStore returns a context that runs as store s.
func WithStore(ctx context.Context, s Store) context.Context {
	return context.WithValue(ctx, storeKey{}, s)
}

// FromContext returns the store of the context, if any.
func FromContext(ctx context.Context) (Store, bool) {
	s, ok := ctx.Value(storeKey{}).(Store)
	return s, ok
}

// ID is the current store's id, or 0 when there is none.
func ID(ctx context.Context) int64 {
	s, _ := FromContext(ctx)
	return s.ID
}

// AsPlatform marks a context as platform-wide: it may read and write the rows
// of every store. Only trusted server code calls this (background jobs, the
// platform admin); it is never derived from request data.
func AsPlatform(ctx context.Context) context.Context {
	return context.WithValue(ctx, platformKey{}, true)
}

// IsPlatform reports whether ctx runs platform-wide.
func IsPlatform(ctx context.Context) bool {
	p, _ := ctx.Value(platformKey{}).(bool)
	return p
}
