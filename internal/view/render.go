// internal/view/render.go
//
// Central view engine: template lookup, override chain, func-map
// injection, and an LRU of parsed *template.Template* sets.
//
// Public helpers
// --------------
//   - Render         – write rendered HTML to an http.ResponseWriter.
//   - RenderToString – return template.HTML (widgets, e-mails).
//
// Lookup precedence (first hit wins):
//  1. sites/<host>/components/<comp>/templates/<tpl>.html
//  2. themes/<theme>/components/<comp>/templates/<tpl>.html
//  3. components/<comp>/templates/<tpl>.html
//
// All templates in the same directory are parsed as one set so sub-templates
// ({{ template "row" . }}) work out-of-the-box.
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

// Parsed template sets per tenant; tweak when perf-testing.
var tmplLRU = cache.New(1024)
var once sync.Once

//
// public helpers
//

// Render executes the template and streams it to w.
func Render(ctx *tenant.Context, w http.ResponseWriter, comp, name string, data any, policy CachePolicy) error {
	t, err := load(ctx, comp, name, policy)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, data)
}

// RenderToString executes and returns HTML (widgets, e-mails).
func RenderToString(ctx *tenant.Context, comp, name string, data any) (template.HTML, CachePolicy, error) {
	t, err := load(ctx, comp, name, CacheDefault)
	if err != nil {
		return "", CacheSkip, err
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return "", CacheSkip, err
	}
	return template.HTML(buf.String()), CacheDefault, nil
}

//
// internal: load
//

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

	// Parse all *.html in the same directory for sub-templates.
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

func buildFuncMap(rctx *tenant.Context) template.FuncMap {
	fm := template.FuncMap{
		"dict":   dict,
		"widget": widgetFunc(rctx),
		"area":   areaFunc(rctx),
	}
	for k, v := range uaFuncMap() { // UA helpers
		fm[k] = v
	}
	return fm
}

//
// helpers
//

// dict builds a map in templates: {{ dict "k" 1 "k2" "v" }}
func dict(kv ...any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, _ := kv[i].(string)
		m[key] = kv[i+1]
	}
	return m
}

// widgetFunc renders a registered widget and returns safe HTML.
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
