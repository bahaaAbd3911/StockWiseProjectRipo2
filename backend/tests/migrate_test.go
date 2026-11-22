package tests

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/migrate"
)

// A database created before the migration tool existed (000001 run by hand,
// like the real one) must get only the NEW migrations; the baseline is adopted.
func TestMigrateAdoptsExistingDatabase(t *testing.T) {
	requireServer(t)
	ctx := context.Background()
	admin, err := database.Connect(ctx, baseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("ecommerce_mig_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)") //nolint:errcheck

	u, _ := url.Parse(baseURL)
	u.Path = "/" + name
	db, err := database.Connect(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	baseline, _ := os.ReadFile("../migrations/000001_baseline.up.sql")
	if _, err := db.Exec(string(baseline)); err != nil { // the "old way"
		t.Fatal(err)
	}
	applied, err := migrate.Up(ctx, db, "../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 1 || applied[0] != "000002_offers_bundles_debts" {
		t.Fatalf("only 000002 should run on an existing database, got %v", applied)
	}
	again, err := migrate.Up(ctx, db, "../migrations")
	if err != nil || len(again) != 0 {
		t.Fatalf("second run must do nothing: %v %v", again, err)
	}
}
