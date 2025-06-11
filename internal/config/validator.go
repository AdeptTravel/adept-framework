// internal/config/validator.go
//
// Thin wrapper around go-playground/validator.
//
// Context
// -------
// `internal/config/loader.go` calls `validateStruct` immediately after it
// unmarshals the merged Koanf tree into a `Config` instance.  Any tag
// mismatch or validation error aborts startup, ensuring the binary never
// runs with partial, malformed, or missing configuration.
//
// The only built-in rule we rely on right now is `required`, attached to
// fields such as `Database.GlobalDSN` and the newly-added
// `Database.GlobalPassword`.  Additional custom rules—e.g., “dsn must
// contain exactly one %s verb” or tenant-name pattern checks—can be
// registered here as the configuration surface grows.
//
// Notes
// -----
//   • Oxford commas, two spaces after periods.
//   • Section dividers use the simple comment style requested.

package config

import "github.com/go-playground/validator/v10"

//
// validator instance (package-level singleton)
//

var v = validator.New()

//
// public API
//

// validateStruct returns the first validation error, or nil on success.
func validateStruct(c *Config) error {
	return v.Struct(c)
}
