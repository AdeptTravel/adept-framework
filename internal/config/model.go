// internal/config/model.go
//
// Typed configuration model for Adept.
//
// Context
// -------
// These structs define the shape of the configuration tree that
// `internal/config/loader.go` builds from three overlay layers:
//
//   • optional `.env`                         – dotenv values,
//   • `conf/global.yaml`                      – primary static file,
//   • `ADEPT_`-prefixed environment overrides – highest precedence.
//
// Any value whose string begins with the prefix `vault:` is resolved
// through the Vault client *before* unmarshalling, so the model never
// stores Vault URIs—only plain strings.
//
// Validation happens immediately after unmarshal; the app fails fast if
// required fields are missing.
//
// Notes
// -----
//   • Struct tags use `koanf:"…"`, not `yaml:"…"`—Koanf ignores `yaml` tags
//     unless configured otherwise.
//   • The `Paths` block is filled at runtime; YAML must not try to set it.
//   • Oxford commas, two spaces after periods.  No em-dash.

package config

//
// HTTP section
//

// HTTP holds web-server tunables.
type HTTP struct {
	ListenAddr string `koanf:"listen_addr" validate:"required,hostname_port"`
	ForceHTTPS bool   `koanf:"force_https"`
}

//
// Database section
//

// Database holds DSN templates and secrets.
//
// The *template* (`GlobalDSN`) is kept in YAML so operators can tweak
// host, port, or flags without touching Vault.  The *secret* portion
// (`GlobalPassword`) is stored in Vault and injected at runtime, keeping
// credentials out of flat files and git history.
type Database struct {
	GlobalDSN      string `koanf:"global_dsn"      validate:"required"`
	GlobalPassword string `koanf:"global_password" validate:"required"`
}

//
// Paths section (runtime only)
//

// Paths is resolved at runtime—never set in YAML or env.  The loader
// discovers `Root` (repo root or ADEPT_ROOT override) so later code can
// build absolute file paths.
type Paths struct {
	Root string // ADEPT_ROOT or discovered parent
}

//
// Root aggregate
//

// Config is the immutable aggregate returned by Load() and cached in an
// atomic.Pointer for lock-free reads throughout the app lifetime.
type Config struct {
	HTTP     HTTP     `koanf:"http"`
	Database Database `koanf:"database"`
	Paths    Paths    `koanf:"-"` // not loaded from config files
}
