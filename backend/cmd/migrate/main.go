// Command migrate applies database migrations without needing psql.
//
//	go run ./cmd/migrate status   # show which migrations are applied
//	go run ./cmd/migrate up       # apply the pending ones
//
// It reads DATABASE_URL from the environment or from .env, and the SQL files
// from ./migrations (override with MIGRATIONS_DIR).
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/migrate"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return errors.New("DATABASE_URL is required")
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}
	cmd := "status"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer db.Close()

	switch cmd {
	case "up":
		applied, err := migrate.Up(ctx, db, dir)
		for _, v := range applied {
			fmt.Println("applied", v)
		}
		if err != nil {
			return err
		}
		if len(applied) == 0 {
			fmt.Println("database is up to date")
		}
		return nil
	case "status":
		list, err := migrate.Status(ctx, db, dir)
		if err != nil {
			return err
		}
		for _, m := range list {
			state := "pending"
			if m.Applied {
				state = "applied"
			}
			fmt.Printf("%-8s %s\n", state, m.Version)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q (use: up, status)", cmd)
	}
}
