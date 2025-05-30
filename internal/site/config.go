// internal/site/config.go
//
// Helpers for fetching key-value settings from the `site_config` table.
// The query runs once when the tenant is loaded, and the resulting map is
// cached in memory alongside the Tenant struct.
package site

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// ConfigBySite returns a map[key]value for one site_id.
func ConfigBySite(ctx context.Context, db *sqlx.DB, siteID uint64) (map[string]string, error) {
	const q = `
	    SELECT  ` + "`key`, value" + `
	    FROM    site_config
	    WHERE   site_id = ?`
	rows := make([]struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}, 0, 8) // small default cap

	if err := db.SelectContext(ctx, &rows, q, siteID); err != nil {
		return nil, err
	}

	cfg := make(map[string]string, len(rows))
	for _, r := range rows {
		cfg[r.Key] = r.Value
	}
	return cfg, nil
}
