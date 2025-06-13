// internal/tenant/router.go
//
// Per-tenant router builder.
//
// * Each request gets a fresh chi.Router (cheap) so side-effect-registered
//   Components are always included without locking.
// * `requestinfo.Enrich` is injected high in the chain so handlers and
//   templates can read UA / Geo / IP from context.
// * If no Component route matches, we fall back to rendering the tenant’s
//   `home.html` template; if that template isn’t present, we return 404.

package tenant

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yanizio/adept/internal/component"
	"github.com/yanizio/adept/internal/requestinfo"
)

// Router builds a chi.Router that mounts every registered Component at “/”.
// The router is created lazily per request; the cost is negligible.
func (t *Tenant) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware: add UA / Geo / URL info to request context.
	r.Use(requestinfo.Enrich)

	// Mount every Component’s routes.
	for _, c := range component.All() {
		r.Mount("/", c.Routes())
	}

	// Fallback handler — try home.html first, else plain 404.
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Renderer.ExecuteTemplate(w, "home.html", map[string]any{
			"Config": t.Config,
		}); err != nil {
			http.NotFound(w, r)
		}
	})

	return r
}
