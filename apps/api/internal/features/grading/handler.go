package grading

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler exposes the grading HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates the grading HTTP handler.
func NewHandler(svc Service, issuer *auth.TokenIssuer) *Handler {
	return &Handler{svc: svc, issuer: issuer}
}

// ListReviewQueue handles GET /api/v1/assessments/{id}/review-queue.
func (h *Handler) ListReviewQueue(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}
	assessmentID := chi.URLParam(r, "id")
	entries, err := h.svc.ListReviewQueue(r.Context(), actor, assessmentID)
	if err != nil {
		mapError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, entries)
}

// GetAttemptForReview handles GET /api/v1/attempts/{attempt_id}/review.
func (h *Handler) GetAttemptForReview(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}
	attemptID := chi.URLParam(r, "attempt_id")
	ctx, err := h.svc.GetAttemptForReview(r.Context(), actor, attemptID)
	if err != nil {
		mapError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, ctx)
}

// GradeItem handles PUT /api/v1/attempts/{attempt_id}/items/{item_id}/grade.
func (h *Handler) GradeItem(w http.ResponseWriter, r *http.Request) {
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
	itemID := chi.URLParam(r, "item_id")

	var req GradeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.svc.GradeItem(r.Context(), actor, attemptID, itemID, req)
	if err != nil {
		mapError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, resp)
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrForbidden):
		writeError(w, r, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, ErrItemNotInAttempt):
		writeError(w, r, http.StatusNotFound, "item_not_in_attempt", err.Error())
	case errors.Is(err, ErrNotGradeable):
		writeError(w, r, http.StatusBadRequest, "not_gradable", err.Error())
	case errors.Is(err, ErrInvalidScore):
		writeError(w, r, http.StatusBadRequest, "invalid_score", err.Error())
	case errors.Is(err, ErrScoreExceedsPoints):
		writeError(w, r, http.StatusBadRequest, "score_exceeds_points", err.Error())
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "grading operation failed")
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(DataEnvelope{Data: data})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorEnvelope{}
	resp.Error.Code = code
	resp.Error.Message = message
	resp.Error.RequestID = middleware.GetReqID(r.Context())
	_ = json.NewEncoder(w).Encode(resp)
}
