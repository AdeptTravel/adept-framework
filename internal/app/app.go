package app

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/adepttraveler/adept-framework/internal/bus"
	"github.com/adepttraveler/adept-framework/internal/outbox"
	"github.com/adepttraveler/adept-framework/modules/site"
)

type Context struct {
	Bus       *bus.InMem
	Outbox    outbox.Store
	Analytics site.AnalyticsWriter // alias to keep deps simple
}

func RegisterMiddlewares(r chi.Router) {
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
}

/*
func RegisterModules(r chi.Router, ctx Context) {
	site.Init(r, ctx)
}
*/

func RegisterModules(r chi.Router, ctx Context) {
	site.Init(r, struct {
		Bus interface {
			Publish(context.Context, string, any)
			Subscribe(string, interface{})
		}
		Analytics site.AnalyticsWriter
	}{
		Bus:       ctx.Bus,
		Analytics: ctx.Analytics,
	})
}
