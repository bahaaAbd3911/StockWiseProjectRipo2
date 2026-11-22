// Package migrate applies the SQL files in migrations/ in order and remembers
// which ones ran in a schema_migrations table.
//
// Files are named NNNNNN_name.up.sql. Each file runs in its own transaction
// together with the row that marks it as applied, so a failing migration
// leaves nothing half-done.
//
// Databases created before this tool existed (by running 000001 by hand) are
// detected: if schema_migrations is empty but the users table already exists,
// 000001 is recorded as applied instead of being run again.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ecommerce-backend/internal/database"
)

const baseline = "000001_baseline"

// Migration is one .up.sql file and whether it has been applied.
type Migration struct {
	Version string
	Path    string
	Applied bool
}

// Status lists all migrations in dir and marks the applied ones. It only
// reads: on a database that has never been migrated it reports what Up would
// do (the baseline counts as applied when the tables already exist).
func Status(ctx context.Context, db *sql.DB, dir string) ([]Migration, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	applied := map[string]bool{}
	var tracked, hasUsers bool
	if err := db.QueryRowContext(ctx,
		`SELECT to_regclass('public.schema_migrations') IS NOT NULL, to_regclass('public.users') IS NOT NULL`,
	).Scan(&tracked, &hasUsers); err != nil {
		return nil, err
	}
	if tracked {
		rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err != nil {
				return nil, err
			}
			applied[v] = true
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	if len(applied) == 0 && hasUsers {
		applied[baseline] = true // will be adopted by Up, not re-run
	}

	out := make([]Migration, 0, len(files))
	for _, f := range files {
		v := strings.TrimSuffix(filepath.Base(f), ".up.sql")
		out = append(out, Migration{Version: v, Path: f, Applied: applied[v]})
	}
	return out, nil
}

// Up applies every pending migration and returns the versions it applied.
func Up(ctx context.Context, db *sql.DB, dir string) ([]string, error) {
	// Only one migrator at a time (e.g. two servers starting together).
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(7342001)`); err != nil {
		return nil, err
	}
	defer conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(7342001)`) //nolint:errcheck

	if err := ensureTable(ctx, db); err != nil {
		return nil, err
	}
	if err := adoptBaseline(ctx, db); err != nil {
		return nil, err
	}

	list, err := Status(ctx, db, dir)
	if err != nil {
		return nil, err
	}
	var done []string
	for _, m := range list {
		if m.Applied {
			continue
		}
		body, err := os.ReadFile(m.Path)
		if err != nil {
			return done, err
		}
		err = database.WithTx(ctx, db, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, string(body)); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.Version)
			return err
		})
		if err != nil {
			return done, fmt.Errorf("migration %s failed: %w", m.Version, err)
		}
		done = append(done, m.Version)
	}
	return done, nil
}

func ensureTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS public.schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamp with time zone DEFAULT now() NOT NULL
		)`)
	return err
}

func adoptBaseline(ctx context.Context, db *sql.DB) error {
	var tracked int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&tracked); err != nil {
		return err
	}
	if tracked > 0 {
		return nil
	}
	var hasUsers bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.users') IS NOT NULL`).Scan(&hasUsers); err != nil {
		return err
	}
	if !hasUsers {
		return nil
	}
	_, err := db.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, baseline)
	return err
}
