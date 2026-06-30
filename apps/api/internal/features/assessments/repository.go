package assessments

import (
	"context"
	"fmt"

	assessmentsqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/assessments/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the assessments feature.
type Repository interface {
	ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error)
}

type sqlcRepository struct {
	queries *assessmentsqlc.Queries
}

// NewRepository creates a new assessments repository backed by generated sqlc
// queries. It preserves the existing Repository interface.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: assessmentsqlc.New(pool)}
}

func (r *sqlcRepository) ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error) {
	var orgUUID pgtype.UUID
	if err := orgUUID.Scan(orgID); err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	rows, err := r.queries.ListPublishedByOrganization(ctx, assessmentsqlc.ListPublishedByOrganizationParams{
		OrganizationID: orgUUID,
		Column2:        opts.Query,
		Column3:        int32(opts.Limit),
		Column4:        int32(opts.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list assessments: %w", err)
	}

	list := make([]Assessment, len(rows))
	for i, row := range rows {
		list[i] = Assessment{
			ID:              row.ID.String(),
			Title:           row.Title,
			Status:          row.Status,
			DurationMinutes: int(row.DurationMinutes),
		}
	}
	return list, nil
}
