package attempts_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/attempts"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type fakeRepo struct {
	getAttempt   func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error)
	getItems     func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error)
	getForUpdate func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error)
	itemExists   func(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error)
	upsertAnswer func(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*attempts.AnswerSaved, error)
	markExpired  func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) error
	submit       func(ctx context.Context, tx pgx.Tx, id, orgID, userID, score, maxScore, gradingStatus string) (*attempts.GradingResult, error)
}

func (f *fakeRepo) GetAttempt(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
	return f.getAttempt(ctx, id, orgID, userID)
}

func (f *fakeRepo) GetAttemptItems(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
	return f.getItems(ctx, id, orgID)
}

func (f *fakeRepo) GetAttemptForUpdate(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
	return f.getForUpdate(ctx, tx, id, orgID, userID)
}

func (f *fakeRepo) ItemExists(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error) {
	return f.itemExists(ctx, tx, itemID, attemptID, orgID)
}

func (f *fakeRepo) UpsertAnswer(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*attempts.AnswerSaved, error) {
	return f.upsertAnswer(ctx, tx, attemptID, itemID, orgID, payload)
}

func (f *fakeRepo) MarkAttemptExpired(ctx context.Context, tx pgx.Tx, id, orgID, userID string) error {
	return f.markExpired(ctx, tx, id, orgID, userID)
}

func (f *fakeRepo) SubmitAttempt(ctx context.Context, tx pgx.Tx, id, orgID, userID, score, maxScore, gradingStatus string) (*attempts.GradingResult, error) {
	return f.submit(ctx, tx, id, orgID, userID, score, maxScore, gradingStatus)
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

func newIssuer() *auth.TokenIssuer {
	return auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
}

func addBearer(req *http.Request, issuer *auth.TokenIssuer) {
	token, _, _ := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1)
	req.Header.Set("Authorization", "Bearer "+token)
}

func addCSRF(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "demo-csrf-token"})
	req.Header.Set(csrf.HeaderName, "demo-csrf-token")
}

func TestService_GetAttempt_OK(t *testing.T) {
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{
				ID:             "attempt-id",
				OrganizationID: "org-id",
				AssessmentID:   "assessment-id",
				Status:         "IN_PROGRESS",
			}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			rev := int64(1)
			answered := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			return []attempts.AttemptItemRow{
				{
					ID:                "item-id",
					QuestionVersionID: "qv-id",
					Position:          1,
					Points:            "1.00",
					AnswerPayload:     json.RawMessage(`{"choice":"A"}`),
					Revision:          &rev,
					AnsweredAt:        &answered,
				},
			}, nil
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	snapshot, err := svc.GetAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if err != nil {
		t.Fatalf("GetAttempt failed: %v", err)
	}
	if snapshot.ID != "attempt-id" {
		t.Errorf("id = %q, want attempt-id", snapshot.ID)
	}
	if len(snapshot.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(snapshot.Items))
	}
	if snapshot.Items[0].Answer == nil {
		t.Fatal("expected answer on item")
	}
	if snapshot.Items[0].Answer.Revision != 1 {
		t.Errorf("revision = %d, want 1", snapshot.Items[0].Answer.Revision)
	}
}

func TestService_GetAttempt_NotFound(t *testing.T) {
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return nil, attempts.ErrAttemptNotFound
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	_, err := svc.GetAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "nope")
	if !errors.Is(err, attempts.ErrAttemptNotFound) {
		t.Fatalf("expected ErrAttemptNotFound, got %v", err)
	}
}

func TestService_SaveAnswer_OK(t *testing.T) {
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			expires := time.Now().Add(time.Hour)
			return &attempts.Attempt{ID: id, Status: "IN_PROGRESS", ExpiresAt: &expires}, nil
		},
		itemExists: func(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error) {
			return true, nil
		},
		upsertAnswer: func(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*attempts.AnswerSaved, error) {
			return &attempts.AnswerSaved{AttemptItemID: itemID, Revision: 2, AnswerPayload: payload, AnsweredAt: time.Now()}, nil
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	saved, err := svc.SaveAnswer(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id", "item-id", json.RawMessage(`{"choice":"B"}`))
	if err != nil {
		t.Fatalf("SaveAnswer failed: %v", err)
	}
	if saved.Revision != 2 {
		t.Errorf("revision = %d, want 2", saved.Revision)
	}
}

func TestService_SaveAnswer_Expired(t *testing.T) {
	expired := false
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			expires := time.Now().Add(-time.Hour)
			return &attempts.Attempt{ID: id, Status: "IN_PROGRESS", ExpiresAt: &expires}, nil
		},
		markExpired: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) error {
			expired = true
			return nil
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	_, err := svc.SaveAnswer(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id", "item-id", json.RawMessage(`{"choice":"B"}`))
	if !errors.Is(err, attempts.ErrAttemptExpired) {
		t.Fatalf("expected ErrAttemptExpired, got %v", err)
	}
	if !expired {
		t.Error("expected attempt to be marked expired")
	}
}

func TestService_Submit_OK(t *testing.T) {
	rev := int64(1)
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			expires := time.Now().Add(time.Hour)
			return &attempts.Attempt{ID: id, Status: "IN_PROGRESS", ExpiresAt: &expires}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{
				{
					ID:            "item-1",
					Position:      1,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"selected_option":"A"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"A"}`),
					Revision:      &rev,
				},
				{
					ID:            "item-2",
					Position:      2,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"selected_option":"C"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"B"}`),
					Revision:      &rev,
				},
			}, nil
		},
		submit: func(ctx context.Context, tx pgx.Tx, id, orgID, userID, score, maxScore, gradingStatus string) (*attempts.GradingResult, error) {
			return &attempts.GradingResult{
				SubmittedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Score:         score,
				MaxScore:      maxScore,
				GradingStatus: gradingStatus,
			}, nil
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.SubmitAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if err != nil {
		t.Fatalf("SubmitAttempt failed: %v", err)
	}
	if result.Status != "SUBMITTED" {
		t.Errorf("status = %q, want SUBMITTED", result.Status)
	}
	if result.Score != "1.00" {
		t.Errorf("score = %q, want 1.00", result.Score)
	}
	if result.MaxScore != "2.00" {
		t.Errorf("max_score = %q, want 2.00", result.MaxScore)
	}
	if result.GradingStatus != "GRADED" {
		t.Errorf("grading_status = %q, want GRADED", result.GradingStatus)
	}
}

func TestService_Submit_AlreadySubmitted(t *testing.T) {
	submitted := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	score := "1.00"
	maxScore := "2.00"
	status := "GRADED"
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: id, Status: "SUBMITTED", SubmittedAt: &submitted, Score: &score, MaxScore: &maxScore, GradingStatus: &status}, nil
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.SubmitAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if err != nil {
		t.Fatalf("SubmitAttempt failed: %v", err)
	}
	if !result.SubmittedAt.Equal(submitted) {
		t.Errorf("submitted_at = %v, want %v", result.SubmittedAt, submitted)
	}
	if result.Score != score {
		t.Errorf("score = %q, want %q", result.Score, score)
	}
}

func TestService_Submit_ZeroScore(t *testing.T) {
	rev := int64(1)
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			expires := time.Now().Add(time.Hour)
			return &attempts.Attempt{ID: id, Status: "IN_PROGRESS", ExpiresAt: &expires}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{
				{
					ID:            "item-1",
					Position:      1,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"selected_option":"C"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"A"}`),
					Revision:      &rev,
				},
			}, nil
		},
		submit: func(ctx context.Context, tx pgx.Tx, id, orgID, userID, score, maxScore, gradingStatus string) (*attempts.GradingResult, error) {
			return &attempts.GradingResult{
				SubmittedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Score:         score,
				MaxScore:      maxScore,
				GradingStatus: gradingStatus,
			}, nil
		},
	}

	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.SubmitAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if err != nil {
		t.Fatalf("SubmitAttempt failed: %v", err)
	}
	if result.Score != "0.00" {
		t.Errorf("score = %q, want 0.00", result.Score)
	}
	if result.MaxScore != "1.00" {
		t.Errorf("max_score = %q, want 1.00", result.MaxScore)
	}
}

func TestHandler_GetAttempt_MissingToken(t *testing.T) {
	svc := &fakeRepo{} // not used
	h := attempts.NewHandler(attempts.NewService(svc, stubTxManager{}), newIssuer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attempts/attempt-id", nil)
	rec := httptest.NewRecorder()

	h.GetAttempt(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetAttempt_OK(t *testing.T) {
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: id, OrganizationID: orgID, AssessmentID: "a", Status: "IN_PROGRESS"}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return nil, nil
		},
	}

	h := attempts.NewHandler(attempts.NewService(repo, stubTxManager{}), newIssuer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attempts/attempt-id", nil)
	addBearer(req, newIssuer())
	rec := httptest.NewRecorder()

	h.GetAttempt(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data attempts.AttemptSnapshot `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Status != "IN_PROGRESS" {
		t.Errorf("status = %q, want IN_PROGRESS", resp.Data.Status)
	}
}

func TestHandler_SaveAnswer_OK(t *testing.T) {
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			expires := time.Now().Add(time.Hour)
			return &attempts.Attempt{ID: id, Status: "IN_PROGRESS", ExpiresAt: &expires}, nil
		},
		itemExists: func(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error) {
			return true, nil
		},
		upsertAnswer: func(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*attempts.AnswerSaved, error) {
			return &attempts.AnswerSaved{AttemptItemID: itemID, Revision: 1, AnswerPayload: payload, AnsweredAt: time.Now()}, nil
		},
	}

	h := attempts.NewHandler(attempts.NewService(repo, stubTxManager{}), newIssuer())

	body := strings.NewReader(`{"answer_payload":{"choice":"A"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/attempts/attempt-id/answers/item-id", body)
	addBearer(req, newIssuer())
	addCSRF(req)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("attempt_id", "attempt-id")
	rctx.URLParams.Add("attempt_item_id", "item-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.SaveAnswer(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_SaveAnswer_MissingCSRF(t *testing.T) {
	h := attempts.NewHandler(attempts.NewService(&fakeRepo{}, stubTxManager{}), newIssuer())

	body := strings.NewReader(`{"answer_payload":{"choice":"A"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/attempts/attempt-id/answers/item-id", body)
	addBearer(req, newIssuer())
	rec := httptest.NewRecorder()

	h.SaveAnswer(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandler_Submit_OK(t *testing.T) {
	rev := int64(1)
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
			expires := time.Now().Add(time.Hour)
			return &attempts.Attempt{ID: id, Status: "IN_PROGRESS", ExpiresAt: &expires}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{
				{
					ID:            "item-1",
					Position:      1,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"selected_option":"A"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"A"}`),
					Revision:      &rev,
				},
			}, nil
		},
		submit: func(ctx context.Context, tx pgx.Tx, id, orgID, userID, score, maxScore, gradingStatus string) (*attempts.GradingResult, error) {
			return &attempts.GradingResult{
				SubmittedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Score:         score,
				MaxScore:      maxScore,
				GradingStatus: gradingStatus,
			}, nil
		},
	}

	h := attempts.NewHandler(attempts.NewService(repo, stubTxManager{}), newIssuer())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/attempts/attempt-id/submit", nil)
	addBearer(req, newIssuer())
	addCSRF(req)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("attempt_id", "attempt-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.SubmitAttempt(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data attempts.AttemptSubmitted `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Score != "1.00" {
		t.Errorf("score = %q, want 1.00", resp.Data.Score)
	}
	if resp.Data.GradingStatus != "GRADED" {
		t.Errorf("grading_status = %q, want GRADED", resp.Data.GradingStatus)
	}
}
