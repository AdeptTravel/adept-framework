package tenant

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/AdeptTravel/adept-framework/internal/theme"
)

// loadSite turns a host into a fully initialised Tenant: site record,
// connection pool, and parsed theme templates.  For now we stub
// enabledModules with “core”; later this will come from the module ACL.
func loadSite(ctx context.Context, global *sqlx.DB, host string) (*Tenant, error) {
	// Site metadata
	rec, err := site.ByHost(ctx, global, host)
	if err != nil {
		return nil, ErrNotFound
	}

	// DB pool (small caps for each tenant)
	opts := database.Options{MaxOpenConns: 5, MaxIdleConns: 2,
		ConnMaxLifetime: 30 * time.Minute, Retries: 2,
		RetryBackoff: 500 * time.Millisecond}
	db, err := database.OpenWithOptions(ctx, rec.DSN, opts)
	if err != nil {
		return nil, err
	}

	// Theme parsing
	enabledModules := []string{"core"} // TODO: pull from DB ACL
	//mgr := theme.Manager{BaseDir: "/themes"}
	mgr := theme.Manager{BaseDir: "themes"}
	th, err := mgr.Load(rec.Theme, enabledModules)
	if err != nil {
		return nil, err
	}

	return &Tenant{
		Meta:     *rec,
		DB:       db,
		Theme:    th,
		Renderer: th.Renderer,
	}, nil
}
