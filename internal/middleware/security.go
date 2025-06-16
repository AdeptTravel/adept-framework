// internal/middleware/security.go
//
// Security-header middleware.
//
// Injects industry-standard headers on every response:
//
//   • Strict-Transport-Security  –  forces HTTPS (2 years + preload)
//   • Content-Security-Policy   –  sane default self-only policy
//   • X-Frame-Options           –  click-jacking defence
//   • X-Content-Type-Options    –  MIME-sniffing defence
//   • Referrer-Policy           –  drops path/query from Referer
//   • Permissions-Policy        –  disables powerful features by default
//
// Notes
// -----
// • Headers are added *after* next.ServeHTTP so handlers may set Content-Type
//   first; the middleware never overwrites an existing value.
// • If Adept is running behind a TLS-terminating proxy, HSTS is still useful
//   because browsers see the tenant’s domain as HTTPS.
// • Oxford commas, two spaces after periods.

package middleware

import "net/http"

// Security sets security headers for every response.
func Security(next http.Handler) http.Handler {
	const (
		hsts = "max-age=63072000; includeSubDomains; preload"
		csp  = "default-src 'self'; img-src 'self' data:; object-src 'none'; " +
			"base-uri 'self'; frame-ancestors 'none'"
		xfo   = "DENY"
		nosn  = "nosniff"
		refer = "strict-origin-when-cross-origin"
		perm  = "geolocation=(), microphone=(), camera=()"
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		add := w.Header().Add // shorthand

		if w.Header().Get("Strict-Transport-Security") == "" {
			add("Strict-Transport-Security", hsts)
		}
		if w.Header().Get("Content-Security-Policy") == "" {
			add("Content-Security-Policy", csp)
		}
		if w.Header().Get("X-Frame-Options") == "" {
			add("X-Frame-Options", xfo)
		}
		if w.Header().Get("X-Content-Type-Options") == "" {
			add("X-Content-Type-Options", nosn)
		}
		if w.Header().Get("Referrer-Policy") == "" {
			add("Referrer-Policy", refer)
		}
		if w.Header().Get("Permissions-Policy") == "" {
			add("Permissions-Policy", perm)
		}
	})
}
