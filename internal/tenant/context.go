// internal/tenant/context.go
//
// Per-request context wrapper + tenant pointer helpers.
//
// Context
// -------
// Components and Widgets need a shared bundle of request-scoped data—URL
// parts, <head> builder, parsed User-Agent—without reaching back into
// *http.Request for every field.  `tenant.Context` carries this data and
// is created once at the top of the handler stack.
//
// In addition, other middleware layers (ACL, alias rewrite) need a way to
// retrieve the *Tenant aggregate from any `context.Context` without causing
// import cycles.  We therefore provide:
//
//     ctx = tenant.WithContext(r.Context(), t)   // set by root router
//     t   = tenant.FromContext(r.Context())      // used downstream
//
// Notes
// -----
// • Oxford commas, two spaces after periods.
// • Max line length 100 columns.

package tenant

import (
	"context"
	"net/http"

	"github.com/yanizio/adept/internal/head"
	"github.com/yanizio/adept/internal/ua"
)

// -----------------------------------------------------------------------------
// Per-request helper bundle
// -----------------------------------------------------------------------------

// Context bundles request-scoped helpers for Components and Widgets.
type Context struct {
	Request *http.Request // Original HTTP request (read-only)
	Head    *head.Builder // Accumulates <title>, meta tags, etc.
	URL     URLInfo       // Canonicalised URL parts
	UA      ua.Info       // Parsed user-agent
	// Geo, User, Session will be added later.
}

// NewContext builds the per-request helper bundle.
func NewContext(r *http.Request) *Context {
	return &Context{
		Request: r,
		Head:    head.New(),
		URL:     newURLInfo(r),
		UA:      ua.Parse(r.UserAgent()),
	}
}

// -----------------------------------------------------------------------------
// Tenant pointer helpers (cycles ↔ safe)
// -----------------------------------------------------------------------------

// ctxTenantKey is unexported to avoid collisions.
type ctxTenantKey struct{}

// WithContext returns a new context carrying the *Tenant pointer.
func WithContext(ctx context.Context, t *Tenant) context.Context {
	return context.WithValue(ctx, ctxTenantKey{}, t)
}

// FromContext retrieves the *Tenant pointer or nil if absent.
func FromContext(ctx context.Context) *Tenant {
	if v := ctx.Value(ctxTenantKey{}); v != nil {
		if ten, ok := v.(*Tenant); ok {
			return ten
		}
	}
	return nil
}
