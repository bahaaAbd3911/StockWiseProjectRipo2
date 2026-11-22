package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce-backend/internal/auth"
)

type fakeAccounts map[int64]struct {
	role    string
	active  bool
	version int
}

func (f fakeAccounts) AuthInfo(_ context.Context, id int64) (auth.AccountInfo, error) {
	a, ok := f[id]
	if !ok {
		return auth.AccountInfo{}, sql.ErrNoRows
	}
	return auth.AccountInfo{Role: a.role, Active: a.active, TokenVersion: a.version}, nil
}

func TestAuthenticateAndRequireAdmin(t *testing.T) {
	tokens := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Hour)
	accounts := fakeAccounts{
		1: {auth.RoleCustomer, true, 0},
		2: {auth.RoleAdmin, true, 0},
		3: {auth.RoleCustomer, false, 0},
		4: {auth.RoleCustomer, true, 2}, // logged out everywhere twice
	}
	tok := func(id int64) string { s, _, _ := tokens.Issue(id, 0); return "Bearer " + s }
	tokV := func(id int64, v int) string { s, _, _ := tokens.Issue(id, v); return "Bearer " + s }

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.UserID(r.Context()) == 0 {
			t.Error("identity missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	userRoute := Authenticate(tokens, accounts)(ok)
	adminRoute := Authenticate(tokens, accounts)(RequireAdmin(ok))

	cases := []struct {
		name   string
		route  http.Handler
		header string
		want   int
	}{
		{"no header", userRoute, "", 401},
		{"wrong scheme", userRoute, "Basic abc", 401},
		{"bad token", userRoute, "Bearer nope", 401},
		{"customer", userRoute, tok(1), 204},
		{"deactivated", userRoute, tok(3), 401},
		{"deleted user", userRoute, tok(99), 401},
		{"revoked (old version)", userRoute, tokV(4, 1), 401},
		{"current version", userRoute, tokV(4, 2), 204},
		{"customer on admin route", adminRoute, tok(1), 403},
		{"admin on admin route", adminRoute, tok(2), 204},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if c.header != "" {
			req.Header.Set("Authorization", c.header)
		}
		rec := httptest.NewRecorder()
		c.route.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s: status %d, want %d (%s)", c.name, rec.Code, c.want, rec.Body)
		}
	}
}

func TestCORS(t *testing.T) {
	h := CORS([]string{"http://localhost:5173"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	pre := httptest.NewRequest(http.MethodOptions, "/api/cart", nil)
	pre.Header.Set("Origin", "http://localhost:5173")
	pre.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, pre)
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("preflight from allowed origin failed: %d %v", rec.Code, rec.Header())
	}

	evil := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
	evil.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, evil)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unknown origin must not get CORS headers")
	}
}

func TestRecoverReturns500(t *testing.T) {
	h := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
}
