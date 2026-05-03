package list_classes

import (
	"context"
	"fmt"

	"classifier-orchestrator/internal/usecase/list_classes/dto"
)

func (u *useCase) Execute(ctx context.Context) (dto.ListClassesResponse, error) {
	repoOutput, err := u.repo.ListClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list classes in repository: %w", err)
	}

	resp := make(dto.ListClassesResponse, 0, len(repoOutput))
	for _, class := range repoOutput {
		resp = append(resp, dto.Class{
			ID:          class.ID,
			Name:        class.Name,
			Description: class.Description,
		})
	}

	return resp, nil
}
