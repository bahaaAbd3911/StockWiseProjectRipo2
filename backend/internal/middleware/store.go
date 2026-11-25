package middleware

import (
	"errors"
	"net/http"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/tenant"
)

// Store finds the store a request belongs to from its host name and puts it
// in the context. Everything after it - authentication, handlers, every
// database query - runs for that store only.
//
// The host is the Host header (a reverse proxy such as Caddy keeps it). With
// trustProxy, X-Forwarded-Host wins, for proxies that rewrite Host.
//
// Paths that are the same for every store (health checks, uploaded images
// with random names) skip the lookup.
func Store(resolver *tenant.Resolver, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if storeless(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			host := r.Host
			if trustProxy {
				if fwd := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); fwd != "" {
					host = fwd
				}
			}
			store, err := resolver.Resolve(r.Context(), host)
			switch {
			case errors.Is(err, tenant.ErrNotFound):
				httpx.WriteError(w, r, apperr.NotFound("store not found"))
				return
			case errors.Is(err, tenant.ErrSuspended):
				httpx.WriteError(w, r, apperr.Forbidden("this store is currently unavailable"))
				return
			case err != nil:
				httpx.WriteError(w, r, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(tenant.WithStore(r.Context(), store)))
		})
	}
}

func storeless(path string) bool {
	return path == "/" || path == "/api/health" || strings.HasPrefix(path, "/uploads/")
}
