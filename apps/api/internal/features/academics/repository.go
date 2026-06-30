package academics

import (
	"context"
	"errors"
	"fmt"
	"time"

	academicssqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/academics/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the academics feature.
type Repository interface {
	ListTerms(ctx context.Context, orgID string) ([]Term, error)
	CreateTerm(ctx context.Context, tx pgx.Tx, orgID, name string, startDate, endDate time.Time) (Term, error)
	UpdateTerm(ctx context.Context, tx pgx.Tx, orgID, termID, name string, startDate, endDate time.Time) (Term, error)
	ArchiveTerm(ctx context.Context, tx pgx.Tx, orgID, termID string) error

	ListSubjects(ctx context.Context, orgID string) ([]Subject, error)
	CreateSubject(ctx context.Context, tx pgx.Tx, orgID, code, name, description string) (Subject, error)
	UpdateSubject(ctx context.Context, tx pgx.Tx, orgID, subjectID, code, name, description string) (Subject, error)
	ArchiveSubject(ctx context.Context, tx pgx.Tx, orgID, subjectID string) error

	ListCourses(ctx context.Context, orgID string) ([]Course, error)
	CreateCourse(ctx context.Context, tx pgx.Tx, orgID, subjectID, termID, code, name string) (Course, error)
	UpdateCourse(ctx context.Context, tx pgx.Tx, orgID, courseID, subjectID, termID, code, name string) (Course, error)
	ArchiveCourse(ctx context.Context, tx pgx.Tx, orgID, courseID string) error

	ListClasses(ctx context.Context, orgID, membershipID string, forTeacher bool) ([]ClassSection, error)
	CreateClass(ctx context.Context, tx pgx.Tx, orgID, courseID, name string) (ClassSection, error)
	UpdateClass(ctx context.Context, tx pgx.Tx, orgID, classID, courseID, name string) (ClassSection, error)
	ArchiveClass(ctx context.Context, tx pgx.Tx, orgID, classID string) error

	ListClassTeachers(ctx context.Context, orgID, classID string) ([]ClassTeacher, error)
	AddClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID, role string) (ClassTeacher, error)
	RemoveClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) error

	ListEnrollments(ctx context.Context, orgID, classID string) ([]Enrollment, error)
	EnrollStudent(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) (Enrollment, error)
	UnenrollStudent(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) error

	GetMembershipByUserID(ctx context.Context, orgID, userID string) (MembershipInfo, error)
	IsClassTeacher(ctx context.Context, orgID, classID, membershipID string) (bool, error)
	ClassExists(ctx context.Context, orgID, classID string) (bool, error)
}

type sqlcRepository struct {
	queries *academicssqlc.Queries
}

// NewRepository creates a new academics repository backed by generated sqlc queries.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: academicssqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return []string{}
	}
	arr, ok := v.([]interface{})
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func toDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func (r *sqlcRepository) ListTerms(ctx context.Context, orgID string) ([]Term, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := r.queries.ListTerms(ctx, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("list terms: %w", err)
	}
	terms := make([]Term, len(rows))
	for i, row := range rows {
		terms[i] = Term{
			ID:        row.ID.String(),
			Name:      row.Name,
			StartDate: formatDate(row.StartDate.Time),
			EndDate:   formatDate(row.EndDate.Time),
			Status:    row.Status,
		}
	}
	return terms, nil
}

func (r *sqlcRepository) CreateTerm(ctx context.Context, tx pgx.Tx, orgID, name string, startDate, endDate time.Time) (Term, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Term{}, fmt.Errorf("invalid organization id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateTerm(ctx, academicssqlc.CreateTermParams{
		OrganizationID: orgUUID,
		Name:           name,
		StartDate:      toDate(startDate),
		EndDate:        toDate(endDate),
	})
	if err != nil {
		return Term{}, fmt.Errorf("create term: %w", err)
	}
	return Term{
		ID:        row.ID.String(),
		Name:      row.Name,
		StartDate: formatDate(row.StartDate.Time),
		EndDate:   formatDate(row.EndDate.Time),
		Status:    row.Status,
	}, nil
}

func (r *sqlcRepository) UpdateTerm(ctx context.Context, tx pgx.Tx, orgID, termID, name string, startDate, endDate time.Time) (Term, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Term{}, fmt.Errorf("invalid organization id: %w", err)
	}
	termUUID, err := toUUID(termID)
	if err != nil {
		return Term{}, fmt.Errorf("invalid term id: %w", err)
	}
	row, err := r.queries.WithTx(tx).UpdateTerm(ctx, academicssqlc.UpdateTermParams{
		ID:             termUUID,
		OrganizationID: orgUUID,
		Name:           name,
		StartDate:      toDate(startDate),
		EndDate:        toDate(endDate),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Term{}, ErrNotFound
	}
	if err != nil {
		return Term{}, fmt.Errorf("update term: %w", err)
	}
	return Term{
		ID:        row.ID.String(),
		Name:      row.Name,
		StartDate: formatDate(row.StartDate.Time),
		EndDate:   formatDate(row.EndDate.Time),
		Status:    row.Status,
	}, nil
}

func (r *sqlcRepository) ArchiveTerm(ctx context.Context, tx pgx.Tx, orgID, termID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	termUUID, err := toUUID(termID)
	if err != nil {
		return fmt.Errorf("invalid term id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveTerm(ctx, academicssqlc.ArchiveTermParams{
		ID:             termUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive term: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) ListSubjects(ctx context.Context, orgID string) ([]Subject, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := r.queries.ListSubjects(ctx, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	subjects := make([]Subject, len(rows))
	for i, row := range rows {
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		subjects[i] = Subject{
			ID:          row.ID.String(),
			Code:        row.Code,
			Name:        row.Name,
			Description: desc,
			Status:      row.Status,
		}
	}
	return subjects, nil
}

func (r *sqlcRepository) CreateSubject(ctx context.Context, tx pgx.Tx, orgID, code, name, description string) (Subject, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Subject{}, fmt.Errorf("invalid organization id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateSubject(ctx, academicssqlc.CreateSubjectParams{
		OrganizationID: orgUUID,
		Code:           code,
		Name:           name,
		Description:    pgtype.Text{String: description, Valid: description != ""},
	})
	if err != nil {
		return Subject{}, fmt.Errorf("create subject: %w", err)
	}
	var desc string
	if row.Description.Valid {
		desc = row.Description.String
	}
	return Subject{
		ID:          row.ID.String(),
		Code:        row.Code,
		Name:        row.Name,
		Description: desc,
		Status:      row.Status,
	}, nil
}

func (r *sqlcRepository) UpdateSubject(ctx context.Context, tx pgx.Tx, orgID, subjectID, code, name, description string) (Subject, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Subject{}, fmt.Errorf("invalid organization id: %w", err)
	}
	subjectUUID, err := toUUID(subjectID)
	if err != nil {
		return Subject{}, fmt.Errorf("invalid subject id: %w", err)
	}
	row, err := r.queries.WithTx(tx).UpdateSubject(ctx, academicssqlc.UpdateSubjectParams{
		ID:             subjectUUID,
		OrganizationID: orgUUID,
		Code:           code,
		Name:           name,
		Description:    pgtype.Text{String: description, Valid: description != ""},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Subject{}, ErrNotFound
	}
	if err != nil {
		return Subject{}, fmt.Errorf("update subject: %w", err)
	}
	var desc string
	if row.Description.Valid {
		desc = row.Description.String
	}
	return Subject{
		ID:          row.ID.String(),
		Code:        row.Code,
		Name:        row.Name,
		Description: desc,
		Status:      row.Status,
	}, nil
}

func (r *sqlcRepository) ArchiveSubject(ctx context.Context, tx pgx.Tx, orgID, subjectID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	subjectUUID, err := toUUID(subjectID)
	if err != nil {
		return fmt.Errorf("invalid subject id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveSubject(ctx, academicssqlc.ArchiveSubjectParams{
		ID:             subjectUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive subject: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) ListCourses(ctx context.Context, orgID string) ([]Course, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := r.queries.ListCourses(ctx, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	courses := make([]Course, len(rows))
	for i, row := range rows {
		courses[i] = Course{
			ID:             row.ID.String(),
			SubjectID:      row.SubjectID.String(),
			AcademicTermID: row.AcademicTermID.String(),
			Code:           row.Code,
			Name:           row.Name,
			Status:         row.Status,
		}
	}
	return courses, nil
}

func (r *sqlcRepository) CreateCourse(ctx context.Context, tx pgx.Tx, orgID, subjectID, termID, code, name string) (Course, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid organization id: %w", err)
	}
	subjectUUID, err := toUUID(subjectID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid subject id: %w", err)
	}
	termUUID, err := toUUID(termID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid term id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateCourse(ctx, academicssqlc.CreateCourseParams{
		OrganizationID: orgUUID,
		SubjectID:      subjectUUID,
		AcademicTermID: termUUID,
		Code:           code,
		Name:           name,
	})
	if err != nil {
		return Course{}, fmt.Errorf("create course: %w", err)
	}
	return Course{
		ID:             row.ID.String(),
		SubjectID:      row.SubjectID.String(),
		AcademicTermID: row.AcademicTermID.String(),
		Code:           row.Code,
		Name:           row.Name,
		Status:         row.Status,
	}, nil
}

func (r *sqlcRepository) UpdateCourse(ctx context.Context, tx pgx.Tx, orgID, courseID, subjectID, termID, code, name string) (Course, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid organization id: %w", err)
	}
	courseUUID, err := toUUID(courseID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid course id: %w", err)
	}
	subjectUUID, err := toUUID(subjectID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid subject id: %w", err)
	}
	termUUID, err := toUUID(termID)
	if err != nil {
		return Course{}, fmt.Errorf("invalid term id: %w", err)
	}
	row, err := r.queries.WithTx(tx).UpdateCourse(ctx, academicssqlc.UpdateCourseParams{
		ID:             courseUUID,
		OrganizationID: orgUUID,
		SubjectID:      subjectUUID,
		AcademicTermID: termUUID,
		Code:           code,
		Name:           name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Course{}, ErrNotFound
	}
	if err != nil {
		return Course{}, fmt.Errorf("update course: %w", err)
	}
	return Course{
		ID:             row.ID.String(),
		SubjectID:      row.SubjectID.String(),
		AcademicTermID: row.AcademicTermID.String(),
		Code:           row.Code,
		Name:           row.Name,
		Status:         row.Status,
	}, nil
}

func (r *sqlcRepository) ArchiveCourse(ctx context.Context, tx pgx.Tx, orgID, courseID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	courseUUID, err := toUUID(courseID)
	if err != nil {
		return fmt.Errorf("invalid course id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveCourse(ctx, academicssqlc.ArchiveCourseParams{
		ID:             courseUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive course: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) ListClasses(ctx context.Context, orgID, membershipID string, forTeacher bool) ([]ClassSection, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	var rows []academicssqlc.ListClassesAdminRow
	if forTeacher {
		membershipUUID, err := toUUID(membershipID)
		if err != nil {
			return nil, fmt.Errorf("invalid membership id: %w", err)
		}
		teacherRows, err := r.queries.ListClassesForTeacher(ctx, academicssqlc.ListClassesForTeacherParams{
			OrganizationID: orgUUID,
			MembershipID:   membershipUUID,
		})
		if err != nil {
			return nil, fmt.Errorf("list classes for teacher: %w", err)
		}
		classes := make([]ClassSection, len(teacherRows))
		for i, row := range teacherRows {
			classes[i] = ClassSection{
				ID:           row.ID.String(),
				CourseID:     row.CourseID.String(),
				Name:         row.Name,
				StudentCount: row.StudentCount,
				TeacherCount: row.TeacherCount,
				Status:       row.Status,
			}
		}
		return classes, nil
	}
	rows, err = r.queries.ListClassesAdmin(ctx, orgUUID)
	if err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	classes := make([]ClassSection, len(rows))
	for i, row := range rows {
		classes[i] = ClassSection{
			ID:           row.ID.String(),
			CourseID:     row.CourseID.String(),
			Name:         row.Name,
			StudentCount: row.StudentCount,
			TeacherCount: row.TeacherCount,
			Status:       row.Status,
		}
	}
	return classes, nil
}

func (r *sqlcRepository) CreateClass(ctx context.Context, tx pgx.Tx, orgID, courseID, name string) (ClassSection, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ClassSection{}, fmt.Errorf("invalid organization id: %w", err)
	}
	courseUUID, err := toUUID(courseID)
	if err != nil {
		return ClassSection{}, fmt.Errorf("invalid course id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateClass(ctx, academicssqlc.CreateClassParams{
		OrganizationID: orgUUID,
		CourseID:       courseUUID,
		Name:           name,
	})
	if err != nil {
		return ClassSection{}, fmt.Errorf("create class: %w", err)
	}
	return ClassSection{
		ID:           row.ID.String(),
		CourseID:     row.CourseID.String(),
		Name:         row.Name,
		StudentCount: row.StudentCount,
		TeacherCount: row.TeacherCount,
		Status:       row.Status,
	}, nil
}

func (r *sqlcRepository) UpdateClass(ctx context.Context, tx pgx.Tx, orgID, classID, courseID, name string) (ClassSection, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ClassSection{}, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return ClassSection{}, fmt.Errorf("invalid class id: %w", err)
	}
	courseUUID, err := toUUID(courseID)
	if err != nil {
		return ClassSection{}, fmt.Errorf("invalid course id: %w", err)
	}
	row, err := r.queries.WithTx(tx).UpdateClass(ctx, academicssqlc.UpdateClassParams{
		ID:             classUUID,
		OrganizationID: orgUUID,
		CourseID:       courseUUID,
		Name:           name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ClassSection{}, ErrNotFound
	}
	if err != nil {
		return ClassSection{}, fmt.Errorf("update class: %w", err)
	}
	return ClassSection{
		ID:           row.ID.String(),
		CourseID:     row.CourseID.String(),
		Name:         row.Name,
		StudentCount: row.StudentCount,
		TeacherCount: row.TeacherCount,
		Status:       row.Status,
	}, nil
}

func (r *sqlcRepository) ArchiveClass(ctx context.Context, tx pgx.Tx, orgID, classID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return fmt.Errorf("invalid class id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveClass(ctx, academicssqlc.ArchiveClassParams{
		ID:             classUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive class: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) ListClassTeachers(ctx context.Context, orgID, classID string) ([]ClassTeacher, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return nil, fmt.Errorf("invalid class id: %w", err)
	}
	rows, err := r.queries.ListClassTeachers(ctx, academicssqlc.ListClassTeachersParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list class teachers: %w", err)
	}
	teachers := make([]ClassTeacher, len(rows))
	for i, row := range rows {
		var displayName string
		if row.DisplayName.Valid {
			displayName = row.DisplayName.String
		}
		teachers[i] = ClassTeacher{
			ID:          row.ID.String(),
			UserID:      row.UserID.String(),
			DisplayName: displayName,
			Role:        row.Role,
			Status:      row.Status,
		}
	}
	return teachers, nil
}

func (r *sqlcRepository) AddClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID, role string) (ClassTeacher, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ClassTeacher{}, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return ClassTeacher{}, fmt.Errorf("invalid class id: %w", err)
	}
	membershipUUID, err := toUUID(membershipID)
	if err != nil {
		return ClassTeacher{}, fmt.Errorf("invalid membership id: %w", err)
	}
	row, err := r.queries.WithTx(tx).AddClassTeacher(ctx, academicssqlc.AddClassTeacherParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
		MembershipID:   membershipUUID,
		Role:           role,
	})
	if err != nil {
		return ClassTeacher{}, fmt.Errorf("add class teacher: %w", err)
	}
	var displayName string
	if row.DisplayName.Valid {
		displayName = row.DisplayName.String
	}
	return ClassTeacher{
		ID:          row.ID.String(),
		UserID:      row.UserID.String(),
		DisplayName: displayName,
		Role:        row.Role,
		Status:      row.Status,
	}, nil
}

func (r *sqlcRepository) RemoveClassTeacher(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return fmt.Errorf("invalid class id: %w", err)
	}
	membershipUUID, err := toUUID(membershipID)
	if err != nil {
		return fmt.Errorf("invalid membership id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).RemoveClassTeacher(ctx, academicssqlc.RemoveClassTeacherParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
		MembershipID:   membershipUUID,
	})
	if err != nil {
		return fmt.Errorf("remove class teacher: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) ListEnrollments(ctx context.Context, orgID, classID string) ([]Enrollment, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return nil, fmt.Errorf("invalid class id: %w", err)
	}
	rows, err := r.queries.ListEnrollments(ctx, academicssqlc.ListEnrollmentsParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	enrollments := make([]Enrollment, len(rows))
	for i, row := range rows {
		var displayName string
		if row.DisplayName.Valid {
			displayName = row.DisplayName.String
		}
		enrollments[i] = Enrollment{
			ID:          row.ID.String(),
			UserID:      row.UserID.String(),
			DisplayName: displayName,
			Status:      row.Status,
		}
	}
	return enrollments, nil
}

func (r *sqlcRepository) EnrollStudent(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) (Enrollment, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Enrollment{}, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return Enrollment{}, fmt.Errorf("invalid class id: %w", err)
	}
	membershipUUID, err := toUUID(membershipID)
	if err != nil {
		return Enrollment{}, fmt.Errorf("invalid membership id: %w", err)
	}
	row, err := r.queries.WithTx(tx).EnrollStudent(ctx, academicssqlc.EnrollStudentParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
		MembershipID:   membershipUUID,
	})
	if err != nil {
		return Enrollment{}, fmt.Errorf("enroll student: %w", err)
	}
	var displayName string
	if row.DisplayName.Valid {
		displayName = row.DisplayName.String
	}
	return Enrollment{
		ID:          row.ID.String(),
		UserID:      row.UserID.String(),
		DisplayName: displayName,
		Status:      row.Status,
	}, nil
}

func (r *sqlcRepository) UnenrollStudent(ctx context.Context, tx pgx.Tx, orgID, classID, membershipID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return fmt.Errorf("invalid class id: %w", err)
	}
	membershipUUID, err := toUUID(membershipID)
	if err != nil {
		return fmt.Errorf("invalid membership id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).UnenrollStudent(ctx, academicssqlc.UnenrollStudentParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
		MembershipID:   membershipUUID,
	})
	if err != nil {
		return fmt.Errorf("unenroll student: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) GetMembershipByUserID(ctx context.Context, orgID, userID string) (MembershipInfo, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return MembershipInfo{}, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return MembershipInfo{}, fmt.Errorf("invalid user id: %w", err)
	}
	row, err := r.queries.GetMembershipByUserID(ctx, academicssqlc.GetMembershipByUserIDParams{
		OrganizationID: orgUUID,
		UserID:         userUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipInfo{}, ErrUserNotFound
	}
	if err != nil {
		return MembershipInfo{}, fmt.Errorf("get membership: %w", err)
	}
	return MembershipInfo{
		ID:     row.ID.String(),
		UserID: row.UserID.String(),
		Roles:  toStringSlice(row.ArrayAgg),
	}, nil
}

func (r *sqlcRepository) IsClassTeacher(ctx context.Context, orgID, classID, membershipID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return false, fmt.Errorf("invalid class id: %w", err)
	}
	membershipUUID, err := toUUID(membershipID)
	if err != nil {
		return false, fmt.Errorf("invalid membership id: %w", err)
	}
	exists, err := r.queries.IsClassTeacher(ctx, academicssqlc.IsClassTeacherParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
		MembershipID:   membershipUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check class teacher: %w", err)
	}
	return exists, nil
}

func (r *sqlcRepository) ClassExists(ctx context.Context, orgID, classID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return false, fmt.Errorf("invalid class id: %w", err)
	}
	exists, err := r.queries.ClassExists(ctx, academicssqlc.ClassExistsParams{
		OrganizationID: orgUUID,
		ID:             classUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check class exists: %w", err)
	}
	return exists, nil
}
