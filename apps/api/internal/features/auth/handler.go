package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
)

// RefreshCookieName is the HttpOnly cookie used to transport refresh tokens.
const RefreshCookieName = "vts_refresh"

// Handler exposes the auth HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates an auth HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	result, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid organization code, username, or password")
		case errors.Is(err, ErrAccountLocked):
			writeError(w, r, http.StatusTooManyRequests, "account_locked", "account temporarily locked due to failed login attempts")
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "login failed")
		}
		return
	}

	setRefreshCookie(w, result.RefreshToken, result.RefreshExpires)
	writeData(w, http.StatusOK, LoginResponse{
		AccessToken:        result.AccessToken,
		ExpiresIn:          result.ExpiresIn,
		User:               result.User,
		Roles:              result.Roles,
		Permissions:        result.Permissions,
		MustChangePassword: result.MustChangePassword,
	})
}

// Me handles GET /api/v1/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid authorization header")
		return
	}

	token := strings.TrimPrefix(header, "Bearer ")
	result, err := h.svc.Me(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load actor")
		return
	}

	writeData(w, http.StatusOK, MeResponse{
		ID:                 result.ID,
		OrganizationID:     result.OrganizationID,
		DisplayName:        result.DisplayName,
		Roles:              result.Roles,
		Permissions:        result.Permissions,
		MustChangePassword: result.MustChangePassword,
	})
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing refresh cookie")
		return
	}

	result, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired refresh session")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "refresh failed")
		return
	}

	setRefreshCookie(w, result.RefreshToken, result.RefreshExpires)
	writeData(w, http.StatusOK, LoginResponse{
		AccessToken:        result.AccessToken,
		ExpiresIn:          result.ExpiresIn,
		User:               result.User,
		Roles:              result.Roles,
		Permissions:        result.Permissions,
		MustChangePassword: result.MustChangePassword,
	})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	var raw string
	if c, err := r.Cookie(RefreshCookieName); err == nil {
		raw = c.Value
	}

	if _, err := h.svc.Logout(r.Context(), raw); err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "logout failed")
		return
	}

	clearRefreshCookie(w)
	writeData(w, http.StatusOK, LogoutResponse{Success: true})
}

// ChangePassword handles POST /api/v1/auth/change-password.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		writeError(w, r, http.StatusForbidden, "invalid_csrf", "invalid csrf token")
		return
	}

	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid authorization header")
		return
	}
	token := strings.TrimPrefix(header, "Bearer ")

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if err := h.svc.ChangePassword(r.Context(), token, req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "current password is incorrect")
		case errors.Is(err, ErrUnauthorized):
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
		case errors.Is(err, ErrWeakPassword), errors.Is(err, ErrPasswordUnchanged), errors.Is(err, ErrPasswordReused):
			writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "change password failed")
		}
		return
	}

	writeData(w, http.StatusOK, ChangePasswordResponse{Success: true})
}
