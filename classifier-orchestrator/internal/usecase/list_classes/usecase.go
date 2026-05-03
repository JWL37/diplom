package list_classes

import (
	"context"

	"classifier-orchestrator/internal/usecase/list_classes/dto"
	"classifier-orchestrator/internal/usecase/list_classes/repository"
)

type UseCase interface {
	Execute(ctx context.Context) (dto.ListClassesResponse, error)
}

type useCase struct {
	repo repository.Repository
}

func New(repo repository.Repository) UseCase {
	return &useCase{
		repo: repo,
	}
}
