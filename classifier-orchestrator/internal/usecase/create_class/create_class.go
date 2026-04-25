package create_class

import (
	"context"
	"fmt"

	"classifier-orchestrator/internal/usecase/create_class/dto"
	repoDto "classifier-orchestrator/internal/usecase/create_class/repository/dto"
)

func (u *useCase) Execute(ctx context.Context, input dto.CreateClassRequest) (dto.CreateClassResponse, error) {
	repoInput := repoDto.CreateClassInput{
		Name:        input.Name,
		Description: input.Description,
	}

	repoOutput, err := u.repo.CreateClass(ctx, repoInput)
	if err != nil {
		return dto.CreateClassResponse{}, fmt.Errorf("failed to create class in repository: %w", err)
	}

	return dto.CreateClassResponse{
		ID: repoOutput.ID,
	}, nil
}
