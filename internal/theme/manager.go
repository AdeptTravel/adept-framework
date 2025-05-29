package theme

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// Manager discovers and loads themes.
type Manager struct {
	BaseDir string // e.g., "themes" (relative) or "/themes" (absolute)
}

// Load parses templates for the given theme name plus all enabled modules.
// Template precedence (high → low):
//  1. /themes/<name>/templates/<module>/... (overrides)
//  2. modules/<module>/widget/**/templates/... (widget defaults)
//  3. modules/<module>/templates/...          (module defaults)
func (m *Manager) Load(name string, modules []string) (*Theme, error) {
	root := filepath.Join(m.BaseDir, name)
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("theme %s not found at %s", name, root)
	}

	// Build base template with a dummy asset helper so early parsing succeeds.
	dummyAsset := func(s string) string { return s }
	tpl := template.New("").Funcs(FuncMap(dummyAsset))

	// 1. Parse module defaults first (lowest precedence).
	for _, mod := range modules {
		//   a) widget templates
		widgetDir := filepath.Join("modules", mod, "widget")
		if files, _ := CollectHTML(widgetDir); len(files) > 0 {
			if _, err := tpl.ParseFiles(files...); err != nil {
				return nil, fmt.Errorf("parse %s widgets: %w", mod, err)
			}
		}

		//   b) module root templates
		modDir := filepath.Join("modules", mod, "templates")
		if files, _ := CollectHTML(modDir); len(files) > 0 {
			if _, err := tpl.ParseFiles(files...); err != nil {
				return nil, fmt.Errorf("parse %s templates: %w", mod, err)
			}
		}
	}

	// 2. Parse theme overrides (highest precedence).
	themeDir := filepath.Join(root, "templates")
	if files, _ := CollectHTML(themeDir); len(files) > 0 {
		if _, err := tpl.ParseFiles(files...); err != nil {
			return nil, fmt.Errorf("parse theme overrides: %w", err)
		}
	}

	// Finalise Theme struct and set real asset helper.
	th := New(name, root, tpl)
	tpl.Funcs(FuncMap(th.AssetFunc)) // replace dummy with real prefix

	return th, nil
}
