package tenant

import (
	"html/template"

	"github.com/jmoiron/sqlx"

	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/AdeptTravel/adept-framework/internal/theme"
)

// entry is the cache value stored in sync.Map.
type entry struct {
	tenant   *Tenant
	lastSeen int64
}

// Tenant holds everything a handler needs for one site.
type Tenant struct {
	Meta     site.Record        // row from `site`
	Config   map[string]string  // key-value from `site_config`
	DB       *sqlx.DB           // per-site connection pool
	Theme    *theme.Theme       // active theme (templates + assets)
	Renderer *template.Template // convenience alias to Theme.Renderer
}

// Close is called by the evictor on idle/LRU eviction.
func (t *Tenant) Close() error { return t.DB.Close() }
