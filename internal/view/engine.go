package view

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// Engine stores parsed templates keyed by theme name.
type Engine struct {
	themes map[string]*template.Template
	funcs  template.FuncMap
}

// New builds an Engine from the embedded theme filesystem.
//
// It expects the theme folders to live under   themes/{theme}/templates/*.html
// in the provided embed.FS.
func New(themeFS embed.FS, funcs template.FuncMap) (*Engine, error) {
	e := &Engine{themes: map[string]*template.Template{}, funcs: funcs}

	entries, err := themeFS.ReadDir(".") // look at top-level dirs inside the embed FS
	if err != nil {
		return nil, err
	}

	for _, dir := range entries {
		if !dir.IsDir() {
			continue
		}
		slug := dir.Name()                            // e.g. "minimal"
		pattern := filepath.Join(slug, "templates", "*.html")

		tpl, err := template.New(slug).Funcs(e.funcs).ParseFS(themeFS, pattern)
		if err != nil {
			return nil, err
		}
		e.themes[slug] = tpl
	}
	return e, nil
}

// ----------------------------------------------------------------------------
// Engine API
// ----------------------------------------------------------------------------

// Exec renders tplName (e.g. "demo/demo") directly to w.
func (e *Engine) Exec(w http.ResponseWriter, theme, tplName string, data any) error {
	tpl := e.lookup(theme)
	return tpl.ExecuteTemplate(w, tplName, data)
}

// ExecuteToHTML renders into a string and returns template.HTML (useful for widgets).
func (e *Engine) ExecuteToHTML(theme, tplName string, data any) (template.HTML, error) {
	tpl := e.lookup(theme)
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, tplName, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// AppendFS allows a module to register its own templates.
//
// Every *.html file found anywhere inside modFS becomes a template named
//   prefix/<base-name-without-ext>
// embedded into *every* theme's Template set.
func (e *Engine) AppendFS(prefix string, modFS embed.FS) error {
	return fs.WalkDir(modFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
			return err
		}
		base := strings.TrimSuffix(filepath.Base(path), ".html") // demo.html → demo
		name := filepath.Join(prefix, base)                      // demo/demo

		b, readErr := modFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		for _, tpl := range e.themes {
			if _, parseErr := tpl.New(name).Parse(string(b)); parseErr != nil {
				return parseErr
			}
		}
		return nil
	})
}

// lookup returns the Template set for theme or falls back to "minimal".
func (e *Engine) lookup(theme string) *template.Template {
	if tpl, ok := e.themes[theme]; ok {
		return tpl
	}
	return e.themes["minimal"]
}
