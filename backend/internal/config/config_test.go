package config

import (
	"strings"
	"testing"
	"time"
)

// allKeys are cleared before each test so the developer's shell cannot leak in.
var allKeys = []string{
	"APP_ENV", "LOG_FORMAT", "APP_PORT", "DATABASE_URL", "DB_MAX_CONNS", "JWT_SECRET", "JWT_TTL",
	"CORS_ALLOWED_ORIGINS", "TRUST_PROXY", "RATE_LIMIT_ENABLED", "TLS_MODE", "TLS_DOMAINS", "TLS_EMAIL",
	"TLS_CACHE_DIR", "TLS_CERT_FILE", "TLS_KEY_FILE", "TLS_PORT", "TLS_HTTP_PORT", "MAIL_DRIVER",
	"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM",
	"PASSWORD_RESET_URL", "PASSWORD_RESET_TTL", "STORE_NAME", "CURRENCY", "FRONTEND_URL", "EMAIL_VERIFY_URL",
	"EMAIL_VERIFY_TTL", "REQUIRE_EMAIL_VERIFICATION", "EMAIL_POLL_INTERVAL", "DEBT_REMINDER_INTERVAL",
	"PUBLIC_URL", "STORAGE_DRIVER", "UPLOAD_DIR", "MAX_UPLOAD_MB", "S3_ENDPOINT", "S3_REGION", "S3_BUCKET",
	"S3_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY", "S3_PUBLIC_URL", "S3_INSECURE",
}

func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for _, k := range allKeys {
		t.Setenv(k, "")
	}
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

// goodProduction is a complete, safe production setup.
func goodProduction() map[string]string {
	return map[string]string{
		"APP_ENV":              "production",
		"DATABASE_URL":         "postgres://shop_api:pw@db.internal.example.com:5432/ecommerce?sslmode=require",
		"JWT_SECRET":           "q8Zr1xV4mK0pT7wL2nB9cY6dF3hJ5sG8aE1uI4oP7zX0",
		"CORS_ALLOWED_ORIGINS": "https://shop.example.com",
		"TRUST_PROXY":          "true",
		"MAIL_DRIVER":          "smtp",
		"SMTP_HOST":            "smtp.example.com",
		"SMTP_FROM":            "Shop <no-reply@example.com>",
		"FRONTEND_URL":         "https://shop.example.com",
		"PUBLIC_URL":           "https://api.example.com",
	}
}

func with(base map[string]string, changes map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range changes {
		out[k] = v
	}
	return out
}

func TestDevelopmentDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL": "postgres://u:p@localhost/ecommerce?sslmode=disable",
		"JWT_SECRET":   "replace_with_at_least_32_random_characters", // allowed outside production
	})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Production() || cfg.LogFormat != "text" || cfg.TLS.Mode != TLSOff || cfg.DBMaxConns != 25 || len(cfg.Warnings) != 0 {
		t.Fatalf("unexpected development config: %+v", cfg)
	}
	if cfg.PasswordResetURL != "http://localhost:5173/reset-password" || cfg.EmailVerifyURL != "http://localhost:5173/verify-email" ||
		cfg.PublicURL != "http://localhost:8080" || cfg.Storage.Driver != "local" || cfg.MaxUploadBytes != 5<<20 ||
		!cfg.RequireEmailVerification || cfg.DebtReminderEvery != 168*time.Hour {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	t.Setenv("DEBT_REMINDER_INTERVAL", "0")
	if cfg, _ := Load(); cfg.DebtReminderEvery != 0 {
		t.Fatal("DEBT_REMINDER_INTERVAL=0 should turn reminders off")
	}
}

func TestGoodProductionConfig(t *testing.T) {
	setEnv(t, goodProduction())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Production() || cfg.LogFormat != "json" {
		t.Fatalf("production should default to JSON logs: %+v", cfg)
	}
	if len(cfg.Warnings) != 0 {
		t.Fatalf("a safe setup should give no warnings: %v", cfg.Warnings)
	}
}

func TestProductionRefusesUnsafeSettings(t *testing.T) {
	cases := map[string]struct {
		changes map[string]string
		want    string
	}{
		"placeholder secret": {map[string]string{"JWT_SECRET": "replace_with_at_least_32_random_characters"}, "example value"},
		"predictable secret": {map[string]string{"JWT_SECRET": strings.Repeat("ab", 20)}, "too predictable"},
		"wildcard origin":    {map[string]string{"CORS_ALLOWED_ORIGINS": "*"}, "must not be *"},
		"http origin":        {map[string]string{"CORS_ALLOWED_ORIGINS": "http://shop.example.com"}, "must use https"},
		"local origin":       {map[string]string{"CORS_ALLOWED_ORIGINS": "https://localhost:5173"}, "local address"},
		"http reset link":    {map[string]string{"PASSWORD_RESET_URL": "http://shop.example.com/reset-password"}, "PASSWORD_RESET_URL"},
		"local frontend":     {map[string]string{"FRONTEND_URL": "http://localhost:5173"}, "FRONTEND_URL"},
		"http image host":    {map[string]string{"PUBLIC_URL": "http://api.example.com"}, "PUBLIC_URL"},
		"s3 incomplete":      {map[string]string{"STORAGE_DRIVER": "s3"}, "S3_ENDPOINT"},
		"log mail driver":    {map[string]string{"MAIL_DRIVER": "log"}, "MAIL_DRIVER must be smtp"},
		"rate limits off":    {map[string]string{"RATE_LIMIT_ENABLED": "false"}, "RATE_LIMIT_ENABLED"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setEnv(t, with(goodProduction(), c.changes))
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want an error containing %q, got %v", c.want, err)
			}
		})
	}
}

func TestEveryProblemIsReportedAtOnce(t *testing.T) {
	setEnv(t, with(goodProduction(), map[string]string{"MAIL_DRIVER": "log", "RATE_LIMIT_ENABLED": "false"}))
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "MAIL_DRIVER") || !strings.Contains(err.Error(), "RATE_LIMIT_ENABLED") {
		t.Fatalf("both problems should be listed: %v", err)
	}
}

func TestProductionWarnings(t *testing.T) {
	setEnv(t, with(goodProduction(), map[string]string{
		"DATABASE_URL": "postgres://postgres:pw@db.example.com/ecommerce?sslmode=disable",
		"TRUST_PROXY":  "false",
	}))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(cfg.Warnings, "\n")
	for _, want := range []string{"sslmode=disable", "superuser", "TRUST_PROXY"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing warning about %s in:\n%s", want, all)
		}
	}
}

func TestDockerDatabaseIsNotWarnedAbout(t *testing.T) {
	setEnv(t, with(goodProduction(), map[string]string{
		"DATABASE_URL": "postgres://shop_api:pw@db:5432/ecommerce?sslmode=disable", // the Compose service
	}))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Warnings) != 0 {
		t.Fatalf("an internal Docker database needs no TLS warning: %v", cfg.Warnings)
	}
}

func TestOriginsAreNormalized(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL":         "postgres://u:p@localhost/x",
		"JWT_SECRET":           strings.Repeat("x", 32),
		"CORS_ALLOWED_ORIGINS": "https://Shop.Example.com/, http://localhost:5173",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(cfg.AllowedOrigins, ","); got != "https://shop.example.com,http://localhost:5173" {
		t.Fatalf("got %s", got)
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://shop.example.com/app")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "no path") {
		t.Fatalf("an origin with a path should be refused, got %v", err)
	}
}

func TestTLSSettings(t *testing.T) {
	base := map[string]string{"DATABASE_URL": "postgres://u:p@localhost/x", "JWT_SECRET": strings.Repeat("x", 32)}

	setEnv(t, with(base, map[string]string{"TLS_MODE": "auto", "TLS_DOMAINS": "API.example.com"}))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TLS.Enabled() || cfg.TLS.Domains[0] != "api.example.com" || cfg.TLS.HTTPSPort != "443" || cfg.TLS.HTTPPort != "80" || cfg.TLS.CacheDir != "certs" {
		t.Fatalf("unexpected TLS config: %+v", cfg.TLS)
	}

	bad := map[string]map[string]string{
		"auto without domains": {"TLS_MODE": "auto"},
		"auto with an IP":      {"TLS_MODE": "auto", "TLS_DOMAINS": "203.0.113.5"},
		"auto without port 80": {"TLS_MODE": "auto", "TLS_DOMAINS": "api.example.com", "TLS_HTTP_PORT": "off"},
		"files without files":  {"TLS_MODE": "files"},
		"unknown mode":         {"TLS_MODE": "on"},
	}
	for name, changes := range bad {
		t.Run(name, func(t *testing.T) {
			setEnv(t, with(base, changes))
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "TLS_") {
				t.Fatalf("want a TLS error, got %v", err)
			}
		})
	}
}
