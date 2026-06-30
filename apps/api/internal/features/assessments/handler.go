package assessments

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
)

// Handler exposes the assessments HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates an assessments HTTP handler.
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

func (h *Handler) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCursor):
		writeError(w, http.StatusBadRequest, "bad_request", "invalid cursor")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, ErrUnauthorized):
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrNotDraft), errors.Is(err, ErrValidationFailed),
		errors.Is(err, ErrDuplicateSection), errors.Is(err, ErrDuplicateItem), errors.Is(err, ErrDuplicateTarget):
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "assessment operation failed")
	}
}

// ListAssessments handles GET /api/v1/assessments for teachers and admins.
func (h *Handler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	if !hasRequiredRole(actor.Roles, []string{"teacher", "admin"}) {
		writeError(w, http.StatusForbidden, "forbidden", "insufficient permissions")
		return
	}

	opts, ok := parseListOptions(w, r)
	if !ok {
		return
	}

	list, page, err := h.svc.ListAssessments(r.Context(), actor.OrgID, opts)
	if err != nil {
		h.mapError(w, err)
		return
	}

	if opts.Limit > 0 {
		writePagedData(w, http.StatusOK, list, page)
		return
	}

	writeData(w, http.StatusOK, list)
}

// CreateAssessment handles POST /api/v1/classes/{class_id}/assessments.
func (h *Handler) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var req CreateAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	assessment, err := h.svc.CreateAssessment(r.Context(), actor, chi.URLParam(r, "class_id"), req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, assessment)
}

// ListAssessmentsByClass handles GET /api/v1/classes/{class_id}/assessments.
func (h *Handler) ListAssessmentsByClass(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	list, err := h.svc.ListAssessmentsByClass(r.Context(), actor, chi.URLParam(r, "class_id"))
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, list)
}

// GetAssessment handles GET /api/v1/assessments/{id}.
func (h *Handler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	assessment, err := h.svc.GetAssessment(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, assessment)
}

// UpdateAssessment handles PATCH /api/v1/assessments/{id}.
func (h *Handler) UpdateAssessment(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var req UpdateAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	assessment, err := h.svc.UpdateAssessment(r.Context(), actor, chi.URLParam(r, "id"), req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, assessment)
}

// CreateSection handles POST /api/v1/assessments/{id}/sections.
func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var req CreateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	section, err := h.svc.CreateSection(r.Context(), actor, chi.URLParam(r, "id"), req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, section)
}

// CreateItem handles POST /api/v1/assessment-sections/{section_id}/items.
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	item, err := h.svc.CreateItem(r.Context(), actor, chi.URLParam(r, "section_id"), req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, item)
}

// CreateTarget handles POST /api/v1/assessments/{id}/targets.
func (h *Handler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var req CreateTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	target, err := h.svc.CreateTarget(r.Context(), actor, chi.URLParam(r, "id"), req)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, target)
}

// ValidateAssessment handles POST /api/v1/assessments/{id}/validate.
func (h *Handler) ValidateAssessment(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	result, err := h.svc.ValidateAssessment(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

// PublishAssessment handles POST /api/v1/assessments/{id}/publish.
func (h *Handler) PublishAssessment(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	result, err := h.svc.PublishAssessment(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

func parseListOptions(w http.ResponseWriter, r *http.Request) (ListOptions, bool) {
	opts := ListOptions{Query: strings.TrimSpace(r.URL.Query().Get("q"))}

	if l := r.URL.Query().Get("limit"); l != "" {
		val, err := strconv.Atoi(l)
		if err != nil || val < 1 {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid limit")
			return ListOptions{}, false
		}
		opts.Limit = val
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		val, err := strconv.Atoi(o)
		if err != nil || val < 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid offset")
			return ListOptions{}, false
		}
		opts.Offset = val
	}

	if cursor := strings.TrimSpace(r.URL.Query().Get("cursor")); cursor != "" {
		opts.Cursor = cursor
		opts.Offset = 0
	}

	if r.URL.Query().Get("count") == "true" {
		opts.Count = true
	}

	return opts, true
}

func hasRequiredRole(roles []string, required []string) bool {
	for _, r := range roles {
		for _, req := range required {
			if r == req {
				return true
			}
		}
	}
	return false
}
