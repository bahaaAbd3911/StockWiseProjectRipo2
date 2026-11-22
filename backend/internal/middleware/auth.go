// Package middleware contains code that runs before (or around) handlers:
// authentication, the admin check, CORS, request logging and panic recovery.
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
)

// AccountLookup returns the current role and active flag of a user. The user
// repository implements it; tests can pass a fake.
type AccountLookup interface {
	AuthInfo(ctx context.Context, userID int64) (role string, active bool, err error)
}

// Authenticate requires a valid "Authorization: Bearer <token>" header.
//
// Steps:
//  1. read and split the header
//  2. verify the token (signature, algorithm, issuer, expiry)
//  3. look the user up in the database: they must still exist and be active
//     (so deactivating an account locks it out immediately, even with a valid
//     token) and we read their CURRENT role
//  4. store auth.Identity in the request context for the handlers
func Authenticate(tokens *auth.TokenManager, accounts AccountLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				httpx.WriteError(w, r, apperr.Unauthorized("missing authorization token"))
				return
			}
			scheme, token, ok := strings.Cut(header, " ")
			token = strings.TrimSpace(token)
			if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
				httpx.WriteError(w, r, apperr.Unauthorized("authorization header must be: Bearer <token>"))
				return
			}

			userID, err := tokens.Parse(token)
			if err != nil {
				httpx.WriteError(w, r, apperr.Unauthorized("invalid or expired token"))
				return
			}

			role, active, err := accounts.AuthInfo(r.Context(), userID)
			if errors.Is(err, sql.ErrNoRows) || (err == nil && !active) {
				httpx.WriteError(w, r, apperr.Unauthorized("account is inactive or does not exist"))
				return
			}
			if err != nil {
				httpx.WriteError(w, r, err)
				return
			}

			ctx := auth.WithIdentity(r.Context(), auth.Identity{UserID: userID, Role: role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin must run after Authenticate. It rejects non-admins with 403.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := auth.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, apperr.Unauthorized("authentication required"))
			return
		}
		if !id.IsAdmin() {
			httpx.WriteError(w, r, apperr.Forbidden("administrator access required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
