// Package database centralises sqlx connection helpers.  The default driver
// is go-sql-driver/mysql, which also works with MariaDB and Cockroach when
// configured for the MySQL wire protocol.
//
// Public entry points:
//
//	Open(dsn)                    – quick helper with conservative pool sizes.
//	OpenWithOptions(dsn, maxOpen, maxIdle) – fine-grained control.
//
// Both helpers Ping the database before returning so callers can fail fast
// during bootstrap.  Callers should Close() the returned *sqlx.DB when no
// longer needed.
package database

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Open returns a *sqlx.DB with sane defaults: 15 max open, 5 idle, and a
// 30-minute connection lifetime.  Suitable for process-wide pools or for test
// setups.
func Open(dsn string) (*sqlx.DB, error) {
	return OpenWithOptions(dsn, 15, 5)
}

// OpenWithOptions lets callers tune maxOpen and maxIdle per pool.  Used by
// the tenant loader to keep per-tenant resource usage small.
func OpenWithOptions(dsn string, maxOpen, maxIdle int) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
