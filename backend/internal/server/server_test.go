package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce-backend/internal/middleware"
)

func TestRedirectToHTTPS(t *testing.T) {
	cases := []struct{ port, host, path, want string }{
		{"443", "api.example.com", "/api/products?page=2", "https://api.example.com/api/products?page=2"},
		{"443", "api.example.com:80", "/", "https://api.example.com/"},
		{"8443", "api.example.com", "/api/health", "https://api.example.com:8443/api/health"},
	}
	for _, c := range cases {
		req := httptest.NewRequest("POST", "http://"+c.host+c.path, strings.NewReader("{}"))
		req.Host = c.host
		rec := httptest.NewRecorder()
		RedirectToHTTPS(c.port).ServeHTTP(rec, req)
		if rec.Code != http.StatusPermanentRedirect || rec.Header().Get("Location") != c.want {
			t.Errorf("%s%s: got %d %q, want 308 %q", c.host, c.path, rec.Code, rec.Header().Get("Location"), c.want)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	get := func(hsts bool, proto string) http.Header {
		req := httptest.NewRequest("GET", "/api/health", nil)
		if proto != "" {
			req.Header.Set("X-Forwarded-Proto", proto)
		}
		rec := httptest.NewRecorder()
		middleware.SecurityHeaders(hsts)(ok).ServeHTTP(rec, req)
		return rec.Header()
	}

	h := get(false, "https")
	if h.Get("X-Content-Type-Options") != "nosniff" || h.Get("X-Frame-Options") != "DENY" || h.Get("Content-Security-Policy") == "" {
		t.Fatalf("basic headers missing: %v", h)
	}
	if h.Get("Strict-Transport-Security") != "" {
		t.Fatal("HSTS must be off outside production")
	}
	if get(true, "https").Get("Strict-Transport-Security") == "" {
		t.Fatal("HSTS expected on HTTPS in production")
	}
	if get(true, "").Get("Strict-Transport-Security") != "" {
		t.Fatal("HSTS must not be sent over plain HTTP")
	}
}
