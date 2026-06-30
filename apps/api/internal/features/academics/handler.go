package academics

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
)

// Handler exposes the academics HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates an academics HTTP handler.
func NewHandler(svc Service, issuer *auth.TokenIssuer) *Handler {
	return &Handler{svc: svc, issuer: issuer}
}

func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (auth.Actor, bool) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return auth.Actor{}, false
	}
	return actor, true
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (auth.Actor, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return auth.Actor{}, false
	}
	for _, role := range actor.Roles {
		if role == "admin" {
			return actor, true
		}
	}
	writeError(w, http.StatusForbidden, "forbidden", "admin access required")
	return auth.Actor{}, false
}

func (h *Handler) requireTeacherOrAdmin(w http.ResponseWriter, r *http.Request) (auth.Actor, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return auth.Actor{}, false
	}
	for _, role := range actor.Roles {
		if role == "teacher" || role == "admin" {
			return actor, true
		}
	}
	writeError(w, http.StatusForbidden, "forbidden", "teacher or admin access required")
	return auth.Actor{}, false
}

func (h *Handler) requireTeacher(w http.ResponseWriter, r *http.Request) (auth.Actor, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return auth.Actor{}, false
	}
	for _, role := range actor.Roles {
		if role == "teacher" {
			return actor, true
		}
	}
	writeError(w, http.StatusForbidden, "forbidden", "teacher access required")
	return auth.Actor{}, false
}

func (h *Handler) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrDuplicateCode), errors.Is(err, ErrDuplicateTeacher), errors.Is(err, ErrDuplicateEnrollment):
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "academics operation failed")
	}
}

// Terms

func (h *Handler) ListTerms(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacherOrAdmin(w, r)
	if !ok {
		return
	}
	terms, err := h.svc.ListTerms(r.Context(), actor.OrgID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, terms)
}

func (h *Handler) CreateTerm(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	var req CreateTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	term, err := h.svc.CreateTerm(r.Context(), actor.OrgID, actor.Roles, req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, term)
}

func (h *Handler) ArchiveTerm(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	termID := chi.URLParam(r, "term_id")
	if err := h.svc.ArchiveTerm(r.Context(), actor.OrgID, actor.Roles, termID); err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// Subjects

func (h *Handler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacherOrAdmin(w, r)
	if !ok {
		return
	}
	subjects, err := h.svc.ListSubjects(r.Context(), actor.OrgID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, subjects)
}

func (h *Handler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	var req CreateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	subject, err := h.svc.CreateSubject(r.Context(), actor.OrgID, actor.Roles, req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, subject)
}

func (h *Handler) ArchiveSubject(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	subjectID := chi.URLParam(r, "subject_id")
	if err := h.svc.ArchiveSubject(r.Context(), actor.OrgID, actor.Roles, subjectID); err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// Courses

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacherOrAdmin(w, r)
	if !ok {
		return
	}
	courses, err := h.svc.ListCourses(r.Context(), actor.OrgID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, courses)
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	var req CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	course, err := h.svc.CreateCourse(r.Context(), actor.OrgID, actor.Roles, req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, course)
}

func (h *Handler) ArchiveCourse(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	courseID := chi.URLParam(r, "course_id")
	if err := h.svc.ArchiveCourse(r.Context(), actor.OrgID, actor.Roles, courseID); err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// Classes

func (h *Handler) ListClasses(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacherOrAdmin(w, r)
	if !ok {
		return
	}
	classes, err := h.svc.ListClasses(r.Context(), actor.OrgID, actor.UserID, actor.Roles)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, classes)
}

func (h *Handler) ListMyTeachingClasses(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacher(w, r)
	if !ok {
		return
	}
	classes, err := h.svc.ListMyTeachingClasses(r.Context(), actor.OrgID, actor.UserID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, classes)
}

func (h *Handler) CreateClass(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	var req CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	class, err := h.svc.CreateClass(r.Context(), actor.OrgID, actor.Roles, req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, class)
}

func (h *Handler) ArchiveClass(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	classID := chi.URLParam(r, "class_id")
	if err := h.svc.ArchiveClass(r.Context(), actor.OrgID, actor.Roles, classID); err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// Class teachers

func (h *Handler) ListClassTeachers(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacherOrAdmin(w, r)
	if !ok {
		return
	}
	classID := chi.URLParam(r, "class_id")
	teachers, err := h.svc.ListClassTeachers(r.Context(), actor.OrgID, actor.UserID, actor.Roles, classID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, teachers)
}

func (h *Handler) AddClassTeacher(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	classID := chi.URLParam(r, "class_id")
	var req AddClassTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	teacher, err := h.svc.AddClassTeacher(r.Context(), actor.OrgID, actor.Roles, classID, req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, teacher)
}

func (h *Handler) RemoveClassTeacher(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	classID := chi.URLParam(r, "class_id")
	userID := chi.URLParam(r, "user_id")
	if err := h.svc.RemoveClassTeacher(r.Context(), actor.OrgID, actor.Roles, classID, userID); err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// Enrollments

func (h *Handler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireTeacherOrAdmin(w, r)
	if !ok {
		return
	}
	classID := chi.URLParam(r, "class_id")
	enrollments, err := h.svc.ListEnrollments(r.Context(), actor.OrgID, actor.UserID, actor.Roles, classID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, enrollments)
}

func (h *Handler) EnrollStudent(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	classID := chi.URLParam(r, "class_id")
	var req EnrollStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	enrollment, err := h.svc.EnrollStudent(r.Context(), actor.OrgID, actor.Roles, classID, req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, enrollment)
}

func (h *Handler) UnenrollStudent(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}
	classID := chi.URLParam(r, "class_id")
	userID := chi.URLParam(r, "user_id")
	if err := h.svc.UnenrollStudent(r.Context(), actor.OrgID, actor.Roles, classID, userID); err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"success": true})
}
