// internal/module/registry.go
//
// A super-light registry: modules call Register(path, handler) in an init()
// function.  The core router looks up the exact URL path (no wildcards yet)
// and, if found, executes the handler.
//
// Handler signature:
//
//	func(ten *tenant.Tenant, ctx *tenant.Context,
//	     w http.ResponseWriter, r *http.Request)
//
// This gives handlers access to both the Tenant (site-wide data) and the
// per-request Context (URLInfo, Head builder, etc.).
package module

import (
	"net/http"
	"sync"

	"github.com/yanizio/adept/internal/tenant"
)

// Handler is what modules register.
type Handler func(*tenant.Tenant, *tenant.Context, http.ResponseWriter, *http.Request)

var (
	mu       sync.RWMutex
	registry = map[string]Handler{}
)

// Register is called from module init() functions.
func Register(path string, h Handler) {
	mu.Lock()
	registry[path] = h
	mu.Unlock()
}

// Lookup returns the handler for an exact path or nil.
func Lookup(path string) Handler {
	mu.RLock()
	defer mu.RUnlock()
	return registry[path]
}
