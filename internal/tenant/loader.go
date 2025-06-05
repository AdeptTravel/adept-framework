// internal/tenant/loader.go
//
// host → Tenant loader.
//
// Context
// -------
// The cache’s slow-path calls `loadSite` to turn a Host header into a live
// Tenant instance.  The loader performs exactly four blocking steps:
//
//  1. Fetch the `site` row (`meta.ByHost`).
//  2. Fetch the key-value pairs from `site_config`.
//  3. Open a small per-site DB pool with retry and sane limits.
//  4. Parse the active Theme’s templates.
//
// All heavy resources (DB pool, parsed templates) are created once here
// and reused for every request until the tenant is evicted.
//
// Notes
// -----
//   - On any error the function returns early; the cache logs and surfaces
//     the error to the caller.
//   - Oxford commas, two spaces after periods, no m-dash.
package tenant

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yanizio/adept/internal/database"
	"github.com/yanizio/adept/internal/tenant/meta"
	"github.com/yanizio/adept/internal/theme"
)

// loadSite turns host → *Tenant in four steps:
//
//  1. Fetch site row.
//  2. Fetch key-value config rows.
//  3. Open small DB pool.
//  4. Parse theme templates.
func loadSite(ctx context.Context, global *sqlx.DB, host string) (*Tenant, error) {
	// 1. site row
	rec, err := meta.ByHost(ctx, global, host)
	if err != nil {
		return nil, ErrNotFound
	}

	// 2. key-value config
	cfg, err := meta.ConfigBySite(ctx, global, rec.ID)
	if err != nil {
		return nil, err
	}

	// 3. tenant DB pool
	opts := database.Options{
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 30 * time.Minute,
		Retries:         2,
		RetryBackoff:    500 * time.Millisecond,
	}
	db, err := database.OpenWithOptions(ctx, rec.DSN, opts)
	if err != nil {
		return nil, err
	}

	// 4. theme parsing
	enabledComponents := []string{"core"} // TODO: pull from ACL table
	mgr := theme.Manager{BaseDir: "themes"}
	th, err := mgr.Load(rec.Theme, enabledComponents)
	if err != nil {
		return nil, err
	}

	return &Tenant{
		Meta:     *rec,
		Config:   cfg,
		DB:       db,
		Theme:    th,
		Renderer: th.Renderer,
	}, nil
}
