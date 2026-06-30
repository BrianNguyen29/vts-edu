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
	loginFunc   func(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error)
	refreshFunc func(ctx context.Context, raw string) (*auth.RefreshResult, error)
	logoutFunc  func(ctx context.Context, raw string) (*auth.LogoutResult, error)
	meFunc      func(ctx context.Context, token string) (*auth.MeResult, error)
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

func addCSRF(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "demo-csrf-token"})
	req.Header.Set(csrf.HeaderName, "demo-csrf-token")
}

func TestHandler_Login_OK(t *testing.T) {
	svc := &fakeService{
		loginFunc: func(ctx context.Context, req auth.LoginRequest) (*auth.LoginResult, error) {
			return &auth.LoginResult{
				AccessToken:    "access-token-123",
				ExpiresIn:      900,
				RefreshToken:   "refresh-token-456",
				RefreshExpires: time.Now().Add(7 * 24 * time.Hour),
				User: auth.UserInfo{
					ID:          "user-id",
					DisplayName: "hs001",
				},
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
				ID:             "user-id",
				OrganizationID: "org-id",
				DisplayName:    "hs001",
				Roles:          []string{"student"},
				Permissions:    []string{},
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
}

func TestHandler_Refresh_OK(t *testing.T) {
	svc := &fakeService{
		refreshFunc: func(ctx context.Context, raw string) (*auth.RefreshResult, error) {
			return &auth.RefreshResult{
				AccessToken:    "new-access-token",
				ExpiresIn:      900,
				RefreshToken:   "new-refresh-token",
				RefreshExpires: time.Now().Add(7 * 24 * time.Hour),
				User: auth.UserInfo{
					ID:          "user-id",
					DisplayName: "hs001",
				},
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
