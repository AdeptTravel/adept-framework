// cmd/web/main.go
//
// Adept – HTTP entry point (Component + View system).
// Component routers are mounted automatically per tenant.
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

	// Side-effect imports: each component registers itself via init().
	_ "github.com/yanizio/adept/components/example"

	"github.com/yanizio/adept/internal/config"
	"github.com/yanizio/adept/internal/database"
	"github.com/yanizio/adept/internal/logger"
	"github.com/yanizio/adept/internal/middleware"
	"github.com/yanizio/adept/internal/tenant"
	"github.com/yanizio/adept/internal/vault"
)

// runningInTTY returns true when stdout is a character device.
func runningInTTY() bool {
	if fi, err := os.Stdout.Stat(); err == nil {
		return fi.Mode()&os.ModeCharDevice != 0
	}
	return false
}

func main() {
	// 0. Early dev logger so boot errors surface.
	devLog, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(devLog)

	// 1. Load configuration.
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("load config", zap.Error(err))
	}

	// 2. Production logger.
	logOut, err := logger.New(cfg.Paths.Root, runningInTTY())
	if err != nil {
		zap.L().Fatal("start logger", zap.Error(err))
	}

	// 3. Vault client (AppRole token already set).
	vaultCli, err := vault.New(context.Background(), logOut.Debugf)
	if err != nil {
		logOut.Fatalw("vault init", zap.Error(err))
	}

	// 4. Global DB connection.
	dsnFunc := func() string {
		c := config.Get()
		return fmt.Sprintf(c.Database.GlobalDSN, c.Database.GlobalPassword)
	}
	globalDB, err := database.OpenProvider(context.Background(), dsnFunc, database.Options{})
	if err != nil {
		logOut.Fatalw("connect global DB", zap.Error(err))
	}
	defer globalDB.Close()

	// 5. Tenant cache.
	cache := tenant.New(globalDB, 30*time.Minute, 100, logOut, vaultCli)

	// 6. Prometheus metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())

	// 7. Root handler: Host → tenant router.
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)
		ten, err := cache.Get(host)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ten.Router().ServeHTTP(w, r) // mounts all Components
	})

	// 8. Optional HTTPS enforcement.
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

// stripPort removes the ":port" suffix so "example.com:443" and
// "example.com" map to the same tenant cache entry.
func stripPort(h string) string {
	if i := strings.IndexByte(h, ':'); i != -1 {
		return h[:i]
	}
	return h
}
