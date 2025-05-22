package tenant

import (
	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/jmoiron/sqlx"
)

// entry pairs a runtime Tenant with its last-seen timestamp.  The timestamp
// is stored as UnixNano and updated atomically.
type entry struct {
	tenant   *Tenant
	lastSeen int64
}

// Tenant is what the HTTP layer uses after cache lookup.
type Tenant struct {
	Meta site.Record
	DB   *sqlx.DB
}

// Close shuts down the tenant’s connection pool.
func (t *Tenant) Close() error {
	return t.DB.Close()
}
