// cmd/web/main.go
//
// Entry point for the Adept Framework HTTP server.
//
// Responsibilities
// ----------------
//
//   - Load environment variables (server-wide file, then .env for dev).
//
//   - Initialise rotating logger (tees to console when running in a TTY).
//
//   - Connect to the global control-plane database and print the count of
//     active sites as an early sanity check.
//
//   - Build the tenant cache (lazy-loads each site on first request).
//
//   - Register Prometheus metrics endpoint.
//
//   - Wrap the root handler with middleware.ForceHTTPS so every non-localhost
//     host is 308-redirected to HTTPS when hit over plain HTTP.
//
//   - Root handler:
//
//     – Looks up the tenant.
//     – Builds a per-request tenant.Context (includes head.Builder).
//     – Seeds the page <title> from site.Record.Title.
//     – Adds two example head tags (charset + favicon).
//     – Renders themes/base/templates/home.html.
//
// Style guide
// -----------
// Blank “//” lines frame large blocks; inline comments use a single “//”.
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
	"github.com/AdeptTravel/adept-framework/internal/tenant"
)

const (
	serverEnvPath = "/usr/local/etc/adept-framework/global.env"
	listenAddr    = ":8080"
)

// loadEnv loads environment variables from a jail-wide file first, then
// falls back to .env in the working directory (dev workflow).
func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

// runningInTTY lets the logger decide whether to tee to stdout.
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
	// Connect to global control-plane database.
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

	// Log the number of active sites to confirm correct DB selection.
	var active int
	_ = globalDB.Get(&active, `
	    SELECT COUNT(*) FROM site
	    WHERE suspended_at IS NULL AND deleted_at IS NULL`)
	logOut.Printf("%d active site(s) found in site table", active)

	//
	// Tenant cache (lazy-loads a site on first request).
	//
	cache := tenant.New(globalDB, tenant.IdleTTL, tenant.MaxEntries, logOut)

	//
	// Metrics endpoint (no middleware).
	//
	http.Handle("/metrics", promhttp.Handler())

	//
	// Root handler wrapped with ForceHTTPS(middleware).
	//
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)

		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		//
		// Build per-request context and seed head defaults.
		//
		ctx := tenant.NewContext(r)

		// Default <title> from site.Record.Title; modules may override.
		if ten.Meta.Title != "" {
			ctx.Head.SetTitle(ten.Meta.Title)
		}

		// Example core tags; modules will push more as needed.
		ctx.Head.Meta(`<meta charset="utf-8">`)
		ctx.Head.Link(`<link rel="icon" href="/favicon.ico">`)

		data := map[string]any{
			"Ctx":  ctx,      // placeholder for future helpers
			"Head": ctx.Head, // theme base layout prints Head slices
		}

		if err := ten.Renderer.ExecuteTemplate(w, "home.html", data); err != nil {
			logOut.Printf("render error: %v", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	})

	// Register root with HTTPS-force middleware (skips localhost).
	http.Handle("/", middleware.ForceHTTPS(cache, root))

	logOut.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logOut.Fatalf("http server: %v", err)
	}
}

// stripPort removes :port from the Host header when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
