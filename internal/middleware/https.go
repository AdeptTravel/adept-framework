// Package middleware holds small, composable HTTP wrappers.
package middleware

import (
	"net/http"
	"strings"

	"github.com/AdeptTravel/adept-framework/internal/tenant"
)

// ForceHTTPS wraps h.  If the request is plain HTTP, the host is not
// “localhost”, and cache.Get confirms the site exists, the wrapper issues a
// 308 Permanent Redirect to the HTTPS version of the same URL.  Otherwise it
// calls the next handler unchanged.
func ForceHTTPS(cache *tenant.Cache, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Already HTTPS or dev host → continue.
		if r.TLS != nil || stripPort(r.Host) == "localhost" {
			h.ServeHTTP(w, r)
			return
		}

		// Only redirect if the host exists in the site table.
		if _, err := cache.Get(stripPort(r.Host)); err == nil {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}

		// Unknown host → keep normal flow (likely 404 later).
		h.ServeHTTP(w, r)
	})
}

// stripPort removes the :port suffix from Host when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
