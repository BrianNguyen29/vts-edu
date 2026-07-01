package gradebook

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/go-chi/chi/v5"
)

// Handler exposes the gradebook HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates a gradebook HTTP handler.
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
	case errors.Is(err, auth.ErrUnauthorized):
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	default:
		log.Printf("gradebook handler error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "gradebook operation failed")
	}
}

// ListAssessmentAttempts handles GET /api/v1/assessments/{id}/attempts.
func (h *Handler) ListAssessmentAttempts(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	assessmentID := chi.URLParam(r, "id")
	attempts, err := h.svc.ListAssessmentAttempts(r.Context(), actor, assessmentID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, attempts)
}

// GetAssessmentResults handles GET /api/v1/assessments/{id}/results.
func (h *Handler) GetAssessmentResults(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	assessmentID := chi.URLParam(r, "id")
	result, err := h.svc.GetAssessmentResults(r.Context(), actor, assessmentID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

// ExportAssessmentAttemptsCSV handles GET /api/v1/assessments/{id}/attempts/export.
func (h *Handler) ExportAssessmentAttemptsCSV(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	assessmentID := chi.URLParam(r, "id")
	data, err := h.svc.ExportAssessmentAttemptsCSV(r.Context(), actor, assessmentID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"assessment-"+assessmentID+"-attempts.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GetClassGradebook handles GET /api/v1/classes/{id}/gradebook.
func (h *Handler) GetClassGradebook(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	classID := chi.URLParam(r, "class_id")
	entries, err := h.svc.GetClassGradebook(r.Context(), actor, classID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	writeData(w, http.StatusOK, entries)
}

// ExportClassGradebookCSV handles GET /api/v1/classes/{class_id}/gradebook/export.
func (h *Handler) ExportClassGradebookCSV(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	classID := chi.URLParam(r, "class_id")
	data, err := h.svc.ExportClassGradebookCSV(r.Context(), actor, classID)
	if err != nil {
		h.mapError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"class-"+classID+"-gradebook.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(DataEnvelope{Data: data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	var env ErrorEnvelope
	env.Error.Code = code
	env.Error.Message = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
