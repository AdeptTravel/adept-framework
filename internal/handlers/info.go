package handlers

import (
	"html/template"
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/config"
	"github.com/AdeptTravel/adept-framework/internal/site"
)

var pageTmpl = template.Must(template.New("info").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Request Debug</title></head>
<body>
    <h1>Request / Session Information</h1>
    <ul>
        <li><strong>Timestamp:</strong> {{.Time}}</li>
        <li><strong>User-Agent:</strong> {{.UserAgent}}</li>
        <li><strong>Client IP:</strong> {{.IP}}</li>
        <li><strong>URL Host:</strong> {{.Host}}</li>
        <li><strong>URL Path:</strong> {{.Path}}</li>
    </ul>
</body>
</html>`))

// Info returns an http.HandlerFunc that writes the ReqInfo page.
func Info(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devHost := ""
		if cfg.App.UseDevHost {
			devHost = cfg.App.DevHost
		}
		ri := site.Extract(r, devHost)
		if err := pageTmpl.Execute(w, ri); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	}
}
