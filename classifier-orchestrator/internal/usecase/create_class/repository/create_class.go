package repository

import (
	"context"
	"fmt"

	"classifier-orchestrator/internal/usecase/create_class/repository/dto"
)

func (r *repository) CreateClass(ctx context.Context, input dto.CreateClassInput) (dto.CreateClassOutput, error) {
	const query = `
		INSERT INTO prediction_classes (name, description)
		VALUES ($1, $2)
		RETURNING id
	`

	var output dto.CreateClassOutput
	err := r.db.QueryRow(ctx, query, input.Name, input.Description).Scan(&output.ID)
	if err != nil {
		return dto.CreateClassOutput{}, fmt.Errorf("failed to insert prediction class: %w", err)
	}

	return output, nil
}
