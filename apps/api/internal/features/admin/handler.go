package admin

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid access token")
		return auth.Actor{}, false
	}

	for _, role := range actor.Roles {
		if role == "admin" {
			return actor, true
		}
	}

	writeError(w, r, http.StatusForbidden, "forbidden", "insufficient permissions")
	return auth.Actor{}, false
}

// ListUsers handles GET /api/v1/users.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	opts, ok := parseListOptions(w, r)
	if !ok {
		return
	}

	users, page, err := h.svc.ListUsers(r.Context(), actor.OrgID, opts)
	if err != nil {
		if errors.Is(err, ErrInvalidCursor) {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid cursor")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}

	if opts.Limit > 0 {
		writePagedData(w, http.StatusOK, users, page)
		return
	}

	writeData(w, http.StatusOK, users)
}

// ListAuditLogs handles GET /api/v1/audit-logs.
func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	opts, ok := parseAuditLogListOptions(w, r)
	if !ok {
		return
	}

	logs, page, err := h.svc.ListAuditLogs(r.Context(), actor.OrgID, opts)
	if err != nil {
		if errors.Is(err, ErrInvalidCursor) {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid cursor")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to list audit logs")
		return
	}

	if opts.Limit > 0 {
		writePagedData(w, http.StatusOK, logs, page)
		return
	}

	writeData(w, http.StatusOK, logs)
}

// ExportAuditLogs handles GET /api/v1/audit-logs/export.
func (h *Handler) ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	opts, ok := parseAuditLogListOptions(w, r)
	if !ok {
		return
	}

	logs, err := h.svc.ExportAuditLogs(r.Context(), actor.OrgID, opts)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to export audit logs")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=audit-logs.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"id", "created_at", "actor_name", "actor_user_id", "action", "resource_type", "resource_id", "before_json", "after_json", "metadata_json"}
	if err := writer.Write(header); err != nil {
		return
	}

	for _, log := range logs {
		row := []string{
			log.ID,
			log.CreatedAt,
			log.ActorName,
			log.ActorUserID,
			log.Action,
			log.ResourceType,
			log.ResourceID,
			log.Before,
			log.After,
			log.Metadata,
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
}

// ImportUsers handles POST /api/v1/users/imports.
func (h *Handler) ImportUsers(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req ImportUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	result, err := h.svc.ImportUsers(r.Context(), actor.OrgID, actor.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to import users")
		}
		return
	}

	if req.DryRun {
		writeData(w, http.StatusOK, result)
		return
	}
	writeData(w, http.StatusCreated, result)
}

// CreateUser handles POST /api/v1/users.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	user, err := h.svc.CreateUser(r.Context(), actor.OrgID, actor.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateLogin):
			writeError(w, r, http.StatusConflict, "duplicate_login", err.Error())
		case errors.Is(err, ErrInvalidInput), errors.Is(err, auth.ErrWeakPassword), errors.Is(err, auth.ErrPasswordReused):
			writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to create user")
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
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	userID := chi.URLParam(r, "user_id")
	if err := h.svc.UpdateRoles(r.Context(), actor.OrgID, actor.UserID, userID, req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to update roles")
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
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	userID := chi.URLParam(r, "user_id")
	if err := h.svc.ResetPassword(r.Context(), actor.OrgID, actor.UserID, userID, req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, ErrInvalidInput), errors.Is(err, auth.ErrWeakPassword), errors.Is(err, auth.ErrPasswordReused):
			writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to reset password")
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
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load organization")
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
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	org, err := h.svc.UpdateOrganization(r.Context(), actor.OrgID, actor.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrganizationNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to update organization")
		}
		return
	}

	writeData(w, http.StatusOK, org)
}

func parseListOptions(w http.ResponseWriter, r *http.Request) (ListOptions, bool) {
	opts := ListOptions{Query: strings.TrimSpace(r.URL.Query().Get("q"))}

	if l := r.URL.Query().Get("limit"); l != "" {
		val, err := strconv.Atoi(l)
		if err != nil || val < 1 {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid limit")
			return ListOptions{}, false
		}
		opts.Limit = val
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		val, err := strconv.Atoi(o)
		if err != nil || val < 0 {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid offset")
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

func parseAuditLogListOptions(w http.ResponseWriter, r *http.Request) (AuditLogListOptions, bool) {
	opts := AuditLogListOptions{
		Action:      strings.TrimSpace(r.URL.Query().Get("action")),
		ActorUserID: strings.TrimSpace(r.URL.Query().Get("actor_user_id")),
		From:        strings.TrimSpace(r.URL.Query().Get("from")),
		To:          strings.TrimSpace(r.URL.Query().Get("to")),
	}

	if opts.From != "" {
		if _, err := time.Parse(time.RFC3339, opts.From); err != nil {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid from timestamp")
			return AuditLogListOptions{}, false
		}
	}
	if opts.To != "" {
		if _, err := time.Parse(time.RFC3339, opts.To); err != nil {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid to timestamp")
			return AuditLogListOptions{}, false
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		val, err := strconv.Atoi(l)
		if err != nil || val < 1 {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid limit")
			return AuditLogListOptions{}, false
		}
		opts.Limit = val
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		val, err := strconv.Atoi(o)
		if err != nil || val < 0 {
			writeError(w, r, http.StatusBadRequest, "bad_request", "invalid offset")
			return AuditLogListOptions{}, false
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
