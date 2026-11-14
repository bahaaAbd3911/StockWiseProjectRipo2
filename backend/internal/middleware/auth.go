package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"ecommerce-backend/internal/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Missing authorization token", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateJWT(parts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		var isActive bool

		err = db.QueryRow(
			"SELECT is_active FROM users WHERE id = $1",
			claims.UserID,
		).Scan(&isActive)

		if err != nil || !isActive {
			http.Error(w, "Account is inactive or does not exist", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
