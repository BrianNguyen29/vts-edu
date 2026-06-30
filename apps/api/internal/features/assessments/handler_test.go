package assessments_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/assessments"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

type fakeService struct {
	listFunc func(ctx context.Context, orgID string, opts assessments.ListOptions) ([]assessments.AssessmentListItem, error)
}

func (f *fakeService) ListAssessments(ctx context.Context, orgID string, opts assessments.ListOptions) ([]assessments.AssessmentListItem, error) {
	return f.listFunc(ctx, orgID, opts)
}

func tokenWithRoles(roles []string) string {
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", roles, 1, false)
	if err != nil {
		panic(err)
	}
	return token
}

func TestHandler_ListAssessments_TeacherAllowed(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts assessments.ListOptions) ([]assessments.AssessmentListItem, error) {
			return []assessments.AssessmentListItem{
				{ID: "a1", Title: "Demo", Status: "PUBLISHED", DurationMinutes: 45},
			}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := assessments.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"teacher"}))
	rec := httptest.NewRecorder()

	h.ListAssessments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data []assessments.AssessmentListItem `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 assessment, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "Demo" {
		t.Errorf("title = %q, want Demo", resp.Data[0].Title)
	}
}

func TestHandler_ListAssessments_PaginationQuery(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts assessments.ListOptions) ([]assessments.AssessmentListItem, error) {
			if opts.Query != "mid" {
				t.Errorf("query = %q, want mid", opts.Query)
			}
			if opts.Limit != 2 {
				t.Errorf("limit = %d, want 2", opts.Limit)
			}
			if opts.Offset != 5 {
				t.Errorf("offset = %d, want 5", opts.Offset)
			}
			return []assessments.AssessmentListItem{
				{ID: "a2", Title: "Midterm", Status: "PUBLISHED", DurationMinutes: 60},
			}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := assessments.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments?q=mid&limit=2&offset=5", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"teacher"}))
	rec := httptest.NewRecorder()

	h.ListAssessments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data []assessments.AssessmentListItem `json:"data"`
		Page *struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		} `json:"page"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 assessment, got %d", len(resp.Data))
	}
	if resp.Page == nil || resp.Page.Limit != 2 || resp.Page.Offset != 5 {
		t.Fatalf("expected page {limit:2 offset:5}, got %+v", resp.Page)
	}
}

func TestHandler_ListAssessments_AdminAllowed(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts assessments.ListOptions) ([]assessments.AssessmentListItem, error) {
			return []assessments.AssessmentListItem{}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := assessments.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.ListAssessments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_ListAssessments_StudentForbidden(t *testing.T) {
	svc := &fakeService{}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := assessments.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"student"}))
	rec := httptest.NewRecorder()

	h.ListAssessments(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandler_ListAssessments_MissingToken(t *testing.T) {
	svc := &fakeService{}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := assessments.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments", nil)
	rec := httptest.NewRecorder()

	h.ListAssessments(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_ListAssessments_ServiceError(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts assessments.ListOptions) ([]assessments.AssessmentListItem, error) {
			return nil, errors.New("boom")
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := assessments.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"teacher"}))
	rec := httptest.NewRecorder()

	h.ListAssessments(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
