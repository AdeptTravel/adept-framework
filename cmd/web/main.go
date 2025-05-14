package main

import (
	"log"
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/config"
	"github.com/AdeptTravel/adept-framework/internal/geo"
	"github.com/AdeptTravel/adept-framework/internal/middleware"
	"github.com/AdeptTravel/adept-framework/internal/modules/demo"
	"github.com/AdeptTravel/adept-framework/internal/requestctx"
	"github.com/AdeptTravel/adept-framework/internal/view"
	"github.com/AdeptTravel/adept-framework/themes"
)

func main() {
	// ------------------------------------------------------------------ load config
	gCfg, _ := config.LoadGlobal("config.yaml")

	// existing DB / GeoIP bootstrap using gCfg …

	sitesRoot := "sites"
	mux := http.NewServeMux()
	// register routes …

	handler := middleware.AttachSiteConfig(sitesRoot)(
		middleware.AttachRequestCtx(gdb, "")(mux))

	log.Fatal(http.ListenAndServe(":8080", handler))

	// ------------------------------------------------------------------ optional Geo-IP
	var gdb *geo.DB
	if path := cfg.GeoIP.DBPath; path != "" {
		gdb, err = geo.Open(path)
		if err != nil {
			log.Fatalf("geoip: %v", err)
		}
		defer gdb.Close()
	}

	// ------------------------------------------------------------------ view engine
	viewEngine, err := view.New(themes.FS, view.FuncMap())
	if err != nil {
		log.Fatalf("view engine: %v", err)
	}
	if err := demo.RegisterTemplates(viewEngine); err != nil {
		log.Fatalf("demo templates: %v", err)
	}

	// ------------------------------------------------------------------ HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", demo.Handler(viewEngine, func(r *http.Request) demo.Data {
		if rc := requestctx.From(r); rc != nil {
			return *rc // *rc is now exactly demo.Data
		}
		return demo.Data{} // fallback
	}))

	// ------------------------------------------------------------------ middleware chain
	devHost := ""
	if cfg.App.UseDevHost {
		devHost = cfg.App.DevHost
	}
	handlerWithCtx := middleware.AttachRequestCtx(gdb, devHost)(mux)

	// ------------------------------------------------------------------ serve
	log.Println("⇢ listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handlerWithCtx))
}
