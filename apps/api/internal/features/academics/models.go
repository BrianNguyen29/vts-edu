package academics

import "time"

// Term is an academic term such as a semester or school year.
type Term struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
}

// CreateTermRequest is the payload for POST /api/v1/academic-terms.
type CreateTermRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// UpdateTermRequest is the payload for PATCH /api/v1/academic-terms/{term_id}.
type UpdateTermRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Subject is a subject taught in an organization.
type Subject struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
}

// CreateSubjectRequest is the payload for POST /api/v1/subjects.
type CreateSubjectRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateSubjectRequest is the payload for PATCH /api/v1/subjects/{subject_id}.
type UpdateSubjectRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Course is a subject offered during a specific term.
type Course struct {
	ID             string `json:"id"`
	SubjectID      string `json:"subject_id"`
	AcademicTermID string `json:"academic_term_id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Status         string `json:"status"`
}

// CreateCourseRequest is the payload for POST /api/v1/courses.
type CreateCourseRequest struct {
	SubjectID      string `json:"subject_id"`
	AcademicTermID string `json:"academic_term_id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
}

// UpdateCourseRequest is the payload for PATCH /api/v1/courses/{course_id}.
type UpdateCourseRequest struct {
	SubjectID      string `json:"subject_id"`
	AcademicTermID string `json:"academic_term_id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
}

// ClassSection is a concrete class section for a course.
type ClassSection struct {
	ID           string `json:"id"`
	CourseID     string `json:"course_id"`
	Name         string `json:"name"`
	StudentCount int64  `json:"student_count"`
	TeacherCount int64  `json:"teacher_count"`
	Status       string `json:"status"`
}

// CreateClassRequest is the payload for POST /api/v1/classes.
type CreateClassRequest struct {
	CourseID string `json:"course_id"`
	Name     string `json:"name"`
}

// UpdateClassRequest is the payload for PATCH /api/v1/classes/{class_id}.
type UpdateClassRequest struct {
	CourseID string `json:"course_id"`
	Name     string `json:"name"`
}

// ClassTeacher is a teacher assignment for a class.
type ClassTeacher struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

// AddClassTeacherRequest is the payload for POST /api/v1/classes/{class_id}/teachers.
type AddClassTeacherRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

// Enrollment is a student enrollment in a class.
type Enrollment struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}

// EnrollStudentRequest is the payload for POST /api/v1/classes/{class_id}/enrollments.
type EnrollStudentRequest struct {
	UserID string `json:"user_id"`
}

// BulkEnrollRequest is the payload for POST /api/v1/classes/{class_id}/enrollments/bulk.
type BulkEnrollRequest struct {
	UserIDs []string `json:"user_ids"`
	DryRun  bool     `json:"dry_run"`
}

// BulkEnrollmentRow is the per-row outcome of a bulk enrollment.
type BulkEnrollmentRow struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// BulkEnrollmentResult is the response for a bulk enrollment operation.
type BulkEnrollmentResult struct {
	Total    int                 `json:"total"`
	Enrolled int                 `json:"enrolled"`
	Failed   int                 `json:"failed"`
	DryRun   bool                `json:"dry_run"`
	Rows     []BulkEnrollmentRow `json:"rows"`
}

// BulkAssignTeacherItem is a single teacher assignment in a bulk operation.
type BulkAssignTeacherItem struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

// BulkAssignTeachersRequest is the payload for POST /api/v1/classes/{class_id}/teachers/bulk.
type BulkAssignTeachersRequest struct {
	Items  []BulkAssignTeacherItem `json:"items"`
	DryRun bool                    `json:"dry_run"`
}

// BulkAssignTeacherRow is the per-row outcome of a bulk teacher assignment.
type BulkAssignTeacherRow struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// BulkAssignTeachersResult is the response for a bulk teacher assignment operation.
type BulkAssignTeachersResult struct {
	Total    int                    `json:"total"`
	Assigned int                    `json:"assigned"`
	Failed   int                    `json:"failed"`
	DryRun   bool                   `json:"dry_run"`
	Rows     []BulkAssignTeacherRow `json:"rows"`
}

// MembershipInfo identifies an organization membership and its roles.
type MembershipInfo struct {
	ID     string
	UserID string
	Roles  []string
}

// PageInfo is returned with paginated list responses.
type PageInfo struct {
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
	TotalCount *int64  `json:"total_count,omitempty"`
}

// DataEnvelope wraps successful API responses.
type DataEnvelope struct {
	Data any       `json:"data"`
	Page *PageInfo `json:"page,omitempty"`
}

// ErrorEnvelope wraps API error responses.
type ErrorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

// AuditLogParams is the persistence input for an audit log row.
type AuditLogParams struct {
	OrganizationID string
	ActorUserID    string
	Action         string
	ResourceType   string
	ResourceID     string
	BeforeJSON     []byte
	AfterJSON      []byte
	MetadataJSON   []byte
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
