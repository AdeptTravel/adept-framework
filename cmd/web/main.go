// Command web boots the Adept Framework multi-tenant HTTP server.
//
// Startup sequence:
//  1. Load environment variables from /usr/local/etc/adept-framework/global.env
//     if it exists, otherwise load .env in the current directory.
//  2. Connect to the global control-plane database using GLOBAL_DB_DSN.
//  3. Query every site with status = 'Active'.
//  4. Open each tenant database from its stored DSN.
//  5. Dispatch requests by the Host header to the correct tenant.
//  6. Listen on :8080 and serve a placeholder handler.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/site"
)

const (
	serverEnvPath = "/usr/local/etc/adept-framework/global.env"
	listenAddr    = ":8080"
	logDir        = "log"                     // ./log/YYYY-MM-DD.log
	geoCity       = "data/GeoLite2-City.mmdb" // GeoLite2 database
)

// tenant groups one site’s metadata with its open connection pool.
type tenant struct {
	meta site.Record
	db   *sqlx.DB
}

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

	/// Load every active site and open each tenant database.

	records, err := site.AllActive(globalDB)
	if err != nil {
		log.Fatalf("query site table: %v", err)
	}
	if len(records) == 0 {
		log.Fatal("no active sites found")
	}

	tenants := make(map[string]tenant, len(records))
	for _, rec := range records {
		db, err := database.Open(rec.DSN)
		if err != nil {
			log.Printf("SKIP %s: cannot connect to tenant DB (%v)", rec.Host, err)
			continue
		}
		tenants[rec.Host] = tenant{meta: rec, db: db}
		log.Printf("loaded tenant %s (theme=%s)", rec.Host, rec.Theme)
	}
	if len(tenants) == 0 {
		log.Fatal("no tenant databases opened successfully")
	}

	/// Register the global HTTP handler that dispatches by Host.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)
		t, ok := tenants[host]
		if !ok {
			http.NotFound(w, r)
			return
		}

		// TODO: replace this with the real router and template engine.
		fmt.Fprintf(w, "Hello from %s (theme=%s)\n", t.meta.Host, t.meta.Theme)
	})

	/// Start the server.

	log.Printf("listening on %s for %d tenants", listenAddr, len(tenants))
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
