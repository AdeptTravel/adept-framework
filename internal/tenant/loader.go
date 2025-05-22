package tenant

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/site"
)

// loadSite queries one active site row and opens its DSN with conservative
// pool sizes.  It returns (*Tenant, nil) on success or (nil, error).
func loadSite(ctx context.Context, global *sqlx.DB, host string) (*Tenant, error) {
	rec, err := site.ByHost(ctx, global, host)
	if err != nil {
		return nil, ErrNotFound
	}

	db, err := database.OpenWithOptions(rec.DSN, 5, 2) // 5 open, 2 idle
	if err != nil {
		return nil, err
	}

	return &Tenant{Meta: *rec, DB: db}, nil
}
