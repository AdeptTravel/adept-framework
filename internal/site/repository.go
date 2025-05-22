// Package site holds thin data-access helpers for the persistent `site`
// table.  Each helper is a single-purpose query, returning a strongly typed
// struct so callers do not repeat column names.
//
// The table schema:
//
//	CREATE TABLE site (
//	    id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
//	    host       VARCHAR(256) NOT NULL UNIQUE,
//	    dsn        VARCHAR(128) NOT NULL,
//	    theme      VARCHAR(256) NOT NULL DEFAULT 'base',
//	    status     ENUM('Active', 'Block', 'Inactive') NOT NULL DEFAULT 'Active',
//	    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
//	    updated_at TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW()
//	);
package site

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// AllActive returns every site with status = 'Active'.  Used by admin
// dashboards or one-off migrations, but no longer required by the HTTP
// bootstrap path.
func AllActive(db *sqlx.DB) ([]Record, error) {
	const q = `
        SELECT id, host, dsn, theme, status
        FROM   site
        WHERE  status = 'Active'`
	var rows []Record
	if err := db.Select(&rows, q); err != nil {
		return nil, err
	}
	return rows, nil
}

// ByHost fetches a single active site row by its host FQDN.  Callers pass a
// context so the lookup respects request deadlines.  Returns (*Record,
// nil) on success, or (nil, error).  The caller should translate sql.ErrNoRows
// into its own domain-specific error (for example, tenant.ErrNotFound).
func ByHost(ctx context.Context, db *sqlx.DB, host string) (*Record, error) {
	const q = `
        SELECT id, host, dsn, theme, status
        FROM   site
        WHERE  host = ? AND status = 'Active'
        LIMIT  1`
	var rec Record
	if err := db.GetContext(ctx, &rec, q, host); err != nil {
		return nil, err
	}
	return &rec, nil
}
