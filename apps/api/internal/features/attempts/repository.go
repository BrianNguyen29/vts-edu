package attempts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	attemptssqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/attempts/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Attempt is the internal attempt row shape.
type Attempt struct {
	ID             string
	OrganizationID string
	AssessmentID   string
	PublicationID  *string
	Status         string
	StartedAt      *time.Time
	ExpiresAt      *time.Time
	SubmittedAt    *time.Time
	Score          *string
	MaxScore       *string
	GradingStatus  *string
}

// AttemptItemRow is a query projection of an item plus its optional answer and answer key.
type AttemptItemRow struct {
	ID                string
	QuestionVersionID string
	Position          int
	Points            string
	Prompt            json.RawMessage
	Choices           json.RawMessage
	AnswerPayload     json.RawMessage
	AnswerKey         json.RawMessage
	Revision          *int64
	AnsweredAt        *time.Time
}

// Repository defines persistence operations for the attempts feature.
type Repository interface {
	GetAttempt(ctx context.Context, attemptID, orgID, userID string) (*Attempt, error)
	GetAttemptItems(ctx context.Context, attemptID, orgID string) ([]AttemptItemRow, error)

	GetAttemptForUpdate(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID string) (*Attempt, error)
	ItemExists(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error)
	UpsertAnswer(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*AnswerSaved, error)

	MarkAttemptExpired(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID string) error
	SubmitAttempt(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID, score, maxScore, gradingStatus string) (*GradingResult, error)
}

// GradingResult is the persisted result of a successful submit.
type GradingResult struct {
	SubmittedAt   time.Time
	Score         string
	MaxScore      string
	GradingStatus string
}

type sqlcRepository struct {
	queries *attemptssqlc.Queries
}

// NewRepository creates a new attempts repository backed by generated sqlc queries.
// It preserves the existing Repository interface.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: attemptssqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
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

func toNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func numericPtr(n pgtype.Numeric) *string {
	if n.Valid {
		f, err := n.Float64Value()
		if err != nil {
			return nil
		}
		s := fmt.Sprintf("%.2f", f.Float64)
		return &s
	}
	return nil
}

func (r *sqlcRepository) GetAttempt(ctx context.Context, attemptID, orgID, userID string) (*Attempt, error) {
	id, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	row, err := r.queries.GetAttempt(ctx, attemptssqlc.GetAttemptParams{
		ID:             id,
		OrganizationID: org,
		StudentUserID:  usr,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}

	return &Attempt{
		ID:             row.ID.String(),
		OrganizationID: row.OrganizationID.String(),
		AssessmentID:   row.AssessmentID.String(),
		PublicationID:  uuidPtr(row.PublicationID),
		Status:         row.Status,
		StartedAt:      tsPtr(row.StartedAt),
		ExpiresAt:      tsPtr(row.ExpiresAt),
		SubmittedAt:    tsPtr(row.SubmittedAt),
		Score:          numericPtr(row.Score),
		MaxScore:       numericPtr(row.MaxScore),
		GradingStatus:  textPtr(row.GradingStatus),
	}, nil
}

func (r *sqlcRepository) GetAttemptItems(ctx context.Context, attemptID, orgID string) ([]AttemptItemRow, error) {
	id, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	rows, err := r.queries.GetAttemptItems(ctx, attemptssqlc.GetAttemptItemsParams{
		AttemptID:      id,
		OrganizationID: org,
	})
	if err != nil {
		return nil, fmt.Errorf("list attempt items: %w", err)
	}

	items := make([]AttemptItemRow, len(rows))
	for i, row := range rows {
		items[i] = AttemptItemRow{
			ID:                row.ID.String(),
			QuestionVersionID: row.QuestionVersionID.String(),
			Position:          int(row.Position),
			Points:            row.AiPoints,
			Prompt:            json.RawMessage(row.PromptJson),
			Choices:           json.RawMessage(row.ChoicesJson),
			AnswerPayload:     json.RawMessage(row.AnswerPayload),
			AnswerKey:         json.RawMessage(row.AnswerKeyJson),
			Revision:          int8Ptr(row.Revision),
			AnsweredAt:        tsPtr(row.AnsweredAt),
		}
	}
	return items, nil
}

func int8Ptr(n pgtype.Int8) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

func (r *sqlcRepository) GetAttemptForUpdate(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID string) (*Attempt, error) {
	id, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	row, err := r.queries.WithTx(tx).GetAttemptForUpdate(ctx, attemptssqlc.GetAttemptForUpdateParams{
		ID:             id,
		OrganizationID: org,
		StudentUserID:  usr,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get attempt for update: %w", err)
	}

	return &Attempt{
		ID:             row.ID.String(),
		OrganizationID: row.OrganizationID.String(),
		AssessmentID:   row.AssessmentID.String(),
		PublicationID:  uuidPtr(row.PublicationID),
		Status:         row.Status,
		StartedAt:      tsPtr(row.StartedAt),
		ExpiresAt:      tsPtr(row.ExpiresAt),
		SubmittedAt:    tsPtr(row.SubmittedAt),
		Score:          numericPtr(row.Score),
		MaxScore:       numericPtr(row.MaxScore),
		GradingStatus:  textPtr(row.GradingStatus),
	}, nil
}

func (r *sqlcRepository) ItemExists(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error) {
	id, err := toUUID(itemID)
	if err != nil {
		return false, fmt.Errorf("invalid item id: %w", err)
	}
	attempt, err := toUUID(attemptID)
	if err != nil {
		return false, fmt.Errorf("invalid attempt id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}

	exists, err := r.queries.WithTx(tx).ItemExists(ctx, attemptssqlc.ItemExistsParams{
		ID:             id,
		AttemptID:      attempt,
		OrganizationID: org,
	})
	if err != nil {
		return false, fmt.Errorf("check item exists: %w", err)
	}
	return exists, nil
}

func (r *sqlcRepository) UpsertAnswer(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*AnswerSaved, error) {
	attempt, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	item, err := toUUID(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	row, err := r.queries.WithTx(tx).UpsertAnswer(ctx, attemptssqlc.UpsertAnswerParams{
		OrganizationID: org,
		AttemptID:      attempt,
		AttemptItemID:  item,
		AnswerPayload:  payload,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert answer: %w", err)
	}

	return &AnswerSaved{
		AttemptItemID: itemID,
		Revision:      row.Revision,
		AnswerPayload: json.RawMessage(row.AnswerPayload),
		AnsweredAt:    row.AnsweredAt.Time,
	}, nil
}

func (r *sqlcRepository) MarkAttemptExpired(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID string) error {
	attempt, err := toUUID(attemptID)
	if err != nil {
		return fmt.Errorf("invalid attempt id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := r.queries.WithTx(tx).MarkAttemptExpired(ctx, attemptssqlc.MarkAttemptExpiredParams{
		ID:             attempt,
		OrganizationID: org,
		StudentUserID:  usr,
	}); err != nil {
		return fmt.Errorf("mark attempt expired: %w", err)
	}
	return nil
}

func (r *sqlcRepository) SubmitAttempt(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID, score, maxScore, gradingStatus string) (*GradingResult, error) {
	attempt, err := toUUID(attemptID)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt id: %w", err)
	}
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	scoreNum, err := toNumeric(score)
	if err != nil {
		return nil, fmt.Errorf("invalid score: %w", err)
	}
	maxScoreNum, err := toNumeric(maxScore)
	if err != nil {
		return nil, fmt.Errorf("invalid max score: %w", err)
	}

	row, err := r.queries.WithTx(tx).SubmitAttempt(ctx, attemptssqlc.SubmitAttemptParams{
		ID:             attempt,
		OrganizationID: org,
		StudentUserID:  usr,
		Score:          scoreNum,
		MaxScore:       maxScoreNum,
		GradingStatus:  toText(gradingStatus),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotInProgress
	}
	if err != nil {
		return nil, fmt.Errorf("submit attempt: %w", err)
	}

	return &GradingResult{
		SubmittedAt:   row.SubmittedAt.Time,
		Score:         row.Score,
		MaxScore:      row.MaxScore,
		GradingStatus: row.GradingStatus.String,
	}, nil
}
