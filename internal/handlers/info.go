package handlers

import (
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/config"
	"github.com/AdeptTravel/adept-framework/internal/geo"
	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/AdeptTravel/adept-framework/internal/view"
)

// Info renders the request debug page through the view engine.
func Info(cfg config.Config, gdb *geo.DB, engine *view.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devHost := ""
		if cfg.App.UseDevHost {
			devHost = cfg.App.DevHost
		}
		ri := site.Extract(r, devHost, gdb)

		if err := engine.Exec(w, "minimal", "debug.html", ri); err != nil {
			http.Error(w, "view error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
