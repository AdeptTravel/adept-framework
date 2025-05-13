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

// Engine parses and stores templates for each theme.
type Engine struct {
	themes map[string]*template.Template // theme slug → template set
	funcs  template.FuncMap
}

// New builds an Engine from an embed.FS that contains:
//
//   <theme>/views/*.html
//
// Each first-level directory in the FS is treated as a separate theme
// (e.g., minimal/views/*.html, modern/views/*.html, …).
func New(themeFS embed.FS, funcs template.FuncMap) (*Engine, error) {
	e := &Engine{
		themes: map[string]*template.Template{},
		funcs:  funcs,
	}

	// discover theme directories at FS root
	rootEntries, err := themeFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, dir := range rootEntries {
		if !dir.IsDir() {
			continue
		}
		slug := dir.Name()                             // "minimal"
		pattern := filepath.Join(slug, "views", "*.html")

		tpl, err := template.New(slug).Funcs(funcs).ParseFS(themeFS, pattern)
		if err != nil {
			return nil, err
		}
		e.themes[slug] = tpl
	}
	return e, nil
}

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

// Exec renders tplName (e.g., "demo/demo") directly to w using the chosen theme.
func (e *Engine) Exec(w http.ResponseWriter, theme, tplName string, data any) error {
	tpl := e.lookup(theme)
	return tpl.ExecuteTemplate(w, tplName, data)
}

// ExecuteToHTML renders tplName and returns it as template.HTML (for widgets).
func (e *Engine) ExecuteToHTML(theme, tplName string, data any) (template.HTML, error) {
	tpl := e.lookup(theme)
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, tplName, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// AppendFS allows a module to register its own templates.  Every *.html file
// found anywhere inside modFS is loaded under the name
//   prefix/<filename-without-ext>
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

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func (e *Engine) lookup(theme string) *template.Template {
	if tpl, ok := e.themes[theme]; ok {
		return tpl
	}
	return e.themes["minimal"] // fallback
}
