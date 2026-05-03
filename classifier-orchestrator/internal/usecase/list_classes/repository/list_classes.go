package repository

import (
	"context"
	"fmt"

	"classifier-orchestrator/internal/usecase/list_classes/repository/dto"
)

func (r *repository) ListClasses(ctx context.Context) (dto.ListClassesOutput, error) {
	const query = `
		SELECT id, name, description
		FROM prediction_classes
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query classes: %w", err)
	}
	defer rows.Close()

	classes := make(dto.ListClassesOutput, 0)
	for rows.Next() {
		var class dto.Class
		if err := rows.Scan(&class.ID, &class.Name, &class.Description); err != nil {
			return nil, fmt.Errorf("failed to scan class: %w", err)
		}
		classes = append(classes, class)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate classes: %w", err)
	}

	return classes, nil
}
