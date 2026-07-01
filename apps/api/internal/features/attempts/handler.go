package attempts

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
)

// Handler exposes the attempts HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates an attempts HTTP handler.
func NewHandler(svc Service, issuer *auth.TokenIssuer) *Handler {
	return &Handler{svc: svc, issuer: issuer}
}

// GetAttempt handles GET /api/v1/attempts/{attempt_id}.
func (h *Handler) GetAttempt(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	snapshot, err := h.svc.GetAttempt(r.Context(), actor, attemptID)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusOK, snapshot)
}

// SaveAnswer handles PUT /api/v1/attempts/{attempt_id}/answers/{attempt_item_id}.
func (h *Handler) SaveAnswer(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	itemID := chi.URLParam(r, "attempt_item_id")

	var req SaveAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	saved, err := h.svc.SaveAnswer(r.Context(), actor, attemptID, itemID, req.AnswerPayload)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusOK, saved)
}

// SubmitAttempt handles POST /api/v1/attempts/{attempt_id}/submit.
func (h *Handler) SubmitAttempt(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	result, err := h.svc.SubmitAttempt(r.Context(), actor, attemptID)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusOK, result)
}

// ListAssignedAssessments handles GET /api/v1/me/assessments.
func (h *Handler) ListAssignedAssessments(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	assessments, err := h.svc.ListAssignedAssessments(r.Context(), actor)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusOK, assessments)
}

// ListAttemptHistory handles GET /api/v1/me/attempts.
func (h *Handler) ListAttemptHistory(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attempts, err := h.svc.ListAttemptHistory(r.Context(), actor)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusOK, attempts)
}

// GetAttemptResult handles GET /api/v1/attempts/{attempt_id}/result.
func (h *Handler) GetAttemptResult(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	result, err := h.svc.GetAttemptResult(r.Context(), actor, attemptID)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusOK, result)
}

// StartAttempt handles POST /api/v1/assessments/{assessment_id}/attempts.
func (h *Handler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	assessmentID := chi.URLParam(r, "assessment_id")
	snapshot, err := h.svc.StartAttempt(r.Context(), actor, assessmentID)
	if err != nil {
		mapServiceError(w, r, err)
		return
	}

	writeData(w, http.StatusCreated, snapshot)
}

func mapServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, auth.ErrUnauthorized):
		writeError(w, r, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrAttemptNotFound), errors.Is(err, ErrAnswerItemNotFound), errors.Is(err, ErrAssessmentNotFound), errors.Is(err, ErrNoPublication):
		writeError(w, r, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, ErrAttemptExpired):
		writeError(w, r, http.StatusConflict, "attempt_expired", err.Error())
	case errors.Is(err, ErrAttemptNotInProgress):
		writeError(w, r, http.StatusConflict, "attempt_not_in_progress", err.Error())
	case errors.Is(err, ErrAttemptNotSubmitted):
		writeError(w, r, http.StatusConflict, "attempt_not_submitted", err.Error())
	case errors.Is(err, ErrAssessmentUnavailable):
		writeError(w, r, http.StatusForbidden, "assessment_unavailable", err.Error())
	case errors.Is(err, ErrAttemptLimitReached):
		writeError(w, r, http.StatusConflict, "attempt_limit_reached", err.Error())
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "request failed")
	}
}
