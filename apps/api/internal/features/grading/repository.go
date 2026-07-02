package grading

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gradingsqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/grading/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the persistence contract for the grading feature.
type Repository interface {
	ListReviewQueue(ctx context.Context, orgID, assessmentID string) ([]ReviewQueueEntry, error)
	GetAttemptForGrading(ctx context.Context, orgID, attemptID string) (*AttemptGradingContext, error)
	GetAttemptItemsForGrading(ctx context.Context, orgID, attemptID string) ([]GradingItemDetail, error)
	GetAttemptItemForGrading(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error)
	UpsertItemGrade(ctx context.Context, tx pgx.Tx, p UpsertItemGradeParams) (ItemGradeRow, error)
	RecomputeAttemptScore(ctx context.Context, tx pgx.Tx, orgID, attemptID string) (RecomputeResult, error)
}

// AttemptItemSnapshot is the minimal attempt_item projection needed to
// validate a manual grade.
type AttemptItemSnapshot struct {
	ID                string
	OrganizationID    string
	AttemptID         string
	QuestionVersionID string
	Position          int
	Points            string
	QuestionType      string
	AnswerPayload     json.RawMessage
}

// UpsertItemGradeParams is the input for UpsertItemGrade.
type UpsertItemGradeParams struct {
	OrganizationID string
	AttemptID      string
	AttemptItemID  string
	GraderUserID   string
	AwardedScore   string
	Feedback       *string
}

// ItemGradeRow is the post-write state of an item_grades row.
type ItemGradeRow struct {
	ID             string
	OrganizationID string
	AttemptID      string
	AttemptItemID  string
	GraderUserID   string
	AwardedScore   string
	Feedback       *string
	GradedAt       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// RecomputeResult is the post-update attempt score snapshot.
type RecomputeResult struct {
	Score         string
	MaxScore      string
	GradingStatus string
}

type sqlcRepository struct {
	queries *gradingsqlc.Queries
}

// NewRepository builds a Repository backed by sqlc + the pgx pool.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: gradingsqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func toOptionalText(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
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

func uuidPtr(u pgtype.UUID) *string {
	if u.Valid {
		s := u.String()
		return &s
	}
	return nil
}

func nonEmptyStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *sqlcRepository) ListReviewQueue(ctx context.Context, orgID, assessmentID string) ([]ReviewQueueEntry, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.ListReviewQueue(ctx, gradingsqlc.ListReviewQueueParams{
		OrganizationID: org,
		AssessmentID:   assessment,
	})
	if err != nil {
		return nil, fmt.Errorf("list review queue: %w", err)
	}
	result := make([]ReviewQueueEntry, len(rows))
	for i, row := range rows {
		entry := ReviewQueueEntry{
			AttemptID:     row.AttemptID.String(),
			StudentUserID: row.StudentUserID.String(),
			Status:        row.Status,
			StartedAt:     tsPtr(row.StartedAt),
			SubmittedAt:   tsPtr(row.SubmittedAt),
			ExpiresAt:     tsPtr(row.ExpiresAt),
			PendingItems:  int(row.PendingItems),
			TotalNonMcq:   int(row.TotalNonMcq),
		}
		if row.StudentName.Valid {
			entry.StudentName = row.StudentName.String
		}
		entry.MaxScore = numericPtr(row.MaxScore)
		result[i] = entry
	}
	return result, nil
}

func (r *sqlcRepository) GetAttemptForGrading(ctx context.Context, orgID, attemptID string) (*AttemptGradingContext, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	row, err := r.queries.GetAttemptForGrading(ctx, gradingsqlc.GetAttemptForGradingParams{
		OrganizationID: org,
		ID:             id,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get attempt for grading: %w", err)
	}
	return &AttemptGradingContext{
		AttemptID:     row.ID.String(),
		AssessmentID:  row.AssessmentID.String(),
		StudentUserID: row.StudentUserID.String(),
		Status:        row.Status,
		Score:         numericPtr(row.Score),
		MaxScore:      numericPtr(row.MaxScore),
		GradingStatus: row.GradingStatus.String,
		SubmittedAt:   tsPtr(row.SubmittedAt),
	}, nil
}

func numericPtr(n pgtype.Numeric) *string {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil {
		return nil
	}
	s := fmt.Sprintf("%.2f", f.Float64)
	return &s
}

func (r *sqlcRepository) GetAttemptItemsForGrading(ctx context.Context, orgID, attemptID string) ([]GradingItemDetail, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	rows, err := r.queries.GetAttemptItemsForGrading(ctx, gradingsqlc.GetAttemptItemsForGradingParams{
		OrganizationID: org,
		AttemptID:      id,
	})
	if err != nil {
		return nil, fmt.Errorf("list attempt items for grading: %w", err)
	}
	items := make([]GradingItemDetail, len(rows))
	for i, row := range rows {
		item := GradingItemDetail{
			ID:                row.ID.String(),
			QuestionVersionID: row.QuestionVersionID.String(),
			Position:          int(row.Position),
			Points:            row.Points,
			QuestionType:      row.QuestionType,
			Prompt:            row.PromptJson,
			Choices:           row.ChoicesJson,
		}
		if row.Revision.Valid {
			item.StudentAnswer = &GradingStudentAnswer{
				AnswerPayload: row.AnswerPayload,
				Revision:      row.Revision.Int64,
				AnsweredAt:    row.AnsweredAt.Time,
			}
		}
		if row.ItemGradeID.Valid {
			item.ItemGrade = &GradingItemGrade{
				ID:           row.ItemGradeID.String(),
				GraderUserID: row.GraderUserID.String(),
				AwardedScore: row.AwardedScore,
				Feedback:     nonEmptyStringPtr(row.Feedback),
				GradedAt:     row.GradedAt.Time,
			}
		}
		items[i] = item
	}
	return items, nil
}

func (r *sqlcRepository) GetAttemptItemForGrading(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}
	row, err := r.queries.GetAttemptItemForGrading(ctx, gradingsqlc.GetAttemptItemForGradingParams{
		OrganizationID: org,
		ID:             id,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get attempt item: %w", err)
	}
	return &AttemptItemSnapshot{
		ID:                row.ID.String(),
		OrganizationID:    row.OrganizationID.String(),
		AttemptID:         row.AttemptID.String(),
		QuestionVersionID: row.QuestionVersionID.String(),
		Position:          int(row.Position),
		Points:            row.Points,
		QuestionType:      row.QuestionType,
		AnswerPayload:     row.AnswerPayload,
	}, nil
}

func (r *sqlcRepository) UpsertItemGrade(ctx context.Context, tx pgx.Tx, p UpsertItemGradeParams) (ItemGradeRow, error) {
	org, err := toUUID(p.OrganizationID)
	if err != nil {
		return ItemGradeRow{}, fmt.Errorf("invalid organization id: %w", err)
	}
	attempt, err := toUUID(p.AttemptID)
	if err != nil {
		return ItemGradeRow{}, fmt.Errorf("invalid attempt id: %w", err)
	}
	item, err := toUUID(p.AttemptItemID)
	if err != nil {
		return ItemGradeRow{}, fmt.Errorf("invalid item id: %w", err)
	}
	grader, err := toUUID(p.GraderUserID)
	if err != nil {
		return ItemGradeRow{}, fmt.Errorf("invalid grader id: %w", err)
	}
	score, err := toNumeric(p.AwardedScore)
	if err != nil {
		return ItemGradeRow{}, fmt.Errorf("invalid awarded score: %w", err)
	}
	row, err := r.queries.WithTx(tx).UpsertItemGrade(ctx, gradingsqlc.UpsertItemGradeParams{
		OrganizationID: org,
		AttemptID:      attempt,
		AttemptItemID:  item,
		GraderUserID:   grader,
		AwardedScore:   score,
		Feedback:       toOptionalText(p.Feedback),
	})
	if err != nil {
		return ItemGradeRow{}, fmt.Errorf("upsert item grade: %w", err)
	}
	return ItemGradeRow{
		ID:             row.ID.String(),
		OrganizationID: row.OrganizationID.String(),
		AttemptID:      row.AttemptID.String(),
		AttemptItemID:  row.AttemptItemID.String(),
		GraderUserID:   row.GraderUserID.String(),
		AwardedScore:   row.AwardedScore,
		Feedback:       textPtr(row.Feedback),
		GradedAt:       row.GradedAt.Time,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}, nil
}

func (r *sqlcRepository) RecomputeAttemptScore(ctx context.Context, tx pgx.Tx, orgID, attemptID string) (RecomputeResult, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return RecomputeResult{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(attemptID)
	if err != nil {
		return RecomputeResult{}, fmt.Errorf("invalid attempt id: %w", err)
	}
	row, err := r.queries.WithTx(tx).RecomputeAttemptScore(ctx, gradingsqlc.RecomputeAttemptScoreParams{
		OrganizationID: org,
		ID:             id,
	})
	if err != nil {
		return RecomputeResult{}, fmt.Errorf("recompute attempt score: %w", err)
	}
	return RecomputeResult{
		Score:         row.Score,
		MaxScore:      row.MaxScore,
		GradingStatus: row.GradingStatus.String,
	}, nil
}

// Ensure compile-time conformance.
var _ Repository = (*sqlcRepository)(nil)
