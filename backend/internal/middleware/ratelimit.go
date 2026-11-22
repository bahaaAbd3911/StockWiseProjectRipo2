package middleware

import (
	"net/http"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/ratelimit"
)

// RealIP works out the caller's IP once per request and stores it in the
// context, so rate limiting and logging all use the same value.
// trustProxy: see config.TrustProxy.
func RealIP(trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httpx.RemoteIP(r, trustProxy)
			next.ServeHTTP(w, r.WithContext(httpx.WithClientIP(r.Context(), ip)))
		})
	}
}

// RateLimit refuses requests from an IP that used up its budget, with
// 429 Too Many Requests and a Retry-After header. A nil limiter = no limit.
// CORS preflight requests are never counted.
func RateLimit(l *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if l == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodOptions {
				if ok, wait := l.Allow(httpx.ClientIP(r.Context())); !ok {
					httpx.WriteError(w, r, apperr.TooManyRequests("too many requests, please try again later", wait))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
