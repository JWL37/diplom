package repository

import (
	"context"
	"errors"

	"classifier-orchestrator/internal/usecase/update_class/repository/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClassNotFound = errors.New("class not found")

type Repository interface {
	UpdateClass(ctx context.Context, input dto.UpdateClassInput) (dto.UpdateClassOutput, error)
}

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}
