// Package theme holds the data structures that describe one visual theme.
// A Theme combines:
//
//   - Name         – the theme directory name (for example, “base”).
//   - Root         – absolute path to that directory on disk.
//   - Renderer     – parsed templates ready for execution.
//   - AssetFunc    – helper injected into templates so they can resolve
//     `{{ asset \"css/main.css\" }}` to a URL.
//
// The first milestone is to load templates and provide an AssetFunc that
// merely prefixes `/themes/<name>/assets/`.  Minification and fingerprinting
// will be added later by the asset manager.
package theme

import (
	"html/template"
	"path/filepath"
)

// Theme is returned by the Manager once all templates are parsed.
type Theme struct {
	Name      string
	Root      string
	Renderer  *template.Template
	AssetFunc func(string) string
}

// New constructs a Theme with an AssetFunc that points to the assets folder.
func New(name, root string, tpl *template.Template) *Theme {
	assetPrefix := filepath.ToSlash("/themes/" + name + "/assets/")
	return &Theme{
		Name:     name,
		Root:     root,
		Renderer: tpl,
		AssetFunc: func(p string) string {
			return assetPrefix + p
		},
	}
}
