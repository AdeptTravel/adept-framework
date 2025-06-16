// internal/component/registry.go
//
// Component registry (cycle-free).
//
// Each concrete component lives under components/<name> and calls
// component.Register() in an init() function.  The tenant loader mounts
// every component’s Routes() at “/” and, after cold-load, invokes Init()
// when the component implements the Initializer interface.

package component

import (
	"sync"

	"github.com/go-chi/chi/v5"
)

// -----------------------------------------------------------------------------
// Contracts
// -----------------------------------------------------------------------------

// Initializer is optional.  If a Component implements it, the tenant loader
// calls Init(info) once after the tenant is loaded.
type Initializer interface {
	Init(TenantInfo) error
}

// Component contract.
//
// Migrations() may return nil if the component has no schema changes.
// Routes() should mount BOTH page and API endpoints, e.g:
//
//	r := chi.NewRouter()
//	r.Get("/login", getLogin)
//	r.Route("/api", func(api chi.Router) { … })
//	return r
type Component interface {
	Name() string
	Routes() chi.Router
	Migrations() []string
	Initializer // embed so Components may omit Init
}

// -----------------------------------------------------------------------------
// Registry implementation
// -----------------------------------------------------------------------------

var (
	mu       sync.RWMutex
	registry = map[string]Component{}
)

// Register is invoked from each component package’s init().
func Register(c Component) {
	mu.Lock()
	registry[c.Name()] = c
	mu.Unlock()
}

// All returns every registered component in arbitrary order.
func All() []Component {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Component, 0, len(registry))
	for _, c := range registry {
		out = append(out, c)
	}
	return out
}

// AllNames returns a set of every registered component name.
// Used by tenant.Router() as a fallback when component_acl is empty.
func AllNames() map[string]struct{} {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]struct{}, len(registry))
	for name := range registry {
		out[name] = struct{}{}
	}
	return out
}
