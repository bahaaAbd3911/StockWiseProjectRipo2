// Package config reads every setting the application needs from environment
// variables, applies defaults, and refuses to start when something required is
// missing. Keeping this in one place means no other package calls os.Getenv.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/ratelimit"
)

// Config holds the validated application settings.
type Config struct {
	Port           string        // APP_PORT, default 8080
	DatabaseURL    string        // DATABASE_URL, required
	JWTSecret      string        // JWT_SECRET, required, at least 32 characters
	JWTTTL         time.Duration // JWT_TTL, default 24h
	AllowedOrigins []string      // CORS_ALLOWED_ORIGINS, comma separated

	// TrustProxy: take the client IP from X-Forwarded-For / X-Real-IP.
	// Turn on ONLY behind a reverse proxy that sets these headers, otherwise
	// clients could fake their IP and dodge the rate limits. TRUST_PROXY.
	TrustProxy bool

	RateLimits RateLimits
	Mail       mail.Config

	// PasswordResetURL is the frontend page that receives ?token=...
	// PASSWORD_RESET_URL, default http://localhost:5173/reset-password
	PasswordResetURL string
	PasswordResetTTL time.Duration // PASSWORD_RESET_TTL, default 30m
}

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
	cfg := Config{
		Port:        getenv("APP_PORT", "8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < minJWTSecretLength {
		return Config{}, fmt.Errorf("JWT_SECRET is required and must be at least %d characters", minJWTSecretLength)
	}

	ttl, err := time.ParseDuration(getenv("JWT_TTL", "24h"))
	if err != nil || ttl <= 0 {
		return Config{}, errors.New("JWT_TTL must be a positive duration such as 24h or 90m")
	}
	cfg.JWTTTL = ttl

	cfg.AllowedOrigins = splitList(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"))

	if cfg.TrustProxy, err = getbool("TRUST_PROXY", false); err != nil {
		return Config{}, err
	}
	cfg.RateLimits = DefaultRateLimits()
	if cfg.RateLimits.Enabled, err = getbool("RATE_LIMIT_ENABLED", true); err != nil {
		return Config{}, err
	}

	cfg.Mail = mail.Config{
		Driver:   strings.ToLower(getenv("MAIL_DRIVER", "log")),
		Host:     getenv("SMTP_HOST", ""),
		Port:     getenv("SMTP_PORT", "587"),
		Username: getenv("SMTP_USERNAME", ""),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     getenv("SMTP_FROM", ""),
	}

	cfg.PasswordResetURL = getenv("PASSWORD_RESET_URL", "http://localhost:5173/reset-password")
	if u, err := url.Parse(cfg.PasswordResetURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Config{}, errors.New("PASSWORD_RESET_URL must be an http(s) URL")
	}
	resetTTL, err := time.ParseDuration(getenv("PASSWORD_RESET_TTL", "30m"))
	if err != nil || resetTTL < time.Minute || resetTTL > 24*time.Hour {
		return Config{}, errors.New("PASSWORD_RESET_TTL must be between 1m and 24h")
	}
	cfg.PasswordResetTTL = resetTTL
	return cfg, nil
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
