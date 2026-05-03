package update_class

import (
	"context"
	"errors"
	"fmt"

	"classifier-orchestrator/internal/usecase/update_class/dto"
	"classifier-orchestrator/internal/usecase/update_class/repository"
	repoDto "classifier-orchestrator/internal/usecase/update_class/repository/dto"
)

func (u *useCase) Execute(ctx context.Context, input dto.UpdateClassRequest) (dto.UpdateClassResponse, error) {
	repoInput := repoDto.UpdateClassInput{
		ID:          input.ID,
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

	repoOutput, err := u.repo.UpdateClass(ctx, repoInput)
	if err != nil {
		if errors.Is(err, repository.ErrClassNotFound) {
			return dto.UpdateClassResponse{}, ErrClassNotFound
		}
		return dto.UpdateClassResponse{}, fmt.Errorf("failed to update class in repository: %w", err)
	}

	return dto.UpdateClassResponse{
		ID:          repoOutput.ID,
		Name:        repoOutput.Name,
		Description: repoOutput.Description,
	}, nil
}
