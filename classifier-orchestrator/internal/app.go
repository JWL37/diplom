package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"classifier-orchestrator/internal/app/database"
	"classifier-orchestrator/internal/app/rest"
	"classifier-orchestrator/internal/config"
	usecaseCreateClass "classifier-orchestrator/internal/usecase/create_class"
	repoCreateClass "classifier-orchestrator/internal/usecase/create_class/repository"
	usecaseListClasses "classifier-orchestrator/internal/usecase/list_classes"
	repoListClasses "classifier-orchestrator/internal/usecase/list_classes/repository"
	usecaseUpdateClass "classifier-orchestrator/internal/usecase/update_class"
	repoUpdateClass "classifier-orchestrator/internal/usecase/update_class/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	logger  *slog.Logger
	httpSrv *http.Server
	db      *pgxpool.Pool
}

func New(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*App, error) {
	db, err := database.NewPostgresDB(ctx, logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("database initialization error: %w", err)
	}

	classRepo := repoCreateClass.New(db)
	createClassUseCase := usecaseCreateClass.New(classRepo)
	listClassesRepo := repoListClasses.New(db)
	listClassesUseCase := usecaseListClasses.New(listClassesRepo)
	updateClassRepo := repoUpdateClass.New(db)
	updateClassUseCase := usecaseUpdateClass.New(updateClassRepo)

	router := rest.NewRouter(logger, createClassUseCase, listClassesUseCase, updateClassUseCase)

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
		db:      db,
	}, nil
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

		a.db.Close()
		a.logger.Info("Database connection pool closed")
		a.logger.Info("Server stopped")
	}

	return nil
}
