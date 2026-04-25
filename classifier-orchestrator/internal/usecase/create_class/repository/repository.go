package repository

import (
	"context"

	"classifier-orchestrator/internal/usecase/create_class/repository/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateClass(ctx context.Context, input dto.CreateClassInput) (dto.CreateClassOutput, error)
}

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}
