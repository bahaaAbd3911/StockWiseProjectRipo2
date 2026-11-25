// Package auth issues and verifies login tokens (JWTs) and carries the
// logged-in user's identity through the request context.
package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Roles stored in users.role.
const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
)

// Identity is who is making the request. It is created by the authentication
// middleware from the token plus a fresh database lookup, so Role is always the
// CURRENT role in the database, never something the client claimed.
type Identity struct {
	UserID        int64
	Role          string
	EmailVerified bool
}

func (i Identity) IsAdmin() bool { return i.Role == RoleAdmin }

// AccountInfo is what authentication needs to know about a user, read fresh
// from the database on every protected request.
type AccountInfo struct {
	Role          string
	Active        bool
	TokenVersion  int // must match the token's "ver" claim
	EmailVerified bool
}

type ctxKey struct{}

// WithIdentity stores the identity in a context.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the identity stored by the middleware, if any.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// UserID returns the logged-in user's ID, or 0 if the request is anonymous.
func UserID(ctx context.Context) int64 {
	id, _ := FromContext(ctx)
	return id.UserID
}

const issuer = "ecommerce-backend"

// Claims is the JWT payload.
type Claims struct {
	UserID int64 `json:"user_id"`
	// Version is users.token_version when the token was issued. Raising the
	// column in the database revokes every older token (see Authenticate).
	Version int `json:"ver"`
	// StoreID is the store the user belongs to. A token is only accepted on
	// that store's host, so a login from one store is useless on another.
	StoreID int64 `json:"sid"`
	jwt.RegisteredClaims
}

// TokenInfo is what a verified token says.
type TokenInfo struct {
	UserID  int64
	Version int
	StoreID int64
}

// TokenManager signs and verifies HS256 tokens with a shared secret.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// Issue creates a signed token for a user of a store at the given token
// version and returns it with its expiry time.
func (m *TokenManager) Issue(storeID, userID int64, version int) (string, time.Time, error) {
	now := m.now()
	expires := now.Add(m.ttl)
	claims := Claims{
		UserID:  userID,
		Version: version,
		StoreID: storeID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

// ErrInvalidToken is returned for any token that fails verification.
var ErrInvalidToken = errors.New("invalid or expired token")

// Parse verifies the signature, algorithm, issuer and expiry, and returns
// what the token says.
func (m *TokenManager) Parse(token string) (TokenInfo, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || claims.UserID <= 0 {
		return TokenInfo{}, ErrInvalidToken
	}
	return TokenInfo{UserID: claims.UserID, Version: claims.Version, StoreID: claims.StoreID}, nil
}
