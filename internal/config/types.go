// internal/config/types.go
package config

import "time"

type Global struct {
	Database struct {
		Type            string
		DSN             string
		MaxOpenConns    int
		MaxIdleConns    int
		MaxConnLifetime time.Duration
	} `yaml:"database"`

	GeoIP struct {
		DBPath string `yaml:"db_path"`
	} `yaml:"geoip"`
}

type Site struct {
	Theme      string `yaml:"theme"` // "minimal", "modern", …
	UseDevHost bool   `yaml:"use_dev_host"`
	DevHost    string `yaml:"dev_host"`
	// add more per-site flags later (feature toggles, brand colors, …)
}
