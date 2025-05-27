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

// loadEnv loads server-level env vars, falling back to .env for dev.
func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

// runningInTTY returns true when stdout is a TTY.
func runningInTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func init() { loadEnv() }

func main() {
	rootDir, _ := os.Getwd()
	logOut, err := logger.New(rootDir, runningInTTY())
	if err != nil {
		log.Fatalf("start logger: %v", err)
	}

	//
	// Connect to the global database.
	//
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

	// Count active sites to confirm correct database selection.
	var activeCount int
	const countSQL = `
	    SELECT COUNT(*)
	    FROM   site
	    WHERE  suspended_at IS NULL
	      AND  deleted_at   IS NULL`
	if err := globalDB.Get(&activeCount, countSQL); err != nil {
		logOut.Printf("could not count active sites: %v", err)
	} else {
		logOut.Printf("%d active site(s) found in site table", activeCount)
	}

	//
	// Initialise the tenant cache.
	//
	cache := tenant.New(globalDB, tenant.IdleTTL, tenant.MaxEntries, logOut)

	// ----------------------------------------------------------------
	// HTTP handlers
	// ----------------------------------------------------------------
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)

		ten, err := cache.Get(host)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintf(w, "Hello from %s (theme=%s)\n", ten.Meta.Host, ten.Meta.Theme)
	})

	logOut.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logOut.Fatalf("http server: %v", err)
	}
}

// stripPort removes any :port suffix from the Host header.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
