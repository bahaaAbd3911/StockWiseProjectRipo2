// Package apperr defines the application's error vocabulary.
//
// Services and repositories return these typed errors ("not found",
// "conflict", ...). The HTTP layer turns each kind into a status code in one
// place (httpx.WriteError), so no handler ever compares error strings. Any
// error that is NOT an *apperr.Error is treated as an unexpected internal
// failure: it is logged and the client only sees "internal server error".
package apperr

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Kind classifies an error so it can be mapped to an HTTP status.
type Kind int

const (
	KindInvalid      Kind = iota + 1 // 400 the request data is wrong
	KindUnauthorized                 // 401 not logged in / bad token
	KindForbidden                    // 403 logged in but not allowed
	KindNotFound                     // 404 the record does not exist (for this user)
	KindConflict                     // 409 valid request, but the current state forbids it
)

// Error is an expected, client-facing error.
type Error struct {
	Kind    Kind
	Message string
	// Fields optionally holds per-field validation messages, e.g.
	// {"email": "must be a valid email address"}.
	Fields map[string]string
}

func (e *Error) Error() string { return e.Message }

func Invalid(msg string) error      { return &Error{Kind: KindInvalid, Message: msg} }
func Unauthorized(msg string) error { return &Error{Kind: KindUnauthorized, Message: msg} }
func Forbidden(msg string) error    { return &Error{Kind: KindForbidden, Message: msg} }
func NotFound(msg string) error     { return &Error{Kind: KindNotFound, Message: msg} }
func Conflict(msg string) error     { return &Error{Kind: KindConflict, Message: msg} }

// Validation returns an Invalid error carrying per-field messages.
func Validation(fields map[string]string) error {
	return &Error{Kind: KindInvalid, Message: "validation failed", Fields: fields}
}

// As extracts an *Error from err, if there is one.
func As(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}

// IsKind reports whether err is an *Error of the given kind.
func IsKind(err error, kind Kind) bool {
	e, ok := As(err)
	return ok && e.Kind == kind
}

// PostgreSQL error codes we translate. Full list:
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
	pgStringTooLong       = "22001"
	pgNumericOutOfRange   = "22003"
)

// FromDB converts low-level database errors into application errors:
//
//   - sql.ErrNoRows            -> NotFound(notFoundMsg)
//   - unique violation          -> Conflict
//   - foreign-key violation     -> NotFound (a referenced row is missing)
//   - check / length / range    -> Invalid
//
// Errors that are already *Error pass through unchanged, and anything else is
// returned as-is (and will become a 500).
func FromDB(err error, notFoundMsg string) error {
	if err == nil {
		return nil
	}
	if _, ok := As(err); ok {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound(notFoundMsg)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return Conflict("a record with the same unique value already exists")
		case pgForeignKeyViolation:
			return NotFound("a referenced record does not exist")
		case pgCheckViolation, pgStringTooLong, pgNumericOutOfRange:
			return Invalid("the data does not satisfy the database constraints")
		}
	}
	return err
}

// IsUniqueViolation reports whether err is a PostgreSQL unique-constraint error,
// optionally for a specific constraint name.
func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgUniqueViolation {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}
