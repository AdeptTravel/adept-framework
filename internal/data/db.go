package data

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql" // MariaDB driver
	_ "github.com/jackc/pgx/v5/stdlib" // Cockroach (later)
)

type DB struct{ *sql.DB }

// Open returns a configured pool.
func Open(driver, dsn string, maxOpen, maxIdle int, lifetime time.Duration) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(lifetime)

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
