package assessments

import (
	"context"
	"fmt"
	"time"

	assessmentsqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/assessments/sqlc"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the assessments feature.
type Repository interface {
	ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error)
	CountPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) (int64, error)
}

type sqlcRepository struct {
	queries *assessmentsqlc.Queries
}

// NewRepository creates a new assessments repository backed by generated sqlc
// queries. It preserves the existing Repository interface.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: assessmentsqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func decodeAssessmentCursor(cursor string) (string, pgtype.UUID, error) {
	if cursor == "" {
		return "", pgtype.UUID{}, nil
	}
	c, err := pagination.Decode(cursor)
	if err != nil {
		return "", pgtype.UUID{}, err
	}
	id, err := toUUID(c.ID)
	if err != nil {
		return "", pgtype.UUID{}, pagination.ErrInvalidCursor
	}
	return c.Key, id, nil
}

func (r *sqlcRepository) ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error) {
	var orgUUID pgtype.UUID
	if err := orgUUID.Scan(orgID); err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	key, cursorID, err := decodeAssessmentCursor(opts.Cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}

	rows, err := r.queries.ListPublishedByOrganization(ctx, assessmentsqlc.ListPublishedByOrganizationParams{
		OrganizationID: orgUUID,
		SearchQuery:    opts.Query,
		CursorKey:      key,
		CursorID:       cursorID,
		PageOffset:     int32(opts.Offset),
		PageLimit:      int32(opts.Limit),
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
			CreatedAt:       row.CreatedAt.Time.Format(time.RFC3339),
		}
	}
	return list, nil
}

func (r *sqlcRepository) CountPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) (int64, error) {
	var orgUUID pgtype.UUID
	if err := orgUUID.Scan(orgID); err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}

	count, err := r.queries.CountPublishedByOrganization(ctx, assessmentsqlc.CountPublishedByOrganizationParams{
		OrganizationID: orgUUID,
		SearchQuery:    opts.Query,
	})
	if err != nil {
		return 0, fmt.Errorf("count assessments: %w", err)
	}
	return count, nil
}
