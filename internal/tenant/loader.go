// internal/tenant/loader.go
//
// host → Tenant loader (Vault-aware).
//
// Context
// -------
// The cache’s slow-path calls `loadSite` to transform an incoming Host header
// into a live *Tenant.  The function performs five blocking steps:
//
//   1. Resolve the lookup host (alias “localhost” if needed) and fetch the
//      `site` row (`meta.ByHost`).
//   2. Fetch key-value pairs from `site_config`.
//   3. Resolve the tenant DB password from Vault and build the DSN.
//   4. Open a small per-site DB pool with retry and sane limits.
//   5. Parse the active Theme’s templates.
//
// Heavy resources (DB pool, parsed templates) are created once per cache
// entry and reused until eviction.
//
// Notes
// -----
// • Host sanitising and DSN construction live in helpers.go
//   (`resolveLookupHost`, `sanitizeHost`, `buildTenantDSN`).
// • Oxford commas, two spaces after periods.

package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yanizio/adept/internal/database"
	"github.com/yanizio/adept/internal/tenant/meta"
	"github.com/yanizio/adept/internal/theme"
	"github.com/yanizio/adept/internal/vault"
)

//
// loader
//

// loadSite executes the slow-path load in five well-defined steps.
func loadSite(
	ctx context.Context,
	global *sqlx.DB,
	host string,
	vcli *vault.Client,
) (*Tenant, error) {

	// 1. resolve alias and fetch site row
	lookupHost := resolveLookupHost(host)
	rec, err := meta.ByHost(ctx, global, lookupHost)
	if err != nil {
		return nil, ErrNotFound
	}

	// 2. key-value config
	cfg, err := meta.ConfigBySite(ctx, global, rec.ID)
	if err != nil {
		return nil, err
	}

	// 3. resolve password and build DSN
	key := sanitizeHost(host) // alias already handled inside
	pw, err := vcli.GetKV(
		ctx,
		fmt.Sprintf("secret/adept/tenant/%s/db", key),
		"password",
		10*time.Minute,
	)
	if err != nil {
		return nil, err
	}
	dsn := buildTenantDSN(key, pw)

	// 4. tenant DB pool (small, single-tenant)
	opts := database.Options{
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 30 * time.Minute,
		Retries:         2,
		RetryBackoff:    500 * time.Millisecond,
	}
	db, err := database.OpenProvider(ctx, func() string { return dsn }, opts)
	if err != nil {
		return nil, err
	}

	// 5. theme parsing
	enabledComponents := []string{"core"} // TODO: pull from ACL table
	mgr := theme.Manager{BaseDir: "themes"}
	th, err := mgr.Load(rec.Theme, enabledComponents)
	if err != nil {
		return nil, err
	}

	return &Tenant{
		Meta:     *rec,
		Config:   cfg,
		DB:       db,
		Theme:    th,
		Renderer: th.Renderer,
	}, nil
}
