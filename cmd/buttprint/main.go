package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/api"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/config"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/jackfruit"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/render"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/scoring"
)

type app struct {
	cfg    *config.Config
	logger *slog.Logger
	server *http.Server
}

func newApp() (*app, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	httpClient := &http.Client{Timeout: 20 * time.Second}

	service := domain.NewService(
		jackfruit.NewClient(httpClient, cfg.JackfruitURL),
		scoring.NewScorer(),
		render.NewSVGRenderer(),
	)

	mux := http.NewServeMux()
	api.NewHandler(service, logger.With("component", "api")).RegisterRoutes(mux)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 25 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &app{
		cfg:    cfg,
		logger: logger,
		server: server,
	}, nil
}

func (a *app) run() {
	go func() {
		a.logger.Info("starting server", "port", a.cfg.Port)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a.shutdown(ctx)
}

func (a *app) shutdown(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("server shutdown error", "error", err)
	}
	a.logger.Info("server stopped")
}

func main() {
	a, err := newApp()
	if err != nil {
		slog.Error("failed to start the app", "error", err)
		os.Exit(1)
	}
	a.run()
}
