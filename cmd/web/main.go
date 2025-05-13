package main

import (
	"log"
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/config"
	"github.com/AdeptTravel/adept-framework/internal/geo"
	"github.com/AdeptTravel/adept-framework/internal/modules/demo"
	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/AdeptTravel/adept-framework/internal/view"
	"github.com/AdeptTravel/adept-framework/themes"
)

func main() {
	// ------------------------------------------------------------------ config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ------------------------------------------------------------------ geo-ip (optional)
	var gdb *geo.DB
	if path := cfg.GeoIP.DBPath; path != "" {
		gdb, err = geo.Open(path)
		if err != nil {
			log.Fatalf("geoip: %v", err)
		}
		defer gdb.Close()
	}

	// ------------------------------------------------------------------ view engine
	vEngine, err := view.New(themes.FS, view.FuncMap())
	if err != nil {
		log.Fatalf("view engine: %v", err)
	}

	// register demo module templates under the "demo/" namespace
	if err := demo.RegisterTemplates(vEngine); err != nil {
		log.Fatalf("demo templates: %v", err)
	}

	// ------------------------------------------------------------------ router
	mux := http.NewServeMux()
	mux.HandleFunc("/", demo.Handler(vEngine, func(r *http.Request) demo.Data {
		devHost := ""
		if cfg.App.UseDevHost {
			devHost = cfg.App.DevHost
		}
		// Extract request/session information as demo.Data
		return site.Extract(r, devHost, gdb)
	}))

	log.Println("⇢ listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
