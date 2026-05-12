package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"vnodex/api/internal/config"
	"vnodex/api/internal/handlers"
	"vnodex/api/internal/store"
)

type App struct {
	cfg      config.Config
	store    *store.Storage
	handlers *handlers.Handler
}

func New(cfg config.Config) (*App, error) {
	st, err := store.New(cfg.DBDSN, cfg.RedisAddr)
	if err != nil {
		return nil, err
	}
	return &App{cfg: cfg, store: st, handlers: handlers.New(st)}, nil
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", a.handlers.Healthz)
	r.Get("/v1/state/{walletID}", a.handlers.GetWalletState)
	r.Get("/v1/records/{walletID}", a.handlers.GetWalletRecords)
	r.Get("/v1/network/head", a.handlers.GetHead)
	r.Post("/v1/operations", a.handlers.PostOperation)
	r.Post("/v1/sync/records", a.handlers.PushRecords)
	return r
}
