// Package httpx holds the small HTTP helpers every handler uses: reading JSON,
// writing JSON, turning errors into responses, and parsing path/query values.
//
// Handlers in this project have the signature
//
//	func(w http.ResponseWriter, r *http.Request) error
//
// and simply `return err` when something goes wrong. Handle() wraps them and
// calls WriteError, so error-to-status mapping lives in exactly one place.
package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-backend/internal/apperr"
)

// HandlerFunc is an HTTP handler that can return an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handle adapts a HandlerFunc to a standard http.Handler.
func Handle(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteError(w, r, err)
		}
	})
}

// MaxBodyBytes limits request bodies so a client cannot exhaust memory.
const MaxBodyBytes = 1 << 20 // 1 MiB

// DecodeJSON reads the request body into dst and returns a client-friendly
// error when the body is missing, too large, or not valid JSON.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return decode(w, r, dst, false)
}

// DecodeOptionalJSON is DecodeJSON for endpoints where the body may be empty.
func DecodeOptionalJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return decode(w, r, dst, true)
}

func decode(w http.ResponseWriter, r *http.Request, dst any, optional bool) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &maxErr):
			return apperr.Invalid("request body is too large")
		case errors.Is(err, io.EOF):
			if optional {
				return nil
			}
			return apperr.Invalid("request body is required")
		case errors.As(err, &typeErr) && typeErr.Field != "":
			return apperr.Validation(map[string]string{typeErr.Field: "has the wrong type"})
		default:
			return apperr.Invalid("request body must be valid JSON")
		}
	}
	return nil
}

// JSON writes v as a JSON response with the given status code. It always
// returns nil so handlers can `return httpx.JSON(...)`; an encoding failure
// after the header is sent cannot be reported to the client, so it is logged.
func JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode JSON response", "error", err)
	}
	return nil
}

// Message writes {"message": msg}.
func Message(w http.ResponseWriter, status int, msg string) error {
	return JSON(w, status, map[string]string{"message": msg})
}

type errorBody struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// StatusFor maps an error kind to an HTTP status code.
func StatusFor(kind apperr.Kind) int {
	switch kind {
	case apperr.KindInvalid:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// WriteError sends an error response. Expected errors (*apperr.Error) are shown
// to the client; anything else is logged and hidden behind a generic message so
// SQL details never leak.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := apperr.As(err); ok {
		if e.RetryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(e.RetryAfter.Seconds()))))
		}
		_ = JSON(w, StatusFor(e.Kind), errorBody{Error: e.Message, Fields: e.Fields})
		return
	}
	if errors.Is(err, context.Canceled) {
		// The client went away; nobody is listening for a response.
		return
	}
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = JSON(w, http.StatusInternalServerError, errorBody{Error: "internal server error"})
}

// PathID parses a positive integer path parameter such as {id}.
func PathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.Invalid(fmt.Sprintf("invalid %s", strings.ReplaceAll(name, "_", " ")))
	}
	return id, nil
}

// QueryInt reads an optional integer query parameter within [min, max].
func QueryInt(r *http.Request, name string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		return 0, apperr.Invalid(fmt.Sprintf("query parameter %q must be a number between %d and %d", name, min, max))
	}
	return n, nil
}

// Pagination reads ?limit= and ?offset= with sensible bounds.
func Pagination(r *http.Request) (limit, offset int, err error) {
	if limit, err = QueryInt(r, "limit", 20, 1, 100); err != nil {
		return 0, 0, err
	}
	if offset, err = QueryInt(r, "offset", 0, 0, 1_000_000); err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

type clientIPKey struct{}

// WithClientIP stores the caller's IP in the context (see middleware.RealIP).
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// ClientIP returns the caller's IP stored by middleware.RealIP, or "".
func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

// RemoteIP extracts the client IP of a request. With trustProxy it uses the
// first address in X-Forwarded-For (or X-Real-IP), as set by a reverse proxy.
func RemoteIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first, _, _ := strings.Cut(xff, ",")
			if ip := net.ParseIP(strings.TrimSpace(first)); ip != nil {
				return ip.String()
			}
		}
		if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); ip != nil {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
