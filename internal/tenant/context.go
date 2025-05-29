// context.go defines the per-request Context passed into templates and
// handlers.  It owns the head.Builder so modules can push tags into the
// eventual <head> section.
package tenant

import (
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/head"
)

// Context is created once per request.
type Context struct {
	Request *http.Request
	Head    *head.Builder
	// Info, User, Session fields will be added later.
}

// NewContext initialises a Context with an empty head builder.
func NewContext(r *http.Request) *Context {
	return &Context{
		Request: r,
		Head:    head.New(),
	}
}
