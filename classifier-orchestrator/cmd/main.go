package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	app "classifier-orchestrator/internal"
	"classifier-orchestrator/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.LoadConfig()

	application, err := app.New(ctx, logger, cfg)
	if err != nil {
		logger.Error("Failed to initialize application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := application.Run(ctx); err != nil {
		logger.Error("Application failed to run", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Application exited cleanly")
}
