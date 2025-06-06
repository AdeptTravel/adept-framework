// internal/config/model.go
//
// Typed configuration model.
/*
Context
--------
These structs hold the global configuration loaded from:

  • optional `.env`
  • `conf/global.yaml`
  • ADEPT_ environment overrides (and, later, Vault secrets)

`loader.go` unmarshals the merged tree into this model, validates it with
go-playground/validator, and caches the pointer atomically for
lock-free reads.

Notes
-----
  • Struct tags use `koanf:"…"`, not `yaml:"…"`, because Koanf ignores
    `yaml` by default.
  • The unexported `Paths` section is filled at load time; callers treat
    it as read-only.
  • Oxford commas, two spaces after periods.  No em dash.
*/
package config

/*──────────────────────────── section structs ──────────────────────────────*/

// HTTP holds web-server tunables.
type HTTP struct {
	ListenAddr string `koanf:"listen_addr" validate:"required,hostname_port"`
	ForceHTTPS bool   `koanf:"force_https"`
}

// Database holds DSNs and pool sizes.
type Database struct {
	GlobalDSN string `koanf:"global_dsn" validate:"required"`
}

// Paths is resolved at runtime—never set in YAML or env.
type Paths struct {
	Root string // ADEPT_ROOT or derived parent of /bin
}

/*──────────────────────────── root aggregate ───────────────────────────────*/

// Config is the aggregate returned by Load() and cached via atomic.Pointer.
type Config struct {
	HTTP     HTTP     `koanf:"http"`
	Database Database `koanf:"database"`
	Paths    Paths    `koanf:"-"`
}
