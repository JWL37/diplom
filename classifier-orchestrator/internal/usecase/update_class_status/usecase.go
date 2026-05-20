package update_class_status

import (
	"context"
	"errors"

	"classifier-orchestrator/internal/usecase/update_class_status/dto"
	"classifier-orchestrator/internal/usecase/update_class_status/repository"
)

var ErrClassNotFound = errors.New("class not found")

type UseCase interface {
	Execute(ctx context.Context, input dto.UpdateClassStatusRequest) (dto.UpdateClassStatusResponse, error)
}

type useCase struct {
	repo repository.Repository
}

func New(repo repository.Repository) UseCase {
	return &useCase{
		repo: repo,
	}
}
