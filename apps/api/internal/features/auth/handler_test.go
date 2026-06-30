package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
)

type fakeService struct {
	loginFunc          func(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error)
	refreshFunc        func(ctx context.Context, raw string) (*auth.RefreshResult, error)
	logoutFunc         func(ctx context.Context, raw string) (*auth.LogoutResult, error)
	meFunc             func(ctx context.Context, token string) (*auth.MeResult, error)
	changePasswordFunc func(ctx context.Context, token string, req auth.ChangePasswordRequest) error
}

func (f *fakeService) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error) {
	return f.loginFunc(ctx, req)
}

func (f *fakeService) Refresh(ctx context.Context, raw string) (*auth.RefreshResult, error) {
	return f.refreshFunc(ctx, raw)
}

func (f *fakeService) Logout(ctx context.Context, raw string) (*auth.LogoutResult, error) {
	return f.logoutFunc(ctx, raw)
}

func (f *fakeService) Me(ctx context.Context, token string) (*auth.MeResult, error) {
	return f.meFunc(ctx, token)
}

func (f *fakeService) ChangePassword(ctx context.Context, token string, req auth.ChangePasswordRequest) error {
	return f.changePasswordFunc(ctx, token, req)
}

func addCSRF(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "demo-csrf-token"})
	req.Header.Set(csrf.HeaderName, "demo-csrf-token")
}

func TestHandler_Login_OK(t *testing.T) {
	svc := &fakeService{
		loginFunc: func(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error) {
			return &auth.LoginResult{
				AccessToken:        "access-token-123",
				ExpiresIn:          900,
				RefreshToken:       "refresh-token-456",
				RefreshExpires:     time.Now().Add(7 * 24 * time.Hour),
				User:               auth.UserInfo{ID: "user-id", DisplayName: "hs001"},
				Roles:              []string{"student"},
				Permissions:        []string{"attempt:read", "attempt:write", "self:read"},
				MustChangePassword: false,
			}, nil
		},
	}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"organization_code":"school-a","username":"hs001","password":"Password123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data auth.LoginResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.AccessToken != "access-token-123" {
		t.Errorf("access_token = %q, want %q", resp.Data.AccessToken, "access-token-123")
	}
	if resp.Data.ExpiresIn != 900 {
		t.Errorf("expires_in = %d, want 900", resp.Data.ExpiresIn)
	}

	cookies := rec.Result().Cookies()
	var refresh *http.Cookie
	for _, c := range cookies {
		if c.Name == auth.RefreshCookieName {
			refresh = c
			break
		}
	}
	if refresh == nil {
		t.Fatal("expected refresh cookie")
	}
	if refresh.Path != "/api/v1/auth" {
		t.Errorf("refresh path = %q, want %q", refresh.Path, "/api/v1/auth")
	}
	if !refresh.HttpOnly {
		t.Error("expected refresh cookie to be HttpOnly")
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &fakeService{
		loginFunc: func(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error) {
			return nil, auth.ErrInvalidCredentials
		},
	}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"organization_code":"school-a","username":"hs001","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var resp auth.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != "invalid_credentials" {
		t.Errorf("error code = %q, want invalid_credentials", resp.Error.Code)
	}
}

func TestHandler_Login_AccountLocked(t *testing.T) {
	svc := &fakeService{
		loginFunc: func(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error) {
			return nil, auth.ErrAccountLocked
		},
	}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"organization_code":"school-a","username":"hs001","password":"Password123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var resp auth.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != "account_locked" {
		t.Errorf("error code = %q, want account_locked", resp.Error.Code)
	}
}

func TestHandler_Me_MissingToken(t *testing.T) {
	svc := &fakeService{}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_Me_ValidToken(t *testing.T) {
	svc := &fakeService{
		meFunc: func(ctx context.Context, token string) (*auth.MeResult, error) {
			if token != "valid-token" {
				return nil, errors.New("unexpected token")
			}
			return &auth.MeResult{
				ID:                 "user-id",
				OrganizationID:     "org-id",
				DisplayName:        "hs001",
				Roles:              []string{"student"},
				Permissions:        []string{"attempt:read"},
				MustChangePassword: true,
			}, nil
		},
	}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data auth.MeResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.ID != "user-id" {
		t.Errorf("id = %q, want user-id", resp.Data.ID)
	}
	if resp.Data.OrganizationID != "org-id" {
		t.Errorf("organization_id = %q, want org-id", resp.Data.OrganizationID)
	}
	if len(resp.Data.Roles) != 1 || resp.Data.Roles[0] != "student" {
		t.Errorf("roles = %v, want [student]", resp.Data.Roles)
	}
	if !resp.Data.MustChangePassword {
		t.Error("expected must_change_password = true")
	}
}

func TestHandler_Refresh_OK(t *testing.T) {
	svc := &fakeService{
		refreshFunc: func(ctx context.Context, raw string) (*auth.RefreshResult, error) {
			return &auth.RefreshResult{
				AccessToken:        "new-access-token",
				ExpiresIn:          900,
				RefreshToken:       "new-refresh-token",
				RefreshExpires:     time.Now().Add(7 * 24 * time.Hour),
				User:               auth.UserInfo{ID: "user-id", DisplayName: "hs001"},
				Roles:              []string{"teacher"},
				Permissions:        []string{"assessment:read", "attempt:read"},
				MustChangePassword: true,
			}, nil
		},
	}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "old-refresh-token"})
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data auth.LoginResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.AccessToken != "new-access-token" {
		t.Errorf("access_token = %q, want new-access-token", resp.Data.AccessToken)
	}

	cookies := rec.Result().Cookies()
	var refresh *http.Cookie
	for _, c := range cookies {
		if c.Name == auth.RefreshCookieName {
			refresh = c
			break
		}
	}
	if refresh == nil || refresh.Value != "new-refresh-token" {
		t.Fatal("expected new refresh cookie")
	}
}

func TestHandler_Refresh_MissingCookie(t *testing.T) {
	svc := &fakeService{}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_Logout_OK(t *testing.T) {
	svc := &fakeService{
		logoutFunc: func(ctx context.Context, raw string) (*auth.LogoutResult, error) {
			return &auth.LogoutResult{Success: true}, nil
		},
	}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "refresh-token"})
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data auth.LogoutResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Data.Success {
		t.Error("expected success")
	}

	cookies := rec.Result().Cookies()
	var cleared *http.Cookie
	for _, c := range cookies {
		if c.Name == auth.RefreshCookieName {
			cleared = c
			break
		}
	}
	if cleared == nil || cleared.MaxAge != -1 {
		t.Error("expected cleared refresh cookie")
	}
}

func TestHandler_Logout_MissingCookie(t *testing.T) {
	svc := &fakeService{
		logoutFunc: func(ctx context.Context, raw string) (*auth.LogoutResult, error) {
			return &auth.LogoutResult{Success: true}, nil
		},
	}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_ChangePassword_OK(t *testing.T) {
	svc := &fakeService{
		changePasswordFunc: func(ctx context.Context, token string, req auth.ChangePasswordRequest) error {
			if token != "valid-token" {
				return errors.New("unexpected token")
			}
			if req.CurrentPassword != "Password123!" || req.NewPassword != "NewPassword123!" {
				return errors.New("unexpected password payload")
			}
			return nil
		},
	}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"current_password":"Password123!","new_password":"NewPassword123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data auth.ChangePasswordResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Data.Success {
		t.Error("expected success")
	}
}

func TestHandler_ChangePassword_InvalidCredentials(t *testing.T) {
	svc := &fakeService{
		changePasswordFunc: func(ctx context.Context, token string, req auth.ChangePasswordRequest) error {
			return auth.ErrInvalidCredentials
		},
	}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"current_password":"WrongPassword","new_password":"NewPassword123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_ChangePassword_ReusedPassword(t *testing.T) {
	svc := &fakeService{
		changePasswordFunc: func(ctx context.Context, token string, req auth.ChangePasswordRequest) error {
			return auth.ErrPasswordReused
		},
	}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"current_password":"Password123!","new_password":"OldPassword123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	addCSRF(req)
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp auth.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != "bad_request" {
		t.Errorf("error code = %q, want bad_request", resp.Error.Code)
	}
}

func TestHandler_ChangePassword_MissingCSRF(t *testing.T) {
	svc := &fakeService{}
	h := auth.NewHandler(svc)

	body := strings.NewReader(`{"current_password":"Password123!","new_password":"NewPassword123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
