// Package config reads every setting the application needs from environment
// variables, applies defaults, and refuses to start when something required is
// missing. Keeping this in one place means no other package calls os.Getenv.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds the validated application settings.
type Config struct {
	Port           string        // APP_PORT, default 8080
	DatabaseURL    string        // DATABASE_URL, required
	JWTSecret      string        // JWT_SECRET, required, at least 32 characters
	JWTTTL         time.Duration // JWT_TTL, default 24h
	AllowedOrigins []string      // CORS_ALLOWED_ORIGINS, comma separated
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
	return cfg, nil
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
