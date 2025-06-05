// internal/tenant/types.go
//
// Tenant cache entry and aggregate.
//
// Context
// -------
// A live Tenant aggregates everything the router and Components need to
// serve a single site: its `site` row, per-site DB pool, in-memory config
// map, active Theme, and a pre-parsed html/template tree for fast
// rendering.  The cache stores a pointer to Tenant inside `entry`, along
// with a `lastSeen` UnixNano timestamp used by the evictor for idle and
// LRU eviction.
//
// Notes
// -----
//   - `Meta` now uses `meta.Record` from internal/tenant/meta.
//   - `Close` is invoked only by the cache evictor; Components must treat
//     Tenant as immutable after initial load.
//   - Oxford commas, two spaces after periods.
package tenant

import (
	"html/template"

	"github.com/jmoiron/sqlx"

	"github.com/yanizio/adept/internal/tenant/meta"
	"github.com/yanizio/adept/internal/theme"
)

//
// Cache entry
//

type entry struct {
	tenant   *Tenant
	lastSeen int64 // UnixNano
}

//
// Tenant aggregate
//

// Tenant groups all per-site runtime assets needed by request handlers.
type Tenant struct {
	Meta     meta.Record        // Row from `site`
	Config   map[string]string  // Key-value pairs from `site_config`
	DB       *sqlx.DB           // Per-site connection pool
	Theme    *theme.Theme       // Active theme (templates + assets)
	Renderer *template.Template // Convenience alias: Theme.Renderer
}

// Close is called by the cache evictor on idle or LRU eviction.
func (t *Tenant) Close() error { return t.DB.Close() }
