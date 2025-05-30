// internal/viewhelpers/helpers.go
//
// Template helpers that pull data out of *tenant.Context.  Imported by the
// theme manager *after* templates are parsed, so every template can call:
//
//	{{ browser .Ctx }} {{ browserVersion .Ctx }}
//	{{ os .Ctx }} – {{ osVersion .Ctx }}
//	{{ device .Ctx }}  {{ platform .Ctx }}
//	{{ if isBot .Ctx }}Robot!{{ end }}
package viewhelpers

import (
	"html/template"

	"github.com/AdeptTravel/adept-framework/internal/tenant"
)

// FuncMap returns UA helpers keyed off *tenant.Context.
func FuncMap() template.FuncMap {
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
