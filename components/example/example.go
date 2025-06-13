// components/example/example.go
//
// Example Component – shows the request’s UA, IP, and Geo details.
package example

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yanizio/adept/internal/component"
	"github.com/yanizio/adept/internal/requestinfo"
)

// compile-time assertions
var (
	_ component.Component   = (*Comp)(nil)
	_ component.Initializer = (*Comp)(nil)
)

// Comp implements component.Component; no per-tenant state needed.
type Comp struct{}

func (c *Comp) Name() string                      { return "example" }
func (c *Comp) Migrations() []string              { return nil }
func (c *Comp) Init(_ component.TenantInfo) error { return nil }

func (c *Comp) Routes() chi.Router {
	r := chi.NewRouter()

	// HTML page
	r.Get("/example", func(w http.ResponseWriter, r *http.Request) {
		ri := requestinfo.FromContext(r.Context())
		if ri == nil {
			http.Error(w, "request info not available", http.StatusInternalServerError)
			return
		}
		tpl := template.Must(template.New("example").Parse(`<!doctype html>
<html>
<head><title>Example – Request Info</title></head>
<body>
  <h1>Request Details</h1>
  <ul>
    <li><strong>IP:</strong> {{.IP}}</li>
    <li><strong>Country:</strong> {{.Country}}</li>
    <li><strong>City:</strong> {{.City}}</li>
    <li><strong>Browser:</strong> {{.Browser}} ({{.Device}})</li>
    <li><strong>OS:</strong> {{.OS}} {{.OSVer}}</li>
    <li><strong>Bot:</strong> {{.IsBot}}</li>
  </ul>
</body>
</html>`))
		data := map[string]any{
			"IP":      ri.Geo.IP,
			"Country": ri.Geo.CountryISO,
			"City":    ri.Geo.City,
			"Browser": ri.UA.Browser,
			"Device":  ri.UA.Device,
			"OS":      ri.UA.OS,
			"OSVer":   ri.UA.OSVersion,
			"IsBot":   ri.UA.IsBot,
		}
		if err := tpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// JSON endpoint
	r.Get("/api/example", func(w http.ResponseWriter, r *http.Request) {
		ri := requestinfo.FromContext(r.Context())
		if ri == nil {
			http.Error(w, "request info not available", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ri); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return r
}

// Register component at package init.
func init() {
	component.Register(&Comp{})
}
