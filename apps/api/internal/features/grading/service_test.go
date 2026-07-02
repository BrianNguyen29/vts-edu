package grading

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/jackc/pgx/v5"
)

type fakeRepo struct {
	listReview     func(ctx context.Context, orgID, assessmentID string) ([]ReviewQueueEntry, error)
	getAttempt     func(ctx context.Context, orgID, attemptID string) (*AttemptGradingContext, error)
	getItems       func(ctx context.Context, orgID, attemptID string) ([]GradingItemDetail, error)
	getItem        func(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error)
	upsert         func(ctx context.Context, tx pgx.Tx, p UpsertItemGradeParams) (ItemGradeRow, error)
	recompute      func(ctx context.Context, tx pgx.Tx, orgID, attemptID string) (RecomputeResult, error)
	upsertCalls    int
	recomputeCalls int
}

func (f *fakeRepo) ListReviewQueue(ctx context.Context, orgID, assessmentID string) ([]ReviewQueueEntry, error) {
	return f.listReview(ctx, orgID, assessmentID)
}

func (f *fakeRepo) GetAttemptForGrading(ctx context.Context, orgID, attemptID string) (*AttemptGradingContext, error) {
	if f.getAttempt == nil {
		return &AttemptGradingContext{
			AttemptID:     attemptID,
			AssessmentID:  "asm",
			StudentUserID: "stu",
			Status:        "SUBMITTED",
		}, nil
	}
	return f.getAttempt(ctx, orgID, attemptID)
}

func (f *fakeRepo) GetAttemptItemsForGrading(ctx context.Context, orgID, attemptID string) ([]GradingItemDetail, error) {
	if f.getItems == nil {
		return nil, nil
	}
	return f.getItems(ctx, orgID, attemptID)
}

func (f *fakeRepo) GetAttemptItemForGrading(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error) {
	return f.getItem(ctx, orgID, itemID)
}

func (f *fakeRepo) UpsertItemGrade(ctx context.Context, tx pgx.Tx, p UpsertItemGradeParams) (ItemGradeRow, error) {
	f.upsertCalls++
	return f.upsert(ctx, tx, p)
}

func (f *fakeRepo) RecomputeAttemptScore(ctx context.Context, tx pgx.Tx, orgID, attemptID string) (RecomputeResult, error) {
	f.recomputeCalls++
	return f.recompute(ctx, tx, orgID, attemptID)
}

type fakeTM struct{}

func (f *fakeTM) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

type fakeAudit struct {
	entries []AuditLogEntry
}

func (f *fakeAudit) InsertAuditLog(ctx context.Context, tx pgx.Tx, p AuditLogEntry) error {
	f.entries = append(f.entries, p)
	return nil
}

func teacherActor() auth.Actor {
	return auth.Actor{UserID: "u-teacher", OrgID: "org-1", Roles: []string{"teacher"}}
}

func studentActor() auth.Actor {
	return auth.Actor{UserID: "u-student", OrgID: "org-1", Roles: []string{"student"}}
}

func newServiceWithFakes() (*service, *fakeRepo, *fakeAudit) {
	repo := &fakeRepo{}
	audit := &fakeAudit{}
	s := &service{repo: repo, tm: &fakeTM{}, audit: audit}
	return s, repo, audit
}

func TestGradeItem_RejectsNonTeacherOrAdmin(t *testing.T) {
	s, _, _ := newServiceWithFakes()
	_, err := s.GradeItem(context.Background(), studentActor(), "att", "itm", GradeItemRequest{AwardedScore: "1.00"})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestGradeItem_RejectsInvalidScore(t *testing.T) {
	s, _, _ := newServiceWithFakes()
	_, err := s.GradeItem(context.Background(), teacherActor(), "att", "itm", GradeItemRequest{AwardedScore: "abc"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGradeItem_RejectsNegativeScore(t *testing.T) {
	s, _, _ := newServiceWithFakes()
	_, err := s.GradeItem(context.Background(), teacherActor(), "att", "itm", GradeItemRequest{AwardedScore: "-1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGradeItem_RejectsScoreAbovePoints(t *testing.T) {
	s, repo, _ := newServiceWithFakes()
	repo.getItem = func(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error) {
		return &AttemptItemSnapshot{
			ID:           "itm",
			AttemptID:    "att",
			Points:       "1.00",
			QuestionType: "essay",
		}, nil
	}
	_, err := s.GradeItem(context.Background(), teacherActor(), "att", "itm", GradeItemRequest{AwardedScore: "2.00"})
	if err == nil {
		t.Fatal("expected ErrScoreExceedsPoints")
	}
}

func TestGradeItem_RejectsMcqItems(t *testing.T) {
	s, repo, _ := newServiceWithFakes()
	repo.getItem = func(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error) {
		return &AttemptItemSnapshot{
			ID:           "itm",
			AttemptID:    "att",
			Points:       "1.00",
			QuestionType: "multiple_choice",
		}, nil
	}
	_, err := s.GradeItem(context.Background(), teacherActor(), "att", "itm", GradeItemRequest{AwardedScore: "1.00"})
	if !errors.Is(err, ErrNotGradeable) {
		t.Fatalf("expected ErrNotGradeable, got %v", err)
	}
}

func TestGradeItem_RejectsItemNotInAttempt(t *testing.T) {
	s, repo, _ := newServiceWithFakes()
	repo.getItem = func(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error) {
		return &AttemptItemSnapshot{
			ID:           "itm",
			AttemptID:    "other-att",
			Points:       "1.00",
			QuestionType: "essay",
		}, nil
	}
	_, err := s.GradeItem(context.Background(), teacherActor(), "att", "itm", GradeItemRequest{AwardedScore: "0.50"})
	if !errors.Is(err, ErrItemNotInAttempt) {
		t.Fatalf("expected ErrItemNotInAttempt, got %v", err)
	}
}

func TestGradeItem_PersistsGradeRecomputesAndAudits(t *testing.T) {
	s, repo, audit := newServiceWithFakes()
	repo.getItem = func(ctx context.Context, orgID, itemID string) (*AttemptItemSnapshot, error) {
		return &AttemptItemSnapshot{
			ID:           "itm",
			AttemptID:    "att",
			Points:       "2.00",
			QuestionType: "essay",
		}, nil
	}
	repo.upsert = func(ctx context.Context, tx pgx.Tx, p UpsertItemGradeParams) (ItemGradeRow, error) {
		fb := "good"
		return ItemGradeRow{
			ID:             "grade-1",
			OrganizationID: p.OrganizationID,
			AttemptID:      p.AttemptID,
			AttemptItemID:  p.AttemptItemID,
			GraderUserID:   p.GraderUserID,
			AwardedScore:   p.AwardedScore,
			Feedback:       &fb,
		}, nil
	}
	repo.recompute = func(ctx context.Context, tx pgx.Tx, orgID, attemptID string) (RecomputeResult, error) {
		return RecomputeResult{Score: "1.50", MaxScore: "3.00", GradingStatus: "PENDING_REVIEW"}, nil
	}
	repo.getItems = func(ctx context.Context, orgID, attemptID string) ([]GradingItemDetail, error) {
		return []GradingItemDetail{
			{ID: "itm", QuestionType: "essay", ItemGrade: &GradingItemGrade{}},
		}, nil
	}
	fb := "good"
	resp, err := s.GradeItem(context.Background(), teacherActor(), "att", "itm", GradeItemRequest{AwardedScore: "1.50", Feedback: &fb})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.AttemptScore != "1.50" || resp.GradingStatus != "PENDING_REVIEW" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("expected 1 upsert, got %d", repo.upsertCalls)
	}
	if repo.recomputeCalls != 1 {
		t.Fatalf("expected 1 recompute, got %d", repo.recomputeCalls)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "attempt.grade" {
		t.Fatalf("expected 1 audit entry with action=attempt.grade, got %+v", audit.entries)
	}
	if audit.entries[0].ResourceID != "itm" {
		t.Fatalf("expected resource_id=itm, got %q", audit.entries[0].ResourceID)
	}
	// Confirm after JSON includes awarded_score.
	var after map[string]any
	if err := json.Unmarshal(audit.entries[0].AfterJSON, &after); err != nil {
		t.Fatalf("after json: %v", err)
	}
	if after["awarded_score"] != "1.50" {
		t.Fatalf("expected awarded_score=1.50, got %v", after["awarded_score"])
	}
}

func TestListReviewQueue_RejectsNonTeacherOrAdmin(t *testing.T) {
	s, _, _ := newServiceWithFakes()
	if _, err := s.ListReviewQueue(context.Background(), studentActor(), "asm"); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestGetAttemptForReview_RejectsNonTeacherOrAdmin(t *testing.T) {
	s, _, _ := newServiceWithFakes()
	if _, err := s.GetAttemptForReview(context.Background(), studentActor(), "att"); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
