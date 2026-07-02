package grading

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/jackc/pgx/v5"
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// AuditLogger is the small subset of the admin audit insert that the grading
// feature depends on. It is satisfied by the admin package's repository and
// keeps the grading package free of a direct dependency on admin.
type AuditLogger interface {
	InsertAuditLog(ctx context.Context, tx pgx.Tx, p AuditLogEntry) error
}

// Notifier is the small seam the grading feature uses to fire
// `attempt.graded` events at the student. It is satisfied by the
// notifications package; nil is a valid value and the service
// degrades to a no-op (no event) when nil.
type Notifier interface {
	Notify(ctx context.Context, orgID, recipientID, eventType, title, body string, metadata map[string]any)
}

// AuditLogEntry is the in-package shape of an audit row.
type AuditLogEntry struct {
	OrganizationID string
	ActorUserID    string
	Action         string
	ResourceType   string
	ResourceID     string
	BeforeJSON     []byte
	AfterJSON      []byte
	MetadataJSON   []byte
}

// Service is the grading application service contract.
type Service interface {
	ListReviewQueue(ctx context.Context, actor auth.Actor, assessmentID string) ([]ReviewQueueEntry, error)
	GetAttemptForReview(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptGradingContext, error)
	GradeItem(ctx context.Context, actor auth.Actor, attemptID, itemID string, req GradeItemRequest) (*GradeItemResponse, error)
}

type service struct {
	repo     Repository
	tm       TransactionManager
	audit    AuditLogger
	notifier Notifier
}

// NewService constructs the grading service.
func NewService(repo Repository, tm TransactionManager, audit AuditLogger) Service {
	return &service{repo: repo, tm: tm, audit: audit}
}

// SetNotifier wires a notifier into the service. nil is allowed and
// disables the notifier path.
func (s *service) SetNotifier(n Notifier) {
	s.notifier = n
}

// ListReviewQueue returns the attempts for an assessment that have at least
// one essay/short_answer item still pending manual review.
func (s *service) ListReviewQueue(ctx context.Context, actor auth.Actor, assessmentID string) ([]ReviewQueueEntry, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(assessmentID) == "" {
		return nil, ErrNotFound
	}
	return s.repo.ListReviewQueue(ctx, actor.OrgID, assessmentID)
}

// GetAttemptForReview returns the full attempt + items context for the
// review detail page.
func (s *service) GetAttemptForReview(ctx context.Context, actor auth.Actor, attemptID string) (*AttemptGradingContext, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(attemptID) == "" {
		return nil, ErrNotFound
	}
	ctx2, err := s.repo.GetAttemptForGrading(ctx, actor.OrgID, attemptID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.GetAttemptItemsForGrading(ctx, actor.OrgID, attemptID)
	if err != nil {
		return nil, err
	}
	ctx2.Items = items
	return ctx2, nil
}

// GradeItem upserts a manual grade for a single item and recomputes the
// attempt's score.
func (s *service) GradeItem(ctx context.Context, actor auth.Actor, attemptID, itemID string, req GradeItemRequest) (*GradeItemResponse, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(attemptID) == "" || strings.TrimSpace(itemID) == "" {
		return nil, ErrNotFound
	}

	score := strings.TrimSpace(req.AwardedScore)
	if score == "" {
		return nil, fmt.Errorf("%w: awarded_score is required", ErrInvalidScore)
	}
	pts, ok := new(big.Rat).SetString(score)
	if !ok {
		return nil, fmt.Errorf("%w: awarded_score must be a decimal string", ErrInvalidScore)
	}
	if pts.Sign() < 0 {
		return nil, fmt.Errorf("%w: awarded_score must be >= 0", ErrInvalidScore)
	}

	var (
		grade    ItemGradeRow
		recomp   RecomputeResult
		prev     *ItemGradeRow
		feedback *string
	)
	if req.Feedback != nil {
		trimmed := strings.TrimSpace(*req.Feedback)
		if trimmed != "" {
			feedback = &trimmed
		}
	}

	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		item, err := s.repo.GetAttemptItemForGrading(ctx, actor.OrgID, itemID)
		if err != nil {
			return err
		}
		if item.AttemptID != attemptID {
			return ErrItemNotInAttempt
		}
		if !isManuallyGradable(item.QuestionType) {
			return fmt.Errorf("%w: %s items are auto-graded", ErrNotGradeable, item.QuestionType)
		}
		points, ok := new(big.Rat).SetString(item.Points)
		if !ok {
			return fmt.Errorf("invalid item points %q", item.Points)
		}
		if pts.Cmp(points) > 0 {
			return fmt.Errorf("%w: awarded_score %s exceeds item points %s", ErrScoreExceedsPoints, pts.FloatString(2), points.FloatString(2))
		}

		grade, err = s.repo.UpsertItemGrade(ctx, tx, UpsertItemGradeParams{
			OrganizationID: actor.OrgID,
			AttemptID:      attemptID,
			AttemptItemID:  itemID,
			GraderUserID:   actor.UserID,
			AwardedScore:   score,
			Feedback:       feedback,
		})
		if err != nil {
			return err
		}

		recomp, err = s.repo.RecomputeAttemptScore(ctx, tx, actor.OrgID, attemptID)
		if err != nil {
			return err
		}

		// Build the audit before/after snapshots.
		if existing, err := s.repo.GetAttemptItemForGrading(ctx, actor.OrgID, itemID); err == nil {
			_ = existing // already loaded above
		}
		prevJSON, _ := json.Marshal(map[string]any{
			"item_id":        itemID,
			"attempt_id":     attemptID,
			"grading_status": recomp.GradingStatus,
		})

		after := map[string]any{
			"item_id":       itemID,
			"attempt_id":    attemptID,
			"question_type": item.QuestionType,
			"awarded_score": grade.AwardedScore,
			"grader_id":     actor.UserID,
			"graded_at":     grade.GradedAt,
		}
		if grade.Feedback != nil {
			after["feedback"] = *grade.Feedback
		}
		afterJSON, _ := json.Marshal(after)

		meta, _ := json.Marshal(map[string]any{
			"attempt_score":  recomp.Score,
			"attempt_max":    recomp.MaxScore,
			"grading_status": recomp.GradingStatus,
		})
		if err := s.audit.InsertAuditLog(ctx, tx, AuditLogEntry{
			OrganizationID: actor.OrgID,
			ActorUserID:    actor.UserID,
			Action:         "attempt.grade",
			ResourceType:   "attempt_item",
			ResourceID:     itemID,
			BeforeJSON:     prevJSON,
			AfterJSON:      afterJSON,
			MetadataJSON:   meta,
		}); err != nil {
			return err
		}
		_ = prev
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Re-read the items to compute still_pending + total_non_mcq counts for the
	// response. This is best-effort and runs after the transaction commits.
	items, _ := s.repo.GetAttemptItemsForGrading(ctx, actor.OrgID, attemptID)
	pending := 0
	nonMcq := 0
	for _, it := range items {
		if it.QuestionType == "essay" || it.QuestionType == "short_answer" {
			nonMcq++
			if it.ItemGrade == nil {
				pending++
			}
		}
	}

	// Fire an `attempt.graded` notification when the attempt just
	// transitioned to GRADED. Best-effort; never affects the response.
	att, _ := s.repo.GetAttemptForGrading(ctx, actor.OrgID, attemptID)
	if att != nil {
		s.notifyAttemptGraded(
			ctx,
			actor.OrgID,
			attemptID,
			att.StudentUserID,
			att.AssessmentID,
			recomp.GradingStatus,
		)
	}

	return &GradeItemResponse{
		ItemGrade: GradingItemGrade{
			ID:           grade.ID,
			GraderUserID: grade.GraderUserID,
			AwardedScore: grade.AwardedScore,
			Feedback:     grade.Feedback,
			GradedAt:     grade.GradedAt,
		},
		AttemptScore:  recomp.Score,
		AttemptMax:    recomp.MaxScore,
		GradingStatus: recomp.GradingStatus,
		StillPending:  pending,
		TotalNonMcq:   nonMcq,
	}, nil
}

// notifyAttemptGraded fires an `attempt.graded` event at the
// recipient student when the attempt was just promoted to GRADED.
// Best-effort: failures are swallowed by the notifier.
func (s *service) notifyAttemptGraded(
	ctx context.Context,
	orgID, attemptID, studentID, assessmentID, status string,
) {
	if s.notifier == nil || studentID == "" {
		return
	}
	if status != "GRADED" {
		// Only notify when the attempt has fully crossed into GRADED.
		return
	}
	s.notifier.Notify(
		ctx, orgID, studentID, "attempt.graded",
		"Bài thi đã được chấm",
		"Điểm của bạn đã được công bố.",
		map[string]any{
			"attempt_id":    attemptID,
			"assessment_id": assessmentID,
		},
	)
}

func isManuallyGradable(questionType string) bool {
	return questionType == "essay" || questionType == "short_answer"
}

func isTeacherOrAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "teacher" || r == "admin" {
			return true
		}
	}
	return false
}

// Ensure conformance at compile-time.
var _ Service = (*service)(nil)
