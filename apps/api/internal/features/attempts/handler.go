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
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	snapshot, err := h.svc.GetAttempt(r.Context(), actor, attemptID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, snapshot)
}

// SaveAnswer handles PUT /api/v1/attempts/{attempt_id}/answers/{attempt_item_id}.
func (h *Handler) SaveAnswer(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	itemID := chi.URLParam(r, "attempt_item_id")

	var req SaveAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	saved, err := h.svc.SaveAnswer(r.Context(), actor, attemptID, itemID, req.AnswerPayload)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, saved)
}

// SubmitAttempt handles POST /api/v1/attempts/{attempt_id}/submit.
func (h *Handler) SubmitAttempt(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	attemptID := chi.URLParam(r, "attempt_id")
	result, err := h.svc.SubmitAttempt(r.Context(), actor, attemptID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, result)
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAttemptNotFound), errors.Is(err, ErrAnswerItemNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, ErrAttemptExpired):
		writeError(w, http.StatusConflict, "attempt_expired", err.Error())
	case errors.Is(err, ErrAttemptNotInProgress):
		writeError(w, http.StatusConflict, "attempt_not_in_progress", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "request failed")
	}
}
