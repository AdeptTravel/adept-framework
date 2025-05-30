package tenant

import (
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/head"
	"github.com/AdeptTravel/adept-framework/internal/ua" // 👈 new import
)

type Context struct {
	Request *http.Request
	Head    *head.Builder
	URL     URLInfo
	UA      ua.Info
	// Geo, User, Session will be added later
}

// NewContext builds the per-request context.
func NewContext(r *http.Request) *Context {
	return &Context{
		Request: r,
		Head:    head.New(),
		URL:     newURLInfo(r),
		UA:      ua.Parse(r.UserAgent()), // 👈 populate here
	}
}
