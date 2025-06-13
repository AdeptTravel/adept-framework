// internal/view/uahelpers.go
//
// User‑Agent‑related template helpers.  Moved from the former
// internal/view package so all view concerns now live under one
// directory.
package view

import (
	"html/template"

	"github.com/yanizio/adept/internal/tenant"
)

// uaFuncMap returns helpers keyed off *tenant.Context.
func uaFuncMap() template.FuncMap {
	return template.FuncMap{
		"browser":        func(c *tenant.Context) string { return c.UA.Browser },
		"browserVersion": func(c *tenant.Context) string { return c.UA.Version },
		"os":             func(c *tenant.Context) string { return c.UA.OS },
		"osVersion":      func(c *tenant.Context) string { return c.UA.OSVersion },
		"device":         func(c *tenant.Context) string { return c.UA.Device },
		"platform":       func(c *tenant.Context) string { return c.UA.Platform },
		"isBot":          func(c *tenant.Context) bool { return c.UA.IsBot },
	}
}
