package httpx

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce-backend/internal/apperr"
)

func TestWriteErrorMapsKindsAndHidesInternals(t *testing.T) {
	cases := []struct {
		err      error
		status   int
		contains string
	}{
		{apperr.Invalid("bad"), 400, "bad"},
		{apperr.Unauthorized("who"), 401, "who"},
		{apperr.Forbidden("no"), 403, "no"},
		{apperr.NotFound("gone"), 404, "gone"},
		{apperr.Conflict("busy"), 409, "busy"},
		{apperr.FromDB(sql.ErrNoRows, "order not found"), 404, "order not found"},
		{errors.New(`pq: relation "users" password_hash secret`), 500, "internal server error"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), c.err)
		if rec.Code != c.status || !strings.Contains(rec.Body.String(), c.contains) {
			t.Errorf("%v -> %d %s; want %d containing %q", c.err, rec.Code, rec.Body, c.status, c.contains)
		}
		if c.status == 500 && strings.Contains(rec.Body.String(), "password_hash") {
			t.Error("internal error details leaked to the client")
		}
	}
}

func TestValidationErrorIncludesFields(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), apperr.Validation(map[string]string{"email": "is required"}))
	var body struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != 400 || body.Fields["email"] != "is required" {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestDecodeJSON(t *testing.T) {
	var dst struct {
		Quantity int `json:"quantity"`
	}
	decodeWith := func(body string) error {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		return DecodeJSON(httptest.NewRecorder(), r, &dst)
	}
	if err := decodeWith(`{"quantity": 2}`); err != nil || dst.Quantity != 2 {
		t.Fatalf("valid JSON failed: %v", err)
	}
	if err := decodeWith(``); !apperr.IsKind(err, apperr.KindInvalid) {
		t.Error("empty body should be a 400")
	}
	if err := decodeWith(`{"quantity": "two"}`); !apperr.IsKind(err, apperr.KindInvalid) {
		t.Error("wrong type should be a 400")
	}
	if err := decodeWith(strings.Repeat(" ", MaxBodyBytes+1) + "{}"); !apperr.IsKind(err, apperr.KindInvalid) {
		t.Error("oversized body should be a 400")
	}
}
