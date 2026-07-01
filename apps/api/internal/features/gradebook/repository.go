package gradebook

import (
	"context"
	"fmt"
	"time"

	gradebooksqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/gradebook/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the gradebook feature.
type Repository interface {
	AssessmentExists(ctx context.Context, orgID, assessmentID string) (bool, error)
	ListAssessmentAttempts(ctx context.Context, orgID, assessmentID string) ([]AssessmentAttempt, error)
	GetAssessmentResults(ctx context.Context, orgID, assessmentID string) (*AssessmentResult, error)
	IsAssessmentTaughtByTeacher(ctx context.Context, orgID, assessmentID, userID string) (bool, error)
	GetClassGradebook(ctx context.Context, orgID, classID string) ([]ClassGradebookEntry, error)
}

// ClassAccessChecker checks whether a teacher can access a class.
type ClassAccessChecker interface {
	GetMembershipByUserID(ctx context.Context, orgID, userID string) (MembershipInfo, error)
	IsClassTeacher(ctx context.Context, orgID, classID, membershipID string) (bool, error)
	ClassExists(ctx context.Context, orgID, classID string) (bool, error)
}

// MembershipInfo is a minimal view of an organization membership.
type MembershipInfo struct {
	ID string
}

type sqlcRepository struct {
	queries *gradebooksqlc.Queries
}

// NewRepository creates a new gradebook repository backed by generated sqlc queries.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: gradebooksqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func nonEmptyStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *sqlcRepository) AssessmentExists(ctx context.Context, orgID, assessmentID string) (bool, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return false, fmt.Errorf("invalid assessment id: %w", err)
	}
	exists, err := r.queries.AssessmentExists(ctx, gradebooksqlc.AssessmentExistsParams{
		ID:             assessment,
		OrganizationID: org,
	})
	if err != nil {
		return false, fmt.Errorf("check assessment exists: %w", err)
	}
	return exists, nil
}

func (r *sqlcRepository) ListAssessmentAttempts(ctx context.Context, orgID, assessmentID string) ([]AssessmentAttempt, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}

	rows, err := r.queries.ListAssessmentAttempts(ctx, gradebooksqlc.ListAssessmentAttemptsParams{
		OrganizationID: org,
		AssessmentID:   assessment,
	})
	if err != nil {
		return nil, fmt.Errorf("list assessment attempts: %w", err)
	}

	result := make([]AssessmentAttempt, len(rows))
	for i, row := range rows {
		attempt := AssessmentAttempt{
			ID:            row.ID.String(),
			AssessmentID:  row.AssessmentID.String(),
			StudentUserID: row.StudentUserID.String(),
			Status:        row.Status,
			StartedAt:     tsPtr(row.StartedAt),
			ExpiresAt:     tsPtr(row.ExpiresAt),
			SubmittedAt:   tsPtr(row.SubmittedAt),
			Score:         nonEmptyStringPtr(row.Score),
			MaxScore:      nonEmptyStringPtr(row.MaxScore),
		}
		if row.StudentName.Valid {
			attempt.StudentName = row.StudentName.String
		}
		attempt.GradingStatus = textPtr(row.GradingStatus)
		result[i] = attempt
	}
	return result, nil
}

func (r *sqlcRepository) GetAssessmentResults(ctx context.Context, orgID, assessmentID string) (*AssessmentResult, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}

	row, err := r.queries.GetAssessmentResults(ctx, gradebooksqlc.GetAssessmentResultsParams{
		OrganizationID: org,
		AssessmentID:   assessment,
	})
	if err != nil {
		return nil, fmt.Errorf("get assessment results: %w", err)
	}

	return &AssessmentResult{
		AssessmentID:    assessmentID,
		TotalAttempts:   row.TotalAttempts,
		SubmittedCount:  row.SubmittedCount,
		InProgressCount: row.InProgressCount,
		ExpiredCount:    row.ExpiredCount,
		AverageScore:    nonEmptyStringPtr(row.AverageScore),
		MaxScore:        nonEmptyStringPtr(row.MaxScore),
	}, nil
}

func (r *sqlcRepository) IsAssessmentTaughtByTeacher(ctx context.Context, orgID, assessmentID, userID string) (bool, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return false, fmt.Errorf("invalid assessment id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user id: %w", err)
	}

	exists, err := r.queries.IsAssessmentTaughtByTeacher(ctx, gradebooksqlc.IsAssessmentTaughtByTeacherParams{
		OrganizationID: org,
		AssessmentID:   assessment,
		UserID:         usr,
	})
	if err != nil {
		return false, fmt.Errorf("check assessment teacher: %w", err)
	}
	return exists, nil
}

func (r *sqlcRepository) GetClassGradebook(ctx context.Context, orgID, classID string) ([]ClassGradebookEntry, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	class, err := toUUID(classID)
	if err != nil {
		return nil, fmt.Errorf("invalid class id: %w", err)
	}

	rows, err := r.queries.GetClassGradebook(ctx, gradebooksqlc.GetClassGradebookParams{
		OrganizationID: org,
		ID:             class,
	})
	if err != nil {
		return nil, fmt.Errorf("get class gradebook: %w", err)
	}

	result := make([]ClassGradebookEntry, len(rows))
	for i, row := range rows {
		entry := ClassGradebookEntry{
			StudentUserID:   row.StudentUserID.String(),
			AssessmentID:    row.AssessmentID.String(),
			AssessmentTitle: row.AssessmentTitle,
		}
		if row.StudentName.Valid {
			entry.StudentName = row.StudentName.String
		}
		if row.AttemptID.Valid {
			id := row.AttemptID.String()
			entry.AttemptID = &id
		}
		if row.Status != "" {
			entry.Status = &row.Status
		}
		entry.Score = nonEmptyStringPtr(row.Score)
		entry.MaxScore = nonEmptyStringPtr(row.MaxScore)
		entry.SubmittedAt = tsPtr(row.SubmittedAt)
		result[i] = entry
	}
	return result, nil
}

// Ensure no accidental interface breakage.
var _ Repository = (*sqlcRepository)(nil)
