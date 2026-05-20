package repository

import (
	"context"
	"errors"
	"fmt"

	"classifier-orchestrator/internal/usecase/update_class_status/repository/dto"

	"github.com/jackc/pgx/v5"
)

func (r *repository) UpdateClassStatus(ctx context.Context, input dto.UpdateClassStatusInput) (dto.UpdateClassStatusOutput, error) {
	const query = `
		UPDATE prediction_classes
		SET status = $1::class_status, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, COALESCE(description, ''), status::text, updated_at
	`

	var output dto.UpdateClassStatusOutput
	err := r.db.QueryRow(ctx, query, input.Status, input.ID).Scan(
		&output.ID,
		&output.Name,
		&output.Description,
		&output.Status,
		&output.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UpdateClassStatusOutput{}, ErrClassNotFound
		}
		return dto.UpdateClassStatusOutput{}, fmt.Errorf("failed to update prediction class status: %w", err)
	}

	return output, nil
}
