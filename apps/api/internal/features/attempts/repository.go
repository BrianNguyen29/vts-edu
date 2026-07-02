package attempts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	attemptssqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/attempts/sqlc"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
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
	QuestionType      string
	Position          int
	Points            string
	Prompt            json.RawMessage
	Choices           json.RawMessage
	AnswerPayload     json.RawMessage
	AnswerKey         json.RawMessage
	Revision          *int64
	AnsweredAt        *time.Time
	AwardedScore      *string
	Feedback          *string
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

	ListAssignedAssessments(ctx context.Context, orgID, userID string) ([]AssignedAssessment, error)
	ListStudentAttempts(ctx context.Context, orgID, userID string, opts ListOptions) ([]StudentAttempt, *PageInfo, error)
	GetLatestPublication(ctx context.Context, orgID, assessmentID string) (*PublicationSnapshot, string, string, error)
	GetInProgressAttempt(ctx context.Context, orgID, userID, assessmentID string) (*Attempt, error)
	CountStudentAttempts(ctx context.Context, orgID, userID, assessmentID string) (int64, error)
	CreateAttempt(ctx context.Context, tx pgx.Tx, orgID, userID, assessmentID, publicationID string, startedAt, expiresAt time.Time) (*Attempt, error)
	CreateAttemptItems(ctx context.Context, tx pgx.Tx, orgID, attemptID string, items []AttemptItemInput) error
}

// AttemptItemInput is the data needed to create an attempt_item row.
type AttemptItemInput struct {
	QuestionVersionID string
	QuestionType      string
	Position          int
	Points            string
	Prompt            json.RawMessage
	Choices           json.RawMessage
	AnswerKey         json.RawMessage
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
			QuestionType:      row.QuestionType,
			Position:          int(row.Position),
			Points:            row.AiPoints,
			Prompt:            json.RawMessage(row.PromptJson),
			Choices:           json.RawMessage(row.ChoicesJson),
			AnswerPayload:     json.RawMessage(row.AnswerPayload),
			AnswerKey:         json.RawMessage(row.AnswerKeyJson),
			Revision:          int8Ptr(row.Revision),
			AnsweredAt:        tsPtr(row.AnsweredAt),
			AwardedScore:      nonEmptyStringPtr(row.AwardedScore),
			Feedback:          nonEmptyStringPtr(row.Feedback),
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
	var scoreNum pgtype.Numeric
	if gradingStatus == "PENDING_REVIEW" {
		scoreNum = pgtype.Numeric{}
	} else {
		scoreNum, err = toNumeric(score)
		if err != nil {
			return nil, fmt.Errorf("invalid score: %w", err)
		}
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

func (r *sqlcRepository) ListAssignedAssessments(ctx context.Context, orgID, userID string) ([]AssignedAssessment, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	rows, err := r.queries.ListAssignedAssessments(ctx, attemptssqlc.ListAssignedAssessmentsParams{
		UserID:         usr,
		OrganizationID: org,
	})
	if err != nil {
		return nil, fmt.Errorf("list assigned assessments: %w", err)
	}

	result := make([]AssignedAssessment, len(rows))
	for i, row := range rows {
		publishedAt := ""
		if row.PublishedAt.Valid {
			publishedAt = row.PublishedAt.Time.Format(time.RFC3339)
		}
		pubID := ""
		if row.PublicationID.Valid {
			pubID = row.PublicationID.String()
		}
		result[i] = AssignedAssessment{
			ID:              row.ID.String(),
			Title:           row.Title,
			Status:          row.Status,
			DurationMinutes: int(row.DurationMinutes),
			MaxAttempts:     int(row.MaxAttempts),
			AttemptsUsed:    int(row.AttemptsUsed),
			Revision:        int(row.Revision),
			PublicationID:   pubID,
			PublishedAt:     publishedAt,
			OpensAt:         tsStringPtr(row.OpensAt),
			ClosesAt:        tsStringPtr(row.ClosesAt),
		}
	}
	return result, nil
}

func tsStringPtr(t pgtype.Timestamptz) *string {
	if t.Valid {
		s := t.Time.Format(time.RFC3339)
		return &s
	}
	return nil
}

func (r *sqlcRepository) ListStudentAttempts(ctx context.Context, orgID, userID string, opts ListOptions) ([]StudentAttempt, *PageInfo, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	cursorKey := ""
	cursorUUID := pgtype.UUID{}
	if opts.Cursor != "" {
		c, err := pagination.Decode(opts.Cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorKey = c.Key
		if c.ID != "" {
			cursorUUID, err = toUUID(c.ID)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid cursor: %w", err)
			}
		}
	}

	// Fetch one extra row to determine whether a next page exists.
	rows, err := r.queries.ListStudentAttempts(ctx, attemptssqlc.ListStudentAttemptsParams{
		OrganizationID: org,
		StudentUserID:  usr,
		PageLimit:      int32(limit + 1),
		CursorKey:      cursorKey,
		CursorID:       cursorUUID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("list student attempts: %w", err)
	}

	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}

	result := make([]StudentAttempt, len(rows))
	for i, row := range rows {
		result[i] = StudentAttempt{
			ID:              row.ID.String(),
			AssessmentID:    row.AssessmentID.String(),
			AssessmentTitle: row.AssessmentTitle,
			Status:          row.Status,
			StartedAt:       tsPtr(row.StartedAt),
			ExpiresAt:       tsPtr(row.ExpiresAt),
			SubmittedAt:     tsPtr(row.SubmittedAt),
			Score:           nonEmptyStringPtr(row.Score),
			MaxScore:        nonEmptyStringPtr(row.MaxScore),
			GradingStatus:   textPtr(row.GradingStatus),
		}
	}

	page := &PageInfo{Limit: limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		if last.CreatedAt.Valid && last.ID.Valid {
			cursor := pagination.Encode(pagination.Cursor{
				Key: last.CreatedAt.Time.Format(time.RFC3339Nano),
				ID:  last.ID.String(),
			})
			page.NextCursor = &cursor
		}
	}

	return result, page, nil
}

func nullableString(t pgtype.Text) *string {
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

func (r *sqlcRepository) GetLatestPublication(ctx context.Context, orgID, assessmentID string) (*PublicationSnapshot, string, string, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid organization id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid assessment id: %w", err)
	}

	row, err := r.queries.GetLatestPublication(ctx, attemptssqlc.GetLatestPublicationParams{
		OrganizationID: org,
		AssessmentID:   assessment,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", "", nil
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("get latest publication: %w", err)
	}

	var snap PublicationSnapshot
	if err := json.Unmarshal(row.SnapshotJson, &snap); err != nil {
		return nil, "", "", fmt.Errorf("parse publication snapshot: %w", err)
	}

	publishedAt := ""
	if row.PublishedAt.Valid {
		publishedAt = row.PublishedAt.Time.Format(time.RFC3339)
	}
	return &snap, row.ID.String(), publishedAt, nil
}

func (r *sqlcRepository) GetInProgressAttempt(ctx context.Context, orgID, userID, assessmentID string) (*Attempt, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}

	row, err := r.queries.GetInProgressAttempt(ctx, attemptssqlc.GetInProgressAttemptParams{
		OrganizationID: org,
		StudentUserID:  usr,
		AssessmentID:   assessment,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get in progress attempt: %w", err)
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

func (r *sqlcRepository) CountStudentAttempts(ctx context.Context, orgID, userID, assessmentID string) (int64, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return 0, fmt.Errorf("invalid assessment id: %w", err)
	}

	count, err := r.queries.CountStudentAttempts(ctx, attemptssqlc.CountStudentAttemptsParams{
		OrganizationID: org,
		StudentUserID:  usr,
		AssessmentID:   assessment,
	})
	if err != nil {
		return 0, fmt.Errorf("count student attempts: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) CreateAttempt(ctx context.Context, tx pgx.Tx, orgID, userID, assessmentID, publicationID string, startedAt, expiresAt time.Time) (*Attempt, error) {
	org, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	usr, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	assessment, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	pub, err := toUUID(publicationID)
	if err != nil {
		return nil, fmt.Errorf("invalid publication id: %w", err)
	}

	row, err := r.queries.WithTx(tx).CreateAttempt(ctx, attemptssqlc.CreateAttemptParams{
		OrganizationID: org,
		AssessmentID:   assessment,
		StudentUserID:  usr,
		PublicationID:  pub,
		StartedAt:      pgtype.Timestamptz{Time: startedAt, Valid: true},
		ExpiresAt:      pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create attempt: %w", err)
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

func (r *sqlcRepository) CreateAttemptItems(ctx context.Context, tx pgx.Tx, orgID, attemptID string, items []AttemptItemInput) error {
	org, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	attempt, err := toUUID(attemptID)
	if err != nil {
		return fmt.Errorf("invalid attempt id: %w", err)
	}

	for _, item := range items {
		qv, err := toUUID(item.QuestionVersionID)
		if err != nil {
			return fmt.Errorf("invalid question version id: %w", err)
		}
		points, err := toNumeric(item.Points)
		if err != nil {
			return fmt.Errorf("invalid points: %w", err)
		}
		questionType := item.QuestionType
		if questionType == "" {
			questionType = "multiple_choice"
		}
		if err := r.queries.WithTx(tx).CreateAttemptItem(ctx, attemptssqlc.CreateAttemptItemParams{
			OrganizationID:    org,
			AttemptID:         attempt,
			QuestionVersionID: qv,
			Position:          int32(item.Position),
			Points:            points,
			PromptJson:        []byte(item.Prompt),
			ChoicesJson:       []byte(item.Choices),
			AnswerKeyJson:     []byte(item.AnswerKey),
			QuestionType:      questionType,
		}); err != nil {
			return fmt.Errorf("create attempt item: %w", err)
		}
	}
	return nil
}
