// Command api starts the HTTP server.
//
// Start-up order: load .env (optional) -> read and validate config -> connect to
// PostgreSQL -> build routes -> listen. On Ctrl+C / SIGTERM it stops accepting
// new connections and gives running requests up to 10 seconds to finish.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           routes.New(db, cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("API listening", "addr", srv.Addr, "cors_origins", cfg.AllowedOrigins)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
