// internal/theme/manager.go
//
// Removed the viewhelpers import so that theme → viewhelpers → tenant
// cycle disappears.  Template helper registration now happens in main.go.
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

// Load parses templates and returns a ready-to-use Theme.
// NOTE: UA helpers are *not* registered here to avoid an import cycle.
//
//	main.go adds them after loading the theme.
func (m *Manager) Load(name string, modules []string) (*Theme, error) {
	root := filepath.Join(m.BaseDir, name)
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("theme %s not found at %s", name, root)
	}

	dummyAsset := func(s string) string { return s } // placeholder
	tpl := template.New("").Funcs(FuncMap(dummyAsset))

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
	tpl.Funcs(FuncMap(th.AssetFunc)) // only asset helper; UA helpers added later
	return th, nil
}
