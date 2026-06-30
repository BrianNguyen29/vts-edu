package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/go-chi/chi/v5"
)

// Handler exposes the admin HTTP endpoints.
type Handler struct {
	svc    Service
	issuer *auth.TokenIssuer
}

// NewHandler creates an admin HTTP handler.
func NewHandler(svc Service, issuer *auth.TokenIssuer) *Handler {
	return &Handler{svc: svc, issuer: issuer}
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (auth.Actor, bool) {
	actor, err := auth.ActorFromRequest(r, h.issuer)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return auth.Actor{}, false
	}

	for _, role := range actor.Roles {
		if role == "admin" {
			return actor, true
		}
	}

	writeError(w, http.StatusForbidden, "forbidden", "insufficient permissions")
	return auth.Actor{}, false
}

// ListUsers handles GET /api/v1/users.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	users, err := h.svc.ListUsers(r.Context(), actor.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}

	writeData(w, http.StatusOK, users)
}

// CreateUser handles POST /api/v1/users.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	user, err := h.svc.CreateUser(r.Context(), actor.OrgID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateLogin):
			writeError(w, http.StatusConflict, "duplicate_login", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create user")
		}
		return
	}

	writeData(w, http.StatusCreated, user)
}

// UpdateRoles handles PUT /api/v1/users/{user_id}/roles.
func (h *Handler) UpdateRoles(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req UpdateRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	userID := chi.URLParam(r, "user_id")
	if err := h.svc.UpdateRoles(r.Context(), actor.OrgID, userID, req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to update roles")
		}
		return
	}

	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// ResetPassword handles POST /api/v1/users/{user_id}/reset-password.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	userID := chi.URLParam(r, "user_id")
	if err := h.svc.ResetPassword(r.Context(), actor.OrgID, userID, req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to reset password")
		}
		return
	}

	writeData(w, http.StatusOK, map[string]bool{"success": true})
}

// GetOrganization handles GET /api/v1/organizations/current.
func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	org, err := h.svc.GetOrganization(r.Context(), actor.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load organization")
		return
	}

	writeData(w, http.StatusOK, org)
}

// UpdateOrganization handles PATCH /api/v1/organizations/current.
func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	org, err := h.svc.UpdateOrganization(r.Context(), actor.OrgID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrganizationNotFound):
			writeError(w, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to update organization")
		}
		return
	}

	writeData(w, http.StatusOK, org)
}
