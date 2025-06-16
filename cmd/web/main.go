// cmd/web/main.go
//
// Adept – HTTP entry point.
//
// Responsibilities
// ----------------
//   1. Parse configuration and bootstrap logging / Vault / DB.
//   2. Maintain an LRU cache of live *tenant.Tenant aggregates.
//   3. Expose Prometheus metrics at /metrics.
//   4. Route every incoming host to its per-tenant chi.Router.
//   5. Enforce optional HTTPS + always set security headers.
//   6. Start an http.Server with sane production timeouts.
//
// Notes
// -----
// • Security headers are added via middleware.Security.
// • Connection timeouts come from internal/server.New.
// • Oxford commas, two spaces after periods.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	// Side-effect import: example component registers itself via init().
	_ "github.com/yanizio/adept/components/example"

	"github.com/yanizio/adept/internal/config"
	"github.com/yanizio/adept/internal/database"
	"github.com/yanizio/adept/internal/logger"
	"github.com/yanizio/adept/internal/middleware"
	"github.com/yanizio/adept/internal/server"
	"github.com/yanizio/adept/internal/tenant"
	"github.com/yanizio/adept/internal/vault"
)

// runningInTTY returns true when stdout is a character device (dev mode).
func runningInTTY() bool {
	if fi, err := os.Stdout.Stat(); err == nil {
		return fi.Mode()&os.ModeCharDevice != 0
	}
	return false
}

func main() {
	// 0. Early dev logger so boot errors appear.
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))

	// 1. Load configuration (yaml + Vault resolution).
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("config load", zap.Error(err))
	}

	// 2. Structured JSON logger (rotates daily).
	logOut, err := logger.New(cfg.Paths.Root, runningInTTY())
	if err != nil {
		zap.L().Fatal("logger init", zap.Error(err))
	}

	// 3. Vault client (AppRole token already exported at startup).
	vaultCli, err := vault.New(context.Background(), logOut.Debugf)
	if err != nil {
		logOut.Fatalw("vault init", zap.Error(err))
	}

	// 4. Global DB pool.
	dsn := func() string {
		c := config.Get()
		return fmt.Sprintf(c.Database.GlobalDSN, c.Database.GlobalPassword)
	}
	globalDB, err := database.OpenProvider(context.Background(), dsn, database.Options{})
	if err != nil {
		logOut.Fatalw("global DB connect", zap.Error(err))
	}
	defer globalDB.Close()

	// 5. Tenant LRU cache (30-min idle TTL, max 100 tenants).
	cache := tenant.New(globalDB, 30*time.Minute, 100, logOut, vaultCli)

	// 6. Prometheus metrics endpoint.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// 7. Root handler: strip :port → cache → tenant.Router().
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)
		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ten.Router().ServeHTTP(w, r)
	})

	// 8. Middleware chain: Security always, ForceHTTPS optional.
	var handler http.Handler = root
	handler = middleware.Security(handler)
	if cfg.HTTP.ForceHTTPS {
		handler = middleware.ForceHTTPS(cache, handler)
	}
	mux.Handle("/", handler)

	// 9. Build server with timeouts, then listen.
	srv := server.New(cfg.HTTP.ListenAddr, mux)

	logOut.Infow("listening", "addr", cfg.HTTP.ListenAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logOut.Fatalw("http server", zap.Error(err))
	}
}

// stripPort("example.com:443") → "example.com".
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
