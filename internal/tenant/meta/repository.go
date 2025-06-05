// internal/tenant/meta/query.go
//
// Site-table query helpers.
//
// Context
// -------
// These functions provide read-only access to the **site** table for
// operations that sit *outside* the main HTTP bootstrap path:
//
//   - `AllActive` — admin dashboards, cron jobs, batch reports.
//   - `ByHost`    — tenant loader on first request.
//
// Both helpers honour the Adept schema (2025-06-05) and exclude suspended
// or deleted rows at SQL level to keep callers simple.
//
// Workflow
// --------
//  1. Callers supply a *sqlx.DB that is already connected to the control-
//     plane database.
//  2. Each helper executes exactly one parameterised SELECT.
//  3. Rows are scanned into `meta.Record`, which mirrors the current
//     schema.
//  4. Errors are returned verbatim so the caller can wrap or log them
//     using the project logger.
//
// Notes
// -----
//   - Column list matches the fields in `Record`; update both together.
//   - All queries carry `LIMIT 1` where appropriate to avoid accidental
//     additional scans.
//   - Oxford commas, two spaces after periods, no m-dash.
package meta

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

// AllActive returns every site that is neither suspended nor deleted.
// Intended for admin dashboards or batch operations, not the HTTP
// bootstrap path.
func AllActive(db *sqlx.DB) ([]Record, error) {
	const q = `
        SELECT id, host, dsn, theme, locale, routing_mode, route_version,
               preload, suspended_at, deleted_at, created_at, updated_at
        FROM   site
        WHERE  suspended_at IS NULL
          AND  deleted_at   IS NULL`
	var rows []Record
	if err := db.Select(&rows, q); err != nil {
		return nil, err
	}
	return rows, nil
}

// ByHost fetches a single site row that is not suspended or deleted.  The
// lookup respects request deadlines via the supplied context.Context.
func ByHost(ctx context.Context, db *sqlx.DB, host string) (*Record, error) {
	const q = `
        SELECT id, host, dsn, theme, locale, routing_mode, route_version,
               preload, suspended_at, deleted_at, created_at, updated_at
        FROM   site
        WHERE  host = ?
          AND  suspended_at IS NULL
          AND  deleted_at   IS NULL
        LIMIT  1`
	var rec Record
	if err := db.GetContext(ctx, &rec, q, host); err != nil {
		// TODO(bjy): replace stdlib log with project logger once observability
		// package lands.
		log.Printf("meta.ByHost: host=%q err=%v", host, err)
		return nil, err
	}
	return &rec, nil
}
