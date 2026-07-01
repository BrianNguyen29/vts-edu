package gradebook

import (
	"context"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

// Service is the gradebook application service contract.
type Service interface {
	ListAssessmentAttempts(ctx context.Context, actor auth.Actor, assessmentID string) ([]AssessmentAttempt, error)
	GetAssessmentResults(ctx context.Context, actor auth.Actor, assessmentID string) (*AssessmentResult, error)
	ExportAssessmentAttemptsCSV(ctx context.Context, actor auth.Actor, assessmentID string) ([]byte, error)
	GetClassGradebook(ctx context.Context, actor auth.Actor, classID string) ([]ClassGradebookEntry, error)
	ExportClassGradebookCSV(ctx context.Context, actor auth.Actor, classID string) ([]byte, error)
}

type service struct {
	repo   Repository
	access ClassAccessChecker
}

// NewService creates the concrete gradebook service.
func NewService(repo Repository, access ClassAccessChecker) Service {
	return &service{repo: repo, access: access}
}

func isAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

func requireTeacherOrAdmin(roles []string) error {
	for _, r := range roles {
		if r == "teacher" || r == "admin" {
			return nil
		}
	}
	return auth.ErrUnauthorized
}

func (s *service) ListAssessmentAttempts(ctx context.Context, actor auth.Actor, assessmentID string) ([]AssessmentAttempt, error) {
	if err := requireTeacherOrAdmin(actor.Roles); err != nil {
		return nil, err
	}
	if err := s.canAccessAssessment(ctx, actor, assessmentID); err != nil {
		return nil, err
	}
	return s.repo.ListAssessmentAttempts(ctx, actor.OrgID, assessmentID)
}

func (s *service) GetAssessmentResults(ctx context.Context, actor auth.Actor, assessmentID string) (*AssessmentResult, error) {
	if err := requireTeacherOrAdmin(actor.Roles); err != nil {
		return nil, err
	}
	if err := s.canAccessAssessment(ctx, actor, assessmentID); err != nil {
		return nil, err
	}
	return s.repo.GetAssessmentResults(ctx, actor.OrgID, assessmentID)
}

func (s *service) ExportAssessmentAttemptsCSV(ctx context.Context, actor auth.Actor, assessmentID string) ([]byte, error) {
	attempts, err := s.ListAssessmentAttempts(ctx, actor, assessmentID)
	if err != nil {
		return nil, err
	}
	return renderAssessmentAttemptsCSV(attempts), nil
}

func (s *service) GetClassGradebook(ctx context.Context, actor auth.Actor, classID string) ([]ClassGradebookEntry, error) {
	if err := requireTeacherOrAdmin(actor.Roles); err != nil {
		return nil, err
	}
	if err := s.canAccessClass(ctx, actor, classID); err != nil {
		return nil, err
	}
	return s.repo.GetClassGradebook(ctx, actor.OrgID, classID)
}

func (s *service) ExportClassGradebookCSV(ctx context.Context, actor auth.Actor, classID string) ([]byte, error) {
	entries, err := s.GetClassGradebook(ctx, actor, classID)
	if err != nil {
		return nil, err
	}
	return renderClassGradebookCSV(entries), nil
}

func (s *service) canAccessAssessment(ctx context.Context, actor auth.Actor, assessmentID string) error {
	if isAdmin(actor.Roles) {
		exists, err := s.repo.AssessmentExists(ctx, actor.OrgID, assessmentID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
		return nil
	}
	exists, err := s.repo.IsAssessmentTaughtByTeacher(ctx, actor.OrgID, assessmentID, actor.UserID)
	if err != nil {
		return err
	}
	if !exists {
		return auth.ErrUnauthorized
	}
	return nil
}

func (s *service) canAccessClass(ctx context.Context, actor auth.Actor, classID string) error {
	if isAdmin(actor.Roles) {
		exists, err := s.access.ClassExists(ctx, actor.OrgID, classID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
		return nil
	}
	membership, err := s.access.GetMembershipByUserID(ctx, actor.OrgID, actor.UserID)
	if err != nil {
		return err
	}
	isTeacher, err := s.access.IsClassTeacher(ctx, actor.OrgID, classID, membership.ID)
	if err != nil {
		return err
	}
	if !isTeacher {
		return auth.ErrUnauthorized
	}
	return nil
}
