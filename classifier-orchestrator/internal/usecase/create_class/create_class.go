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

	repoInput.PositiveGoldens = make([]repoDto.GoldenInput, 0, len(input.PositiveGoldens))
	for _, golden := range input.PositiveGoldens {
		repoInput.PositiveGoldens = append(repoInput.PositiveGoldens, repoDto.GoldenInput{Text: golden.Text})
	}

	repoInput.NegativeGoldens = make([]repoDto.GoldenInput, 0, len(input.NegativeGoldens))
	for _, golden := range input.NegativeGoldens {
		repoInput.NegativeGoldens = append(repoInput.NegativeGoldens, repoDto.GoldenInput{Text: golden.Text})
	}

	repoOutput, err := u.repo.CreateClass(ctx, repoInput)
	if err != nil {
		return dto.CreateClassResponse{}, fmt.Errorf("failed to create class in repository: %w", err)
	}

	return dto.CreateClassResponse{
		ID: repoOutput.ID,
	}, nil
}
