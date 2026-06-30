package academics_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/academics"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

type fakeService struct {
	listMyTeachingClassesFunc func(ctx context.Context, orgID, userID string) ([]academics.ClassSection, error)
}

func (f *fakeService) ListTerms(ctx context.Context, orgID string) ([]academics.Term, error) {
	return nil, nil
}

func (f *fakeService) CreateTerm(ctx context.Context, orgID string, roles []string, req academics.CreateTermRequest) (academics.Term, error) {
	return academics.Term{}, nil
}

func (f *fakeService) UpdateTerm(ctx context.Context, orgID string, roles []string, termID string, req academics.UpdateTermRequest) (academics.Term, error) {
	return academics.Term{}, nil
}

func (f *fakeService) ArchiveTerm(ctx context.Context, orgID string, roles []string, termID string) error {
	return nil
}

func (f *fakeService) ListSubjects(ctx context.Context, orgID string) ([]academics.Subject, error) {
	return nil, nil
}

func (f *fakeService) CreateSubject(ctx context.Context, orgID string, roles []string, req academics.CreateSubjectRequest) (academics.Subject, error) {
	return academics.Subject{}, nil
}

func (f *fakeService) UpdateSubject(ctx context.Context, orgID string, roles []string, subjectID string, req academics.UpdateSubjectRequest) (academics.Subject, error) {
	return academics.Subject{}, nil
}

func (f *fakeService) ArchiveSubject(ctx context.Context, orgID string, roles []string, subjectID string) error {
	return nil
}

func (f *fakeService) ListCourses(ctx context.Context, orgID string) ([]academics.Course, error) {
	return nil, nil
}

func (f *fakeService) CreateCourse(ctx context.Context, orgID string, roles []string, req academics.CreateCourseRequest) (academics.Course, error) {
	return academics.Course{}, nil
}

func (f *fakeService) UpdateCourse(ctx context.Context, orgID string, roles []string, courseID string, req academics.UpdateCourseRequest) (academics.Course, error) {
	return academics.Course{}, nil
}

func (f *fakeService) ArchiveCourse(ctx context.Context, orgID string, roles []string, courseID string) error {
	return nil
}

func (f *fakeService) ListClasses(ctx context.Context, orgID, userID string, roles []string) ([]academics.ClassSection, error) {
	return nil, nil
}

func (f *fakeService) ListMyTeachingClasses(ctx context.Context, orgID, userID string) ([]academics.ClassSection, error) {
	if f.listMyTeachingClassesFunc != nil {
		return f.listMyTeachingClassesFunc(ctx, orgID, userID)
	}
	return nil, nil
}

func (f *fakeService) CreateClass(ctx context.Context, orgID string, roles []string, req academics.CreateClassRequest) (academics.ClassSection, error) {
	return academics.ClassSection{}, nil
}

func (f *fakeService) UpdateClass(ctx context.Context, orgID string, roles []string, classID string, req academics.UpdateClassRequest) (academics.ClassSection, error) {
	return academics.ClassSection{}, nil
}

func (f *fakeService) ArchiveClass(ctx context.Context, orgID string, roles []string, classID string) error {
	return nil
}

func (f *fakeService) ListClassTeachers(ctx context.Context, orgID, userID string, roles []string, classID string) ([]academics.ClassTeacher, error) {
	return nil, nil
}

func (f *fakeService) AddClassTeacher(ctx context.Context, orgID string, roles []string, classID string, req academics.AddClassTeacherRequest) (academics.ClassTeacher, error) {
	return academics.ClassTeacher{}, nil
}

func (f *fakeService) RemoveClassTeacher(ctx context.Context, orgID string, roles []string, classID, teacherUserID string) error {
	return nil
}

func (f *fakeService) ListEnrollments(ctx context.Context, orgID, userID string, roles []string, classID string) ([]academics.Enrollment, error) {
	return nil, nil
}

func (f *fakeService) EnrollStudent(ctx context.Context, orgID string, roles []string, classID string, req academics.EnrollStudentRequest) (academics.Enrollment, error) {
	return academics.Enrollment{}, nil
}

func (f *fakeService) UnenrollStudent(ctx context.Context, orgID string, roles []string, classID, studentUserID string) error {
	return nil
}

func tokenWithRoles(roles []string) string {
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", roles, 1, false)
	if err != nil {
		panic(err)
	}
	return token
}

func TestHandler_ListMyTeachingClasses_TeacherAllowed(t *testing.T) {
	svc := &fakeService{
		listMyTeachingClassesFunc: func(ctx context.Context, orgID, userID string) ([]academics.ClassSection, error) {
			if orgID != "org-id" {
				t.Errorf("orgID = %q, want org-id", orgID)
			}
			if userID != "user-id" {
				t.Errorf("userID = %q, want user-id", userID)
			}
			return []academics.ClassSection{
				{ID: "class-1", CourseID: "course-1", Name: "8A1", StudentCount: 30, TeacherCount: 1, Status: "ACTIVE"},
			}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := academics.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/teaching/classes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"teacher"}))
	rec := httptest.NewRecorder()

	h.ListMyTeachingClasses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data []academics.ClassSection `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 class, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "8A1" || resp.Data[0].StudentCount != 30 {
		t.Errorf("class = %+v, want 8A1 with 30 students", resp.Data[0])
	}
}

func TestHandler_ListMyTeachingClasses_AdminForbidden(t *testing.T) {
	svc := &fakeService{}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := academics.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/teaching/classes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.ListMyTeachingClasses(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandler_ListMyTeachingClasses_StudentForbidden(t *testing.T) {
	svc := &fakeService{}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := academics.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/teaching/classes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"student"}))
	rec := httptest.NewRecorder()

	h.ListMyTeachingClasses(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandler_ListMyTeachingClasses_MissingToken(t *testing.T) {
	svc := &fakeService{}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := academics.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/teaching/classes", nil)
	rec := httptest.NewRecorder()

	h.ListMyTeachingClasses(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_ListMyTeachingClasses_ServiceError(t *testing.T) {
	svc := &fakeService{
		listMyTeachingClassesFunc: func(ctx context.Context, orgID, userID string) ([]academics.ClassSection, error) {
			return nil, errors.New("boom")
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := academics.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/teaching/classes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"teacher"}))
	rec := httptest.NewRecorder()

	h.ListMyTeachingClasses(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
