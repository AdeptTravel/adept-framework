// Package database centralises sqlx connection helpers.  The wrapper
// intentionally stays thin, but it now offers context‑aware dialing, retry
// logic, and configurable pool sizes so callers can tailor resource usage
// without rewriting boilerplate.

package database

import (
	"context"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql" // side‑effect import keeps driver pluggable
	"github.com/jmoiron/sqlx"
)

//
// Options lets callers tune connection‑pool behaviour and retry policy.
// Zero values fall back to sensible defaults.
//

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

// merge fills zero values in dst with defaults.
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

// Open is a convenience wrapper that dials with background context and
// default pool sizes.
func Open(dsn string) (*sqlx.DB, error) {
	return OpenWithOptions(context.Background(), dsn, Options{})
}

// OpenWithPool lets callers override maxOpen and maxIdle while keeping all
// other defaults intact.
func OpenWithPool(dsn string, maxOpen, maxIdle int) (*sqlx.DB, error) {
	return OpenWithOptions(context.Background(), dsn, Options{
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
	})
}

// OpenWithOptions dials using the provided context and options.  If Retries
// > 0, it will re‑attempt the dial + ping loop with exponential backoff.
func OpenWithOptions(ctx context.Context, dsn string, opts Options) (*sqlx.DB, error) {
	opts.merge()

	var lastErr error
	backoff := opts.RetryBackoff

	for attempt := 0; attempt <= opts.Retries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		db, err := sqlx.Open("mysql", dsn)
		if err != nil {
			lastErr = err
			goto retry
		}

		db.SetMaxOpenConns(opts.MaxOpenConns)
		db.SetMaxIdleConns(opts.MaxIdleConns)
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)

		if err = db.PingContext(ctx); err == nil {
			return db, nil // success
		}

		// ping failed, record error and retry after close
		_ = db.Close()
		lastErr = err

	retry:
		if attempt < opts.Retries {
			select {
			case <-time.After(backoff):
				backoff *= 2 // simple exponential
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
