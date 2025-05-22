// Command web boots the Adept Framework multi-tenant HTTP server.
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
	"github.com/AdeptTravel/adept-framework/internal/logger"
	"github.com/AdeptTravel/adept-framework/internal/tenant"
)

const serverEnvPath = "/usr/local/etc/adept-framework/global.env"
const listenAddr = ":8080"

// loadEnv first tries the server-level env file.  If absent, it falls back
// to .env in the working directory.
func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

// runningInTTY returns true when stdout is a character device.  The result is
// used to decide whether to tee log output to the console.
func runningInTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func init() {
	loadEnv()
}

func main() {
	rootDir, _ := os.Getwd()
	logOut, err := logger.New(rootDir, runningInTTY())
	if err != nil {
		log.Fatalf("start logger: %v", err)
	}

	/// Open the global (control-plane) database.

	globalDSN := os.Getenv("GLOBAL_DB_DSN")
	if globalDSN == "" {
		logOut.Fatal("GLOBAL_DB_DSN is not set")
	}

	logOut.Println("connecting to global DB …")
	globalDB, err := database.Open(globalDSN)
	if err != nil {
		logOut.Fatalf("connect global DB: %v", err)
	}
	defer globalDB.Close()
	logOut.Println("global DB online")

	/// Initialise the tenant cache.

	cache := tenant.New(
		globalDB,
		tenant.IdleTTL,
		tenant.MaxEntries,
	)

	/// Register HTTP handlers.

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)

		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Placeholder response
		fmt.Fprintf(w, "Hello from %s (theme=%s)\n", ten.Meta.Host, ten.Meta.Theme)
	})

	/// Start the HTTP server.

	logOut.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logOut.Fatalf("http server: %v", err)
	}
}

// stripPort removes the :port suffix from a Host header when present.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
