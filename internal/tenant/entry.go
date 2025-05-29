package tenant

import (
	"html/template"

	"github.com/jmoiron/sqlx"

	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/AdeptTravel/adept-framework/internal/theme"
)

// entry is the cache value: runtime Tenant plus last-seen timestamp.
type entry struct {
	tenant   *Tenant
	lastSeen int64
}

// Tenant wraps everything a request handler needs for one site.
type Tenant struct {
	Meta     site.Record        // site-table metadata
	DB       *sqlx.DB           // per-tenant connection pool
	Theme    *theme.Theme       // active theme (templates + assets)
	Renderer *template.Template // convenience alias for Theme.Renderer
}

// Close shuts down the tenant’s DB pool.
func (t *Tenant) Close() error { return t.DB.Close() }
