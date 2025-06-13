// internal/widget/registry.go
//
// Widget registry and lookup helpers.
//
// A **Widget** is a reusable view fragment rendered inside a page or a
// widget *area*.  Each concrete widget lives under its Component folder
// (`components/<comp>/widgets/<name>.go`) and registers itself by calling
// `widget.Register(&MyWidget{})` in an init() func.
//
// The key used for registration is `<component>/<widget>` – e.g.
// "auth/login" – and must be returned by the widget’s `ID` method.
//
// Template authors can embed a widget with:
//
//	{{ widget "auth/login" (dict "limit" 5) }}
//
// Params are optional.  The helper looks up the widget, invokes
// `Render`, and returns `template.HTML`.
package widget

import (
	"sync"
)

// Widget represents a view fragment that can be embedded inside any page
// template.  Render returns the generated HTML and a cache policy hint.
// Params are an arbitrary key‑value map passed from the template.
//
// Implementations should treat missing params defensively (nil map).
// Template resolution MUST respect the override chain (site → theme →
// component) by delegating to internal/view.
//
// CachePolicy mirrors the enum defined in internal/view.
// A widget may decide at runtime – e.g., CacheSkip when a CSRF token is
// injected.
//
// Errors should be returned, not written to http.ResponseWriter, so the
// calling helper can decide how to surface the failure.
//
// Render MUST be concurrency‑safe; multiple goroutines may call it.
type Widget interface {
	ID() string
	Render(rctx any, params map[string]any) (html string, policy int, err error)
}

var (
	mu       sync.RWMutex
	registry = map[string]Widget{}
)

// Register a widget during init().  If a duplicate key is registered the
// latter entry overwrites the former; duplicates are logged by the caller.
func Register(w Widget) {
	mu.Lock()
	registry[w.ID()] = w
	mu.Unlock()
}

// Lookup returns the widget or nil.
func Lookup(key string) Widget {
	mu.RLock()
	defer mu.RUnlock()
	return registry[key]
}

// All returns a copy of the registry slice – useful for tests or auto‑documentation.
func All() []Widget {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Widget, 0, len(registry))
	for _, w := range registry {
		out = append(out, w)
	}
	return out
}
