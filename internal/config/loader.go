// internal/config/loader.go
//
// Configuration loader and hot-reloader.
//
/*
Context
--------
`Load()` builds one immutable `Config` struct from three layers (highest
precedence last):

  1. Optional `.env` file — first `<root>/conf/.env`, then jail-wide fallback.
  2. `conf/global.yaml`.
  3. Environment variables prefixed `ADEPT_`, where `__` maps to “.”
     (e.g., `ADEPT_HTTP__LISTEN_ADDR → http.listen_addr`).

After merging, the tree is unmarshalled into strongly-typed structs,
validated, enriched with the runtime root path, and cached in an
`atomic.Pointer` for lock-free reads.  `Reload()` simply calls `Load()`
again and swaps the pointer.

Instrumentation
---------------
  • DEBUG spans — root discovery, YAML read, env overlay.
  • ERROR spans — YAML parse, env overlay, unmarshal, validation failures.
  • INFO  span  — final “config loaded” with key highlights.
  • Logs use the global *sugared* logger (`zap.S()`) so early boot issues
    surface even before the file logger is installed (bootstrap console).

Notes
-----
  • `rootDir()` now climbs the cwd tree until it finds `conf/global.yaml`;
    this lets `go run ./cmd/web` work from any sub-directory.
  • Vault integration will hook into this file later.
  • Oxford commas, two spaces after periods.
*/
package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	koanf "github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

var current atomic.Pointer[Config]

/*──────────────────────────── root discovery ───────────────────────────────*/

// rootDir resolves ADEPT_ROOT or climbs directories until conf/global.yaml
// is found.  Falls back to executable heuristic for production layout.
func rootDir() string {
	if r := os.Getenv("ADEPT_ROOT"); r != "" {
		return r
	}

	wd, _ := os.Getwd()
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "conf", "global.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			break
		}
		dir = parent
	}

	exe, _ := os.Executable()
	if filepath.Base(filepath.Dir(exe)) == "bin" {
		return filepath.Dir(filepath.Dir(exe))
	}
	return wd
}

/*─────────────────────────────── loader ───────────────────────────────────*/

// Load reads .env, YAML, env overrides, validates, and caches Config.
func Load() (*Config, error) {
	root := rootDir()
	zap.S().Debugw("config root resolved", "root", root)

	// .env (optional, no error if missing)
	_ = godotenv.Load(filepath.Join(root, "conf", ".env"))

	k := koanf.New(".")

	yamlPath := filepath.Join(root, "conf", "global.yaml")
	if err := k.Load(file.Provider(yamlPath), yaml.Parser()); err != nil {
		zap.S().Errorw("config yaml load failed", "file", yamlPath, "err", err)
		return nil, err
	}
	zap.S().Debugw("config yaml loaded", "file", yamlPath)

	// Env overrides: ADEPT_HTTP__LISTEN_ADDR → http.listen_addr
	if err := k.Load(env.Provider("ADEPT_", ".", func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, "__", "."))
	}), nil); err != nil {
		zap.S().Errorw("config env overlay failed", "err", err)
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		zap.S().Errorw("config unmarshal failed", "err", err)
		return nil, err
	}

	cfg.Paths.Root = root
	if err := validateStruct(&cfg); err != nil {
		zap.S().Errorw("config validation failed", "err", err)
		return nil, err
	}

	current.Store(&cfg)
	zap.S().Infow("config loaded",
		"listen_addr", cfg.HTTP.ListenAddr,
		"force_https", cfg.HTTP.ForceHTTPS,
		"root", cfg.Paths.Root,
	)
	return &cfg, nil
}

/*──────────────────────────── helpers ─────────────────────────────────────*/

func Get() *Config  { return current.Load() }
func Reload() error { _, err := Load(); return err }
