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
	const insertPositiveGolden = `
		INSERT INTO positive_goldens (class_id, text_content)
		VALUES ($1, $2)
	`
	const upsertNegativeGolden = `
		INSERT INTO negative_goldens (text_content)
		VALUES ($1)
		ON CONFLICT (text_content) DO UPDATE SET text_content = EXCLUDED.text_content
		RETURNING id
	`
	const linkNegativeGolden = `
		INSERT INTO class_negative_goldens (class_id, negative_golden_id)
		VALUES ($1, $2)
		ON CONFLICT (class_id, negative_golden_id) DO NOTHING
	`

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return dto.CreateClassOutput{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	var output dto.CreateClassOutput
	err = tx.QueryRow(ctx, query, input.Name, input.Description).Scan(&output.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return dto.CreateClassOutput{}, fmt.Errorf("failed to insert prediction class: %w", err)
	}

	for _, golden := range input.PositiveGoldens {
		if _, err := tx.Exec(ctx, insertPositiveGolden, output.ID, golden.Text); err != nil {
			_ = tx.Rollback(ctx)
			return dto.CreateClassOutput{}, fmt.Errorf("failed to insert positive golden: %w", err)
		}
	}

	for _, golden := range input.NegativeGoldens {
		var negativeID int32
		if err := tx.QueryRow(ctx, upsertNegativeGolden, golden.Text).Scan(&negativeID); err != nil {
			_ = tx.Rollback(ctx)
			return dto.CreateClassOutput{}, fmt.Errorf("failed to upsert negative golden: %w", err)
		}
		if _, err := tx.Exec(ctx, linkNegativeGolden, output.ID, negativeID); err != nil {
			_ = tx.Rollback(ctx)
			return dto.CreateClassOutput{}, fmt.Errorf("failed to link negative golden: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.CreateClassOutput{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return output, nil
}
