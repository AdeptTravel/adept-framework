// internal/routing/alias.go
//
// Alias-resolution cache and middleware (import-cycle safe).
//
// Context
// -------
// Tenants in alias or both routing modes expose friendly paths that must be
// rewritten to absolute Component paths.  A lightweight interface—AliasTenant—
// keeps this package independent of *tenant*, avoiding cyclic imports.
//
// Workflow
// --------
//   1. Tenant cold-load constructs AliasCache via routing.NewAliasCache().
//   2. tenant.Router() wires routing.Middleware(tenant) early in the chain.
//   3. Middleware rewrites r.URL.Path on cache hit; otherwise falls through or
//      404s per routing mode.
//
// Notes
// -----
// • Oxford commas, two spaces after periods.
// • Max line length 100 columns.

package routing

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------
// AliasCache
// -----------------------------------------------------------------------------

// AliasCache stores alias→target pairs plus TTL/version state.  Zero value is
// unusable; construct with NewAliasCache.
type AliasCache struct {
	mu       sync.RWMutex
	data     map[string]string
	loadedAt time.Time
	ttl      time.Duration
	version  int
	db       *sql.DB
}

// NewAliasCache returns a ready cache with the specified TTL.
func NewAliasCache(db *sql.DB, ttl time.Duration) *AliasCache {
	return &AliasCache{data: map[string]string{}, db: db, ttl: ttl}
}

// Load refreshes all aliases from route_alias.
func (c *AliasCache) Load(ctx context.Context) error {
	rows, err := c.db.QueryContext(ctx, `SELECT alias_path, target_path FROM route_alias`)
	if err != nil {
		return err
	}
	defer rows.Close()

	fresh := make(map[string]string)
	for rows.Next() {
		var alias, target string
		if err := rows.Scan(&alias, &target); err != nil {
			return err
		}
		fresh[alias] = target
	}
	if err := rows.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	c.data = fresh
	c.loadedAt = time.Now()
	c.mu.Unlock()

	zap.L().Debug("alias cache load",
		zap.Int("count", len(fresh)))
	return nil
}

func (c *AliasCache) lookup(path string) (string, bool) {
	c.mu.RLock()
	target, ok := c.data[path]
	stale := time.Since(c.loadedAt) > c.ttl
	c.mu.RUnlock()
	return target, ok && !stale
}

func (c *AliasCache) needsRefresh(currentVer int) bool {
	c.mu.RLock()
	stale := time.Since(c.loadedAt) > c.ttl || currentVer != c.version
	c.mu.RUnlock()
	return stale
}

func (c *AliasCache) setVersion(v int) {
	c.mu.Lock()
	c.version = v
	c.mu.Unlock()
}

// -----------------------------------------------------------------------------
// Middleware factory
// -----------------------------------------------------------------------------

// AliasTenant is the minimal contract a *tenant.Tenant must fulfil.  Defined
// here to avoid importing the full tenant package and thus prevent import
// cycles.
type AliasTenant interface {
	RoutingMode() string
	RouteVersion() int
	AliasCache() *AliasCache
}

const (
	RouteModeAbsolute  = "absolute"
	RouteModeAliasOnly = "alias"
	RouteModeBoth      = "both"
)

// Middleware returns a tenant-bound Chi middleware that rewrites alias paths.
func Middleware(t AliasTenant) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			mode := t.RoutingMode()
			if mode == RouteModeAbsolute {
				next.ServeHTTP(w, r)
				return
			}

			cache := t.AliasCache()
			if cache.needsRefresh(t.RouteVersion()) {
				if err := cache.Load(r.Context()); err == nil {
					cache.setVersion(t.RouteVersion())
				} else {
					zap.L().Warn("alias cache reload failed", zap.Error(err))
				}
			}

			if target, ok := cache.lookup(r.URL.Path); ok {
				original := r.URL.Path
				r.URL.Path = target
				r.RequestURI = target
				zap.L().Debug("alias rewrite",
					zap.String("from", original),
					zap.String("to", target))

				next.ServeHTTP(w, r)
				return
			}

			if mode == RouteModeAliasOnly {
				http.NotFound(w, r)
				return
			}

			// mode == both and alias miss
			next.ServeHTTP(w, r)
		})
	}
}
