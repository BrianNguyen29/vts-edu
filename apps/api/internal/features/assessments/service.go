package assessments

import (
	"context"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
)

// Service is the assessments application service contract.
type Service interface {
	ListAssessments(ctx context.Context, orgID string, opts ListOptions) ([]AssessmentListItem, *PageInfo, error)
}

type service struct {
	repo Repository
}

// NewService creates the concrete assessments service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListAssessments returns the tenant-scoped published/open assessment list.
func (s *service) ListAssessments(ctx context.Context, orgID string, opts ListOptions) ([]AssessmentListItem, *PageInfo, error) {
	queryOpts := opts
	if opts.Limit > 0 {
		queryOpts.Limit = opts.Limit + 1
	}

	rows, err := s.repo.ListPublishedByOrganization(ctx, orgID, queryOpts)
	if err != nil {
		return nil, nil, err
	}

	page := &PageInfo{Limit: opts.Limit, Offset: opts.Offset}
	if opts.Limit > 0 {
		if len(rows) > opts.Limit {
			page.HasMore = true
			last := rows[opts.Limit-1]
			cursor := pagination.Encode(pagination.Cursor{Key: last.CreatedAt, ID: last.ID})
			page.NextCursor = &cursor
			rows = rows[:opts.Limit]
		}
	}

	if opts.Count {
		count, err := s.repo.CountPublishedByOrganization(ctx, orgID, opts)
		if err != nil {
			return nil, nil, err
		}
		page.TotalCount = &count
	}

	items := make([]AssessmentListItem, len(rows))
	for i, r := range rows {
		items[i] = AssessmentListItem{
			ID:              r.ID,
			Title:           r.Title,
			Status:          r.Status,
			DurationMinutes: r.DurationMinutes,
		}
	}
	return items, page, nil
}
