// internal/server/timeouts.go
//
// HTTP server helper with robust timeouts.
//
// Production hardening recommends:
//
//   • ReadTimeout   – abort slow-loris headers (10 s)
//   • WriteTimeout  – cap total response time (15 s)
//   • IdleTimeout   – close keep-alives on idle clients (60 s)
//
// This helper centralises those defaults so cmd/web doesn’t repeat boilerplate.
//

package server

import (
	"net/http"
	"time"
)

// New constructs an *http.Server with sensible defaults.
func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// TLSConfig may be injected by callers (e.g., autocert).
	}
}
