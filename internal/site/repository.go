// Package site provides helpers to fetch site metadata from the global DB.
package site

import "github.com/jmoiron/sqlx"

// Record mirrors one row from the `site` table.
type Record struct {
	ID     uint64 `db:"id"`
	Host   string `db:"host"`
	DSN    string `db:"dsn"`
	Theme  string `db:"theme"`
	Status string `db:"status"`
}

// ByHost returns the first active site whose host matches the FQDN.
func ByHost(db *sqlx.DB, host string) (*Record, error) {
	const q = `
        SELECT id, host, dsn, theme, status
        FROM   site
        WHERE  host = ? AND status = 'Active'
        LIMIT  1`
	var rec Record
	if err := db.Get(&rec, q, host); err != nil {
		return nil, err
	}
	return &rec, nil
}
