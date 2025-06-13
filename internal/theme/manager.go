// internal/theme/manager.go
//
// Theme.Manager parses template trees.  To keep the build graph acyclic,
// we register only stub helpers here; the real helpers are injected by
// the view engine just before ExecuteTemplate.
package theme

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

type Manager struct {
	BaseDir string // e.g., "themes"
}

// Load parses the theme’s templates and returns a ready-to-use Theme.
// Stub helpers (dict, widget, area) are defined so Parse succeeds;
// real helpers overwrite them at render time.
func (m *Manager) Load(name string, modules []string) (*Theme, error) {
	root := filepath.Join(m.BaseDir, name)
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("theme %s not found at %s", name, root)
	}

	dummyAsset := func(s string) string { return s }
	stubs := template.FuncMap{
		"dict":   func(...any) map[string]any { return map[string]any{} },
		"widget": func(string, ...any) string { return "" },
		"area":   func(string) string { return "" },
	}

	// Base template with asset helper and stub funcs.
	tpl := template.New("").Funcs(FuncMap(dummyAsset)).Funcs(stubs)

	// ------------------------------------------------------------------
	// Parse module defaults (lowest precedence).
	// ------------------------------------------------------------------
	for _, mod := range modules {
		if files, _ := CollectHTML(filepath.Join("modules", mod, "widget")); len(files) > 0 {
			if _, err := tpl.ParseFiles(files...); err != nil {
				return nil, fmt.Errorf("parse %s widgets: %w", mod, err)
			}
		}
		if files, _ := CollectHTML(filepath.Join("modules", mod, "templates")); len(files) > 0 {
			if _, err := tpl.ParseFiles(files...); err != nil {
				return nil, fmt.Errorf("parse %s templates: %w", mod, err)
			}
		}
	}

	// ------------------------------------------------------------------
	// Parse theme overrides (highest precedence).
	// ------------------------------------------------------------------
	if files, _ := CollectHTML(filepath.Join(root, "templates")); len(files) > 0 {
		if _, err := tpl.ParseFiles(files...); err != nil {
			return nil, fmt.Errorf("parse theme overrides: %w", err)
		}
	}

	// Finalise theme and swap in real asset helper.
	th := New(name, root, tpl)
	tpl.Funcs(FuncMap(th.AssetFunc)) // only asset helper; real view helpers added later
	return th, nil
}
