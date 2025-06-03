// modules/debug/debug.go
//
// Demo module that echoes URLInfo, remote IP, and user-agent data.
package debug

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/yanizio/adept/internal/module"
	"github.com/yanizio/adept/internal/tenant"
)

func init() {
	// Register at exact path /debug
	module.Register("/debug", handler)
}

// handler writes a JSON blob with selected context fields.
func handler(ten *tenant.Tenant, ctx *tenant.Context, w http.ResponseWriter, r *http.Request) {
	out := map[string]any{
		"host":      ctx.URL.Host,
		"route":     ctx.URL.Route,
		"ext":       ctx.URL.Ext,
		"mime":      ctx.URL.MIME,
		"query":     ctx.URL.QueryRaw,
		"ip":        clientIP(r),
		"ua":        r.UserAgent(),
		"ua_parsed": ctx.UA,
		"config":    ten.Config, // show default_ext etc.
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

// clientIP grabs the remote address without port.
func clientIP(r *http.Request) string {
	h, _, _ := net.SplitHostPort(r.RemoteAddr)
	return h
}
