package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"time"

	"ecommerce-backend/internal/tenant"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// AppRole is the database role the API runs as (created by migration 000007).
// It is not the table owner and not a superuser, so PostgreSQL's row-level
// security applies to it: every query only sees the current store's rows.
const AppRole = "stockwise_app"

// ConnectApp opens the pool the API uses. Every connection:
//   - switches to AppRole right after connecting (SET ROLE), and
//   - before each statement makes sure app.store_id / app.platform match the
//     store in the statement's context (see tenant.WithStore / AsPlatform).
//
// The settings are sent only when they change on that connection, so most
// statements cost nothing extra. Changing store inside a transaction is
// refused: transactions belong to one store.
//
// Use Connect (no role switch, no store) for migrations and maintenance.
func ConnectApp(ctx context.Context, databaseURL string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db := sql.OpenDB(&tenantConnector{inner: stdlib.GetConnector(*cfg)})
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to PostgreSQL as %s (is migration 000007 applied?): %w", AppRole, err)
	}
	return db, nil
}

type tenantConnector struct {
	inner driver.Connector
}

func (c *tenantConnector) Driver() driver.Driver { return c.inner.Driver() }

func (c *tenantConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.inner.Connect(ctx)
	if err != nil {
		return nil, err
	}
	exec, ok := conn.(driver.ExecerContext)
	if !ok {
		_ = conn.Close()
		return nil, errors.New("database driver does not support ExecContext")
	}
	if _, err := exec.ExecContext(ctx, "SET ROLE "+pgx.Identifier{AppRole}.Sanitize(), nil); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("SET ROLE %s: %w", AppRole, err)
	}
	return &tenantConn{inner: conn}, nil
}

// tenantConn forwards everything to the pgx connection, setting the store first.
type tenantConn struct {
	inner    driver.Conn
	store    string // value of app.store_id on this connection
	platform string // value of app.platform ("on" / "")
	known    bool   // false until the settings were sent once
	inTx     bool
}

var errStoreChangeInTx = errors.New("database: a transaction cannot switch to another store")

func wanted(ctx context.Context) (store, platform string) {
	if id := tenant.ID(ctx); id > 0 {
		store = strconv.FormatInt(id, 10)
	}
	if tenant.IsPlatform(ctx) {
		platform = "on"
	}
	return store, platform
}

func (c *tenantConn) ensure(ctx context.Context) error {
	store, platform := wanted(ctx)
	if c.known && store == c.store && platform == c.platform {
		return nil
	}
	if c.inTx {
		return errStoreChangeInTx
	}
	exec := c.inner.(driver.ExecerContext)
	if _, err := exec.ExecContext(ctx,
		`SELECT set_config('app.store_id', $1, false), set_config('app.platform', $2, false)`,
		[]driver.NamedValue{{Ordinal: 1, Value: store}, {Ordinal: 2, Value: platform}}); err != nil {
		c.known = false
		return err
	}
	c.store, c.platform, c.known = store, platform, true
	return nil
}

func (c *tenantConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *tenantConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if err := c.ensure(ctx); err != nil {
		return nil, err
	}
	return c.inner.(driver.ConnPrepareContext).PrepareContext(ctx, query)
}

func (c *tenantConn) Close() error { return c.inner.Close() }

func (c *tenantConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *tenantConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if err := c.ensure(ctx); err != nil {
		return nil, err
	}
	tx, err := c.inner.(driver.ConnBeginTx).BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	c.inTx = true
	return &tenantTx{inner: tx, conn: c}, nil
}

func (c *tenantConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := c.ensure(ctx); err != nil {
		return nil, err
	}
	return c.inner.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func (c *tenantConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := c.ensure(ctx); err != nil {
		return nil, err
	}
	return c.inner.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func (c *tenantConn) Ping(ctx context.Context) error {
	if p, ok := c.inner.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

// CheckNamedValue lets pgx convert Go values itself (arrays, decimals...).
func (c *tenantConn) CheckNamedValue(nv *driver.NamedValue) error {
	if ch, ok := c.inner.(driver.NamedValueChecker); ok {
		return ch.CheckNamedValue(nv)
	}
	return driver.ErrSkip
}

func (c *tenantConn) ResetSession(ctx context.Context) error {
	if c.inTx { // a transaction left open is a bug; drop the connection
		return driver.ErrBadConn
	}
	if r, ok := c.inner.(driver.SessionResetter); ok {
		return r.ResetSession(ctx)
	}
	return nil
}

func (c *tenantConn) IsValid() bool {
	if v, ok := c.inner.(driver.Validator); ok {
		return v.IsValid()
	}
	return true
}

type tenantTx struct {
	inner driver.Tx
	conn  *tenantConn
}

func (t *tenantTx) Commit() error {
	t.conn.inTx = false
	return t.inner.Commit()
}

// A rollback also undoes a set_config made inside the transaction; settings
// are only ever changed outside transactions, so the cached values stay true.
func (t *tenantTx) Rollback() error {
	t.conn.inTx = false
	return t.inner.Rollback()
}
