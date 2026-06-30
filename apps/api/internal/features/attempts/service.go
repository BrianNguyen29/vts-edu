package attempts

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/jackc/pgx/v5"
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// Service is the attempts application service contract.
type Service interface {
	GetAttempt(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptSnapshot, error)
	SaveAnswer(ctx context.Context, actor auth.Actor, attemptID, itemID string, payload json.RawMessage) (*AnswerSaved, error)
	SubmitAttempt(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptSubmitted, error)
}

type service struct {
	repo Repository
	tm   TransactionManager
}

// NewService creates the concrete attempts service.
func NewService(repo Repository, tm TransactionManager) Service {
	return &service{repo: repo, tm: tm}
}

// GetAttempt returns the authenticated actor's attempt snapshot.
func (s *service) GetAttempt(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptSnapshot, error) {
	attempt, err := s.repo.GetAttempt(ctx, attemptID, actor.OrgID, actor.UserID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetAttemptItems(ctx, attemptID, actor.OrgID)
	if err != nil {
		return nil, err
	}

	snapshot := &AttemptSnapshot{
		ID:             attempt.ID,
		OrganizationID: attempt.OrganizationID,
		AssessmentID:   attempt.AssessmentID,
		PublicationID:  attempt.PublicationID,
		Status:         attempt.Status,
		StartedAt:      attempt.StartedAt,
		ExpiresAt:      attempt.ExpiresAt,
		SubmittedAt:    attempt.SubmittedAt,
		Items:          make([]AttemptItem, len(items)),
	}

	for i, it := range items {
		item := AttemptItem{
			ID:                it.ID,
			QuestionVersionID: it.QuestionVersionID,
			Position:          it.Position,
			Points:            it.Points,
			Prompt:            it.Prompt,
			Choices:           it.Choices,
		}
		if it.Revision != nil {
			item.Answer = &AnswerSnapshot{
				AnswerPayload: it.AnswerPayload,
				Revision:      *it.Revision,
				AnsweredAt:    *it.AnsweredAt,
			}
		}
		snapshot.Items[i] = item
	}

	return snapshot, nil
}

// SaveAnswer persists an answer for an in-progress, non-expired attempt item.
func (s *service) SaveAnswer(ctx context.Context, actor auth.Actor, attemptID, itemID string, payload json.RawMessage) (*AnswerSaved, error) {
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}

	var saved *AnswerSaved
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		attempt, err := s.repo.GetAttemptForUpdate(ctx, tx, attemptID, actor.OrgID, actor.UserID)
		if err != nil {
			return err
		}

		if attempt.Status != "IN_PROGRESS" {
			return ErrAttemptNotInProgress
		}

		if attempt.ExpiresAt != nil && attempt.ExpiresAt.Before(time.Now().UTC()) {
			if err := s.repo.MarkAttemptExpired(ctx, tx, attemptID, actor.OrgID, actor.UserID); err != nil {
				return err
			}
			return ErrAttemptExpired
		}

		exists, err := s.repo.ItemExists(ctx, tx, itemID, attemptID, actor.OrgID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrAnswerItemNotFound
		}

		saved, err = s.repo.UpsertAnswer(ctx, tx, attemptID, itemID, actor.OrgID, payload)
		return err
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// SubmitAttempt transitions an owned in-progress attempt to SUBMITTED or EXPIRED.
// It grades MCQ answers synchronously and persists the result.
func (s *service) SubmitAttempt(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptSubmitted, error) {
	var result *AttemptSubmitted
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		attempt, err := s.repo.GetAttemptForUpdate(ctx, tx, attemptID, actor.OrgID, actor.UserID)
		if err != nil {
			return err
		}

		switch attempt.Status {
		case "SUBMITTED":
			result = &AttemptSubmitted{
				ID:            attemptID,
				Status:        "SUBMITTED",
				SubmittedAt:   *attempt.SubmittedAt,
				Score:         defaultString(attempt.Score, "0.00"),
				MaxScore:      defaultString(attempt.MaxScore, "0.00"),
				GradingStatus: defaultString(attempt.GradingStatus, "GRADED"),
			}
			return nil
		case "EXPIRED":
			return ErrAttemptExpired
		case "IN_PROGRESS":
			if attempt.ExpiresAt != nil && attempt.ExpiresAt.Before(time.Now().UTC()) {
				if err := s.repo.MarkAttemptExpired(ctx, tx, attemptID, actor.OrgID, actor.UserID); err != nil {
					return err
				}
				return ErrAttemptExpired
			}

			items, err := s.repo.GetAttemptItems(ctx, attemptID, actor.OrgID)
			if err != nil {
				return err
			}
			score, maxScore := gradeAttempt(items)

			grading, err := s.repo.SubmitAttempt(ctx, tx, attemptID, actor.OrgID, actor.UserID, score, maxScore, "GRADED")
			if err != nil {
				return err
			}
			result = &AttemptSubmitted{
				ID:            attemptID,
				Status:        "SUBMITTED",
				SubmittedAt:   grading.SubmittedAt,
				Score:         grading.Score,
				MaxScore:      grading.MaxScore,
				GradingStatus: grading.GradingStatus,
			}
			return nil
		default:
			return ErrAttemptNotInProgress
		}
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func gradeAttempt(items []AttemptItemRow) (string, string) {
	score := big.NewRat(0, 1)
	maxScore := big.NewRat(0, 1)

	for _, it := range items {
		points, ok := new(big.Rat).SetString(it.Points)
		if !ok {
			points = big.NewRat(0, 1)
		}
		maxScore.Add(maxScore, points)

		if it.Revision != nil && answerMatches(it.AnswerPayload, it.AnswerKey) {
			score.Add(score, points)
		}
	}

	return score.FloatString(2), maxScore.FloatString(2)
}

func answerMatches(answer json.RawMessage, answerKey json.RawMessage) bool {
	var selected struct {
		SelectedOption string `json:"selected_option"`
	}
	if err := json.Unmarshal(answer, &selected); err != nil {
		return false
	}

	var key struct {
		CorrectOption string `json:"correct_option"`
	}
	if err := json.Unmarshal(answerKey, &key); err != nil {
		return false
	}

	return selected.SelectedOption != "" && selected.SelectedOption == key.CorrectOption
}

func defaultString(s *string, fallback string) string {
	if s == nil || *s == "" {
		return fallback
	}
	return *s
}
