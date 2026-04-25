package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"classifier-orchestrator/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL", slog.String("host", cfg.DBHost), slog.String("port", cfg.DBPort))

	return pool, nil
}
