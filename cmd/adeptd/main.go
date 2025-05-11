package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/AdeptTravel/adept-framework/internal/analytics"
	"github.com/AdeptTravel/adept-framework/internal/app"
	"github.com/AdeptTravel/adept-framework/internal/bus"
	"github.com/AdeptTravel/adept-framework/internal/db"
	"github.com/AdeptTravel/adept-framework/internal/outbox"
)

func main() {
	_ = godotenv.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// infra
	mariadb, err := db.Open()
	if err != nil {
		slog.Error("db", "err", err)
		return
	}
	b := bus.New()
	ob := outbox.Store{DB: mariadb}
	go ob.Pump(ctx, b)

	// context passed to modules
	ax := analytics.Writer{Add: ob.Add}
	modCtx := app.Context{Bus: b, Outbox: ob, Analytics: ax}

	// router
	r := chi.NewRouter()
	app.RegisterMiddlewares(r)
	app.RegisterModules(r, modCtx)

	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() { _ = srv.ListenAndServe() }()
	slog.Info("running on :8080")
	<-ctx.Done()
	_ = srv.Shutdown(ctx)
}
