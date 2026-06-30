package assessments

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	opts, ok := parseListOptions(w, r)
	if !ok {
		return
	}

	list, page, err := h.svc.ListAssessments(r.Context(), actor.OrgID, opts)
	if err != nil {
		if errors.Is(err, ErrInvalidCursor) {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid cursor")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list assessments")
		return
	}

	if opts.Limit > 0 {
		writePagedData(w, http.StatusOK, list, page)
		return
	}

	writeData(w, http.StatusOK, list)
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
