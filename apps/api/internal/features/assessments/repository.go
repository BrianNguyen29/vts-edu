package assessments

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the assessments feature.
type Repository interface {
	ListPublishedByOrganization(ctx context.Context, orgID string) ([]Assessment, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new assessments repository backed by a pgx pool.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) ListPublishedByOrganization(ctx context.Context, orgID string) ([]Assessment, error) {
	query := `
		SELECT id, title, status, duration_minutes
		FROM assessments
		WHERE organization_id = $1
		  AND status IN ('OPEN', 'PUBLISHED')
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list assessments: %w", err)
	}
	defer rows.Close()

	var list []Assessment
	for rows.Next() {
		var a Assessment
		if err := rows.Scan(&a.ID, &a.Title, &a.Status, &a.DurationMinutes); err != nil {
			return nil, fmt.Errorf("scan assessment: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assessments: %w", err)
	}
	if list == nil {
		list = []Assessment{}
	}
	return list, nil
}
