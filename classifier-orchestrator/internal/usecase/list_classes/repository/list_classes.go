package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"classifier-orchestrator/internal/usecase/list_classes/repository/dto"
)

func (r *repository) ListClasses(ctx context.Context) (dto.ListClassesOutput, error) {
	const query = `
		SELECT
			pc.id,
			pc.name,
			COALESCE(pc.description, ''),
			pc.status::text,
			pc.created_at,
			pc.updated_at,
			COALESCE(positive.positive_goldens, '[]'::jsonb),
			COALESCE(negative.negative_goldens, '[]'::jsonb)
		FROM prediction_classes pc
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', pg.id,
					'text', pg.text_content
				)
				ORDER BY pg.id
			) AS positive_goldens
			FROM positive_goldens pg
			WHERE pg.class_id = pc.id
		) positive ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', ng.id,
					'text', ng.text_content
				)
				ORDER BY ng.id
			) AS negative_goldens
			FROM class_negative_goldens cng
			JOIN negative_goldens ng ON ng.id = cng.negative_golden_id
			WHERE cng.class_id = pc.id
		) negative ON TRUE
		ORDER BY pc.id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query classes: %w", err)
	}
	defer rows.Close()

	classes := make(dto.ListClassesOutput, 0)
	for rows.Next() {
		var class dto.Class
		var positiveGoldens []byte
		var negativeGoldens []byte
		if err := rows.Scan(
			&class.ID,
			&class.Name,
			&class.Description,
			&class.Status,
			&class.CreatedAt,
			&class.UpdatedAt,
			&positiveGoldens,
			&negativeGoldens,
		); err != nil {
			return nil, fmt.Errorf("failed to scan class: %w", err)
		}
		if err := json.Unmarshal(positiveGoldens, &class.PositiveGoldens); err != nil {
			return nil, fmt.Errorf("failed to unmarshal positive goldens: %w", err)
		}
		if err := json.Unmarshal(negativeGoldens, &class.NegativeGoldens); err != nil {
			return nil, fmt.Errorf("failed to unmarshal negative goldens: %w", err)
		}
		classes = append(classes, class)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate classes: %w", err)
	}

	return classes, nil
}
