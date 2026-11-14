package middleware

import (
	"database/sql"
	"net/http"
)

func AdminMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value(UserIDKey).(int64)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var role string

		err := db.QueryRow(
			"SELECT role FROM users WHERE id = $1 AND is_active = TRUE",
			userID,
		).Scan(&role)

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if role != "admin" {
			http.Error(w, "Forbidden: administrator access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
