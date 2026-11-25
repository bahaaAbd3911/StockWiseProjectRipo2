package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/tenant"
)

// PlatformAccounts looks up a platform admin for authentication.
type PlatformAccounts interface {
	AuthInfo(ctx context.Context, adminID int64) (active bool, tokenVersion int, err error)
}

// AuthenticatePlatform requires a platform-admin token (issued by
// /api/platform/login; store logins are refused because they are signed by
// another issuer). The admin must exist, be active, and the token version
// must be current. The request then runs in platform mode: it may work with
// every store's data.
func AuthenticatePlatform(tokens *auth.TokenManager, accounts PlatformAccounts) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, token, ok := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
				httpx.WriteError(w, r, apperr.Unauthorized("missing authorization token"))
				return
			}
			claims, err := tokens.Parse(strings.TrimSpace(token))
			if err != nil {
				httpx.WriteError(w, r, apperr.Unauthorized("invalid or expired token"))
				return
			}
			active, version, err := accounts.AuthInfo(r.Context(), claims.UserID)
			if errors.Is(err, sql.ErrNoRows) || (err == nil && !active) {
				httpx.WriteError(w, r, apperr.Unauthorized("account is inactive or does not exist"))
				return
			}
			if err != nil {
				httpx.WriteError(w, r, err)
				return
			}
			if version != claims.Version {
				httpx.WriteError(w, r, apperr.Unauthorized("this login has been revoked; please log in again"))
				return
			}
			ctx := tenant.AsPlatform(auth.WithPlatformAdmin(r.Context(), claims.UserID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
