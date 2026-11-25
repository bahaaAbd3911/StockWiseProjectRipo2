// Package config reads every setting the application needs from environment
// variables, applies defaults, and refuses to start when something required is
// missing. Keeping this in one place means no other package calls os.Getenv.
//
// With APP_ENV=production it also refuses settings that are fine on a laptop
// but unsafe on the internet (a placeholder JWT secret, localhost origins,
// emails printed to the log, rate limits off...). Every problem is reported at
// once, so a broken deploy can be fixed in one go.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/ratelimit"
	"ecommerce-backend/internal/storage"

	_ "time/tzdata" // time zone data built into the binary (minimal Docker images have none)
)

// Config holds the validated application settings.
type Config struct {
	// Env is "development" (default) or "production". APP_ENV.
	Env string
	// LogFormat is "text" (default in development) or "json" (default in
	// production, for log collectors). LOG_FORMAT.
	LogFormat string

	Port           string        // APP_PORT, default 8080 (plain HTTP; unused when TLS is on)
	DatabaseURL    string        // DATABASE_URL, required
	DBMaxConns     int           // DB_MAX_CONNS, default 25; keep below the database's connection limit
	JWTSecret      string        // JWT_SECRET, required, at least 32 characters
	JWTTTL         time.Duration // JWT_TTL, default 24h
	AllowedOrigins []string      // CORS_ALLOWED_ORIGINS, comma separated

	// TrustProxy: take the client IP from X-Forwarded-For / X-Real-IP.
	// Turn on ONLY behind a reverse proxy that sets these headers, otherwise
	// clients could fake their IP and dodge the rate limits. TRUST_PROXY.
	TrustProxy bool

	TLS        TLS
	RateLimits RateLimits
	Mail       mail.Config

	// The shop as customers see it (used in emails).
	StoreName   string // STORE_NAME, default StockWise
	FrontendURL string // FRONTEND_URL, default http://localhost:5173
	Currency    string // CURRENCY, printed after amounts in emails, e.g. JOD (default none)

	// Stores (multi-tenant). PlatformDomain: stores live at <slug>.PLATFORM_DOMAIN
	// ("" = no subdomain stores). DefaultStore: the store for every other host
	// (localhost, an IP, the bare domain); "" = such hosts get 404. A single-shop
	// setup keeps the default "main"; a platform sets PLATFORM_DOMAIN and
	// DEFAULT_STORE= (empty) so unknown hosts are refused.
	PlatformDomain string // PLATFORM_DOMAIN, default ""
	DefaultStore   string // DEFAULT_STORE, default main
	// PlatformAdminHost is where the platform super admin works (the
	// /api/platform API only answers there). PLATFORM_ADMIN_HOST, default
	// admin.PLATFORM_DOMAIN; "" = platform admin off.
	PlatformAdminHost string
	// StoreURL is a store's public frontend address, with {slug}, e.g.
	// https://{slug}.myplatform.com (used in emails). STORE_URL, default
	// FRONTEND_URL's scheme and port with {slug}.PLATFORM_DOMAIN.
	StoreURL string

	// Admin dashboard.
	StoreTimezone     *time.Location // STORE_TIMEZONE, default Asia/Amman: when a "day" starts
	LowStockThreshold int            // LOW_STOCK_THRESHOLD, default 5: products at or below are "low stock"

	// PasswordResetURL is the frontend page that receives ?token=...
	// PASSWORD_RESET_URL, default FRONTEND_URL/reset-password
	PasswordResetURL string
	PasswordResetTTL time.Duration // PASSWORD_RESET_TTL, default 30m

	// Email verification at sign-up.
	EmailVerifyURL           string        // EMAIL_VERIFY_URL, default FRONTEND_URL/verify-email
	EmailVerifyTTL           time.Duration // EMAIL_VERIFY_TTL, default 48h (1h - 168h)
	RequireEmailVerification bool          // REQUIRE_EMAIL_VERIFICATION, default true: checkout needs a verified email

	// Notification emails (queued in email_outbox).
	EmailPollInterval time.Duration // EMAIL_POLL_INTERVAL, default 5s
	DebtReminderEvery time.Duration // DEBT_REMINDER_INTERVAL, default 168h (weekly); 0 = off

	// Uploads (product and profile images).
	PublicURL      string         // PUBLIC_URL: the API's own public address, default http://localhost:APP_PORT
	Storage        storage.Config // STORAGE_DRIVER, UPLOAD_DIR, S3_*
	MaxUploadBytes int64          // MAX_UPLOAD_MB, default 5

	// Warnings are settings that work but are probably wrong. They are logged
	// at start-up and never stop the server.
	Warnings []string
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Production reports whether APP_ENV=production.
func (c Config) Production() bool { return c.Env == EnvProduction }

// TLS selects who handles HTTPS.
//
//	off   (default) a reverse proxy or the hosting platform terminates HTTPS
//	      and forwards plain HTTP to APP_PORT
//	auto  the API gets free certificates from Let's Encrypt for TLS_DOMAINS
//	      and renews them itself; needs ports 80 and 443 reachable
//	files the API uses TLS_CERT_FILE / TLS_KEY_FILE
//
// With auto/files the API listens on TLS_PORT (443) for HTTPS and on
// TLS_HTTP_PORT (80), where it redirects to HTTPS (and, for auto, answers
// Let's Encrypt's checks).
type TLS struct {
	Mode      string   // TLS_MODE
	Domains   []string // TLS_DOMAINS, comma separated (auto)
	Email     string   // TLS_EMAIL, optional: Let's Encrypt expiry notices (auto)
	CacheDir  string   // TLS_CACHE_DIR, default "certs": where certificates are kept (auto)
	CertFile  string   // TLS_CERT_FILE (files)
	KeyFile   string   // TLS_KEY_FILE (files)
	HTTPSPort string   // TLS_PORT, default 443
	HTTPPort  string   // TLS_HTTP_PORT, default 80; "off" disables the redirect listener
}

const (
	TLSOff   = "off"
	TLSAuto  = "auto"
	TLSFiles = "files"
)

// Enabled reports whether the API serves HTTPS itself.
func (t TLS) Enabled() bool { return t.Mode == TLSAuto || t.Mode == TLSFiles }

// RateLimits groups every limit. Enabled=false (the zero value) switches all
// of them off, which is what the tests use.
type RateLimits struct {
	Enabled       bool
	General       ratelimit.Rule // per IP, every request
	Auth          ratelimit.Rule // per IP, login/register/password endpoints
	LoginFailures ratelimit.Rule // per email, failed logins only
	ResetRequests ratelimit.Rule // per email, "forgot password" requests
}

// DefaultRateLimits are the production limits.
func DefaultRateLimits() RateLimits {
	return RateLimits{
		Enabled:       true,
		General:       ratelimit.Rule{Requests: 300, Per: time.Minute},
		Auth:          ratelimit.Rule{Requests: 10, Per: time.Minute},
		LoginFailures: ratelimit.Rule{Requests: 5, Per: 15 * time.Minute},
		ResetRequests: ratelimit.Rule{Requests: 3, Per: time.Hour},
	}
}

const minJWTSecretLength = 32

// Load builds a Config from the environment.
func Load() (Config, error) {
	var problems []string
	bad := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	cfg := Config{
		Env:         strings.ToLower(getenv("APP_ENV", EnvDevelopment)),
		Port:        getenv("APP_PORT", "8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
	if cfg.Env != EnvDevelopment && cfg.Env != EnvProduction {
		bad("APP_ENV must be development or production")
	}

	cfg.LogFormat = strings.ToLower(getenv("LOG_FORMAT", map[bool]string{true: "json", false: "text"}[cfg.Production()]))
	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		bad("LOG_FORMAT must be text or json")
	}

	if cfg.DatabaseURL == "" {
		bad("DATABASE_URL is required")
	}
	var err error
	if cfg.DBMaxConns, err = getint("DB_MAX_CONNS", 25); err != nil || cfg.DBMaxConns < 1 || cfg.DBMaxConns > 500 {
		bad("DB_MAX_CONNS must be a number between 1 and 500")
	}
	if len(cfg.JWTSecret) < minJWTSecretLength {
		bad("JWT_SECRET is required and must be at least %d characters", minJWTSecretLength)
	}

	if ttl, err := time.ParseDuration(getenv("JWT_TTL", "24h")); err != nil || ttl <= 0 {
		bad("JWT_TTL must be a positive duration such as 24h or 90m")
	} else {
		cfg.JWTTTL = ttl
	}

	for _, o := range splitList(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")) {
		origin, err := normalizeOrigin(o)
		if err != nil {
			bad("CORS_ALLOWED_ORIGINS: %v", err)
			continue
		}
		cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin)
	}

	if cfg.TrustProxy, err = getbool("TRUST_PROXY", false); err != nil {
		bad("%v", err)
	}
	cfg.RateLimits = DefaultRateLimits()
	if cfg.RateLimits.Enabled, err = getbool("RATE_LIMIT_ENABLED", true); err != nil {
		bad("%v", err)
	}

	cfg.TLS = TLS{
		Mode:      strings.ToLower(getenv("TLS_MODE", TLSOff)),
		Domains:   splitList(strings.ToLower(getenv("TLS_DOMAINS", ""))),
		Email:     getenv("TLS_EMAIL", ""),
		CacheDir:  getenv("TLS_CACHE_DIR", "certs"),
		CertFile:  getenv("TLS_CERT_FILE", ""),
		KeyFile:   getenv("TLS_KEY_FILE", ""),
		HTTPSPort: getenv("TLS_PORT", "443"),
		HTTPPort:  strings.ToLower(getenv("TLS_HTTP_PORT", "80")),
	}
	switch cfg.TLS.Mode {
	case TLSOff:
	case TLSAuto:
		if len(cfg.TLS.Domains) == 0 {
			bad("TLS_MODE=auto needs TLS_DOMAINS (for example api.example.com)")
		}
		for _, d := range cfg.TLS.Domains {
			if strings.ContainsAny(d, "/:*") || net.ParseIP(d) != nil || !strings.Contains(d, ".") {
				bad("TLS_DOMAINS: %q is not a domain name (Let's Encrypt cannot issue certificates for IPs, ports or wildcards here)", d)
			}
		}
		if cfg.TLS.HTTPPort == "off" {
			bad("TLS_MODE=auto needs TLS_HTTP_PORT (80): Let's Encrypt checks the domain over HTTP")
		}
	case TLSFiles:
		if cfg.TLS.CertFile == "" || cfg.TLS.KeyFile == "" {
			bad("TLS_MODE=files needs TLS_CERT_FILE and TLS_KEY_FILE")
		}
	default:
		bad("TLS_MODE must be off, auto or files")
	}

	cfg.Mail = mail.Config{
		Driver:   strings.ToLower(getenv("MAIL_DRIVER", "log")),
		Host:     getenv("SMTP_HOST", ""),
		Port:     getenv("SMTP_PORT", "587"),
		Username: getenv("SMTP_USERNAME", ""),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     getenv("SMTP_FROM", ""),
	}

	cfg.PlatformDomain = strings.ToLower(strings.Trim(getenv("PLATFORM_DOMAIN", ""), "."))
	if cfg.PlatformDomain != "" && (strings.ContainsAny(cfg.PlatformDomain, "/: ") || !strings.Contains(cfg.PlatformDomain, ".") && cfg.PlatformDomain != "localhost") {
		bad("PLATFORM_DOMAIN must be a domain name such as myplatform.com (no scheme, port or path)")
	}
	cfg.DefaultStore = strings.ToLower(os.Getenv("DEFAULT_STORE"))
	if _, set := os.LookupEnv("DEFAULT_STORE"); !set {
		cfg.DefaultStore = "main"
	}

	// Empty = admin.PLATFORM_DOMAIN; "off" = no platform admin.
	cfg.PlatformAdminHost = strings.ToLower(getenv("PLATFORM_ADMIN_HOST", ""))
	switch {
	case cfg.PlatformAdminHost == "off":
		cfg.PlatformAdminHost = ""
	case cfg.PlatformAdminHost == "" && cfg.PlatformDomain != "":
		cfg.PlatformAdminHost = "admin." + cfg.PlatformDomain
	}

	cfg.StoreName = getenv("STORE_NAME", "StockWise")
	cfg.Currency = getenv("CURRENCY", "")
	if loc, err := time.LoadLocation(getenv("STORE_TIMEZONE", "Asia/Amman")); err != nil {
		bad("STORE_TIMEZONE must be a time zone name such as Asia/Amman, Asia/Riyadh or UTC")
	} else {
		cfg.StoreTimezone = loc
	}
	if n, err := getint("LOW_STOCK_THRESHOLD", 5); err != nil || n < 0 || n > 10000 {
		bad("LOW_STOCK_THRESHOLD must be a number between 0 and 10000")
	} else {
		cfg.LowStockThreshold = n
	}
	cfg.FrontendURL = strings.TrimRight(getenv("FRONTEND_URL", "http://localhost:5173"), "/")
	if !storage.ValidPublicURL(cfg.FrontendURL) {
		bad("FRONTEND_URL must be an http(s) URL such as https://shop.example.com")
	}
	cfg.StoreURL = strings.TrimRight(getenv("STORE_URL", ""), "/")
	if cfg.StoreURL == "" && cfg.PlatformDomain != "" {
		if u, err := url.Parse(cfg.FrontendURL); err == nil {
			port := ""
			if p := u.Port(); p != "" {
				port = ":" + p
			}
			cfg.StoreURL = u.Scheme + "://{slug}." + cfg.PlatformDomain + port
		}
	}
	if cfg.StoreURL != "" && (!strings.Contains(cfg.StoreURL, "{slug}") || !storage.ValidPublicURL(strings.ReplaceAll(cfg.StoreURL, "{slug}", "x"))) {
		bad("STORE_URL must be an http(s) address containing {slug}, e.g. https://{slug}.myplatform.com")
	}

	cfg.PasswordResetURL = getenv("PASSWORD_RESET_URL", cfg.FrontendURL+"/reset-password")
	if !storage.ValidPublicURL(cfg.PasswordResetURL) {
		bad("PASSWORD_RESET_URL must be an http(s) URL")
	}
	if resetTTL, err := time.ParseDuration(getenv("PASSWORD_RESET_TTL", "30m")); err != nil || resetTTL < time.Minute || resetTTL > 24*time.Hour {
		bad("PASSWORD_RESET_TTL must be between 1m and 24h")
	} else {
		cfg.PasswordResetTTL = resetTTL
	}

	cfg.EmailVerifyURL = getenv("EMAIL_VERIFY_URL", cfg.FrontendURL+"/verify-email")
	if !storage.ValidPublicURL(cfg.EmailVerifyURL) {
		bad("EMAIL_VERIFY_URL must be an http(s) URL")
	}
	if ttl, err := time.ParseDuration(getenv("EMAIL_VERIFY_TTL", "48h")); err != nil || ttl < time.Hour || ttl > 7*24*time.Hour {
		bad("EMAIL_VERIFY_TTL must be between 1h and 168h")
	} else {
		cfg.EmailVerifyTTL = ttl
	}
	if cfg.RequireEmailVerification, err = getbool("REQUIRE_EMAIL_VERIFICATION", true); err != nil {
		bad("%v", err)
	}

	if d, err := time.ParseDuration(getenv("EMAIL_POLL_INTERVAL", "5s")); err != nil || d < 100*time.Millisecond || d > 10*time.Minute {
		bad("EMAIL_POLL_INTERVAL must be between 100ms and 10m")
	} else {
		cfg.EmailPollInterval = d
	}
	if v := getenv("DEBT_REMINDER_INTERVAL", "168h"); v != "0" {
		if d, err := time.ParseDuration(v); err != nil || d < time.Hour {
			bad("DEBT_REMINDER_INTERVAL must be at least 1h (for example 168h for weekly), or 0 to turn reminders off")
		} else {
			cfg.DebtReminderEvery = d
		}
	}

	cfg.PublicURL = strings.TrimRight(getenv("PUBLIC_URL", "http://localhost:"+cfg.Port), "/")
	if !storage.ValidPublicURL(cfg.PublicURL) {
		bad("PUBLIC_URL must be the API's public address, e.g. https://api.example.com")
	}
	insecureS3, err := getbool("S3_INSECURE", false)
	if err != nil {
		bad("%v", err)
	}
	cfg.Storage = storage.Config{
		Driver:    strings.ToLower(getenv("STORAGE_DRIVER", "local")),
		Dir:       getenv("UPLOAD_DIR", "data/uploads"),
		PublicURL: cfg.PublicURL,
		S3: storage.S3Config{
			Endpoint:        getenv("S3_ENDPOINT", ""),
			Region:          getenv("S3_REGION", ""),
			Bucket:          getenv("S3_BUCKET", ""),
			AccessKeyID:     getenv("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
			PublicURL:       strings.TrimRight(getenv("S3_PUBLIC_URL", ""), "/"),
			Insecure:        insecureS3,
		},
	}
	switch cfg.Storage.Driver {
	case "local":
	case "s3":
		if s := cfg.Storage.S3; s.Endpoint == "" || s.Bucket == "" || s.AccessKeyID == "" || s.SecretAccessKey == "" || !storage.ValidPublicURL(s.PublicURL) {
			bad("STORAGE_DRIVER=s3 needs S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY and S3_PUBLIC_URL")
		}
	default:
		bad("STORAGE_DRIVER must be local or s3")
	}
	if mb, err := getint("MAX_UPLOAD_MB", 5); err != nil || mb < 1 || mb > 50 {
		bad("MAX_UPLOAD_MB must be between 1 and 50")
	} else {
		cfg.MaxUploadBytes = int64(mb) << 20
	}

	if len(problems) == 0 && cfg.Production() {
		var warnings []string
		problems, warnings = checkProduction(cfg)
		cfg.Warnings = append(cfg.Warnings, warnings...)
	}
	if len(problems) > 0 {
		return Config{}, Error{Problems: problems}
	}
	return cfg, nil
}

// Error lists every invalid setting found by Load.
type Error struct{ Problems []string }

func (e Error) Error() string {
	return "invalid configuration:\n  - " + strings.Join(e.Problems, "\n  - ")
}

// placeholderSecrets are values from the examples that must never reach a
// real server.
var placeholderSecrets = []string{"replace_with", "change_me", "changeme", "your_secret", "secret_key"}

// checkProduction returns the settings that are unsafe on the internet
// (problems: the server refuses to start) and the ones that are probably
// wrong (warnings: logged).
func checkProduction(cfg Config) (problems, warnings []string) {
	bad := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	warn := func(format string, args ...any) { warnings = append(warnings, fmt.Sprintf(format, args...)) }

	// Login tokens: anyone who knows the secret can sign in as any user.
	lower := strings.ToLower(cfg.JWTSecret)
	for _, p := range placeholderSecrets {
		if strings.Contains(lower, p) {
			bad("JWT_SECRET is still the example value; generate a random one (see docs/production.md)")
			break
		}
	}
	if distinct(cfg.JWTSecret) < 12 {
		bad("JWT_SECRET is too predictable; generate a random one (see docs/production.md)")
	}
	if cfg.JWTTTL > 7*24*time.Hour {
		warn("JWT_TTL is %s; a stolen token stays valid that long (24h is usual)", cfg.JWTTTL)
	}

	// Browsers: only the real frontend, only over HTTPS.
	if len(cfg.AllowedOrigins) == 0 {
		bad("CORS_ALLOWED_ORIGINS must list the frontend's address, e.g. https://shop.example.com")
	}
	for _, o := range cfg.AllowedOrigins {
		switch {
		case o == "*":
			bad("CORS_ALLOWED_ORIGINS must not be * in production; list the frontend's address")
		case !strings.HasPrefix(o, "https://"):
			bad("CORS_ALLOWED_ORIGINS: %s must use https://", o)
		case isLocal(hostOf(o)):
			bad("CORS_ALLOWED_ORIGINS: %s is a local address; use the frontend's public address", o)
		}
	}

	// Links in emails point at the frontend. Reset and verification links
	// carry a secret, so: HTTPS, a public host, and the frontend's domain.
	for _, link := range []struct{ name, value string }{
		{"FRONTEND_URL", cfg.FrontendURL},
		{"PASSWORD_RESET_URL", cfg.PasswordResetURL},
		{"EMAIL_VERIFY_URL", cfg.EmailVerifyURL},
	} {
		if u, _ := url.Parse(link.value); u.Scheme != "https" || isLocal(u.Hostname()) {
			bad("%s must be a public https:// address of the frontend (got %s)", link.name, link.value)
		} else if !contains(cfg.AllowedOrigins, u.Scheme+"://"+u.Host) {
			warn("%s (%s) is not on a domain listed in CORS_ALLOWED_ORIGINS; check it points at your frontend", link.name, u.Host)
		}
	}

	// Uploaded images are loaded by browsers from these addresses.
	imageBase, imageVar := cfg.PublicURL, "PUBLIC_URL"
	if cfg.Storage.Driver == "s3" {
		imageBase, imageVar = cfg.Storage.S3.PublicURL, "S3_PUBLIC_URL"
	}
	if u, _ := url.Parse(imageBase); u.Scheme != "https" || isLocal(u.Hostname()) {
		bad("%s must be a public https:// address (uploaded images are served from it; got %s)", imageVar, imageBase)
	}
	if cfg.Storage.S3.Insecure {
		bad("S3_INSECURE must be false in production")
	}

	// Email: the log driver would print reset links into the logs.
	if cfg.Mail.Driver != "smtp" {
		bad("MAIL_DRIVER must be smtp in production (the log driver writes password reset links into the logs)")
	}

	if !cfg.RateLimits.Enabled {
		bad("RATE_LIMIT_ENABLED must be true in production")
	}

	// HTTPS: either the API does it (TLS_MODE) or something in front of it.
	switch {
	case cfg.TLS.Enabled() && cfg.TrustProxy:
		warn("TLS_MODE=%s and TRUST_PROXY=true: if nothing sits in front of the API, clients can fake their IP; turn TRUST_PROXY off", cfg.TLS.Mode)
	case !cfg.TLS.Enabled() && !cfg.TrustProxy:
		warn("TLS_MODE=off and TRUST_PROXY=false: make sure a proxy or the platform serves HTTPS; if one does, set TRUST_PROXY=true, otherwise every visitor shares the proxy's rate limit")
	}

	// Database: encrypted connection and its own user.
	if u, err := url.Parse(cfg.DatabaseURL); err == nil {
		if u.Query().Get("sslmode") == "disable" && !isLocal(u.Hostname()) {
			warn("DATABASE_URL has sslmode=disable for a remote server: passwords and data travel unencrypted; use sslmode=require")
		}
		if u.User != nil && u.User.Username() == "postgres" {
			warn("DATABASE_URL uses the postgres superuser; create a dedicated user for the API (see docs/production.md)")
		}
	}

	// Stores: on a platform, a host that is no store should get 404, not a shop.
	if cfg.PlatformDomain != "" && cfg.DefaultStore != "" {
		warn("PLATFORM_DOMAIN is set and DEFAULT_STORE=%s: every unknown host name shows that store; set DEFAULT_STORE= (empty) to refuse them", cfg.DefaultStore)
	}
	return problems, warnings
}

// normalizeOrigin checks one CORS origin and strips a trailing slash (browsers
// send "https://shop.example.com", never with a slash, so a slash in the
// setting would silently block every request).
func normalizeOrigin(o string) (string, error) {
	if o == "*" {
		return o, nil
	}
	o = strings.TrimRight(o, "/")
	u, err := url.Parse(o)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" || u.RawQuery != "" {
		return "", fmt.Errorf("%q must look like https://shop.example.com (scheme and host only, no path)", o)
	}
	return strings.ToLower(u.Scheme + "://" + u.Host), nil
}

func hostOf(origin string) string {
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// isLocal reports addresses that only work on the same machine or network.
func isLocal(host string) bool {
	host = strings.ToLower(host)
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return true
	}
	// A name without a dot ("db", "postgres") is a Docker / internal network
	// name, never a public internet address.
	if net.ParseIP(host) == nil && !strings.Contains(host, ".") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast())
}

func distinct(s string) int {
	seen := map[rune]bool{}
	for _, r := range s {
		seen[r] = true
	}
	return len(seen)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func getbool(key string, fallback bool) (bool, error) {
	v := getenv(key, "")
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return b, nil
}

func getint(key string, fallback int) (int, error) {
	v := getenv(key, "")
	if v == "" {
		return fallback, nil
	}
	return strconv.Atoi(v)
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
