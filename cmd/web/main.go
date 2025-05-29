// cmd/web/main.go
//
// Entry point for the Adept Framework HTTP server.  This file:
//
//   - Loads env vars (server-level file, then .env).
//   - Opens the global control-plane database.
//   - Prints the count of active sites for easy troubleshooting.
//   - Builds the tenant cache (lazy-loaded sites).
//   - Starts an http.Server on :8080 with a minimal handler that
//     renders home.html through the per-tenant template renderer.
//
// The handler now constructs a request-scoped tenant.Context that owns
// a head.Builder, letting core or modules push <meta>, <link>, <script>,
// and JSON-LD blocks into the eventual <head> section.
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
	"github.com/AdeptTravel/adept-framework/internal/tenant"
)

const (
	serverEnvPath = "/usr/local/etc/adept-framework/global.env"
	listenAddr    = ":8080"
)

// loadEnv attempts to load env vars from the server-wide file first, then
// falls back to .env in the working directory (dev workflow).
func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

// runningInTTY detects whether stdout is a character device.  Used by the
// logger to decide if it should tee to the console.
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
	// Open global database.
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

	// Quick sanity check on active sites.
	var active int
	_ = globalDB.Get(&active, `
	    SELECT COUNT(*) FROM site
	    WHERE suspended_at IS NULL AND deleted_at IS NULL`)
	logOut.Printf("%d active site(s) found in site table", active)

	//
	// Tenant cache (lazy-loads sites on first request).
	//
	cache := tenant.New(globalDB, tenant.IdleTTL, tenant.MaxEntries, logOut)

	//
	// HTTP routes.
	//
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)

		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Build request context with a head.Builder so modules and
		// core can push tags into the page head.
		ctx := tenant.NewContext(r)
		ctx.Head.Meta(`<meta charset="utf-8">`)                // example
		ctx.Head.Link(`<link rel="icon" href="/favicon.ico">`) // example
		// Default <title> from the site record.
		//if ten.Meta.Title != "" {
		ctx.Head.SetTitle("Test")
		//}

		data := map[string]any{
			"Ctx":  ctx,      // future template helpers
			"Head": ctx.Head, // used by theme base layout
		}

		if err := ten.Renderer.ExecuteTemplate(w, "home.html", data); err != nil {
			logOut.Printf("render error: %v", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	})

	logOut.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logOut.Fatalf("http server: %v", err)
	}
}

// stripPort trims the :port suffix from the Host header when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
