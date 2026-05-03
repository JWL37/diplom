package repository

import (
	"context"

	"classifier-orchestrator/internal/usecase/list_classes/repository/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListClasses(ctx context.Context) (dto.ListClassesOutput, error)
}

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}
