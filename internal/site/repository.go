package site

import "github.com/jmoiron/sqlx"

// Record already defined earlier.

const baseQuery = `SELECT id, host, dsn, theme, status FROM site`

// AllActive returns every site row with status = 'Active'.
func AllActive(db *sqlx.DB) ([]Record, error) {
	rows := make([]Record, 0, 8)
	err := db.Select(&rows, baseQuery+" WHERE status = 'Active'")
	return rows, err
}
