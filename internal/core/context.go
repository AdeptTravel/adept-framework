//
//  internal/core/context.go
//
//  Central request context passed to every module, widget, and template.
//

package core

import (
	"net/http"

	"github.com/yanizio/adept/internal/requestinfo"
	"github.com/yanizio/adept/internal/tenant"
)

type Context struct {
	Site    *tenant.Site             // Lazy-loaded tenant record
	Request *http.Request            // Raw request (already SafeRequest-wrapped soon)
	Writer  http.ResponseWriter      // Convenience writer
	Params  map[string]string        // Route params (“slug”, etc.)
	Info    *requestinfo.RequestInfo // UA, Geo, URL, timestamp
}
