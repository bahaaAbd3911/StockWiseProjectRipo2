package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/ratelimit"
	"ecommerce-backend/internal/routes"
	"ecommerce-backend/internal/services"
	"ecommerce-backend/internal/storage"
)

// ---------- test mailer ----------

// captureMailer keeps emails in memory instead of sending them. failNext
// simulates a mail-provider outage for one address.
type captureMailer struct {
	mu       sync.Mutex
	msgs     []mail.Message
	failNext map[string]int
}

func (c *captureMailer) Send(_ context.Context, m mail.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failNext[m.To] > 0 {
		c.failNext[m.To]--
		return errors.New("simulated SMTP outage")
	}
	c.msgs = append(c.msgs, m)
	return nil
}

func (c *captureMailer) failFor(addr string, times int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failNext == nil {
		c.failNext = map[string]int{}
	}
	c.failNext[addr] = times
}

// to returns every email sent to an address.
func (c *captureMailer) to(addr string) []mail.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []mail.Message
	for _, m := range c.msgs {
		if m.To == addr {
			out = append(out, m)
		}
	}
	return out
}

var outbox = &captureMailer{}

// testPublicURL is the API address used in uploaded-image URLs during tests.
const testPublicURL = "http://api.test"

// testPlatformDomain: other stores are reached with Host: <slug>.shop.test.
const testPlatformDomain = "shop.test"

func testConfig() config.Config {
	return config.Config{
		JWTSecret:                "integration-test-secret-0123456789abcdef",
		JWTTTL:                   time.Hour,
		PlatformDomain:           testPlatformDomain, // stores at <slug>.shop.test
		DefaultStore:             "main",             // 127.0.0.1 (the test server) is the main store
		PlatformAdminHost:        "admin." + testPlatformDomain,
		StoreURL:                 "http://{slug}." + testPlatformDomain,
		AllowedOrigins:           []string{"http://localhost:5173"},
		StoreName:                "TestShop",
		FrontendURL:              "http://localhost:5173",
		Currency:                 "JOD",
		PasswordResetURL:         "http://localhost:5173/reset-password",
		PasswordResetTTL:         30 * time.Minute,
		EmailVerifyURL:           "http://localhost:5173/verify-email",
		EmailVerifyTTL:           48 * time.Hour,
		RequireEmailVerification: true,
		EmailPollInterval:        50 * time.Millisecond,
		DebtReminderEvery:        0, // the debt test triggers reminders itself
		PublicURL:                testPublicURL,
		Storage:                  storage.Config{Driver: "local", Dir: uploadDir, PublicURL: testPublicURL},
		MaxUploadBytes:           1 << 20,
		StoreTimezone:            mustLocation("Asia/Amman"),
		LowStockThreshold:        5,
	}
}

var tokenInLink = regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)

// about returns the emails to addr whose subject contains part.
func (c *captureMailer) about(addr, part string) []mail.Message {
	var out []mail.Message
	for _, m := range c.to(addr) {
		if strings.Contains(m.Subject, part) {
			out = append(out, m)
		}
	}
	return out
}

const (
	resetSubject  = "Reset your password"
	verifySubject = "Confirm your email"
)

// waitForMail waits until `n` emails about `subject` reached addr (they are
// sent in the background) and returns the token from the newest one.
func waitForMail(t *testing.T, box *captureMailer, addr, subject string, n int) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if msgs := box.about(addr, subject); len(msgs) >= n {
			m := tokenInLink.FindStringSubmatch(msgs[len(msgs)-1].Text)
			if m == nil {
				t.Fatalf("reset email has no link: %s", msgs[len(msgs)-1].Text)
			}
			return m[1]
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected %d %q email(s) to %s, got %d", n, subject, addr, len(box.about(addr, subject)))
	return ""
}

// ---------- login revocation ----------

func TestLoginRevocation(t *testing.T) {
	requireServer(t)
	register(t, "dave@example.com")
	phone := login(t, "dave@example.com", "password123")
	laptop := login(t, "dave@example.com", "password123")

	res := obj(must(t, 200, "PUT", "/api/profile/password", laptop,
		J{"current_password": "password123", "new_password": "newpassword1"}))
	fresh := str(res["token"])
	if fresh == "" || res["expires_at"] == nil {
		t.Fatalf("changing the password should return a new token: %v", res)
	}
	out := obj(must(t, 401, "GET", "/api/profile", phone, nil))
	if !strings.Contains(str(out["error"]), "revoked") {
		t.Fatalf("other devices should be logged out with a clear message: %v", out)
	}
	must(t, 401, "GET", "/api/profile", laptop, nil) // the old token of this device too
	must(t, 200, "GET", "/api/profile", fresh, nil)  // ...but the new token works

	second := login(t, "dave@example.com", "newpassword1")
	must(t, 200, "POST", "/api/logout-all", fresh, nil)
	must(t, 401, "GET", "/api/profile", fresh, nil)
	must(t, 401, "GET", "/api/profile", second, nil)
	must(t, 401, "POST", "/api/logout-all", "", nil)

	again := login(t, "dave@example.com", "newpassword1") // logging in again works
	must(t, 200, "GET", "/api/profile", again, nil)
}

// ---------- password reset ----------

func TestPasswordReset(t *testing.T) {
	requireServer(t)
	register(t, "erin@example.com")
	session := login(t, "erin@example.com", "password123")

	// Unknown and known emails get the SAME answer.
	unknown := obj(must(t, 200, "POST", "/api/password/forgot", "", J{"email": "nobody@example.com"}))
	known := obj(must(t, 200, "POST", "/api/password/forgot", "", J{"email": "  ERIN@example.com "}))
	if unknown["message"] != services.ForgotPasswordMessage || known["message"] != unknown["message"] {
		t.Fatalf("responses must not reveal whether an account exists: %v / %v", unknown, known)
	}
	must(t, 400, "POST", "/api/password/forgot", "", J{"email": "not-an-email"})

	token1 := waitForMail(t, outbox, "erin@example.com", resetSubject, 1)
	if len(outbox.about("nobody@example.com", resetSubject)) != 0 {
		t.Fatal("no email may be sent for an unknown address")
	}
	var leaked int
	testDB.QueryRow(`SELECT COUNT(*) FROM password_reset_tokens WHERE token_hash = $1`, token1).Scan(&leaked)
	if leaked != 0 {
		t.Fatal("the raw token must never be stored, only its hash")
	}

	// Requesting again invalidates the first link.
	must(t, 200, "POST", "/api/password/forgot", "", J{"email": "erin@example.com"})
	token2 := waitForMail(t, outbox, "erin@example.com", resetSubject, 2)
	must(t, 400, "POST", "/api/password/reset", "", J{"token": token1, "new_password": "brandnewpass1"})

	must(t, 400, "POST", "/api/password/reset", "", J{"token": "made-up", "new_password": "brandnewpass1"})
	weak := obj(must(t, 400, "POST", "/api/password/reset", "", J{"token": token2, "new_password": "short"}))
	if obj(weak["fields"])["new_password"] == nil {
		t.Fatalf("weak password should be reported on the field: %v", weak)
	}

	must(t, 200, "POST", "/api/password/reset", "", J{"token": token2, "new_password": "brandnewpass1"})
	must(t, 401, "GET", "/api/profile", session, nil) // logged out everywhere
	must(t, 401, "POST", "/api/login", "", J{"email": "erin@example.com", "password": "password123"})
	login(t, "erin@example.com", "brandnewpass1")
	must(t, 400, "POST", "/api/password/reset", "", J{"token": token2, "new_password": "anotherpass1"}) // single use

	// Expired links do not work.
	must(t, 200, "POST", "/api/password/forgot", "", J{"email": "erin@example.com"})
	token3 := waitForMail(t, outbox, "erin@example.com", resetSubject, 3)
	if _, err := testDB.Exec(`UPDATE password_reset_tokens SET expires_at = NOW() - INTERVAL '1 minute'
		WHERE user_id = (SELECT id FROM users WHERE email = 'erin@example.com') AND used_at IS NULL`); err != nil {
		t.Fatal(err)
	}
	must(t, 400, "POST", "/api/password/reset", "", J{"token": token3, "new_password": "anotherpass1"})

	// Deactivated accounts get no email.
	register(t, "frank@example.com")
	must(t, 200, "DELETE", "/api/profile", login(t, "frank@example.com", "password123"), nil)
	must(t, 200, "POST", "/api/password/forgot", "", J{"email": "frank@example.com"})
	time.Sleep(300 * time.Millisecond)
	if len(outbox.about("frank@example.com", resetSubject)) != 0 {
		t.Fatal("a deactivated account must not receive a reset link")
	}
}

// ---------- rate limiting ----------

// callFrom sends a request as if it came from `ip` (via X-Forwarded-For).
func callFrom(t *testing.T, base, method, path, ip string, body any) (int, http.Header, J) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, base+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", ip+", 10.255.255.1")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out J
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, res.Header, out
}

func TestRateLimits(t *testing.T) {
	requireServer(t)
	cfg := testConfig()
	cfg.TrustProxy = true // lets the test pretend to be different IPs
	cfg.RateLimits = config.RateLimits{
		Enabled:       true,
		General:       ratelimit.Rule{Requests: 1000, Per: time.Minute},
		Auth:          ratelimit.Rule{Requests: 3, Per: time.Minute},
		LoginFailures: ratelimit.Rule{Requests: 2, Per: 15 * time.Minute},
		ResetRequests: ratelimit.Rule{Requests: 1, Per: time.Hour},
	}
	box := &captureMailer{}
	app, err := routes.New(appDB, cfg, box)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(app.Handler)
	defer srv.Close()
	defer app.Shutdown()

	register(t, "gina@example.com")
	register(t, "hank@example.com")
	creds := func(email, pw string) J { return J{"email": email, "password": pw} }

	// Per IP: 3 account requests per minute, then 429 with Retry-After.
	// (A different email each time, so the per-email limit is not what trips.)
	for i := 0; i < 3; i++ {
		email := "nobody" + strconv.Itoa(i) + "@example.com"
		if s, _, _ := callFrom(t, srv.URL, "POST", "/api/login", "192.0.2.1", creds(email, "whatever1")); s == 429 {
			t.Fatalf("request %d should not be limited yet", i+1)
		}
	}
	status, header, body := callFrom(t, srv.URL, "POST", "/api/login", "192.0.2.1", creds("nobody9@example.com", "whatever1"))
	retry, _ := strconv.Atoi(header.Get("Retry-After"))
	if status != 429 || retry < 1 || body["error"] == nil {
		t.Fatalf("4th request: status %d, Retry-After %q, body %v; want 429 with a wait time", status, header.Get("Retry-After"), body)
	}
	if s, _, _ := callFrom(t, srv.URL, "POST", "/api/login", "192.0.2.2", creds("gina@example.com", "password123")); s != 200 {
		t.Fatalf("another IP must not be affected, got %d", s)
	}
	if s, _, _ := callFrom(t, srv.URL, "GET", "/api/products", "192.0.2.1", nil); s != 200 {
		t.Fatalf("normal pages are not covered by the strict account limit, got %d", s)
	}

	// Per email: 2 failed logins lock the account for a while, from ANY IP.
	callFrom(t, srv.URL, "POST", "/api/login", "198.51.100.1", creds("gina@example.com", "wrong-pass1"))
	callFrom(t, srv.URL, "POST", "/api/login", "198.51.100.2", creds("gina@example.com", "wrong-pass2"))
	status, _, body = callFrom(t, srv.URL, "POST", "/api/login", "198.51.100.3", creds("gina@example.com", "password123"))
	if status != 429 || !strings.Contains(str(body["error"]), "failed login attempts") {
		t.Fatalf("account should be locked after 2 failures even with the right password: %d %v", status, body)
	}

	// A successful login clears the failure counter.
	callFrom(t, srv.URL, "POST", "/api/login", "203.0.113.1", creds("hank@example.com", "wrong-pass1"))
	if s, _, _ := callFrom(t, srv.URL, "POST", "/api/login", "203.0.113.2", creds("hank@example.com", "password123")); s != 200 {
		t.Fatalf("login should still work after one failure, got %d", s)
	}
	if s, _, _ := callFrom(t, srv.URL, "POST", "/api/login", "203.0.113.3", creds("hank@example.com", "wrong-pass1")); s != 401 {
		t.Fatalf("counter should have been reset by the successful login, got %d", s)
	}

	// Reset emails per address are capped silently (same reply, no extra email).
	for _, ip := range []string{"203.0.113.10", "203.0.113.11"} {
		s, _, b := callFrom(t, srv.URL, "POST", "/api/password/forgot", ip, J{"email": "hank@example.com"})
		if s != 200 || b["message"] != services.ForgotPasswordMessage {
			t.Fatalf("forgot-password must always answer the same: %d %v", s, b)
		}
	}
	waitForMail(t, box, "hank@example.com", resetSubject, 1)
	app.Shutdown()
	if n := len(box.about("hank@example.com", resetSubject)); n != 1 {
		t.Fatalf("only 1 reset email per hour allowed, got %d", n)
	}
}

func mustLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
