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
	getAttempt    func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error)
	getItems      func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error)
	getForUpdate  func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error)
	itemExists    func(ctx context.Context, tx pgx.Tx, itemID, attemptID, orgID string) (bool, error)
	upsertAnswer  func(ctx context.Context, tx pgx.Tx, attemptID, itemID, orgID string, payload json.RawMessage) (*attempts.AnswerSaved, error)
	markExpired   func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) error
	submit        func(ctx context.Context, tx pgx.Tx, id, orgID, userID, score, maxScore, gradingStatus string) (*attempts.GradingResult, error)
	listAssigned  func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error)
	listHistory   func(ctx context.Context, orgID, userID string, opts attempts.ListOptions) ([]attempts.StudentAttempt, *attempts.PageInfo, error)
	getLatestPub  func(ctx context.Context, orgID, assessmentID string) (*attempts.PublicationSnapshot, string, string, error)
	getInProgress func(ctx context.Context, orgID, userID, assessmentID string) (*attempts.Attempt, error)
	countAttempts func(ctx context.Context, orgID, userID, assessmentID string) (int64, error)
	createAttempt func(ctx context.Context, tx pgx.Tx, orgID, userID, assessmentID, publicationID string, startedAt, expiresAt time.Time) (*attempts.Attempt, error)
	createItems   func(ctx context.Context, tx pgx.Tx, orgID, attemptID string, items []attempts.AttemptItemInput) error
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

func (f *fakeRepo) ListAssignedAssessments(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
	if f.listAssigned != nil {
		return f.listAssigned(ctx, orgID, userID)
	}
	return nil, nil
}

func (f *fakeRepo) ListStudentAttempts(ctx context.Context, orgID, userID string, opts attempts.ListOptions) ([]attempts.StudentAttempt, *attempts.PageInfo, error) {
	if f.listHistory != nil {
		return f.listHistory(ctx, orgID, userID, opts)
	}
	return nil, nil, nil
}

func (f *fakeRepo) GetLatestPublication(ctx context.Context, orgID, assessmentID string) (*attempts.PublicationSnapshot, string, string, error) {
	if f.getLatestPub != nil {
		return f.getLatestPub(ctx, orgID, assessmentID)
	}
	return nil, "", "", nil
}

func (f *fakeRepo) GetInProgressAttempt(ctx context.Context, orgID, userID, assessmentID string) (*attempts.Attempt, error) {
	if f.getInProgress != nil {
		return f.getInProgress(ctx, orgID, userID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) CountStudentAttempts(ctx context.Context, orgID, userID, assessmentID string) (int64, error) {
	if f.countAttempts != nil {
		return f.countAttempts(ctx, orgID, userID, assessmentID)
	}
	return 0, nil
}

func (f *fakeRepo) CreateAttempt(ctx context.Context, tx pgx.Tx, orgID, userID, assessmentID, publicationID string, startedAt, expiresAt time.Time) (*attempts.Attempt, error) {
	if f.createAttempt != nil {
		return f.createAttempt(ctx, tx, orgID, userID, assessmentID, publicationID, startedAt, expiresAt)
	}
	return &attempts.Attempt{ID: "new-attempt-id", OrganizationID: orgID, AssessmentID: assessmentID, Status: "IN_PROGRESS"}, nil
}

func (f *fakeRepo) CreateAttemptItems(ctx context.Context, tx pgx.Tx, orgID, attemptID string, items []attempts.AttemptItemInput) error {
	if f.createItems != nil {
		return f.createItems(ctx, tx, orgID, attemptID, items)
	}
	return nil
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

func newIssuer() *auth.TokenIssuer {
	return auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
}

func addBearer(req *http.Request, issuer *auth.TokenIssuer) {
	token, _, _ := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, false)
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
					Prompt:            json.RawMessage(`{"text":"Demo prompt"}`),
					Choices:           json.RawMessage(`[{"id":"A","text":"One"},{"id":"B","text":"Two"}]`),
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
	if len(snapshot.Items[0].Prompt) == 0 {
		t.Error("expected prompt on item")
	}
	if len(snapshot.Items[0].Choices) == 0 {
		t.Error("expected choices on item")
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
	expires := time.Now().Add(time.Hour).UTC()
	repo := &fakeRepo{
		getForUpdate: func(ctx context.Context, tx pgx.Tx, id, orgID, userID string) (*attempts.Attempt, error) {
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
	before := time.Now().UTC().Add(-time.Second)
	saved, err := svc.SaveAnswer(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id", "item-id", json.RawMessage(`{"choice":"B"}`))
	after := time.Now().UTC().Add(time.Second)
	if err != nil {
		t.Fatalf("SaveAnswer failed: %v", err)
	}
	if saved.Revision != 2 {
		t.Errorf("revision = %d, want 2", saved.Revision)
	}
	// server_time must be populated with a recent authoritative clock value
	if saved.ServerTime.Before(before) || saved.ServerTime.After(after) {
		t.Errorf("server_time = %v, want between %v and %v", saved.ServerTime, before, after)
	}
	// expires_at must mirror the loaded attempt so the client can recalibrate
	if saved.ExpiresAt == nil {
		t.Fatal("expires_at = nil, want populated")
	}
	if !saved.ExpiresAt.Equal(expires) {
		t.Errorf("expires_at = %v, want %v", *saved.ExpiresAt, expires)
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
			return []attempts.AttemptItemRow{
				{
					ID:                "item-id",
					QuestionVersionID: "qv-id",
					Position:          1,
					Points:            "1.00",
					Prompt:            json.RawMessage(`{"text":"Handler prompt"}`),
					Choices:           json.RawMessage(`[{"id":"A","text":"Alpha"}]`),
				},
			}, nil
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
	if len(resp.Data.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(resp.Data.Items))
	}
	if string(resp.Data.Items[0].Prompt) != `{"text":"Handler prompt"}` {
		t.Errorf("prompt = %s, want handler prompt", resp.Data.Items[0].Prompt)
	}
	if len(resp.Data.Items[0].Choices) == 0 {
		t.Error("expected choices on item")
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

func TestService_ListAssignedAssessments_OK(t *testing.T) {
	repo := &fakeRepo{
		listAssigned: func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
			return []attempts.AssignedAssessment{
				{ID: "assessment-id", Title: "Quiz", Status: "OPEN", DurationMinutes: 30, MaxAttempts: 2, Revision: 1, PublicationID: "pub-id"},
			}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.ListAssignedAssessments(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}})
	if err != nil {
		t.Fatalf("ListAssignedAssessments failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 assessment, got %d", len(result))
	}
	if result[0].Title != "Quiz" {
		t.Errorf("title = %q, want Quiz", result[0].Title)
	}
}

func TestService_ListAssignedAssessments_Forbidden(t *testing.T) {
	svc := attempts.NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.ListAssignedAssessments(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}})
	if !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_StartAttempt_OK(t *testing.T) {
	created := false
	repo := &fakeRepo{
		listAssigned: func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
			return []attempts.AssignedAssessment{
				{ID: "assessment-id", Title: "Quiz", Status: "OPEN", DurationMinutes: 30, MaxAttempts: 2, Revision: 1, PublicationID: "pub-id"},
			}, nil
		},
		getInProgress: func(ctx context.Context, orgID, userID, assessmentID string) (*attempts.Attempt, error) {
			return nil, nil
		},
		countAttempts: func(ctx context.Context, orgID, userID, assessmentID string) (int64, error) {
			return 0, nil
		},
		getLatestPub: func(ctx context.Context, orgID, assessmentID string) (*attempts.PublicationSnapshot, string, string, error) {
			return &attempts.PublicationSnapshot{
				ID:              assessmentID,
				Title:           "Quiz",
				DurationMinutes: 30,
				MaxAttempts:     2,
				Sections: []attempts.PublicationSection{
					{
						ID:    "section-1",
						Title: "Section A",
						Items: []attempts.PublicationItem{
							{ID: "item-1", QuestionVersionID: "qv-1", Points: "1.00", Prompt: json.RawMessage(`{"text":"Q1"}`), Choices: json.RawMessage(`[]`), AnswerKey: json.RawMessage(`{"correct_option":"A"}`)},
						},
					},
				},
			}, "pub-id", "", nil
		},
		createAttempt: func(ctx context.Context, tx pgx.Tx, orgID, userID, assessmentID, publicationID string, startedAt, expiresAt time.Time) (*attempts.Attempt, error) {
			created = true
			return &attempts.Attempt{ID: "attempt-id", OrganizationID: orgID, AssessmentID: assessmentID, Status: "IN_PROGRESS", StartedAt: &startedAt, ExpiresAt: &expiresAt}, nil
		},
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: id, OrganizationID: orgID, AssessmentID: "assessment-id", Status: "IN_PROGRESS"}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{
				{ID: "item-1", QuestionVersionID: "qv-1", Position: 1, Points: "1.00", Prompt: json.RawMessage(`{"text":"Q1"}`), Choices: json.RawMessage(`[]`)},
			}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	snapshot, err := svc.StartAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, "assessment-id")
	if err != nil {
		t.Fatalf("StartAttempt failed: %v", err)
	}
	if !created {
		t.Error("expected attempt to be created")
	}
	if snapshot.ID != "attempt-id" {
		t.Errorf("id = %q, want attempt-id", snapshot.ID)
	}
	if len(snapshot.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(snapshot.Items))
	}
	if snapshot.Items[0].Position != 1 {
		t.Errorf("position = %d, want 1", snapshot.Items[0].Position)
	}
}

func TestService_StartAttempt_ResumeExisting(t *testing.T) {
	created := false
	repo := &fakeRepo{
		listAssigned: func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
			return []attempts.AssignedAssessment{
				{ID: "assessment-id", Title: "Quiz", Status: "OPEN", DurationMinutes: 30, MaxAttempts: 2, Revision: 1, PublicationID: "pub-id"},
			}, nil
		},
		getInProgress: func(ctx context.Context, orgID, userID, assessmentID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: "existing-id", OrganizationID: orgID, AssessmentID: assessmentID, Status: "IN_PROGRESS"}, nil
		},
		createAttempt: func(ctx context.Context, tx pgx.Tx, orgID, userID, assessmentID, publicationID string, startedAt, expiresAt time.Time) (*attempts.Attempt, error) {
			created = true
			return nil, nil
		},
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: id, OrganizationID: orgID, AssessmentID: "assessment-id", Status: "IN_PROGRESS"}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{{ID: "item-1", Position: 1, Points: "1.00"}}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	snapshot, err := svc.StartAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, "assessment-id")
	if err != nil {
		t.Fatalf("StartAttempt failed: %v", err)
	}
	if created {
		t.Error("expected no new attempt to be created")
	}
	if snapshot.ID != "existing-id" {
		t.Errorf("id = %q, want existing-id", snapshot.ID)
	}
}

func TestService_StartAttempt_LimitReached(t *testing.T) {
	repo := &fakeRepo{
		listAssigned: func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
			return []attempts.AssignedAssessment{
				{ID: "assessment-id", Title: "Quiz", Status: "OPEN", DurationMinutes: 30, MaxAttempts: 1, Revision: 1, PublicationID: "pub-id"},
			}, nil
		},
		getInProgress: func(ctx context.Context, orgID, userID, assessmentID string) (*attempts.Attempt, error) {
			return nil, nil
		},
		countAttempts: func(ctx context.Context, orgID, userID, assessmentID string) (int64, error) {
			return 1, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	_, err := svc.StartAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, "assessment-id")
	if !errors.Is(err, attempts.ErrAttemptLimitReached) {
		t.Fatalf("expected ErrAttemptLimitReached, got %v", err)
	}
}

func TestService_StartAttempt_Unavailable(t *testing.T) {
	repo := &fakeRepo{
		listAssigned: func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
			return nil, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	_, err := svc.StartAttempt(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, "assessment-id")
	if !errors.Is(err, attempts.ErrAssessmentUnavailable) {
		t.Fatalf("expected ErrAssessmentUnavailable, got %v", err)
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

func TestService_ListAssignedAssessments_Availability(t *testing.T) {
	upcoming := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	open := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	closed := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	repo := &fakeRepo{
		listAssigned: func(ctx context.Context, orgID, userID string) ([]attempts.AssignedAssessment, error) {
			return []attempts.AssignedAssessment{
				{ID: "a-upcoming", Title: "Future", Status: "OPEN", PublicationID: "pub-1", OpensAt: &upcoming},
				{ID: "a-open", Title: "Current", Status: "OPEN", PublicationID: "pub-2", OpensAt: &open},
				{ID: "a-closed", Title: "Past", Status: "PUBLISHED", PublicationID: "pub-3", ClosesAt: &closed},
			}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.ListAssignedAssessments(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}})
	if err != nil {
		t.Fatalf("ListAssignedAssessments failed: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 assessments, got %d", len(result))
	}
	want := map[string]string{"a-upcoming": "upcoming", "a-open": "open", "a-closed": "closed"}
	for _, a := range result {
		if a.Availability != want[a.ID] {
			t.Errorf("availability for %s = %q, want %q", a.ID, a.Availability, want[a.ID])
		}
	}
}

func TestService_ListAttemptHistory_OK(t *testing.T) {
	started := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		listHistory: func(ctx context.Context, orgID, userID string, opts attempts.ListOptions) ([]attempts.StudentAttempt, *attempts.PageInfo, error) {
			return []attempts.StudentAttempt{
				{ID: "attempt-1", AssessmentID: "assessment-id", AssessmentTitle: "Quiz", Status: "SUBMITTED", StartedAt: &started},
			}, &attempts.PageInfo{Limit: 10, HasMore: false}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	result, page, err := svc.ListAttemptHistory(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, attempts.ListOptions{})
	if err != nil {
		t.Fatalf("ListAttemptHistory failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 attempt, got %d", len(result))
	}
	if result[0].AssessmentTitle != "Quiz" {
		t.Errorf("title = %q, want Quiz", result[0].AssessmentTitle)
	}
	if page == nil || page.HasMore {
		t.Errorf("expected no next page, got %+v", page)
	}
}

func TestService_ListAttemptHistory_Forbidden(t *testing.T) {
	svc := attempts.NewService(&fakeRepo{}, stubTxManager{})
	_, _, err := svc.ListAttemptHistory(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}}, attempts.ListOptions{})
	if !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_GetAttemptResult_OK(t *testing.T) {
	rev := int64(1)
	answered := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	score := "1.00"
	maxScore := "2.00"
	grading := "GRADED"
	submitted := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{
				ID: id, OrganizationID: orgID, AssessmentID: "assessment-id", Status: "SUBMITTED",
				SubmittedAt: &submitted, Score: &score, MaxScore: &maxScore, GradingStatus: &grading,
			}, nil
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
					AnsweredAt:    &answered,
				},
				{
					ID:            "item-2",
					Position:      2,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"selected_option":"C"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"B"}`),
					Revision:      &rev,
					AnsweredAt:    &answered,
				},
			}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.GetAttemptResult(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if err != nil {
		t.Fatalf("GetAttemptResult failed: %v", err)
	}
	if result.Status != "SUBMITTED" {
		t.Errorf("status = %q, want SUBMITTED", result.Status)
	}
	if result.Score != "1.00" {
		t.Errorf("score = %q, want 1.00", result.Score)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].IsCorrect == nil || !*result.Items[0].IsCorrect {
		t.Error("expected first item to be correct")
	}
	if result.Items[1].IsCorrect != nil && *result.Items[1].IsCorrect {
		t.Error("expected second item to be incorrect")
	}
	if result.Items[0].StudentAnswer == nil {
		t.Fatal("expected student answer on first item")
	}
}

func TestService_GetAttemptResult_NotSubmitted(t *testing.T) {
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: id, OrganizationID: orgID, AssessmentID: "assessment-id", Status: "IN_PROGRESS"}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	_, err := svc.GetAttemptResult(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if !errors.Is(err, attempts.ErrAttemptNotSubmitted) {
		t.Fatalf("expected ErrAttemptNotSubmitted, got %v", err)
	}
}

func TestHandler_ListAttemptHistory_OK(t *testing.T) {
	started := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		listHistory: func(ctx context.Context, orgID, userID string, opts attempts.ListOptions) ([]attempts.StudentAttempt, *attempts.PageInfo, error) {
			return []attempts.StudentAttempt{{ID: "attempt-1", AssessmentID: "a", AssessmentTitle: "Quiz", Status: "SUBMITTED", StartedAt: &started}}, &attempts.PageInfo{Limit: 10, HasMore: false}, nil
		},
	}
	h := attempts.NewHandler(attempts.NewService(repo, stubTxManager{}), newIssuer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/attempts", nil)
	addBearer(req, newIssuer())
	rec := httptest.NewRecorder()

	h.ListAttemptHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_GetAttemptResult_OK(t *testing.T) {
	rev := int64(1)
	answered := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	score := "1.00"
	maxScore := "2.00"
	grading := "GRADED"
	submitted := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{
				ID: id, OrganizationID: orgID, AssessmentID: "assessment-id", Status: "SUBMITTED",
				SubmittedAt: &submitted, Score: &score, MaxScore: &maxScore, GradingStatus: &grading,
			}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{
				{
					ID:            "item-1",
					Position:      1,
					Points:        "1.00",
					Prompt:        json.RawMessage(`{"text":"Q"}`),
					Choices:       json.RawMessage(`[]`),
					AnswerPayload: json.RawMessage(`{"selected_option":"A"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"A"}`),
					Revision:      &rev,
					AnsweredAt:    &answered,
				},
			}, nil
		},
	}
	h := attempts.NewHandler(attempts.NewService(repo, stubTxManager{}), newIssuer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attempts/attempt-id/result", nil)
	addBearer(req, newIssuer())
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("attempt_id", "attempt-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetAttemptResult(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestService_Submit_EssayAlwaysPendingReview(t *testing.T) {
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
					QuestionType:  "essay",
					Position:      1,
					Points:        "2.00",
					AnswerPayload: json.RawMessage(`{"text":"My essay answer"}`),
					AnswerKey:     json.RawMessage(`{}`),
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
	if result.GradingStatus != "PENDING_REVIEW" {
		t.Errorf("grading_status = %q, want PENDING_REVIEW", result.GradingStatus)
	}
	if result.Score != "0.00" {
		t.Errorf("score = %q, want 0.00 (pending review placeholder)", result.Score)
	}
	if result.MaxScore != "2.00" {
		t.Errorf("max_score = %q, want 2.00", result.MaxScore)
	}
}

func TestService_Submit_ShortAnswerExactMatch(t *testing.T) {
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
					QuestionType:  "short_answer",
					Position:      1,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"text":"7"}`),
					AnswerKey:     json.RawMessage(`{"accepted_answers":["7","bảy"]}`),
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
	if result.GradingStatus != "GRADED" {
		t.Errorf("grading_status = %q, want GRADED", result.GradingStatus)
	}
	if result.Score != "1.00" {
		t.Errorf("score = %q, want 1.00", result.Score)
	}
}

func TestService_Submit_MixedMCQAndEssayPending(t *testing.T) {
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
					QuestionType:  "multiple_choice",
					Position:      1,
					Points:        "1.00",
					AnswerPayload: json.RawMessage(`{"selected_option":"A"}`),
					AnswerKey:     json.RawMessage(`{"correct_option":"A"}`),
					Revision:      &rev,
				},
				{
					ID:            "item-2",
					QuestionType:  "essay",
					Position:      2,
					Points:        "2.00",
					AnswerPayload: json.RawMessage(`{"text":"essay"}`),
					AnswerKey:     json.RawMessage(`{}`),
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
	if result.GradingStatus != "PENDING_REVIEW" {
		t.Errorf("grading_status = %q, want PENDING_REVIEW", result.GradingStatus)
	}
}

func TestService_GetAttemptResult_PendingItemNoIsCorrect(t *testing.T) {
	rev := int64(1)
	answered := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		getAttempt: func(ctx context.Context, id, orgID, userID string) (*attempts.Attempt, error) {
			return &attempts.Attempt{ID: id, OrganizationID: orgID, AssessmentID: "assessment-id", Status: "SUBMITTED"}, nil
		},
		getItems: func(ctx context.Context, id, orgID string) ([]attempts.AttemptItemRow, error) {
			return []attempts.AttemptItemRow{
				{
					ID:            "item-1",
					QuestionType:  "essay",
					Position:      1,
					Points:        "2.00",
					AnswerPayload: json.RawMessage(`{"text":"essay"}`),
					AnswerKey:     json.RawMessage(`{}`),
					Revision:      &rev,
					AnsweredAt:    &answered,
				},
			}, nil
		},
	}
	svc := attempts.NewService(repo, stubTxManager{})
	result, err := svc.GetAttemptResult(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id"}, "attempt-id")
	if err != nil {
		t.Fatalf("GetAttemptResult failed: %v", err)
	}
	if result.Items[0].GradingStatus != "PENDING_REVIEW" {
		t.Errorf("grading_status = %q, want PENDING_REVIEW", result.Items[0].GradingStatus)
	}
	if result.Items[0].IsCorrect != nil {
		t.Error("expected IsCorrect to be nil for PENDING_REVIEW item")
	}
}
