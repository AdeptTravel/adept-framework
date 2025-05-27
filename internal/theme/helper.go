//
//  internal/theme/helpers.go
//
//  Theme functions that expose RequestInfo fields with short,
//  ergonomic names.  These helpers prevent HTML authors from poking
//  through nested structs repeatedly.
//

package theme

import (
	"html/template"
	"net/url"

	"github.com/adepttravel/adept-framework/internal/tenant"
)

// FuncMap returns the global template function map.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		// Geo helpers
		"clientIP": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.Geo.IP.String()
		},
		"country": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.Geo.CountryISO
		},
		"city": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.Geo.City
		},

		// UA helpers
		"browser": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.UA.Browser
		},
		"browserVersion": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.UA.Version
		},
		"os": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.UA.OS
		},
		"osVersion": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.UA.OSVersion
		},
		"device": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.UA.Device
		},
		"platform": func(c *tenant.Context) string {
			if c.Info == nil {
				return ""
			}
			return c.Info.UA.Platform
		},
		"isBot": func(c *tenant.Context) bool {
			return c.Info != nil && c.Info.UA.IsBot
		},

		// URL helper
		"url": func(c *tenant.Context) *url.URL {
			if c.Info == nil {
				return nil
			}
			return c.Info.URL
		},
	}
}
