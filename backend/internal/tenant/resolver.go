package tenant

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound  = errors.New("store not found")
	ErrSuspended = errors.New("store suspended")
)

// Resolver maps a host name to its store:
//
//	<slug>.PLATFORM_DOMAIN          -> the store with that slug
//	a verified custom domain         -> the store that owns it (store_domains)
//	anything else (localhost, an IP,
//	the bare platform domain...)     -> DEFAULT_STORE, if one is configured
//
// Answers are cached for a short time so most requests need no query.
type Resolver struct {
	db             *sql.DB
	platformDomain string // e.g. "myplatform.com" ("" = subdomains off)
	defaultSlug    string // e.g. "main" ("" = unknown hosts get 404)
	ttl            time.Duration

	mu    sync.Mutex
	cache map[string]cached
}

type cached struct {
	store   Store
	err     error
	expires time.Time
}

func NewResolver(db *sql.DB, platformDomain, defaultSlug string) *Resolver {
	return &Resolver{
		db:             db,
		platformDomain: strings.ToLower(strings.Trim(platformDomain, ".")),
		defaultSlug:    defaultSlug,
		ttl:            30 * time.Second,
		cache:          map[string]cached{},
	}
}

// NormalizeHost lower-cases a Host header and removes the port and a final dot.
func NormalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.Trim(strings.TrimSuffix(host, "."), "[]")
}

// SlugFor returns the subdomain label when host is <label>.platformDomain.
func (r *Resolver) SlugFor(host string) (string, bool) {
	if r.platformDomain == "" {
		return "", false
	}
	label, ok := strings.CutSuffix(host, "."+r.platformDomain)
	if !ok || label == "" || strings.Contains(label, ".") {
		return "", false
	}
	return label, true
}

// Resolve returns the active store for a Host header value.
func (r *Resolver) Resolve(ctx context.Context, rawHost string) (Store, error) {
	host := NormalizeHost(rawHost)
	r.mu.Lock()
	c, ok := r.cache[host]
	r.mu.Unlock()
	if ok && time.Now().Before(c.expires) {
		return c.store, c.err
	}

	s, err := r.lookup(ctx, host)
	if err == nil && s.Status != StatusActive {
		err = ErrSuspended
	}
	if err == nil || errors.Is(err, ErrNotFound) || errors.Is(err, ErrSuspended) {
		r.mu.Lock()
		if len(r.cache) > 10_000 { // a flood of made-up host names cannot grow memory forever
			r.cache = map[string]cached{}
		}
		r.cache[host] = cached{store: s, err: err, expires: time.Now().Add(r.ttl)}
		r.mu.Unlock()
	}
	return s, err
}

// Forget drops cached answers (after a store is renamed, suspended...).
func (r *Resolver) Forget() {
	r.mu.Lock()
	r.cache = map[string]cached{}
	r.mu.Unlock()
}

const storeColumns = `s.id, s.slug, s.name, s.status, s.currency, s.timezone,
	COALESCE((SELECT p.domain FROM store_domains p WHERE p.store_id = s.id AND p.is_primary AND p.verified_at IS NOT NULL), '')`

func (r *Resolver) lookup(ctx context.Context, host string) (Store, error) {
	if slug, ok := r.SlugFor(host); ok {
		return r.bySlug(ctx, slug)
	}
	if host != "" && net.ParseIP(host) == nil && host != "localhost" && host != r.platformDomain {
		s, err := scan(r.db.QueryRowContext(ctx, `
			SELECT `+storeColumns+` FROM store_domains d JOIN stores s ON s.id = d.store_id
			WHERE d.domain = $1 AND d.verified_at IS NOT NULL`, host))
		if !errors.Is(err, ErrNotFound) {
			return s, err
		}
	}
	if r.defaultSlug != "" {
		return r.bySlug(ctx, r.defaultSlug)
	}
	return Store{}, ErrNotFound
}

func (r *Resolver) bySlug(ctx context.Context, slug string) (Store, error) {
	return scan(r.db.QueryRowContext(ctx, `SELECT `+storeColumns+` FROM stores s WHERE s.slug = $1`, slug))
}

func scan(row *sql.Row) (Store, error) {
	var s Store
	err := row.Scan(&s.ID, &s.Slug, &s.Name, &s.Status, &s.Currency, &s.Timezone, &s.PrimaryDomain)
	if errors.Is(err, sql.ErrNoRows) {
		return Store{}, ErrNotFound
	}
	return s, err
}

// IsServed reports whether the platform serves this host name: a store's
// subdomain, a verified custom domain of an active store, or the platform
// admin host. Caddy asks this before getting an HTTPS certificate on demand,
// so certificates are only ever requested for real stores.
func (r *Resolver) IsServed(ctx context.Context, rawHost, platformHost string) (bool, error) {
	host := NormalizeHost(rawHost)
	if host == "" {
		return false, nil
	}
	if host == NormalizeHost(platformHost) || (r.platformDomain != "" && host == r.platformDomain && r.defaultSlug != "") {
		return true, nil // the platform admin, and the bare domain when it shows the default store
	}
	var s Store
	var err error
	if slug, ok := r.SlugFor(host); ok {
		s, err = r.bySlug(ctx, slug)
	} else {
		s, err = scan(r.db.QueryRowContext(ctx, `
			SELECT `+storeColumns+` FROM store_domains d JOIN stores s ON s.id = d.store_id
			WHERE d.domain = $1 AND d.verified_at IS NOT NULL`, host))
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil && s.Status == StatusActive, err
}
