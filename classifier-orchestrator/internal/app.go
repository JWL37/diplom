package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"classifier-orchestrator/internal/app/rest"
	"classifier-orchestrator/internal/config"
)

type App struct {
	logger  *slog.Logger
	httpSrv *http.Server
}

func New(logger *slog.Logger, cfg *config.Config) *App {
	router := rest.NewRouter(logger)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return &App{
		logger:  logger,
		httpSrv: srv,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("Starting HTTP server", slog.String("addr", a.httpSrv.Addr))

	errCh := make(chan error, 1)
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server start error: %w", err)
	case <-ctx.Done():
		a.logger.Info("Shutting down the server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown error: %w", err)
		}
		a.logger.Info("Server stopped")
	}

	return nil
}
