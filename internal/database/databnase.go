// internal/database/database.go
//
// Package database centralises sqlx connection helpers.  The wrapper stays
// thin, but it offers context-aware dialing, retry logic, configurable pool
// sizes, **and now a lazy DSN provider** so secrets fetched from Vault (or any
// other rotating source) can be injected at call-time.  Existing helpers that
// accept a raw DSN are preserved as shims so downstream code keeps compiling.
//
// Oxford commas, two spaces after periods.

package database

import (
	"context"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql" // side-effect import keeps driver pluggable
	"github.com/jmoiron/sqlx"
)

//
// configurable options
//

// Options lets callers tune connection-pool behaviour and retry policy.  Zero
// values fall back to sensible defaults.
type Options struct {
	MaxOpenConns    int           // default 15
	MaxIdleConns    int           // default 5
	ConnMaxLifetime time.Duration // default 30m
	Retries         int           // dial retries, default 0 (no retry)
	RetryBackoff    time.Duration // sleep between retries, default 1s
}

var defaultOpts = Options{
	MaxOpenConns:    15,
	MaxIdleConns:    5,
	ConnMaxLifetime: 30 * time.Minute,
	Retries:         0,
	RetryBackoff:    time.Second,
}

func (dst *Options) merge() {
	if dst.MaxOpenConns == 0 {
		dst.MaxOpenConns = defaultOpts.MaxOpenConns
	}
	if dst.MaxIdleConns == 0 {
		dst.MaxIdleConns = defaultOpts.MaxIdleConns
	}
	if dst.ConnMaxLifetime == 0 {
		dst.ConnMaxLifetime = defaultOpts.ConnMaxLifetime
	}
	if dst.RetryBackoff == 0 {
		dst.RetryBackoff = defaultOpts.RetryBackoff
	}
}

//
// DSN provider glue
//

// DSNProvider returns a valid DSN string every time it is called.  The function
// may hit Vault, perform in-memory rotation, or simply return a constant.
type DSNProvider func() string

//
// public helpers
//

func Open(dsn string) (*sqlx.DB, error) {
	return OpenProvider(context.Background(), func() string { return dsn }, Options{})
}

func OpenProvider(ctx context.Context, dsnFunc DSNProvider, opts Options) (*sqlx.DB, error) {
	return openWithOptions(ctx, dsnFunc, opts)
}

func OpenWithPool(dsn string, maxOpen, maxIdle int) (*sqlx.DB, error) {
	return OpenProvider(
		context.Background(),
		func() string { return dsn },
		Options{MaxOpenConns: maxOpen, MaxIdleConns: maxIdle},
	)
}

// Deprecated: migrate to OpenProvider for rotational credentials.
func OpenWithOptions(ctx context.Context, dsn string, opts Options) (*sqlx.DB, error) {
	return openWithOptions(ctx, func() string { return dsn }, opts)
}

//
// internal dial + retry loop
//

func openWithOptions(ctx context.Context, dsnFunc DSNProvider, opts Options) (*sqlx.DB, error) {
	opts.merge()

	var lastErr error
	backoff := opts.RetryBackoff

	for attempt := 0; attempt <= opts.Retries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		db, err := sqlx.Open("mysql", dsnFunc())
		if err != nil {
			lastErr = err
			goto retry
		}

		db.SetMaxOpenConns(opts.MaxOpenConns)
		db.SetMaxIdleConns(opts.MaxIdleConns)
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)

		if err = db.PingContext(ctx); err == nil {
			return db, nil
		}

		_ = db.Close()
		lastErr = err

	retry:
		if attempt < opts.Retries {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if lastErr == nil {
		lastErr = errors.New("database: open failed with unknown error")
	}
	return nil, lastErr
}
