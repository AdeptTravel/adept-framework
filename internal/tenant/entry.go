// internal/tenant/entry.go
//
// Tenant cache entry and aggregate.
//
// A live Tenant aggregates everything the router and Components need to
// serve a single site: its `site` row, per-site DB pool, in-memory config
// map, active Theme, Vault client, and a pre-parsed *template.Template tree
// for fast rendering.

package tenant

import (
	"html/template"

	"github.com/jmoiron/sqlx"

	"github.com/yanizio/adept/internal/tenant/meta"
	"github.com/yanizio/adept/internal/theme"
	"github.com/yanizio/adept/internal/vault"
)

// Cache entry wrapper
type entry struct {
	tenant   *Tenant
	lastSeen int64 // UnixNano
}

// Tenant aggregate
type Tenant struct {
	Meta     meta.Record        // Row from `site`
	Config   map[string]string  // site_config key→value
	DB       *sqlx.DB           // Per-site connection pool
	Theme    *theme.Theme       // Active theme
	Renderer *template.Template // Convenience alias: Theme.Renderer
	Vault    *vault.Client      // Vault client for secret lookup
}

// Close is called by the cache evictor on idle or LRU eviction.
func (t *Tenant) Close() error { return t.DB.Close() }

// -- component.TenantInfo implementations ----------------------------------
func (t *Tenant) GetDB() *sqlx.DB              { return t.DB }
func (t *Tenant) GetConfig() map[string]string { return t.Config }
func (t *Tenant) GetTheme() *theme.Theme       { return t.Theme }
func (t *Tenant) GetVault() *vault.Client      { return t.Vault }
