package academics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// Service is the academics application service contract.
type Service interface {
	ListTerms(ctx context.Context, orgID string) ([]Term, error)
	CreateTerm(ctx context.Context, orgID string, roles []string, req CreateTermRequest) (Term, error)
	UpdateTerm(ctx context.Context, orgID string, roles []string, termID string, req UpdateTermRequest) (Term, error)
	ArchiveTerm(ctx context.Context, orgID string, roles []string, termID string) error

	ListSubjects(ctx context.Context, orgID string) ([]Subject, error)
	CreateSubject(ctx context.Context, orgID string, roles []string, req CreateSubjectRequest) (Subject, error)
	UpdateSubject(ctx context.Context, orgID string, roles []string, subjectID string, req UpdateSubjectRequest) (Subject, error)
	ArchiveSubject(ctx context.Context, orgID string, roles []string, subjectID string) error

	ListCourses(ctx context.Context, orgID string) ([]Course, error)
	CreateCourse(ctx context.Context, orgID string, roles []string, req CreateCourseRequest) (Course, error)
	UpdateCourse(ctx context.Context, orgID string, roles []string, courseID string, req UpdateCourseRequest) (Course, error)
	ArchiveCourse(ctx context.Context, orgID string, roles []string, courseID string) error

	ListClasses(ctx context.Context, orgID string, userID string, roles []string) ([]ClassSection, error)
	ListMyTeachingClasses(ctx context.Context, orgID string, userID string) ([]ClassSection, error)
	CreateClass(ctx context.Context, orgID string, roles []string, req CreateClassRequest) (ClassSection, error)
	UpdateClass(ctx context.Context, orgID string, roles []string, classID string, req UpdateClassRequest) (ClassSection, error)
	ArchiveClass(ctx context.Context, orgID string, roles []string, classID string) error

	ListClassTeachers(ctx context.Context, orgID string, userID string, roles []string, classID string) ([]ClassTeacher, error)
	AddClassTeacher(ctx context.Context, orgID string, roles []string, classID string, req AddClassTeacherRequest) (ClassTeacher, error)
	RemoveClassTeacher(ctx context.Context, orgID string, roles []string, classID, teacherUserID string) error

	ListEnrollments(ctx context.Context, orgID string, userID string, roles []string, classID string) ([]Enrollment, error)
	EnrollStudent(ctx context.Context, orgID string, roles []string, classID string, req EnrollStudentRequest) (Enrollment, error)
	UnenrollStudent(ctx context.Context, orgID string, roles []string, classID, studentUserID string) error

	BulkEnrollStudents(ctx context.Context, orgID string, roles []string, classID string, req BulkEnrollRequest) (BulkEnrollmentResult, error)
	BulkAssignTeachers(ctx context.Context, orgID string, roles []string, classID string, req BulkAssignTeachersRequest) (BulkAssignTeachersResult, error)
}

const maxBulkItems = 100

type service struct {
	repo Repository
	tm   TransactionManager
}

// NewService creates the concrete academics service.
func NewService(repo Repository, tm TransactionManager) Service {
	return &service{repo: repo, tm: tm}
}

func isAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

func requireAdmin(roles []string) error {
	if !isAdmin(roles) {
		return ErrUnauthorized
	}
	return nil
}

func requireTeacherOrAdmin(roles []string) error {
	for _, r := range roles {
		if r == "teacher" || r == "admin" {
			return nil
		}
	}
	return ErrUnauthorized
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format %q: %w", s, err)
	}
	return t, nil
}

func (s *service) ListTerms(ctx context.Context, orgID string) ([]Term, error) {
	return s.repo.ListTerms(ctx, orgID)
}

func (s *service) CreateTerm(ctx context.Context, orgID string, roles []string, req CreateTermRequest) (Term, error) {
	if err := requireAdmin(roles); err != nil {
		return Term{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Term{}, ErrInvalidInput
	}
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return Term{}, ErrInvalidInput
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return Term{}, ErrInvalidInput
	}
	if !endDate.After(startDate) {
		return Term{}, ErrInvalidInput
	}

	var term Term
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		term, err = s.repo.CreateTerm(ctx, tx, orgID, name, startDate, endDate)
		return err
	})
	return term, err
}

func (s *service) UpdateTerm(ctx context.Context, orgID string, roles []string, termID string, req UpdateTermRequest) (Term, error) {
	if err := requireAdmin(roles); err != nil {
		return Term{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Term{}, ErrInvalidInput
	}
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return Term{}, ErrInvalidInput
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return Term{}, ErrInvalidInput
	}
	if !endDate.After(startDate) {
		return Term{}, ErrInvalidInput
	}

	var term Term
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		term, err = s.repo.UpdateTerm(ctx, tx, orgID, termID, name, startDate, endDate)
		return err
	})
	return term, err
}

func (s *service) ArchiveTerm(ctx context.Context, orgID string, roles []string, termID string) error {
	if err := requireAdmin(roles); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveTerm(ctx, tx, orgID, termID)
	})
}

func (s *service) ListSubjects(ctx context.Context, orgID string) ([]Subject, error) {
	return s.repo.ListSubjects(ctx, orgID)
}

func (s *service) CreateSubject(ctx context.Context, orgID string, roles []string, req CreateSubjectRequest) (Subject, error) {
	if err := requireAdmin(roles); err != nil {
		return Subject{}, err
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return Subject{}, ErrInvalidInput
	}

	var subject Subject
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		subject, err = s.repo.CreateSubject(ctx, tx, orgID, code, name, strings.TrimSpace(req.Description))
		if isDuplicateError(err) {
			return ErrDuplicateCode
		}
		return err
	})
	return subject, err
}

func (s *service) UpdateSubject(ctx context.Context, orgID string, roles []string, subjectID string, req UpdateSubjectRequest) (Subject, error) {
	if err := requireAdmin(roles); err != nil {
		return Subject{}, err
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return Subject{}, ErrInvalidInput
	}

	var subject Subject
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		subject, err = s.repo.UpdateSubject(ctx, tx, orgID, subjectID, code, name, strings.TrimSpace(req.Description))
		if isDuplicateError(err) {
			return ErrDuplicateCode
		}
		return err
	})
	return subject, err
}

func (s *service) ArchiveSubject(ctx context.Context, orgID string, roles []string, subjectID string) error {
	if err := requireAdmin(roles); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveSubject(ctx, tx, orgID, subjectID)
	})
}

func (s *service) ListCourses(ctx context.Context, orgID string) ([]Course, error) {
	return s.repo.ListCourses(ctx, orgID)
}

func (s *service) CreateCourse(ctx context.Context, orgID string, roles []string, req CreateCourseRequest) (Course, error) {
	if err := requireAdmin(roles); err != nil {
		return Course{}, err
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" || req.SubjectID == "" || req.AcademicTermID == "" {
		return Course{}, ErrInvalidInput
	}

	var course Course
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		course, err = s.repo.CreateCourse(ctx, tx, orgID, req.SubjectID, req.AcademicTermID, code, name)
		if isDuplicateError(err) {
			return ErrDuplicateCode
		}
		return err
	})
	return course, err
}

func (s *service) UpdateCourse(ctx context.Context, orgID string, roles []string, courseID string, req UpdateCourseRequest) (Course, error) {
	if err := requireAdmin(roles); err != nil {
		return Course{}, err
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" || req.SubjectID == "" || req.AcademicTermID == "" {
		return Course{}, ErrInvalidInput
	}

	var course Course
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		course, err = s.repo.UpdateCourse(ctx, tx, orgID, courseID, req.SubjectID, req.AcademicTermID, code, name)
		if isDuplicateError(err) {
			return ErrDuplicateCode
		}
		return err
	})
	return course, err
}

func (s *service) ArchiveCourse(ctx context.Context, orgID string, roles []string, courseID string) error {
	if err := requireAdmin(roles); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveCourse(ctx, tx, orgID, courseID)
	})
}

func (s *service) ListClasses(ctx context.Context, orgID string, userID string, roles []string) ([]ClassSection, error) {
	if err := requireTeacherOrAdmin(roles); err != nil {
		return nil, err
	}
	if isAdmin(roles) {
		return s.repo.ListClasses(ctx, orgID, "", false)
	}
	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListClasses(ctx, orgID, membership.ID, true)
}

func (s *service) ListMyTeachingClasses(ctx context.Context, orgID string, userID string) ([]ClassSection, error) {
	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if !hasRole(membership.Roles, "teacher") {
		return nil, ErrUnauthorized
	}
	return s.repo.ListClasses(ctx, orgID, membership.ID, true)
}

func (s *service) CreateClass(ctx context.Context, orgID string, roles []string, req CreateClassRequest) (ClassSection, error) {
	if err := requireAdmin(roles); err != nil {
		return ClassSection{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || req.CourseID == "" {
		return ClassSection{}, ErrInvalidInput
	}

	var class ClassSection
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		class, err = s.repo.CreateClass(ctx, tx, orgID, req.CourseID, name)
		return err
	})
	return class, err
}

func (s *service) UpdateClass(ctx context.Context, orgID string, roles []string, classID string, req UpdateClassRequest) (ClassSection, error) {
	if err := requireAdmin(roles); err != nil {
		return ClassSection{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || req.CourseID == "" {
		return ClassSection{}, ErrInvalidInput
	}

	var class ClassSection
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		class, err = s.repo.UpdateClass(ctx, tx, orgID, classID, req.CourseID, name)
		return err
	})
	return class, err
}

func (s *service) ArchiveClass(ctx context.Context, orgID string, roles []string, classID string) error {
	if err := requireAdmin(roles); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveClass(ctx, tx, orgID, classID)
	})
}

func (s *service) ListClassTeachers(ctx context.Context, orgID string, userID string, roles []string, classID string) ([]ClassTeacher, error) {
	if err := requireTeacherOrAdmin(roles); err != nil {
		return nil, err
	}
	if err := s.canAccessClass(ctx, orgID, userID, roles, classID); err != nil {
		return nil, err
	}
	return s.repo.ListClassTeachers(ctx, orgID, classID)
}

func (s *service) AddClassTeacher(ctx context.Context, orgID string, roles []string, classID string, req AddClassTeacherRequest) (ClassTeacher, error) {
	if err := requireAdmin(roles); err != nil {
		return ClassTeacher{}, err
	}
	if req.UserID == "" {
		return ClassTeacher{}, ErrInvalidInput
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = "teacher"
	}
	if role != "teacher" && role != "assistant" {
		return ClassTeacher{}, ErrInvalidInput
	}

	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, req.UserID)
	if err != nil {
		return ClassTeacher{}, err
	}

	var teacher ClassTeacher
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		teacher, err = s.repo.AddClassTeacher(ctx, tx, orgID, classID, membership.ID, role)
		if isDuplicateError(err) {
			return ErrDuplicateTeacher
		}
		return err
	})
	return teacher, err
}

func (s *service) RemoveClassTeacher(ctx context.Context, orgID string, roles []string, classID, teacherUserID string) error {
	if err := requireAdmin(roles); err != nil {
		return err
	}
	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, teacherUserID)
	if err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.RemoveClassTeacher(ctx, tx, orgID, classID, membership.ID)
	})
}

func (s *service) ListEnrollments(ctx context.Context, orgID string, userID string, roles []string, classID string) ([]Enrollment, error) {
	if err := requireTeacherOrAdmin(roles); err != nil {
		return nil, err
	}
	if err := s.canAccessClass(ctx, orgID, userID, roles, classID); err != nil {
		return nil, err
	}
	return s.repo.ListEnrollments(ctx, orgID, classID)
}

func (s *service) EnrollStudent(ctx context.Context, orgID string, roles []string, classID string, req EnrollStudentRequest) (Enrollment, error) {
	if err := requireAdmin(roles); err != nil {
		return Enrollment{}, err
	}
	if req.UserID == "" {
		return Enrollment{}, ErrInvalidInput
	}
	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, req.UserID)
	if err != nil {
		return Enrollment{}, err
	}
	if !hasRole(membership.Roles, "student") {
		return Enrollment{}, ErrInvalidInput
	}

	var enrollment Enrollment
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		enrollment, err = s.repo.EnrollStudent(ctx, tx, orgID, classID, membership.ID)
		if isDuplicateError(err) {
			return ErrDuplicateEnrollment
		}
		return err
	})
	return enrollment, err
}

func (s *service) UnenrollStudent(ctx context.Context, orgID string, roles []string, classID, studentUserID string) error {
	if err := requireAdmin(roles); err != nil {
		return err
	}
	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, studentUserID)
	if err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.UnenrollStudent(ctx, tx, orgID, classID, membership.ID)
	})
}

func (s *service) BulkEnrollStudents(ctx context.Context, orgID string, roles []string, classID string, req BulkEnrollRequest) (BulkEnrollmentResult, error) {
	if err := requireAdmin(roles); err != nil {
		return BulkEnrollmentResult{}, err
	}
	if len(req.UserIDs) == 0 || len(req.UserIDs) > maxBulkItems {
		return BulkEnrollmentResult{}, ErrInvalidInput
	}
	exists, err := s.repo.ClassExists(ctx, orgID, classID)
	if err != nil {
		return BulkEnrollmentResult{}, err
	}
	if !exists {
		return BulkEnrollmentResult{}, ErrNotFound
	}

	result := BulkEnrollmentResult{
		Total:  len(req.UserIDs),
		DryRun: req.DryRun,
		Rows:   make([]BulkEnrollmentRow, len(req.UserIDs)),
	}
	enrolledCount := 0
	failedCount := 0

	for i, userID := range req.UserIDs {
		membership, err := s.repo.GetMembershipByUserID(ctx, orgID, userID)
		var rowErr error
		switch {
		case err != nil:
			rowErr = err
		case !hasRole(membership.Roles, "student"):
			rowErr = ErrInvalidInput
		}

		if rowErr == nil && !req.DryRun {
			err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
				_, err := s.repo.EnrollStudent(ctx, tx, orgID, classID, membership.ID)
				return err
			})
			if err != nil {
				if isDuplicateError(err) {
					rowErr = ErrDuplicateEnrollment
				} else {
					rowErr = err
				}
			}
		}

		status := "enrolled"
		if rowErr != nil {
			status = "error"
			failedCount++
		} else if req.DryRun {
			status = "valid"
		} else {
			enrolledCount++
		}

		result.Rows[i] = BulkEnrollmentRow{
			UserID: userID,
			Status: status,
			Error:  errorString(rowErr),
		}
	}

	result.Enrolled = enrolledCount
	result.Failed = failedCount

	if !req.DryRun {
		after, _ := json.Marshal(map[string]any{
			"enrolled": enrolledCount,
			"failed":   failedCount,
			"total":    len(req.UserIDs),
		})
		meta, _ := json.Marshal(map[string]any{
			"class_id": classID,
			"dry_run":  false,
		})
		if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
				OrganizationID: orgID,
				ActorUserID:    "", // bulk actions do not have a single actor user
				Action:         "class.enroll_bulk",
				ResourceType:   "class",
				ResourceID:     classID,
				AfterJSON:      after,
				MetadataJSON:   meta,
			})
		}); err != nil {
			return result, fmt.Errorf("audit log: %w", err)
		}
	}

	return result, nil
}

func (s *service) BulkAssignTeachers(ctx context.Context, orgID string, roles []string, classID string, req BulkAssignTeachersRequest) (BulkAssignTeachersResult, error) {
	if err := requireAdmin(roles); err != nil {
		return BulkAssignTeachersResult{}, err
	}
	if len(req.Items) == 0 || len(req.Items) > maxBulkItems {
		return BulkAssignTeachersResult{}, ErrInvalidInput
	}
	exists, err := s.repo.ClassExists(ctx, orgID, classID)
	if err != nil {
		return BulkAssignTeachersResult{}, err
	}
	if !exists {
		return BulkAssignTeachersResult{}, ErrNotFound
	}

	result := BulkAssignTeachersResult{
		Total:  len(req.Items),
		DryRun: req.DryRun,
		Rows:   make([]BulkAssignTeacherRow, len(req.Items)),
	}
	assignedCount := 0
	failedCount := 0

	for i, item := range req.Items {
		role := strings.ToLower(strings.TrimSpace(item.Role))
		if role == "" {
			role = "teacher"
		}

		membership, err := s.repo.GetMembershipByUserID(ctx, orgID, item.UserID)
		var rowErr error
		switch {
		case item.UserID == "":
			rowErr = ErrInvalidInput
		case role != "teacher" && role != "assistant":
			rowErr = ErrInvalidInput
		case err != nil:
			rowErr = err
		case !hasRole(membership.Roles, "teacher"):
			rowErr = ErrInvalidInput
		}

		if rowErr == nil && !req.DryRun {
			err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
				_, err := s.repo.AddClassTeacher(ctx, tx, orgID, classID, membership.ID, role)
				return err
			})
			if err != nil {
				if isDuplicateError(err) {
					rowErr = ErrDuplicateTeacher
				} else {
					rowErr = err
				}
			}
		}

		status := "assigned"
		if rowErr != nil {
			status = "error"
			failedCount++
		} else if req.DryRun {
			status = "valid"
		} else {
			assignedCount++
		}

		result.Rows[i] = BulkAssignTeacherRow{
			UserID: item.UserID,
			Status: status,
			Error:  errorString(rowErr),
		}
	}

	result.Assigned = assignedCount
	result.Failed = failedCount

	if !req.DryRun {
		after, _ := json.Marshal(map[string]any{
			"assigned": assignedCount,
			"failed":   failedCount,
			"total":    len(req.Items),
		})
		meta, _ := json.Marshal(map[string]any{
			"class_id": classID,
			"dry_run":  false,
		})
		if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
				OrganizationID: orgID,
				ActorUserID:    "",
				Action:         "class.teacher_bulk",
				ResourceType:   "class",
				ResourceID:     classID,
				AfterJSON:      after,
				MetadataJSON:   meta,
			})
		}); err != nil {
			return result, fmt.Errorf("audit log: %w", err)
		}
	}

	return result, nil
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *service) canAccessClass(ctx context.Context, orgID, userID string, roles []string, classID string) error {
	if isAdmin(roles) {
		exists, err := s.repo.ClassExists(ctx, orgID, classID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
		return nil
	}
	membership, err := s.repo.GetMembershipByUserID(ctx, orgID, userID)
	if err != nil {
		return err
	}
	isTeacher, err := s.repo.IsClassTeacher(ctx, orgID, classID, membership.ID)
	if err != nil {
		return err
	}
	if !isTeacher {
		return ErrUnauthorized
	}
	return nil
}
