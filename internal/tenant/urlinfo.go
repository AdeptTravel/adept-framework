// internal/tenant/urlinfo.go
//
// Canonicalised URL snapshot.
//
// Context
// -------
// Components and Widgets often need only the clean path, file extension,
// or query map—*not* the full `*url.URL` with scheme, raw path, and port.
// `URLInfo` captures just the fields requested by the architecture notes:
//
//   - Host      — `r.Host` without the `:port` suffix.
//   - Route     — path trimmed of leading/trailing “/”; root (“/”) → ""
//   - Ext       — file extension (e.g., “.html” or “.jpg”).
//   - MIME      — result of `mime.TypeByExtension(Ext)`; empty when Ext = "".
//   - QueryRaw  — raw query string (`RawQuery`).
//   - Query     — parsed `url.Values`.
//   - Fragment  — URL fragment without the “#”.
//
// These keys let downstream code derive canonical URLs and
// content-type hints without duplicating string parsing.
//
// Notes
// -----
//   - `URLInfo` is stored inside `tenant.Context` and therefore lives for
//     exactly one HTTP request.
//   - Oxford commas, two spaces after periods, no m-dash.
package tenant

import (
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// URLInfo is attached to tenant.Context and exposed to templates.
type URLInfo struct {
	Host     string     // "blog.example.com"
	Route    string     // "blog/2025/05" from path "/blog/2025/05"
	Ext      string     // ".html"
	MIME     string     // "text/html"
	QueryRaw string     // "tag=go"
	Query    url.Values // Parsed query args
	Fragment string     // "top"
}

// newURLInfo builds URLInfo from the incoming request.
func newURLInfo(r *http.Request) URLInfo {
	host := stripPort(r.Host)
	route := strings.Trim(r.URL.Path, "/") // "" when root
	ext := filepath.Ext(r.URL.Path)
	mimeType := mime.TypeByExtension(ext)

	return URLInfo{
		Host:     host,
		Route:    route,
		Ext:      ext,
		MIME:     mimeType,
		QueryRaw: r.URL.RawQuery,
		Query:    r.URL.Query(),
		Fragment: r.URL.Fragment,
	}
}

// stripPort removes the :port suffix from Host when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
