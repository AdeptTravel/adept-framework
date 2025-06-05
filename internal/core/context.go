// internal/core/context.go
//
// Central per-request context.
//
// Context
// -------
// Every handler builds a *core.Context and passes it down to Components
// and Widgets.  It bundles:
//
//   - Tenant  — live tenant aggregate (site meta, DB pool, theme).
//   - Request — the original *http.Request (safe-wrapped upstream).
//   - Writer  — convenience http.ResponseWriter.
//   - Params  — route params such as “slug”.
//   - Info    — parsed UA, geo, URL, and timestamp.
//
// Notes
// -----
// • Components must treat Tenant as read-only.
// • Oxford commas, two spaces after periods.
package core

import (
	"net/http"

	"github.com/yanizio/adept/internal/requestinfo"
	"github.com/yanizio/adept/internal/tenant"
)

// Context is passed to Components, Widgets, and templates.
type Context struct {
	Tenant  *tenant.Tenant           // Live tenant aggregate
	Request *http.Request            // Original request
	Writer  http.ResponseWriter      // Convenience writer
	Params  map[string]string        // Route params (“slug”, etc.)
	Info    *requestinfo.RequestInfo // UA, Geo, URL, timestamp
}
