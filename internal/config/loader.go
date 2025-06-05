// internal/config/loader.go
//
// Configuration loader and hot-reloader.
//
// Context
// -------
// `Load()` builds a Config instance from three layers:
//
//  1. Optional `.env` file (jail-wide or repo root).
//  2. `conf/global.yaml`.
//  3. Environment variables prefixed `ADEPT_` that override YAML keys.
//
// The resulting struct is validated, enriched with the resolved
// ADEPT_ROOT path, then stored in an `atomic.Pointer` for lock-free
// reads.  `Reload()` simply calls `Load()` again and swaps the pointer.
//
// Notes
// -----
// • Vault support will hook into this file later.
// • The root-discovery logic matches the one-directory deployment model.
// • Oxford commas, two spaces after periods.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"

	// koanf core (v2) – give it an explicit alias so “koanf.New” resolves.
	"github.com/knadh/koanf/parsers/yaml"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	koanf "github.com/knadh/koanf/v2"
)

var current atomic.Pointer[Config]

// rootDir resolves ADEPT_ROOT or derives it from the executable’s parent.
func rootDir() string {
	if r := os.Getenv("ADEPT_ROOT"); r != "" {
		return r
	}
	exe, _ := os.Executable()
	if filepath.Base(filepath.Dir(exe)) == "bin" {
		return filepath.Dir(filepath.Dir(exe))
	}
	wd, _ := os.Getwd()
	return wd
}

// Load reads .env, YAML, and env overrides, validates, and caches Config.
func Load() (*Config, error) {
	root := rootDir()

	// .env (optional, jail-wide preferred)
	_ = godotenv.Load(filepath.Join(root, "conf", ".env"))

	k := koanf.New(".")
	if err := k.Load(file.Provider(filepath.Join(root, "conf", "global.yaml")), yaml.Parser()); err != nil {
		return nil, err
	}

	// Environment overrides: ADEPT_HTTP__LISTEN_ADDR → http.listen_addr
	if err := k.Load(env.Provider("ADEPT_", ".", func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, "__", "."))
	}), nil); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	cfg.Paths.Root = root

	if err := validateStruct(&cfg); err != nil {
		return nil, err
	}

	current.Store(&cfg)
	return &cfg, nil
}

// Get returns the last successfully loaded Config pointer.
func Get() *Config { return current.Load() }

// Reload re-reads config files and swaps the atomic pointer.
func Reload() error {
	_, err := Load()
	return err
}
