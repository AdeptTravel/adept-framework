// cmd/web/main.go
//
// Adept – HTTP entry point.
//
// Request life-cycle
// ------------------
//
//  0. Load configuration (dotenv → YAML → env) via internal/config.
//  1. Start daily rotating logger (tees to console when running in a TTY).
//  2. Open global control-plane DB and log active-site count.
//  3. Build tenant cache (lazy-loads each site on first hit).
//  4. Expose Prometheus /metrics endpoint.
//  5. Build the root handler and wrap it with ForceHTTPS middleware
//     when cfg.HTTP.ForceHTTPS is true.
//  6. Root-handler flow:
//     • tenant lookup            – cache.Get(host)
//     • per-request Context      – Head builder, URLInfo, UA
//     • default <title>          – host name
//     • component dispatch       – module.Lookup(path)
//     • fallback template render – home.html
package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/yanizio/adept/internal/config"
	"github.com/yanizio/adept/internal/database"
	"github.com/yanizio/adept/internal/logger"
	"github.com/yanizio/adept/internal/middleware"
	"github.com/yanizio/adept/internal/module"
	"github.com/yanizio/adept/internal/tenant"
	"github.com/yanizio/adept/internal/viewhelpers"
)

// runningInTTY returns true when stdout is a character device.
func runningInTTY() bool {
	if fi, err := os.Stdout.Stat(); err == nil {
		return fi.Mode()&os.ModeCharDevice != 0
	}
	return false
}

func main() {

	tempLog, _ := zap.NewDevelopment() // *zap.Logger
	zap.ReplaceGlobals(tempLog)
	tempLog.Sugar().Infow("bootstrap", "step", "init")

	//
	// ── 0.  Load configuration ─────────────────────────────────────────
	//
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("load config", zap.Error(err))
	}

	logOut, err := logger.New(cfg.Paths.Root, runningInTTY())
	if err != nil {
		zap.L().Fatal("start logger", zap.Error(err))
	}

	//
	// ── 1.  Global DB connect ───────────────────────────────────────────
	//
	logOut.Infow("connecting to global DB")
	globalDB, err := database.Open(cfg.Database.GlobalDSN)
	if err != nil {
		logOut.Fatalw("connect global DB", zap.Error(err))
	}
	defer globalDB.Close()
	logOut.Infow("global DB online")

	// Log active-site count as an early sanity check.
	var active int
	_ = globalDB.Get(&active, `
	    SELECT COUNT(*) FROM site
	    WHERE suspended_at IS NULL AND deleted_at IS NULL`)
	logOut.Infof("%d active site(s) found", active)

	//
	// ── 2.  Tenant cache (lazy site loader) ─────────────────────────────
	//
	cache := tenant.New(globalDB, tenant.IdleTTL, tenant.MaxEntries, logOut)

	//
	// ── 3.  Metrics endpoint ────────────────────────────────────────────
	//
	http.Handle("/metrics", promhttp.Handler())

	//
	// ── 4.  Root handler: tenant lookup → component dispatch → render ───
	//
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)

		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Register UA helpers once per renderer instance (idempotent).
		ten.Renderer.Funcs(viewhelpers.FuncMap())

		//
		// Build per-request Context and seed head defaults.
		//
		ctx := tenant.NewContext(r)

		ctx.Head.SetTitle(host) // default title = host name
		ctx.Head.Meta(`<meta charset="utf-8">`)
		ctx.Head.Link(`<link rel="icon" href="/favicon.ico">`)

		//
		// Component dispatch – exact path match.
		//
		if h := module.Lookup(r.URL.Path); h != nil {
			h(ten, ctx, w, r)
			return
		}

		//
		// Fallback: render home.html via theme renderer.
		//
		data := map[string]any{
			"Ctx":  ctx,
			"Head": ctx.Head,
		}
		if err := ten.Renderer.ExecuteTemplate(w, "home.html", data); err != nil {
			logOut.Errorw("render error", zap.Error(err))
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	})

	//
	// ── 5.  Wrap with HTTPS-enforcement middleware (skip localhost) ────
	//
	var handler http.Handler = root
	if cfg.HTTP.ForceHTTPS {
		handler = middleware.ForceHTTPS(cache, root)
	}
	http.Handle("/", handler)

	logOut.Infow("listening", "addr", cfg.HTTP.ListenAddr)
	if err := http.ListenAndServe(cfg.HTTP.ListenAddr, nil); err != nil {
		logOut.Fatalw("http server", zap.Error(err))
	}
}

// stripPort removes any “:port” suffix from the Host header.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
