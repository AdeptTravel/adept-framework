// internal/tenant/context.go
//
// Per-request context wrapper.
//
// Context
// -------
// Components and Widgets need a shared bundle of request-scoped data—URL
// parts, <head> builder, parsed User-Agent—without reaching back into
// *http.Request for every field.  `tenant.Context` carries this data and
// is created once at the top of the handler stack.
//
// Fields
// ------
// • `Request` – the original *http.Request (read-only).
// • `Head`    – HTML <head> helper used by templates.
// • `URL`     – small struct with path, query, and canonical link helpers.
// • `UA`      – parsed user-agent info (browser, device, bot flag).
// • Geo/User/Session placeholders will be filled in future milestones.
//
// Notes
// -----
//   - The struct is *not* exported outside the tenant package; Components
//     receive a pointer via Dependency Injection.
//   - Oxford commas, two spaces after periods.
package tenant

import (
	"net/http"

	"github.com/yanizio/adept/internal/head"
	"github.com/yanizio/adept/internal/ua"
)

// Context bundles request-scoped helpers for Components and Widgets.
type Context struct {
	Request *http.Request // Original HTTP request (read-only)
	Head    *head.Builder // Accumulates <title>, meta tags, etc.
	URL     URLInfo       // Canonicalised URL parts
	UA      ua.Info       // Parsed user-agent
	// Geo, User, Session will be added later
}

// NewContext builds the per-request context.
func NewContext(r *http.Request) *Context {
	return &Context{
		Request: r,
		Head:    head.New(),
		URL:     newURLInfo(r),
		UA:      ua.Parse(r.UserAgent()),
	}
}
