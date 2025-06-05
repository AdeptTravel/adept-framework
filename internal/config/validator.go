// internal/config/validator.go
//
// Thin wrapper around go-playground/validator.
//
// Context
// -------
// The loader calls `validateStruct` after unmarshalling YAML + env into a
// Config instance.  Any tag mismatch aborts startup, ensuring the binary
// never runs with partial or unknown configuration.
//
// Notes
// -----
// • Additional custom validation rules can be registered here.
package config

import "github.com/go-playground/validator/v10"

var v = validator.New()

func validateStruct(c *Config) error { return v.Struct(c) }
