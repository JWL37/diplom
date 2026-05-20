package list_classes

import (
	"context"
	"fmt"

	"classifier-orchestrator/internal/usecase/list_classes/dto"
	repoDto "classifier-orchestrator/internal/usecase/list_classes/repository/dto"
)

func (u *useCase) Execute(ctx context.Context) (dto.ListClassesResponse, error) {
	repoOutput, err := u.repo.ListClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list classes in repository: %w", err)
	}

	resp := make(dto.ListClassesResponse, 0, len(repoOutput))
	for _, class := range repoOutput {
		resp = append(resp, dto.Class{
			ID:              class.ID,
			Name:            class.Name,
			Description:     class.Description,
			Status:          class.Status,
			CreatedAt:       class.CreatedAt,
			UpdatedAt:       class.UpdatedAt,
			PositiveGoldens: mapGoldens(class.PositiveGoldens),
			NegativeGoldens: mapGoldens(class.NegativeGoldens),
		})
	}

	return resp, nil
}

func mapGoldens(goldens []repoDto.Golden) []dto.Golden {
	resp := make([]dto.Golden, 0, len(goldens))
	for _, golden := range goldens {
		resp = append(resp, dto.Golden{
			ID:   golden.ID,
			Text: golden.Text,
		})
	}

	return resp
}
