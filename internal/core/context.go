//
//  internal/core/context.go
//
//  Central request context passed to every module, widget, and template.
//

package core

import (
	"net/http"

	"github.com/adepttravel/adept-framework/internal/requestinfo"
	"github.com/adepttravel/adept-framework/internal/tenant"
)

type Context struct {
	Site    *tenant.Site             // Lazy-loaded tenant record
	Request *http.Request            // Raw request (already SafeRequest-wrapped soon)
	Writer  http.ResponseWriter      // Convenience writer
	Params  map[string]string        // Route params (“slug”, etc.)
	Info    *requestinfo.RequestInfo // UA, Geo, URL, timestamp
}
