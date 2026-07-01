package gradebook_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/gradebook"
	"github.com/go-chi/chi/v5"
)

type fakeRepo struct {
	assessmentExists            func(ctx context.Context, orgID, assessmentID string) (bool, error)
	listAssessmentAttempts      func(ctx context.Context, orgID, assessmentID string) ([]gradebook.AssessmentAttempt, error)
	getAssessmentResults        func(ctx context.Context, orgID, assessmentID string) (*gradebook.AssessmentResult, error)
	isAssessmentTaughtByTeacher func(ctx context.Context, orgID, assessmentID, userID string) (bool, error)
	getClassGradebook           func(ctx context.Context, orgID, classID string) ([]gradebook.ClassGradebookEntry, error)
}

func (f *fakeRepo) AssessmentExists(ctx context.Context, orgID, assessmentID string) (bool, error) {
	if f.assessmentExists != nil {
		return f.assessmentExists(ctx, orgID, assessmentID)
	}
	return false, nil
}

func (f *fakeRepo) ListAssessmentAttempts(ctx context.Context, orgID, assessmentID string) ([]gradebook.AssessmentAttempt, error) {
	if f.listAssessmentAttempts != nil {
		return f.listAssessmentAttempts(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) GetAssessmentResults(ctx context.Context, orgID, assessmentID string) (*gradebook.AssessmentResult, error) {
	if f.getAssessmentResults != nil {
		return f.getAssessmentResults(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) IsAssessmentTaughtByTeacher(ctx context.Context, orgID, assessmentID, userID string) (bool, error) {
	if f.isAssessmentTaughtByTeacher != nil {
		return f.isAssessmentTaughtByTeacher(ctx, orgID, assessmentID, userID)
	}
	return false, nil
}

func (f *fakeRepo) GetClassGradebook(ctx context.Context, orgID, classID string) ([]gradebook.ClassGradebookEntry, error) {
	if f.getClassGradebook != nil {
		return f.getClassGradebook(ctx, orgID, classID)
	}
	return nil, nil
}

type fakeAccess struct {
	classExists           func(ctx context.Context, orgID, classID string) (bool, error)
	getMembershipByUserID func(ctx context.Context, orgID, userID string) (gradebook.MembershipInfo, error)
	isClassTeacher        func(ctx context.Context, orgID, classID, membershipID string) (bool, error)
}

func (f *fakeAccess) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	if f.classExists != nil {
		return f.classExists(ctx, orgID, classID)
	}
	return false, nil
}

func (f *fakeAccess) GetMembershipByUserID(ctx context.Context, orgID, userID string) (gradebook.MembershipInfo, error) {
	if f.getMembershipByUserID != nil {
		return f.getMembershipByUserID(ctx, orgID, userID)
	}
	return gradebook.MembershipInfo{}, errors.New("membership not found")
}

func (f *fakeAccess) IsClassTeacher(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
	if f.isClassTeacher != nil {
		return f.isClassTeacher(ctx, orgID, classID, membershipID)
	}
	return false, nil
}

func newIssuer() *auth.TokenIssuer {
	return auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
}

func addBearer(req *http.Request, issuer *auth.TokenIssuer, roles ...string) {
	if len(roles) == 0 {
		roles = []string{"teacher"}
	}
	token, _, _ := issuer.IssueAccessToken("user-id", "org-id", "session-id", roles, 1, false)
	req.Header.Set("Authorization", "Bearer "+token)
}

func TestService_ListAssessmentAttempts_TeacherAllowed(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentTaughtByTeacher: func(ctx context.Context, orgID, assessmentID, userID string) (bool, error) {
			return true, nil
		},
		listAssessmentAttempts: func(ctx context.Context, orgID, assessmentID string) ([]gradebook.AssessmentAttempt, error) {
			return []gradebook.AssessmentAttempt{{ID: "attempt-id", AssessmentID: assessmentID, StudentUserID: "student-id", Status: "SUBMITTED"}}, nil
		},
	}
	svc := gradebook.NewService(repo, &fakeAccess{})
	result, err := svc.ListAssessmentAttempts(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}}, "assessment-id")
	if err != nil {
		t.Fatalf("ListAssessmentAttempts failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 attempt, got %d", len(result))
	}
}

func TestService_ListAssessmentAttempts_StudentForbidden(t *testing.T) {
	svc := gradebook.NewService(&fakeRepo{}, &fakeAccess{})
	_, err := svc.ListAssessmentAttempts(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, "assessment-id")
	if !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_ListAssessmentAttempts_NonTeachingTeacherForbidden(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentTaughtByTeacher: func(ctx context.Context, orgID, assessmentID, userID string) (bool, error) {
			return false, nil
		},
	}
	svc := gradebook.NewService(repo, &fakeAccess{})
	_, err := svc.ListAssessmentAttempts(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}}, "assessment-id")
	if !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_GetAssessmentResults_AdminAllowed(t *testing.T) {
	repo := &fakeRepo{
		assessmentExists: func(ctx context.Context, orgID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentResults: func(ctx context.Context, orgID, assessmentID string) (*gradebook.AssessmentResult, error) {
			return &gradebook.AssessmentResult{AssessmentID: assessmentID, TotalAttempts: 5, SubmittedCount: 4}, nil
		},
	}
	svc := gradebook.NewService(repo, &fakeAccess{})
	result, err := svc.GetAssessmentResults(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"admin"}}, "assessment-id")
	if err != nil {
		t.Fatalf("GetAssessmentResults failed: %v", err)
	}
	if result.TotalAttempts != 5 {
		t.Errorf("total_attempts = %d, want 5", result.TotalAttempts)
	}
}

func TestService_ExportAssessmentAttemptsCSV(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentTaughtByTeacher: func(ctx context.Context, orgID, assessmentID, userID string) (bool, error) {
			return true, nil
		},
		listAssessmentAttempts: func(ctx context.Context, orgID, assessmentID string) ([]gradebook.AssessmentAttempt, error) {
			return []gradebook.AssessmentAttempt{{ID: "attempt-id", AssessmentID: assessmentID, StudentUserID: "student-id", Status: "SUBMITTED"}}, nil
		},
	}
	svc := gradebook.NewService(repo, &fakeAccess{})
	csv, err := svc.ExportAssessmentAttemptsCSV(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}}, "assessment-id")
	if err != nil {
		t.Fatalf("ExportAssessmentAttemptsCSV failed: %v", err)
	}
	if len(csv) == 0 {
		t.Fatal("expected non-empty csv")
	}
}

func TestService_GetClassGradebook_TeacherAllowed(t *testing.T) {
	access := &fakeAccess{
		getMembershipByUserID: func(ctx context.Context, orgID, userID string) (gradebook.MembershipInfo, error) {
			return gradebook.MembershipInfo{ID: "membership-id"}, nil
		},
		isClassTeacher: func(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
			return true, nil
		},
	}
	repo := &fakeRepo{
		getClassGradebook: func(ctx context.Context, orgID, classID string) ([]gradebook.ClassGradebookEntry, error) {
			return []gradebook.ClassGradebookEntry{{StudentUserID: "student-id", StudentName: "Student", AssessmentID: "a", AssessmentTitle: "Quiz"}}, nil
		},
	}
	svc := gradebook.NewService(repo, access)
	result, err := svc.GetClassGradebook(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}}, "class-id")
	if err != nil {
		t.Fatalf("GetClassGradebook failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
}

func TestService_GetClassGradebook_StudentForbidden(t *testing.T) {
	svc := gradebook.NewService(&fakeRepo{}, &fakeAccess{})
	_, err := svc.GetClassGradebook(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"student"}}, "class-id")
	if !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_ExportClassGradebookCSV(t *testing.T) {
	access := &fakeAccess{
		getMembershipByUserID: func(ctx context.Context, orgID, userID string) (gradebook.MembershipInfo, error) {
			return gradebook.MembershipInfo{ID: "membership-id"}, nil
		},
		isClassTeacher: func(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
			return true, nil
		},
	}
	repo := &fakeRepo{
		getClassGradebook: func(ctx context.Context, orgID, classID string) ([]gradebook.ClassGradebookEntry, error) {
			return []gradebook.ClassGradebookEntry{{StudentUserID: "student-id", StudentName: "Student", AssessmentID: "a", AssessmentTitle: "Quiz"}}, nil
		},
	}
	svc := gradebook.NewService(repo, access)
	csv, err := svc.ExportClassGradebookCSV(context.Background(), auth.Actor{UserID: "user-id", OrgID: "org-id", Roles: []string{"teacher"}}, "class-id")
	if err != nil {
		t.Fatalf("ExportClassGradebookCSV failed: %v", err)
	}
	if len(csv) == 0 {
		t.Fatal("expected non-empty csv")
	}
}

func TestHandler_ListAssessmentAttempts_OK(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentTaughtByTeacher: func(ctx context.Context, orgID, assessmentID, userID string) (bool, error) {
			return true, nil
		},
		listAssessmentAttempts: func(ctx context.Context, orgID, assessmentID string) ([]gradebook.AssessmentAttempt, error) {
			return []gradebook.AssessmentAttempt{{ID: "attempt-id", AssessmentID: assessmentID, StudentUserID: "student-id", Status: "SUBMITTED"}}, nil
		},
	}
	h := gradebook.NewHandler(gradebook.NewService(repo, &fakeAccess{}), newIssuer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assessments/assessment-id/attempts", nil)
	addBearer(req, newIssuer(), "teacher")
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "assessment-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.ListAssessmentAttempts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_GetClassGradebook_OK(t *testing.T) {
	access := &fakeAccess{
		getMembershipByUserID: func(ctx context.Context, orgID, userID string) (gradebook.MembershipInfo, error) {
			return gradebook.MembershipInfo{ID: "membership-id"}, nil
		},
		isClassTeacher: func(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
			return true, nil
		},
	}
	repo := &fakeRepo{
		getClassGradebook: func(ctx context.Context, orgID, classID string) ([]gradebook.ClassGradebookEntry, error) {
			return []gradebook.ClassGradebookEntry{{StudentUserID: "student-id", StudentName: "Student", AssessmentID: "a", AssessmentTitle: "Quiz"}}, nil
		},
	}
	h := gradebook.NewHandler(gradebook.NewService(repo, access), newIssuer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes/class-id/gradebook", nil)
	addBearer(req, newIssuer(), "teacher")
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("class_id", "class-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetClassGradebook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
