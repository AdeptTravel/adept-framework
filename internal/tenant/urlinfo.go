// urlinfo.go
//
// URLInfo captures only the fields requested:
//
//   - Host   – r.Host without the :port suffix.
//   - Route  – the URL path stripped of leading/trailing “/”.  Root ("/")
//     becomes the empty string.
//   - Ext    – file extension from the path (empty when absent).
//   - MIME   – mime.TypeByExtension(Ext).  Empty string when Ext == "".
//   - QueryRaw, Query, Fragment – unchanged from *url.URL.
//
// These keys are enough for modules and widgets to derive canonical URLs,
// file types, and routing context without the extra noise of scheme,
// directory, or isHTTPS.
package tenant

import (
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// URLInfo is stored in tenant.Context.
type URLInfo struct {
	Host     string     // blog.example.com
	Route    string     // "blog/2025/05" from path "/blog/2025/05"
	Ext      string     // ".html"
	MIME     string     // "text/html"
	QueryRaw string     // "tag=go"
	Query    url.Values // parsed query args
	Fragment string     // "top"
}

// newURLInfo builds URLInfo from the incoming request.
func newURLInfo(r *http.Request) URLInfo {
	host := stripPort(r.Host)

	// Derive Route by trimming leading/trailing slashes from the path.
	route := strings.Trim(r.URL.Path, "/")

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

// stripPort removes :port from the Host header when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
