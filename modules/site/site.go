package website

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/adepttraveler/adept-framework/internal/analytics"
)

type AnalyticsWriter = analytics.Writer

func Init(r chi.Router, cfg struct {
	Bus       interface {
		Publish(context.Context, string, any)
		Subscribe[T any](string, func(context.Context, T))
	}
	Analytics AnalyticsWriter
}) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		cfg.Analytics.Count(r.Context(), "page.view", map[string]string{"path": "/"})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"msg": "hello adept"})
	})

	// Example subscriber that prints analytics events.
	cfg.Bus.Subscribe[map[string]any]("analytics", func(ctx context.Context, e map[string]any) {
		// just log for now
	})
}
