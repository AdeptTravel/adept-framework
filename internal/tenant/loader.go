// loader.go contains the single responsibility of translating a host FQDN
// into a fully initialised Tenant object.  The function:
//
//	loadSite(ctx, globalDB, host) (*Tenant, error)
//
// performs three steps:
//
//  1. Query the global `site` table for an active row whose host matches.
//  2. Open a small, capped connection pool to the tenant’s own database,
//     honouring context deadlines and retrying transient dial errors.
//  3. Wrap the metadata and pool into a Tenant struct that callers can
//     store in the runtime cache.
//
// A Tenant opened here is ready for immediate use by HTTP handlers.
package tenant

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/site"
)

// loadSite is invoked by cache.Get on a cache miss.  It returns ErrNotFound
// when the host has no active row, propagates any connection errors, and
// otherwise returns a fully initialised Tenant.
func loadSite(ctx context.Context, global *sqlx.DB, host string) (*Tenant, error) {
	//
	// 1.  Fetch metadata.
	//
	rec, err := site.ByHost(ctx, global, host)
	if err != nil {
		return nil, ErrNotFound
	}

	//
	// 2.  Open capped connection pool with retry.
	//
	opts := database.Options{
		MaxOpenConns:    5,                      // per-tenant upper bound
		MaxIdleConns:    2,                      // keep-alive idles
		ConnMaxLifetime: 30 * time.Minute,       // hygiene
		Retries:         2,                      // dial retries
		RetryBackoff:    500 * time.Millisecond, // wait between retries
	}

	db, err := database.OpenWithOptions(ctx, rec.DSN, opts)
	if err != nil {
		return nil, err
	}

	//
	// 3.  Build Tenant wrapper and return.
	//
	return &Tenant{
		Meta: *rec,
		DB:   db,
	}, nil
}
