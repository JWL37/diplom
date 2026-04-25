package create_class

import (
	"context"

	"classifier-orchestrator/internal/usecase/create_class/dto"
	"classifier-orchestrator/internal/usecase/create_class/repository"
)

type UseCase interface {
	Execute(ctx context.Context, input dto.CreateClassRequest) (dto.CreateClassResponse, error)
}

type useCase struct {
	repo repository.Repository
}

func New(repo repository.Repository) UseCase {
	return &useCase{
		repo: repo,
	}
}
