// internal/config/model.go
//
// Typed configuration model.
//
// Context
// -------
// This file declares the strongly-typed structs that hold global
// configuration loaded from `.env`, `conf/global.yaml`, and (in the
// future) HashiCorp Vault.  Callers import `internal/config` and access
// fields directly instead of poking into generic maps.
//
// Validation tags are applied in `loader.go` using go-playground/validator.
//
// Notes
// -----
//   - Unexported `Paths` section is filled at load time; callers treat it as
//     read-only.
//   - Oxford commas, two spaces after periods.
package config

// HTTP holds web-server tunables.
type HTTP struct {
	ListenAddr string `yaml:"listen_addr" validate:"required,hostname_port"`
	ForceHTTPS bool   `yaml:"force_https"`
}

// Database holds DSNs and pool sizes.
type Database struct {
	GlobalDSN string `yaml:"global_dsn" validate:"required"`
}

// Paths is resolved at runtime—never set in YAML.
type Paths struct {
	Root string // ADEPT_ROOT or derived parent of /bin
}

// Config is the aggregate returned by Load() and cached via atomic.Pointer.
type Config struct {
	HTTP     HTTP     `yaml:"http"`
	Database Database `yaml:"database"`
	Paths    Paths    `yaml:"-"`
}
