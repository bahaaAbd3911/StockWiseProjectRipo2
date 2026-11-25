// Command platform manages platform super admins (the people who create and
// run stores). There is no sign-up for them on purpose: they are created here,
// on the server.
//
//	go run ./cmd/platform create-admin -email you@example.com -first Sara -last Ahmad
//
// A strong random password is generated and printed once. Store it in a
// password manager; it is not shown again.
//
// Reads DATABASE_URL from the environment or from .env.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] != "create-admin" {
		return errors.New("usage: platform create-admin -email EMAIL [-first NAME] [-last NAME]")
	}
	fs := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	email := fs.String("email", "", "the admin's email (required)")
	first := fs.String("first", "", "first name")
	last := fs.String("last", "", "last name")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	addr := validators.NormalizeEmail(*email)
	if !validators.IsEmail(addr) {
		return errors.New("-email must be a valid email address")
	}

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return errors.New("DATABASE_URL is required")
	}
	ctx := context.Background()
	db, err := database.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer db.Close()

	password, err := randomPassword()
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	id, err := repository.NewPlatformRepository(db).CreateAdmin(ctx, addr,
		strings.TrimSpace(*first), strings.TrimSpace(*last), string(hash))
	if err != nil {
		return err
	}
	fmt.Printf("platform admin #%d created\n  email:    %s\n  password: %s\n\nSign in on the platform admin address (PLATFORM_ADMIN_HOST). This password is not shown again.\n",
		id, addr, password)
	return nil
}

// randomPassword: 18 random bytes, URL-safe base64 (24 characters).
func randomPassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
