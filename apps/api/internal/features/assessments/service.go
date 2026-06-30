package assessments

import (
	"context"
)

// Service is the assessments application service contract.
type Service interface {
	ListAssessments(ctx context.Context, orgID string, opts ListOptions) ([]AssessmentListItem, error)
}

type service struct {
	repo Repository
}

// NewService creates the concrete assessments service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListAssessments returns the tenant-scoped published/open assessment list.
func (s *service) ListAssessments(ctx context.Context, orgID string, opts ListOptions) ([]AssessmentListItem, error) {
	rows, err := s.repo.ListPublishedByOrganization(ctx, orgID, opts)
	if err != nil {
		return nil, err
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
	return items, nil
}
