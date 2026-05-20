package update_class_status

import (
	"context"
	"errors"
	"fmt"

	"classifier-orchestrator/internal/usecase/update_class_status/dto"
	"classifier-orchestrator/internal/usecase/update_class_status/repository"
	repoDto "classifier-orchestrator/internal/usecase/update_class_status/repository/dto"
)

func (u *useCase) Execute(ctx context.Context, input dto.UpdateClassStatusRequest) (dto.UpdateClassStatusResponse, error) {
	repoOutput, err := u.repo.UpdateClassStatus(ctx, repoDto.UpdateClassStatusInput{
		ID:     input.ID,
		Status: input.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrClassNotFound) {
			return dto.UpdateClassStatusResponse{}, ErrClassNotFound
		}
		return dto.UpdateClassStatusResponse{}, fmt.Errorf("failed to update class status in repository: %w", err)
	}

	return dto.UpdateClassStatusResponse{
		ID:          repoOutput.ID,
		Name:        repoOutput.Name,
		Description: repoOutput.Description,
		Status:      repoOutput.Status,
		UpdatedAt:   repoOutput.UpdatedAt,
	}, nil
}
