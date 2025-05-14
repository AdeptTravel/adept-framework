package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AdeptTravel/adept-framework/internal/geo"
	"github.com/AdeptTravel/adept-framework/internal/requestctx"
)

// AttachRequestCtx populates a *requestctx.RequestCtx and stores it in the
// request's context for downstream code.
//
//	geoDB   – nil to skip Geo-IP lookup
//	devHost – override Host header when running locally; empty string in prod
func AttachRequestCtx(geoDB *geo.DB, devHost string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ------------------------- client IP
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if xf := r.Header.Get("X-Forward-For"); xf != "" {
				ip = strings.TrimSpace(strings.Split(xf, ",")[0])
			}

			// ------------------------- host (override in dev if needed)
			host := r.Host
			if devHost != "" {
				host = devHost
			}

			rc := &requestctx.RequestCtx{
				Timestamp: time.Now(),
				IP:        ip,
				Host:      host,
				Path:      r.URL.Path,
				UserAgent: r.UserAgent(),
			}

			// ------------------------- Geo-IP enrichment
			if geoDB != nil {
				if loc, _ := geoDB.Lookup(ip); loc.CountryISO != "" {
					rc.Geo = &requestctx.GeoInfo{
						CountryISO: loc.CountryISO,
						Country:    loc.Country,
						City:       loc.City,
						Lat:        loc.Lat,
						Lon:        loc.Lon,
					}
				}
			}

			// ------------------------- pass to next handler
			next.ServeHTTP(w, requestctx.With(r, rc))
		})
	}
}
