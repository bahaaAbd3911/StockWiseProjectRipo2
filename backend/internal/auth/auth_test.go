package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "0123456789abcdef0123456789abcdef"

func TestIssueAndParse(t *testing.T) {
	m := NewTokenManager(secret, time.Hour)
	token, expires, err := m.Issue(7, 42, 3)
	if err != nil {
		t.Fatal(err)
	}
	if time.Until(expires) < 59*time.Minute {
		t.Fatalf("unexpected expiry %v", expires)
	}
	info, err := m.Parse(token)
	if err != nil || info.UserID != 42 || info.Version != 3 || info.StoreID != 7 {
		t.Fatalf("Parse = %+v, %v; want user 42, version 3, store 7", info, err)
	}
}

func TestRejectsBadTokens(t *testing.T) {
	m := NewTokenManager(secret, time.Hour)
	good, _, _ := m.Issue(1, 1, 0)

	other := NewTokenManager("ffffffffffffffffffffffffffffffff", time.Hour)
	if _, err := other.Parse(good); err == nil {
		t.Error("token signed with another secret was accepted")
	}

	expired := NewTokenManager(secret, time.Hour)
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	old, _, _ := expired.Issue(1, 1, 0)
	if _, err := m.Parse(old); err == nil {
		t.Error("expired token was accepted")
	}

	none := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: 1, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: issuer, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}})
	unsigned, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := m.Parse(unsigned); err == nil {
		t.Error(`"alg: none" token was accepted`)
	}

	if _, err := m.Parse("garbage"); err == nil {
		t.Error("garbage was accepted")
	}
}
