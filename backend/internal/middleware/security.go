package middleware

import "net/http"

// SecurityHeaders adds the browser protections that make sense for a JSON API:
//
//   - X-Content-Type-Options: nosniff  browsers never guess a JSON body is HTML/JS
//   - Content-Security-Policy / X-Frame-Options  responses cannot run scripts or
//     be shown inside another site's frame
//   - Referrer-Policy: no-referrer  URLs are never leaked to other sites
//   - Strict-Transport-Security (when hsts is true, i.e. in production)  once a
//     browser has seen the API over HTTPS it refuses plain HTTP for two years,
//     so a network attacker cannot downgrade the connection. Only sent on
//     HTTPS responses (directly, or via a proxy's X-Forwarded-Proto); browsers
//     ignore it over HTTP anyway.
func SecurityHeaders(hsts bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			if hsts && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
				h.Set("Strict-Transport-Security", "max-age=63072000")
			}
			next.ServeHTTP(w, r)
		})
	}
}
