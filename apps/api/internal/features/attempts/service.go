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
	ListAssignedAssessments(ctx context.Context, actor auth.Actor) ([]AssignedAssessment, error)
	ListAttemptHistory(ctx context.Context, actor auth.Actor, opts ListOptions) ([]StudentAttempt, *PageInfo, error)
	GetAttemptResult(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptResult, error)
	StartAttempt(ctx context.Context, actor auth.Actor, assessmentID string) (*AttemptSnapshot, error)
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
		ServerTime:     time.Now().UTC(),
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

// ListAttemptHistory returns the current student's submitted/in-progress attempts.
func (s *service) ListAttemptHistory(ctx context.Context, actor auth.Actor, opts ListOptions) ([]StudentAttempt, *PageInfo, error) {
	if !hasRole(actor.Roles, "student") {
		return nil, nil, auth.ErrUnauthorized
	}
	return s.repo.ListStudentAttempts(ctx, actor.OrgID, actor.UserID, opts)
}

// GetAttemptResult returns the graded review view for a submitted or expired attempt.
func (s *service) GetAttemptResult(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptResult, error) {
	attempt, err := s.repo.GetAttempt(ctx, attemptID, actor.OrgID, actor.UserID)
	if err != nil {
		return nil, err
	}
	if attempt.Status != "SUBMITTED" && attempt.Status != "EXPIRED" {
		return nil, ErrAttemptNotSubmitted
	}

	items, err := s.repo.GetAttemptItems(ctx, attemptID, actor.OrgID)
	if err != nil {
		return nil, err
	}

	result := &AttemptResult{
		ID:            attempt.ID,
		AssessmentID:  attempt.AssessmentID,
		Status:        attempt.Status,
		SubmittedAt:   attempt.SubmittedAt,
		Score:         defaultString(attempt.Score, "0.00"),
		MaxScore:      defaultString(attempt.MaxScore, "0.00"),
		GradingStatus: defaultString(attempt.GradingStatus, "GRADED"),
		ServerTime:    time.Now().UTC(),
		Items:         make([]AttemptResultItem, len(items)),
	}

	for i, it := range items {
		item := AttemptResultItem{
			ID:                it.ID,
			QuestionVersionID: it.QuestionVersionID,
			Position:          it.Position,
			Points:            it.Points,
			Prompt:            it.Prompt,
			Choices:           it.Choices,
			CorrectAnswer:     it.AnswerKey,
		}
		if it.Revision != nil {
			item.StudentAnswer = &AttemptResultAnswer{
				AnswerPayload: it.AnswerPayload,
				AnsweredAt:    *it.AnsweredAt,
			}
		}
		item.IsCorrect = it.Revision != nil && answerMatches(it.AnswerPayload, it.AnswerKey)
		result.Items[i] = item
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

// ListAssignedAssessments returns all assessments assigned to the current student with availability.
func (s *service) ListAssignedAssessments(ctx context.Context, actor auth.Actor) ([]AssignedAssessment, error) {
	if !hasRole(actor.Roles, "student") {
		return nil, auth.ErrUnauthorized
	}
	rows, err := s.repo.ListAssignedAssessments(ctx, actor.OrgID, actor.UserID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for i := range rows {
		rows[i].Availability = availabilityAt(now, rows[i].OpensAt, rows[i].ClosesAt)
	}
	return rows, nil
}

func availabilityAt(now time.Time, opensAt, closesAt *string) string {
	var opens, closes *time.Time
	if opensAt != nil && *opensAt != "" {
		if t, err := time.Parse(time.RFC3339, *opensAt); err == nil {
			opens = &t
		}
	}
	if closesAt != nil && *closesAt != "" {
		if t, err := time.Parse(time.RFC3339, *closesAt); err == nil {
			closes = &t
		}
	}
	if closes != nil && !now.Before(*closes) {
		return "closed"
	}
	if opens != nil && now.Before(*opens) {
		return "upcoming"
	}
	return "open"
}

// StartAttempt begins a new attempt or resumes an existing in-progress attempt for an assigned assessment.
func (s *service) StartAttempt(ctx context.Context, actor auth.Actor, assessmentID string) (*AttemptSnapshot, error) {
	if !hasRole(actor.Roles, "student") {
		return nil, auth.ErrUnauthorized
	}

	assigned, err := s.ListAssignedAssessments(ctx, actor)
	if err != nil {
		return nil, err
	}
	var target *AssignedAssessment
	for i := range assigned {
		if assigned[i].ID == assessmentID {
			target = &assigned[i]
			break
		}
	}
	if target == nil || target.Availability != "open" {
		return nil, ErrAssessmentUnavailable
	}

	if target.PublicationID == "" {
		return nil, ErrNoPublication
	}

	existing, err := s.repo.GetInProgressAttempt(ctx, actor.OrgID, actor.UserID, assessmentID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.GetAttempt(ctx, actor, existing.ID)
	}

	count, err := s.repo.CountStudentAttempts(ctx, actor.OrgID, actor.UserID, assessmentID)
	if err != nil {
		return nil, err
	}
	if count >= int64(target.MaxAttempts) {
		return nil, ErrAttemptLimitReached
	}

	snap, pubID, _, err := s.repo.GetLatestPublication(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return nil, ErrNoPublication
	}
	if pubID != target.PublicationID {
		return nil, ErrNoPublication
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(snap.DurationMinutes) * time.Minute)

	var attempt *Attempt
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		attempt, err = s.repo.CreateAttempt(ctx, tx, actor.OrgID, actor.UserID, assessmentID, pubID, now, expiresAt)
		if err != nil {
			return err
		}

		items := flattenSnapshotItems(snap)
		if len(items) > 0 {
			if err := s.repo.CreateAttemptItems(ctx, tx, actor.OrgID, attempt.ID, items); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetAttempt(ctx, actor, attempt.ID)
}

func flattenSnapshotItems(snap *PublicationSnapshot) []AttemptItemInput {
	var inputs []AttemptItemInput
	pos := 1
	for _, section := range snap.Sections {
		for _, item := range section.Items {
			prompt := item.Prompt
			if prompt == nil {
				prompt = json.RawMessage("{}")
			}
			choices := item.Choices
			if choices == nil {
				choices = json.RawMessage("{}")
			}
			answerKey := item.AnswerKey
			if answerKey == nil {
				answerKey = json.RawMessage("{}")
			}
			inputs = append(inputs, AttemptItemInput{
				QuestionVersionID: item.QuestionVersionID,
				Position:          pos,
				Points:            item.Points,
				Prompt:            prompt,
				Choices:           choices,
				AnswerKey:         answerKey,
			})
			pos++
		}
	}
	return inputs
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
