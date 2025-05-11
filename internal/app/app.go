package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/adepttraveladept-framework/internal/bus"
	"github.com/adepttraveladept-framework/internal/outbox"
)

type Context struct {
	Bus       *bus.InMem
	Outbox    outbox.Store
	Analytics website.AnalyticsWriter // alias to keep deps simple
}

func RegisterMiddlewares(r chi.Router) {
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
}

func RegisterModules(r chi.Router, ctx Context) {
	website.Init(r, ctx)
}
