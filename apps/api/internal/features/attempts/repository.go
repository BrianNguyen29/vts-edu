package attempts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new attempts repository backed by a pgx pool.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetAttempt(ctx context.Context, attemptID, orgID, userID string) (*Attempt, error) {
	query := `
		SELECT
			a.id,
			a.organization_id,
			a.assessment_id,
			a.publication_id,
			a.status,
			a.started_at,
			a.expires_at,
			a.submitted_at,
			a.score::text,
			a.max_score::text,
			a.grading_status
		FROM attempts a
		WHERE a.id = $1
		  AND a.organization_id = $2
		  AND a.student_user_id = $3
		LIMIT 1
	`

	var a Attempt
	err := r.pool.QueryRow(ctx, query, attemptID, orgID, userID).Scan(
		&a.ID,
		&a.OrganizationID,
		&a.AssessmentID,
		&a.PublicationID,
		&a.Status,
		&a.StartedAt,
		&a.ExpiresAt,
		&a.SubmittedAt,
		&a.Score,
		&a.MaxScore,
		&a.GradingStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}
	return &a, nil
}

func (r *repository) GetAttemptItems(ctx context.Context, attemptID, orgID string) ([]AttemptItemRow, error) {
	query := `
		SELECT
			ai.id,
			ai.question_version_id,
			ai.position,
			ai.points::text,
			aa.answer_payload,
			ai.answer_key_json,
			aa.revision,
			aa.answered_at
		FROM attempt_items ai
		LEFT JOIN attempt_answers aa
			ON aa.attempt_item_id = ai.id
			AND aa.organization_id = ai.organization_id
			AND aa.attempt_id = ai.attempt_id
		WHERE ai.attempt_id = $1
		  AND ai.organization_id = $2
		ORDER BY ai.position
	`

	rows, err := r.pool.Query(ctx, query, attemptID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list attempt items: %w", err)
	}
	defer rows.Close()

	var items []AttemptItemRow
	for rows.Next() {
		var it AttemptItemRow
		if err := rows.Scan(
			&it.ID,
			&it.QuestionVersionID,
			&it.Position,
			&it.Points,
			&it.AnswerPayload,
			&it.AnswerKey,
			&it.Revision,
			&it.AnsweredAt,
		); err != nil {
			return nil, fmt.Errorf("scan attempt item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attempt items: %w", err)
	}
	return items, nil
}

func (r *repository) GetAttemptForUpdate(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID string) (*Attempt, error) {
	query := `
		SELECT
			id,
			organization_id,
			assessment_id,
			publication_id,
			status,
			started_at,
			expires_at,
			submitted_at,
			score::text,
			max_score::text,
			grading_status
		FROM attempts
		WHERE id = $1
		  AND organization_id = $2
		  AND student_user_id = $3
		FOR UPDATE
	`

	var a Attempt
	err := tx.QueryRow(ctx, query, attemptID, orgID, userID).Scan(
		&a.ID,
		&a.OrganizationID,
		&a.AssessmentID,
		&a.PublicationID,
		&a.Status,
		&a.StartedAt,
		&a.ExpiresAt,
		&a.SubmittedAt,
		&a.Score,
		&a.MaxScore,
		&a.GradingStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get attempt for update: %w", err)
	}
	return &a, nil
}

func (r *repository) ItemExists(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error) {
	query := `
		SELECT 1
		FROM attempt_items
		WHERE id = $1
		  AND attempt_id = $2
		  AND organization_id = $3
		LIMIT 1
	`

	var n int
	err := tx.QueryRow(ctx, query, itemID, attemptID, orgID).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check item exists: %w", err)
	}
	return true, nil
}

func (r *repository) UpsertAnswer(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*AnswerSaved, error) {
	query := `
		INSERT INTO attempt_answers (
			organization_id,
			attempt_id,
			attempt_item_id,
			answer_payload,
			revision,
			answered_at,
			updated_at
		) VALUES ($1, $2, $3, $4, 1, now(), now())
		ON CONFLICT (organization_id, attempt_id, attempt_item_id)
		DO UPDATE SET
			answer_payload = EXCLUDED.answer_payload,
			revision = attempt_answers.revision + 1,
			answered_at = now(),
			updated_at = now()
		RETURNING revision, answered_at, answer_payload
	`

	var saved AnswerSaved
	saved.AttemptItemID = itemID
	err := tx.QueryRow(ctx, query, orgID, attemptID, itemID, payload).Scan(
		&saved.Revision,
		&saved.AnsweredAt, // not used in response, but scanned for completeness
		&saved.AnswerPayload,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert answer: %w", err)
	}
	return &saved, nil
}

func (r *repository) MarkAttemptExpired(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID string) error {
	query := `
		UPDATE attempts
		SET status = 'EXPIRED', updated_at = now()
		WHERE id = $1
		  AND organization_id = $2
		  AND student_user_id = $3
	`
	_, err := tx.Exec(ctx, query, attemptID, orgID, userID)
	if err != nil {
		return fmt.Errorf("mark attempt expired: %w", err)
	}
	return nil
}

func (r *repository) SubmitAttempt(ctx context.Context, tx pgx.Tx, attemptID, orgID, userID, score, maxScore, gradingStatus string) (*GradingResult, error) {
	query := `
		UPDATE attempts
		SET status = 'SUBMITTED',
		    submitted_at = now(),
		    score = $4,
		    max_score = $5,
		    grading_status = $6,
		    updated_at = now()
		WHERE id = $1
		  AND organization_id = $2
		  AND student_user_id = $3
		  AND status = 'IN_PROGRESS'
		RETURNING submitted_at, score::text, max_score::text, grading_status
	`

	var result GradingResult
	err := tx.QueryRow(ctx, query, attemptID, orgID, userID, score, maxScore, gradingStatus).Scan(
		&result.SubmittedAt,
		&result.Score,
		&result.MaxScore,
		&result.GradingStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Not IN_PROGRESS anymore; caller will reconcile.
		return nil, ErrAttemptNotInProgress
	}
	if err != nil {
		return nil, fmt.Errorf("submit attempt: %w", err)
	}
	return &result, nil
}
