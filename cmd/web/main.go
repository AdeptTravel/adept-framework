// cmd/web/main.go
//
// Adept Framework – HTTP entry point.
//
// Request life-cycle
// ------------------
//
//  1. Load env vars (jail-wide file → .env fallback).
//
//  2. Start daily rotating logger (tees to console when running in a TTY).
//
//  3. Open global control-plane DB and log active-site count.
//
//  4. Build tenant-cache (lazy-loads each site on first hit).
//
//  5. Expose Prometheus /metrics endpoint.
//
//  6. Build the root handler and wrap it with ForceHTTPS middleware
//     so every non-localhost HTTP request is 308-redirected to HTTPS.
//
//  7. Root-handler flow:
//
//     • tenant lookup            – cache.Get(host)
//     • per-request Context      – head.Builder + URLInfo + UA
//     • default <title>          – site.Record.Title
//     • module dispatch          – module.Lookup(path)
//     • fallback template render – home.html
//
// Large comment blocks are framed by blank “//” lines; inline comments use
// a single “//”.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/logger"
	"github.com/AdeptTravel/adept-framework/internal/middleware"
	"github.com/AdeptTravel/adept-framework/internal/module"
	"github.com/AdeptTravel/adept-framework/internal/tenant"
	"github.com/AdeptTravel/adept-framework/internal/viewhelpers"

	_ "github.com/AdeptTravel/adept-framework/modules/debug" // demo module
)

const (
	serverEnvPath = "/usr/local/etc/adept-framework/global.env"
	listenAddr    = ":8080"
)

// loadEnv prefers the jail-wide env file; on dev it falls back to .env.
func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

// runningInTTY returns true when stdout is a character device.
func runningInTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func init() { loadEnv() }

func main() {
	rootDir, _ := os.Getwd()
	logOut, err := logger.New(rootDir, runningInTTY())
	if err != nil {
		log.Fatalf("start logger: %v", err)
	}

	//
	// ── 1.  Global DB connect ───────────────────────────────────────────
	//
	dsn := os.Getenv("GLOBAL_DB_DSN")
	if dsn == "" {
		logOut.Fatal("GLOBAL_DB_DSN is not set")
	}
	logOut.Println("connecting to global DB …")
	globalDB, err := database.Open(dsn)
	if err != nil {
		logOut.Fatalf("connect global DB: %v", err)
	}
	defer globalDB.Close()
	logOut.Println("global DB online")

	// Log active-site count as an early sanity check.
	var active int
	_ = globalDB.Get(&active, `
	    SELECT COUNT(*) FROM site
	    WHERE suspended_at IS NULL AND deleted_at IS NULL`)
	logOut.Printf("%d active site(s) found", active)

	//
	// ── 2.  Tenant cache (lazy site loader) ─────────────────────────────
	//
	cache := tenant.New(globalDB, tenant.IdleTTL, tenant.MaxEntries, logOut)

	//
	// ── 3.  Metrics endpoint ────────────────────────────────────────────
	//
	http.Handle("/metrics", promhttp.Handler())

	//
	// ── 4.  Root handler: tenant lookup → module dispatch → render ─────
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

		if ten.Meta.Title != "" {
			ctx.Head.SetTitle(ten.Meta.Title)
		}
		ctx.Head.Meta(`<meta charset="utf-8">`)
		ctx.Head.Link(`<link rel="icon" href="/favicon.ico">`)

		//
		// Module dispatch – exact path match (e.g., /debug).
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
			logOut.Printf("render error: %v", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	})

	//
	// ── 5.  Wrap with HTTPS-enforcement middleware (skip localhost) ────
	//
	http.Handle("/", middleware.ForceHTTPS(cache, root))

	logOut.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logOut.Fatalf("http server: %v", err)
	}
}

// stripPort removes any “:port” suffix from the Host header.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
