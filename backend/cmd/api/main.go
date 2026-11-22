// Command api starts the HTTP server.
//
// Start-up order: load .env (optional) -> read and validate config (in
// production, unsafe settings stop it here) -> connect to PostgreSQL -> build
// routes -> listen (plain HTTP, or HTTPS when TLS_MODE is set). On Ctrl+C /
// SIGTERM it stops accepting new connections and gives running requests up to
// 10 seconds to finish.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/routes"
	"ecommerce-backend/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// "api healthcheck" asks the running server whether it is healthy (exit 0)
	// or not (exit 1). Docker's HEALTHCHECK uses it: the image has no curl.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		var cfgErr config.Error
		if errors.As(err, &cfgErr) {
			fmt.Fprintln(os.Stderr, err) // a readable list, one problem per line
		} else {
			slog.Error("server stopped", "error", err)
		}
		os.Exit(1)
	}
}

func run() error {
	// A local .env file is optional; in production the variables are usually
	// provided by the hosting platform instead.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("could not read .env file", "error", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	for _, w := range cfg.Warnings {
		slog.Warn("config: " + w)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.DBMaxConns)
	db.SetMaxIdleConns(min(10, cfg.DBMaxConns))

	mailer, err := mail.New(cfg.Mail)
	if err != nil {
		return err
	}
	app, err := routes.New(db, cfg, mailer)
	if err != nil {
		return err
	}

	slog.Info("starting", "env", cfg.Env, "cors_origins", cfg.AllowedOrigins,
		"rate_limits", cfg.RateLimits.Enabled, "mail_driver", cfg.Mail.Driver, "trust_proxy", cfg.TrustProxy)
	err = server.Run(ctx, cfg, app.Handler)
	app.Shutdown() // finish account emails being sent; stop the email worker
	return err
}

func healthcheck() int {
	url := "http://127.0.0.1:" + envOr("APP_PORT", "8080") + "/api/health"
	client := &http.Client{Timeout: 3 * time.Second}
	if m := os.Getenv("TLS_MODE"); m == config.TLSAuto || m == config.TLSFiles {
		// Same machine: the certificate is for the public domain, not 127.0.0.1.
		url = "https://127.0.0.1:" + envOr("TLS_PORT", "443") + "/api/health"
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} //nolint:gosec // loopback only
	}
	res, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unhealthy:", err)
		return 1
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "unhealthy: status", res.StatusCode)
		return 1
	}
	return 0
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
