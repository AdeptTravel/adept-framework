// internal/component/tenantinfo.go
//
// Exposes per-tenant resources to Components during Init() without
// importing the concrete tenant package—avoids an import cycle.

package tenant

import (
	"github.com/jmoiron/sqlx"
	"github.com/yanizio/adept/internal/theme"
	"github.com/yanizio/adept/internal/vault"
)

// TenantInfo provides read-only access to runtime assets a Component
// may need at Init time.  The concrete *tenant.Tenant satisfies it.
type TenantInfo interface {
	GetDB() *sqlx.DB
	GetConfig() map[string]string
	GetTheme() *theme.Theme
	GetVault() *vault.Client
}
