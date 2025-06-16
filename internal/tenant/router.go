// internal/tenant/router.go
//
// Cached per-tenant router.
//
// The router is built once per tenant (lazy) and cached.  It mounts only
// Components enabled in `component_acl` and wires the alias-rewrite and
// request-info middleware in the correct order.

package tenant

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/yanizio/adept/internal/component"
	"github.com/yanizio/adept/internal/requestinfo"
	"github.com/yanizio/adept/internal/routing"
)

const (
	RouteModeAbsolute  = routing.RouteModeAbsolute
	RouteModeAliasOnly = routing.RouteModeAliasOnly
	RouteModeBoth      = routing.RouteModeBoth
)

// Router builds (once) and returns the http.Handler for this tenant.
func (t *Tenant) Router() http.Handler {
	t.routerOnce.Do(func() {
		r := chi.NewRouter()

		// Alias rewrite must run before request-info enrichment.
		r.Use(routing.Middleware(t))
		r.Use(requestinfo.Enrich)

		// Query component ACL at build time.
		enabled := t.fetchEnabledComponents(context.Background())
		if len(enabled) == 0 {
			zap.L().Warn("component_acl empty – mounting all components",
				zap.String("host", t.Host()))
			enabled = component.AllNames()
		}

		for _, c := range component.All() {
			if _, ok := enabled[c.Name()]; ok {
				r.Mount("/", c.Routes())
			}
		}

		// Fallback: render home.html or 404.
		r.NotFound(func(w http.ResponseWriter, req *http.Request) {
			err := t.Renderer.ExecuteTemplate(w, "home.html",
				map[string]any{"Config": t.Config})
			if err != nil {
				http.NotFound(w, req)
			}
		})

		t.router = r
	})
	return t.router
}

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

// fetchEnabledComponents returns a set[name] for components enabled in ACL.
func (t *Tenant) fetchEnabledComponents(ctx context.Context) map[string]struct{} {
	db := t.GetDB()
	if db == nil {
		return nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT component FROM component_acl WHERE enabled = 1`)
	if err != nil {
		if isUnknownTable(err) {
			return nil // ACL table not yet migrated—treat as “all enabled”.
		}
		zap.L().Error("component_acl query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	set := make(map[string]struct{}, 8)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			zap.L().Error("component_acl scan", zap.Error(err))
			return nil
		}
		set[name] = struct{}{}
	}
	return set
}

// isUnknownTable recognises MariaDB (error 1146) and Cockroach/Postgres (42P01)
// “table does not exist” errors without importing driver-specific types.
func isUnknownTable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "1146") || strings.Contains(msg, "42P01")

}
