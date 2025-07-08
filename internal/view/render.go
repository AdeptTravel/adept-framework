// internal/view/render.go
//
// Central view engine: template lookup, override chain, func-map injection,
// and an LRU of parsed *template.Template* sets.
//
// Public helpers
// --------------
//   - Render         – write rendered HTML to an http.ResponseWriter.
//   - RenderToString – return template.HTML (widgets, e-mails).
//
// Lookup precedence (first hit wins):
//   1. sites/<host>/components/<comp>/templates/<tpl>.html
//   2. themes/<theme>/components/<comp>/templates/<tpl>.html
//   3. components/<comp>/templates/<tpl>.html
//
// All templates in the same directory are parsed as one set so sub-templates
// ({{ template "row" . }}) work out-of-the-box.
//
// New in July 2025
// ----------------
//   • execName() chooses the best template to execute:
//       – If the set contains "<name>.html", we run that (file has no define).
//       – Else we fall back to "<name>" (root template defined via {{ define }}).
//   • Callers now pass the logical name (e.g. "login"); view.Render figures
//     out the concrete template automatically.
//
// Style
// -----
// • Oxford commas, two spaces after periods.

package view

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yanizio/adept/internal/cache"
	"github.com/yanizio/adept/internal/tenant"
	"github.com/yanizio/adept/internal/widget"
)

//
// cache definitions
//

// CachePolicy hints how the caller wants this template cached.
type CachePolicy int

const (
	CacheDefault CachePolicy = iota // obey global TTL
	CacheSkip                       // never cache
	CacheForce                      // always cache (long TTL, reserved)
)

// Parsed template sets per tenant; tweak capacity when perf-testing.
var tmplLRU = cache.New(1024)
var once sync.Once

//
// public helpers
//

// Render executes the template set and streams it to w.
//
// We first load (or parse) the appropriate template set, then execute the
// concrete template determined by execName().  This allows both:
//
//   - A simple file "login.html" with no {{ define }} block.  In that case
//     execName runs "login.html" automatically.
//   - A file that wraps markup in {{ define "login" }} … {{ end }} and relies
//     on that root template name.
//
// Either style works; developers can choose per component.
func Render(ctx *tenant.Context, w http.ResponseWriter, comp, name string, data any, policy CachePolicy) error {
	t, err := load(ctx, comp, name, policy)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, execName(t, name), data)
}

// RenderToString executes and returns HTML (used by widgets and e-mail
// generators).  It mirrors Render, but writes to a buffer instead of w.
func RenderToString(ctx *tenant.Context, comp, name string, data any) (template.HTML, CachePolicy, error) {
	t, err := load(ctx, comp, name, CacheDefault)
	if err != nil {
		return "", CacheSkip, err
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, execName(t, name), data); err != nil {
		return "", CacheSkip, err
	}
	return template.HTML(buf.String()), CacheDefault, nil
}

//
// internal: load
//

// load finds and (if necessary) parses the template set for the given tenant,
// component, and base name, obeying the provided cache policy.
func load(ctx *tenant.Context, comp, name string, policy CachePolicy) (*template.Template, error) {
	theme := "default" // TODO: derive from tenant once theme support lands
	key := strings.Join([]string{ctx.Request.Host, theme, comp, name}, "::")

	if policy != CacheSkip {
		if v, ok := tmplLRU.Get(key); ok {
			return v.(*template.Template), nil
		}
	}

	paths := []string{
		filepath.Join("sites", ctx.Request.Host, "components", comp, "templates", name+".html"),
		filepath.Join("themes", theme, "components", comp, "templates", name+".html"),
		filepath.Join("components", comp, "templates", name+".html"),
	}

	var base string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			base = p
			break
		}
	}
	if base == "" {
		return nil, os.ErrNotExist
	}

	// Parse all *.html in the same directory so sub-templates work.
	dir := filepath.Dir(base)
	pattern := filepath.Join(dir, "*.html")

	t, err := template.New(name).Funcs(buildFuncMap(ctx)).ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	if policy != CacheSkip {
		tmplLRU.Add(key, t)
	}
	return t, nil
}

//
// func-map builders
//

func buildFuncMap(rctx *tenant.Context) template.FuncMap {
	fm := template.FuncMap{
		"dict":   dict,
		"widget": widgetFunc(rctx),
		"area":   areaFunc(rctx),
	}
	for k, v := range uaFuncMap() { // UA helpers (browser/os parsing)
		fm[k] = v
	}
	return fm
}

//
// helpers
//

// execName picks the template name to execute.
//
// Priority:
//  1. If the set has "<name>.html" (file-based template), run that.
//  2. Otherwise, fall back to "<name>" (root template defined in code).
func execName(t *template.Template, name string) string {
	if tmpl := t.Lookup(name + ".html"); tmpl != nil {
		return name + ".html"
	}
	return name
}

// dict builds a map in templates: {{ dict "k" 1 "k2" "v" }}.
func dict(kv ...any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, _ := kv[i].(string)
		m[key] = kv[i+1]
	}
	return m
}

// widgetFunc renders a registered widget and returns safe HTML.  Errors are
// hidden behind <!-- comments --> so end-users never see stack traces.
func widgetFunc(rctx *tenant.Context) func(string, map[string]any) template.HTML {
	return func(key string, params map[string]any) template.HTML {
		w := widget.Lookup(key)
		if w == nil {
			return template.HTML("<!-- widget not found -->")
		}
		html, _, err := w.Render(rctx, params)
		if err != nil {
			return template.HTML("<!-- widget error -->")
		}
		return template.HTML(html)
	}
}

// areaFunc is a stub until the widget-area feature lands.
func areaFunc(_ *tenant.Context) func(string) template.HTML {
	return func(string) template.HTML { return "" }
}
