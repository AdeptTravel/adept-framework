package site

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

// AllActive returns every site that is neither suspended nor deleted.  This
// helper is used by admin dashboards or batch operations, not by the HTTP
// bootstrap path.
func AllActive(db *sqlx.DB) ([]Record, error) {
	const q = `
        SELECT id, host, dsn, theme, title, locale,
               suspended_at, deleted_at
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
// caller supplies a context so the lookup respects request deadlines.
func ByHost(ctx context.Context, db *sqlx.DB, host string) (*Record, error) {

	const q = `
        SELECT id, host, dsn, theme, title, locale,
               suspended_at, deleted_at
        FROM   site
        WHERE  host = ?
          AND  suspended_at IS NULL
          AND  deleted_at   IS NULL
        LIMIT  1;`
	var rec Record

	if err := db.GetContext(ctx, &rec, q, host); err != nil {
		log.Printf("ByHost result for '%s': %+v", host, rec)
		log.Printf("ByHost error for '%s': %v", host, err)
		return nil, err
	}

	return &rec, nil
}
