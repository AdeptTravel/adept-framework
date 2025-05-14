package requestctx

import (
	"context"
	"net/http"
	"time"
)

// RequestCtx carries every per-request fact we might need across modules,
// widgets, templates, and security checks.
type RequestCtx struct {
	Timestamp time.Time

	// network
	IP        string
	Host      string
	Path      string
	UserAgent string

	// geo (nil until we enrich it)
	Geo *GeoInfo

	// FUTURE: Auth, AB-test bucket, device, etc.
}

// GeoInfo is a friendly subset of the MaxMind record.
type GeoInfo struct {
	CountryISO string
	Country    string
	City       string
	Lat, Lon   float64
}

// -----------------------------------------------------------------------------
// context helpers
// -----------------------------------------------------------------------------

type ctxKey struct{}

// From extracts *RequestCtx from an http.Request; returns nil if missing.
func From(r *http.Request) *RequestCtx {
	if v, ok := r.Context().Value(ctxKey{}).(*RequestCtx); ok {
		return v
	}
	return nil
}

// With attaches rc to the request’s context and returns a new *http.Request.
func With(r *http.Request, rc *RequestCtx) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, rc))
}
