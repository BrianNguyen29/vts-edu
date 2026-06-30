package assessments

import (
	"net/http"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
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

// ListAssessments handles GET /api/v1/assessments for teachers and admins.
func (h *Handler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return
	}

	if !hasRequiredRole(actor.Roles, []string{"teacher", "admin"}) {
		writeError(w, http.StatusForbidden, "forbidden", "insufficient permissions")
		return
	}

	list, err := h.svc.ListAssessments(r.Context(), actor.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list assessments")
		return
	}

	writeData(w, http.StatusOK, list)
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
