package repository

import (
	"context"
	"errors"
	"fmt"

	"classifier-orchestrator/internal/usecase/update_class/repository/dto"

	"github.com/jackc/pgx/v5"
)

func (r *repository) UpdateClass(ctx context.Context, input dto.UpdateClassInput) (dto.UpdateClassOutput, error) {
	const updateClass = `
		UPDATE prediction_classes
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description
	`
	const deletePositiveGoldens = `
		DELETE FROM positive_goldens
		WHERE class_id = $1
	`
	const insertPositiveGolden = `
		INSERT INTO positive_goldens (class_id, text_content)
		VALUES ($1, $2)
	`
	const deleteNegativeLinks = `
		DELETE FROM class_negative_goldens
		WHERE class_id = $1
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
		return dto.UpdateClassOutput{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	var output dto.UpdateClassOutput
	err = tx.QueryRow(ctx, updateClass, input.Name, input.Description, input.ID).Scan(&output.ID, &output.Name, &output.Description)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UpdateClassOutput{}, ErrClassNotFound
		}
		return dto.UpdateClassOutput{}, fmt.Errorf("failed to update prediction class: %w", err)
	}

	if _, err := tx.Exec(ctx, deletePositiveGoldens, input.ID); err != nil {
		_ = tx.Rollback(ctx)
		return dto.UpdateClassOutput{}, fmt.Errorf("failed to delete positive goldens: %w", err)
	}
	for _, golden := range input.PositiveGoldens {
		if _, err := tx.Exec(ctx, insertPositiveGolden, input.ID, golden.Text); err != nil {
			_ = tx.Rollback(ctx)
			return dto.UpdateClassOutput{}, fmt.Errorf("failed to insert positive golden: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, deleteNegativeLinks, input.ID); err != nil {
		_ = tx.Rollback(ctx)
		return dto.UpdateClassOutput{}, fmt.Errorf("failed to delete negative golden links: %w", err)
	}
	for _, golden := range input.NegativeGoldens {
		var negativeID int32
		if err := tx.QueryRow(ctx, upsertNegativeGolden, golden.Text).Scan(&negativeID); err != nil {
			_ = tx.Rollback(ctx)
			return dto.UpdateClassOutput{}, fmt.Errorf("failed to upsert negative golden: %w", err)
		}
		if _, err := tx.Exec(ctx, linkNegativeGolden, input.ID, negativeID); err != nil {
			_ = tx.Rollback(ctx)
			return dto.UpdateClassOutput{}, fmt.Errorf("failed to link negative golden: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.UpdateClassOutput{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return output, nil
}
