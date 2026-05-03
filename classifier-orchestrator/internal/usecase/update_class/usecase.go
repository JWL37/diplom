package update_class

import (
	"context"
	"errors"

	"classifier-orchestrator/internal/usecase/update_class/dto"
	"classifier-orchestrator/internal/usecase/update_class/repository"
)

var ErrClassNotFound = errors.New("class not found")

type UseCase interface {
	Execute(ctx context.Context, input dto.UpdateClassRequest) (dto.UpdateClassResponse, error)
}

type useCase struct {
	repo repository.Repository
}

func New(repo repository.Repository) UseCase {
	return &useCase{
		repo: repo,
	}
}
