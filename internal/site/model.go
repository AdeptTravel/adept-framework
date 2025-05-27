package site

import (
	"time"
)

// Record mirrors one row in the persistent `site` table.
//
// Non‑nullable timestamps (created_at, updated_at) are stored in
// sql.NullTime so rows that still carry MySQL's legacy zero date
// ("0000‑00‑00 00:00:00") do not cause scan errors.  Callers that need a
// time.Time value should check the Valid flag before using the Time field.
//
//	if rec.CreatedAt.Valid {
//	    doSomething(rec.CreatedAt.Time)
//	}
//
// This change fixes the tenant‑loader failure you observed when the row was
// found yet sqlx returned a scan error, leaving the struct half‑populated.
type Record struct {
	ID           uint64     `db:"id"`
	Host         string     `db:"host"`
	DSN          string     `db:"dsn"`
	Theme        string     `db:"theme"`
	Title        string     `db:"title"`
	Locale       string     `db:"locale"`
	RoutingMode  string     `db:"routing_mode"`
	RouteVersion string     `db:"route_version"`
	SuspendedAt  *time.Time `db:"suspended_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
	//CreatedAt   *time.Time `db:"created_at"`
	//UpdatedAt   sql.NullTime `db:"updated_at"`
}
