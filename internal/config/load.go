// internal/config/load.go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadGlobal(path string) (Global, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Global{}, err
	}
	var g Global
	return g, yaml.Unmarshal(b, &g)
}

func LoadSite(sitesRoot, host string) (Site, error) {
	path := filepath.Join(sitesRoot, host, "config.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return Site{}, err
	}
	var s Site
	return s, yaml.Unmarshal(b, &s)
}
