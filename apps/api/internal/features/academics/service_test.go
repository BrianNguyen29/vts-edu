package academics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type fakeRepository struct {
	listClassesFunc           func(ctx context.Context, orgID, membershipID string, forTeacher bool) ([]ClassSection, error)
	listEnrollmentsFunc       func(ctx context.Context, orgID, classID string) ([]Enrollment, error)
	enrollStudentFunc         func(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) (Enrollment, error)
	addClassTeacherFunc       func(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID, role string) (ClassTeacher, error)
	getMembershipByUserIDFunc func(ctx context.Context, orgID, userID string) (MembershipInfo, error)
	isClassTeacherFunc        func(ctx context.Context, orgID, classID, membershipID string) (bool, error)
	classExistsFunc           func(ctx context.Context, orgID, classID string) (bool, error)
}

func (f *fakeRepository) ListTerms(ctx context.Context, orgID string) ([]Term, error) {
	return nil, nil
}

func (f *fakeRepository) CreateTerm(ctx context.Context, tx pgx.Tx, orgID, name string, startDate, endDate time.Time) (Term, error) {
	return Term{}, nil
}

func (f *fakeRepository) UpdateTerm(ctx context.Context, tx pgx.Tx, orgID, termID, name string, startDate, endDate time.Time) (Term, error) {
	return Term{}, nil
}

func (f *fakeRepository) ArchiveTerm(ctx context.Context, tx pgx.Tx, orgID, termID string) error {
	return nil
}

func (f *fakeRepository) ListSubjects(ctx context.Context, orgID string) ([]Subject, error) {
	return nil, nil
}

func (f *fakeRepository) CreateSubject(ctx context.Context, tx pgx.Tx, orgID, code, name, description string) (Subject, error) {
	return Subject{}, nil
}

func (f *fakeRepository) UpdateSubject(ctx context.Context, tx pgx.Tx, orgID, subjectID, code, name, description string) (Subject, error) {
	return Subject{}, nil
}

func (f *fakeRepository) ArchiveSubject(ctx context.Context, tx pgx.Tx, orgID, subjectID string) error {
	return nil
}

func (f *fakeRepository) ListCourses(ctx context.Context, orgID string) ([]Course, error) {
	return nil, nil
}

func (f *fakeRepository) CreateCourse(ctx context.Context, tx pgx.Tx, orgID, subjectID, termID, code, name string) (Course, error) {
	return Course{}, nil
}

func (f *fakeRepository) UpdateCourse(ctx context.Context, tx pgx.Tx, orgID, courseID, subjectID, termID, code, name string) (Course, error) {
	return Course{}, nil
}

func (f *fakeRepository) ArchiveCourse(ctx context.Context, tx pgx.Tx, orgID, courseID string) error {
	return nil
}

func (f *fakeRepository) ListClasses(ctx context.Context, orgID, membershipID string, forTeacher bool) ([]ClassSection, error) {
	if f.listClassesFunc != nil {
		return f.listClassesFunc(ctx, orgID, membershipID, forTeacher)
	}
	return nil, nil
}

func (f *fakeRepository) CreateClass(ctx context.Context, tx pgx.Tx, orgID, courseID, name string) (ClassSection, error) {
	return ClassSection{}, nil
}

func (f *fakeRepository) UpdateClass(ctx context.Context, tx pgx.Tx, orgID, classID, courseID, name string) (ClassSection, error) {
	return ClassSection{}, nil
}

func (f *fakeRepository) ArchiveClass(ctx context.Context, tx pgx.Tx, orgID, classID string) error {
	return nil
}

func (f *fakeRepository) ListClassTeachers(ctx context.Context, orgID, classID string) ([]ClassTeacher, error) {
	return nil, nil
}

func (f *fakeRepository) AddClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID, role string) (ClassTeacher, error) {
	if f.addClassTeacherFunc != nil {
		return f.addClassTeacherFunc(ctx, tx, orgID, classID, membershipID, role)
	}
	return ClassTeacher{}, nil
}

func (f *fakeRepository) RemoveClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) error {
	return nil
}

func (f *fakeRepository) ListEnrollments(ctx context.Context, orgID, classID string) ([]Enrollment, error) {
	if f.listEnrollmentsFunc != nil {
		return f.listEnrollmentsFunc(ctx, orgID, classID)
	}
	return nil, nil
}

func (f *fakeRepository) EnrollStudent(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) (Enrollment, error) {
	if f.enrollStudentFunc != nil {
		return f.enrollStudentFunc(ctx, tx, orgID, classID, membershipID)
	}
	return Enrollment{}, nil
}

func (f *fakeRepository) UnenrollStudent(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) error {
	return nil
}

func (f *fakeRepository) GetMembershipByUserID(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
	if f.getMembershipByUserIDFunc != nil {
		return f.getMembershipByUserIDFunc(ctx, orgID, userID)
	}
	return MembershipInfo{}, nil
}

func (f *fakeRepository) IsClassTeacher(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
	if f.isClassTeacherFunc != nil {
		return f.isClassTeacherFunc(ctx, orgID, classID, membershipID)
	}
	return false, nil
}

func (f *fakeRepository) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	if f.classExistsFunc != nil {
		return f.classExistsFunc(ctx, orgID, classID)
	}
	return false, nil
}

func (f *fakeRepository) InsertAuditLog(ctx context.Context, tx pgx.Tx, p AuditLogParams) error {
	return nil
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

func TestService_ListMyTeachingClasses_Teacher(t *testing.T) {
	repo := &fakeRepository{
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "membership-1", UserID: userID, Roles: []string{"teacher"}}, nil
		},
		listClassesFunc: func(ctx context.Context, orgID, membershipID string, forTeacher bool) ([]ClassSection, error) {
			if membershipID != "membership-1" {
				t.Errorf("membershipID = %q, want membership-1", membershipID)
			}
			if !forTeacher {
				t.Error("expected forTeacher=true")
			}
			return []ClassSection{{ID: "class-1", Name: "8A1", StudentCount: 30}}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	classes, err := svc.ListMyTeachingClasses(context.Background(), "org-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(classes) != 1 || classes[0].Name != "8A1" {
		t.Errorf("classes = %+v", classes)
	}
}

func TestService_ListMyTeachingClasses_NonTeacher(t *testing.T) {
	repo := &fakeRepository{
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "membership-1", UserID: userID, Roles: []string{"student"}}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	_, err := svc.ListMyTeachingClasses(context.Background(), "org-1", "user-1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_ListMyTeachingClasses_MembershipNotFound(t *testing.T) {
	repo := &fakeRepository{
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{}, ErrUserNotFound
		},
	}

	svc := NewService(repo, stubTxManager{})
	_, err := svc.ListMyTeachingClasses(context.Background(), "org-1", "user-1")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestService_ListEnrollments_Admin(t *testing.T) {
	repo := &fakeRepository{
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "admin-membership", UserID: userID, Roles: []string{"admin"}}, nil
		},
		classExistsFunc: func(ctx context.Context, orgID, classID string) (bool, error) {
			return true, nil
		},
		listEnrollmentsFunc: func(ctx context.Context, orgID, classID string) ([]Enrollment, error) {
			return []Enrollment{{ID: "enroll-1", UserID: "student-1", DisplayName: "Student A"}}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	enrollments, err := svc.ListEnrollments(context.Background(), "org-1", "admin-user", []string{"admin"}, "class-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(enrollments) != 1 {
		t.Errorf("enrollments = %+v", enrollments)
	}
}

func TestService_ListEnrollments_TeacherAssigned(t *testing.T) {
	repo := &fakeRepository{
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "teacher-membership", UserID: userID, Roles: []string{"teacher"}}, nil
		},
		isClassTeacherFunc: func(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
			return membershipID == "teacher-membership", nil
		},
		listEnrollmentsFunc: func(ctx context.Context, orgID, classID string) ([]Enrollment, error) {
			return []Enrollment{{ID: "enroll-1", UserID: "student-1", DisplayName: "Student A"}}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	enrollments, err := svc.ListEnrollments(context.Background(), "org-1", "teacher-user", []string{"teacher"}, "class-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(enrollments) != 1 {
		t.Errorf("enrollments = %+v", enrollments)
	}
}

func TestService_ListEnrollments_TeacherNotAssigned(t *testing.T) {
	repo := &fakeRepository{
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "teacher-membership", UserID: userID, Roles: []string{"teacher"}}, nil
		},
		isClassTeacherFunc: func(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
			return false, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	_, err := svc.ListEnrollments(context.Background(), "org-1", "teacher-user", []string{"teacher"}, "class-1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_UpdateTerm_Unauthorized(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.UpdateTerm(context.Background(), "org-1", []string{"teacher"}, "term-1", UpdateTermRequest{Name: "HK1", StartDate: "2026-09-01", EndDate: "2027-01-31"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_UpdateTerm_InvalidInput(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.UpdateTerm(context.Background(), "org-1", []string{"admin"}, "term-1", UpdateTermRequest{Name: "", StartDate: "2026-09-01", EndDate: "2027-01-31"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_UpdateSubject_InvalidInput(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.UpdateSubject(context.Background(), "org-1", []string{"admin"}, "subject-1", UpdateSubjectRequest{Code: "", Name: "Math"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_UpdateCourse_InvalidInput(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.UpdateCourse(context.Background(), "org-1", []string{"admin"}, "course-1", UpdateCourseRequest{SubjectID: "subject-1", AcademicTermID: "term-1", Code: "C1", Name: ""})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_UpdateClass_InvalidInput(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.UpdateClass(context.Background(), "org-1", []string{"admin"}, "class-1", UpdateClassRequest{CourseID: "course-1", Name: ""})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_BulkEnrollStudents_DryRun(t *testing.T) {
	repo := &fakeRepository{
		classExistsFunc: func(ctx context.Context, orgID, classID string) (bool, error) {
			return true, nil
		},
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			if userID == "student-1" || userID == "student-2" {
				return MembershipInfo{ID: "m-" + userID, UserID: userID, Roles: []string{"student"}}, nil
			}
			return MembershipInfo{ID: "m-" + userID, UserID: userID, Roles: []string{"teacher"}}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	result, err := svc.BulkEnrollStudents(context.Background(), "org-1", []string{"admin"}, "class-1", BulkEnrollRequest{
		UserIDs: []string{"student-1", "student-2", "teacher-1"},
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("BulkEnrollStudents failed: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("total = %d, want 3", result.Total)
	}
	if result.Enrolled != 0 {
		t.Errorf("enrolled = %d, want 0 on dry run", result.Enrolled)
	}
	if result.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Failed)
	}
}

func TestService_BulkEnrollStudents_Confirm(t *testing.T) {
	enrolled := 0
	repo := &fakeRepository{
		classExistsFunc: func(ctx context.Context, orgID, classID string) (bool, error) {
			return true, nil
		},
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "m-" + userID, UserID: userID, Roles: []string{"student"}}, nil
		},
		enrollStudentFunc: func(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) (Enrollment, error) {
			enrolled++
			return Enrollment{ID: "e-1", UserID: "student-1"}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	result, err := svc.BulkEnrollStudents(context.Background(), "org-1", []string{"admin"}, "class-1", BulkEnrollRequest{
		UserIDs: []string{"student-1"},
		DryRun:  false,
	})
	if err != nil {
		t.Fatalf("BulkEnrollStudents failed: %v", err)
	}
	if result.Enrolled != 1 {
		t.Errorf("enrolled = %d, want 1", result.Enrolled)
	}
}

func TestService_BulkAssignTeachers_DryRun(t *testing.T) {
	repo := &fakeRepository{
		classExistsFunc: func(ctx context.Context, orgID, classID string) (bool, error) {
			return true, nil
		},
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			if userID == "teacher-1" {
				return MembershipInfo{ID: "m-" + userID, UserID: userID, Roles: []string{"teacher"}}, nil
			}
			return MembershipInfo{ID: "m-" + userID, UserID: userID, Roles: []string{"student"}}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	result, err := svc.BulkAssignTeachers(context.Background(), "org-1", []string{"admin"}, "class-1", BulkAssignTeachersRequest{
		Items: []BulkAssignTeacherItem{
			{UserID: "teacher-1", Role: "teacher"},
			{UserID: "student-1", Role: "assistant"},
		},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("BulkAssignTeachers failed: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("total = %d, want 2", result.Total)
	}
	if result.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Failed)
	}
}

func TestService_BulkAssignTeachers_Confirm(t *testing.T) {
	assigned := 0
	repo := &fakeRepository{
		classExistsFunc: func(ctx context.Context, orgID, classID string) (bool, error) {
			return true, nil
		},
		getMembershipByUserIDFunc: func(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
			return MembershipInfo{ID: "m-" + userID, UserID: userID, Roles: []string{"teacher"}}, nil
		},
		addClassTeacherFunc: func(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID, role string) (ClassTeacher, error) {
			assigned++
			return ClassTeacher{ID: "ct-1", UserID: "teacher-1", Role: role}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	result, err := svc.BulkAssignTeachers(context.Background(), "org-1", []string{"admin"}, "class-1", BulkAssignTeachersRequest{
		Items: []BulkAssignTeacherItem{
			{UserID: "teacher-1"},
		},
		DryRun: false,
	})
	if err != nil {
		t.Fatalf("BulkAssignTeachers failed: %v", err)
	}
	if result.Assigned != 1 {
		t.Errorf("assigned = %d, want 1", result.Assigned)
	}
}
