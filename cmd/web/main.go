// Command web boots the Adept Framework multi-tenant HTTP server.
//
// Workflow
//   • Load environment variables from /usr/local/etc/adept-framework/global.env
//     if present, else fall back to .env in the working directory.
//   • Connect to the global control-plane database using GLOBAL_DB_DSN.
//   • Initialise a tenant.Cache that lazy-loads site records and opens per-
//     tenant connection pools on demand.
//   • Handle every HTTP request by mapping r.Host to a cached tenant,
//     loading it if necessary, and updating the LastSeen timestamp.
//   • Expose Prometheus metrics on /metrics and a placeholder root handler.
/*
   Required imports
     go get golang.org/x/sync
     go get github.com/prometheus/client_golang
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/tenant"
)

const (
	serverEnvPath = "/usr/local/etc/adept-framework/global.env"
	listenAddr    = ":8080"
)

// loadEnv first tries the server-level env file.  If absent, it falls back
// to .env in the working directory.
func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

func init() {
	loadEnv()
}

func main() {
	/// Open the global (control-plane) database.

	globalDSN := os.Getenv("GLOBAL_DB_DSN")
	if globalDSN == "" {
		log.Fatal("GLOBAL_DB_DSN is not set")
	}
	globalDB, err := database.Open(globalDSN)
	if err != nil {
		log.Fatalf("connect global DB: %v", err)
	}
	defer globalDB.Close()

	/// Initialise the tenant cache with default TTL and capacity.

	cache := tenant.New(
		globalDB,
		tenant.IdleTTL,    // 30-minute idle timeout
		tenant.MaxEntries, // 100 tenants before LRU eviction
	)

	/// Register HTTP handlers.

	// Metrics endpoint for Prometheus scraping.
	http.Handle("/metrics", promhttp.Handler())

	// Main catch-all handler that dispatches by Host header.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)

		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// TODO: replace with real router and template engine.
		fmt.Fprintf(w, "Hello from %s (theme=%s)\n", ten.Meta.Host, ten.Meta.Theme)
	})

	/// Start the HTTP server.

	log.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// stripPort removes the :port suffix from a Host header when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
