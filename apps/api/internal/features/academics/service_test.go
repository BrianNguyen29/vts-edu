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

func (f *fakeRepository) ArchiveTerm(ctx context.Context, tx pgx.Tx, orgID, termID string) error {
	return nil
}

func (f *fakeRepository) ListSubjects(ctx context.Context, orgID string) ([]Subject, error) {
	return nil, nil
}

func (f *fakeRepository) CreateSubject(ctx context.Context, tx pgx.Tx, orgID, code, name, description string) (Subject, error) {
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

func (f *fakeRepository) ArchiveClass(ctx context.Context, tx pgx.Tx, orgID, classID string) error {
	return nil
}

func (f *fakeRepository) ListClassTeachers(ctx context.Context, orgID, classID string) ([]ClassTeacher, error) {
	return nil, nil
}

func (f *fakeRepository) AddClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID, role string) (ClassTeacher, error) {
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
