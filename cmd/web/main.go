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

const serverEnvPath = "/usr/local/etc/adept-framework/global.env"
const listenAddr = ":8080"

func loadEnv() {
	if _, err := os.Stat(serverEnvPath); err == nil {
		_ = godotenv.Load(serverEnvPath)
		return
	}
	_ = godotenv.Load()
}

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

	// Global DB
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

	var active int
	_ = globalDB.Get(&active, `
	    SELECT COUNT(*) FROM site
	    WHERE suspended_at IS NULL AND deleted_at IS NULL`)
	logOut.Printf("%d active site(s) found in site table", active)

	// Tenant cache
	cache := tenant.New(globalDB, tenant.IdleTTL, tenant.MaxEntries, logOut)

	// HTTP handlers
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)
		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Render home page; ctx=nil until RequestInfo is wired
		if err := ten.Renderer.ExecuteTemplate(w, "home.html", nil); err != nil {
			logOut.Printf("render error: %v", err)
			http.Error(w, "template error", 500)
		}
	})

	logOut.Printf("listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logOut.Fatalf("http server: %v", err)
	}
}

func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
