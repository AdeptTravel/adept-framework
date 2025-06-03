package tenant

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yanizio/adept/internal/database"
	"github.com/yanizio/adept/internal/site"
	"github.com/yanizio/adept/internal/theme"
)

// loadSite turns host → *Tenant.  Steps:
//
//  1. Fetch site row.
//  2. Fetch key-value config rows.
//  3. Open small DB pool.
//  4. Parse theme templates.
func loadSite(ctx context.Context, global *sqlx.DB, host string) (*Tenant, error) {
	// 1. site row
	rec, err := site.ByHost(ctx, global, host)
	if err != nil {
		return nil, ErrNotFound
	}

	// 2. key-value config
	cfg, err := site.ConfigBySite(ctx, global, rec.ID)
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
	enabledModules := []string{"core"} // TODO: pull from ACL table
	mgr := theme.Manager{BaseDir: "themes"}
	th, err := mgr.Load(rec.Theme, enabledModules)
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
