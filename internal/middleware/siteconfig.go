// internal/middleware/siteconfig.go
package middleware

import (
	"context"
	"net/http"
	"sync"

	"github.com/AdeptTravel/adept-framework/internal/config"
)

type ctxKey struct{}

func AttachSiteConfig(sitesRoot string) func(http.Handler) http.Handler {
	var mu sync.RWMutex
	cache := map[string]config.Site{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			mu.RLock()
			siteCfg, ok := cache[host]
			mu.RUnlock()

			if !ok {
				cfg, err := config.LoadSite(sitesRoot, host)
				if err != nil {
					http.Error(w, "unknown site", 404)
					return
				}
				mu.Lock()
				cache[host] = cfg
				mu.Unlock()
				siteCfg = cfg
			}

			ctx := context.WithValue(r.Context(), ctxKey{}, siteCfg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func From(r *http.Request) config.Site {
	if v, ok := r.Context().Value(ctxKey{}).(config.Site); ok {
		return v
	}
	return config.Site{} // zero value if missing
}
