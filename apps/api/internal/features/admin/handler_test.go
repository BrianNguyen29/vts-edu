package admin_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/admin"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/go-chi/chi/v5"
)

type fakeService struct {
	listFunc          func(ctx context.Context, orgID string, opts admin.ListOptions) ([]admin.User, error)
	createFunc        func(ctx context.Context, orgID, actorID string, req admin.CreateUserRequest) (admin.User, error)
	updateRolesFunc   func(ctx context.Context, orgID, actorID, userID string, req admin.UpdateRolesRequest) error
	resetPasswordFunc func(ctx context.Context, orgID, actorID, userID string, req admin.ResetPasswordRequest) error
	getOrgFunc        func(ctx context.Context, orgID string) (admin.Organization, error)
	updateOrgFunc     func(ctx context.Context, orgID, actorID string, req admin.UpdateOrganizationRequest) (admin.Organization, error)
}

func (f *fakeService) ListUsers(ctx context.Context, orgID string, opts admin.ListOptions) ([]admin.User, error) {
	return f.listFunc(ctx, orgID, opts)
}

func (f *fakeService) CreateUser(ctx context.Context, orgID, actorID string, req admin.CreateUserRequest) (admin.User, error) {
	return f.createFunc(ctx, orgID, actorID, req)
}

func (f *fakeService) UpdateRoles(ctx context.Context, orgID, actorID, userID string, req admin.UpdateRolesRequest) error {
	return f.updateRolesFunc(ctx, orgID, actorID, userID, req)
}

func (f *fakeService) ResetPassword(ctx context.Context, orgID, actorID, userID string, req admin.ResetPasswordRequest) error {
	return f.resetPasswordFunc(ctx, orgID, actorID, userID, req)
}

func (f *fakeService) GetOrganization(ctx context.Context, orgID string) (admin.Organization, error) {
	return f.getOrgFunc(ctx, orgID)
}

func (f *fakeService) UpdateOrganization(ctx context.Context, orgID, actorID string, req admin.UpdateOrganizationRequest) (admin.Organization, error) {
	return f.updateOrgFunc(ctx, orgID, actorID, req)
}

func tokenWithRoles(roles []string) string {
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", roles, 1, false)
	if err != nil {
		panic(err)
	}
	return token
}

func TestHandler_ListUsers_AdminAllowed(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts admin.ListOptions) ([]admin.User, error) {
			return []admin.User{
				{ID: "u1", LoginName: "admin001", Roles: []string{"admin"}},
			}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data []admin.User `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Data))
	}
}

func TestHandler_ListUsers_PaginationQuery(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts admin.ListOptions) ([]admin.User, error) {
			if opts.Query != "alice" {
				t.Errorf("query = %q, want alice", opts.Query)
			}
			if opts.Limit != 5 {
				t.Errorf("limit = %d, want 5", opts.Limit)
			}
			if opts.Offset != 10 {
				t.Errorf("offset = %d, want 10", opts.Offset)
			}
			return []admin.User{{ID: "u1", LoginName: "alice01"}}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?q=alice&limit=5&offset=10", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data []admin.User `json:"data"`
		Page *struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		} `json:"page"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Data))
	}
	if resp.Page == nil || resp.Page.Limit != 5 || resp.Page.Offset != 10 {
		t.Fatalf("expected page {limit:5 offset:10}, got %+v", resp.Page)
	}
}

func TestHandler_ListUsers_StudentForbidden(t *testing.T) {
	svc := &fakeService{}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"student"}))
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandler_CreateUser_OK(t *testing.T) {
	svc := &fakeService{
		createFunc: func(ctx context.Context, orgID, actorID string, req admin.CreateUserRequest) (admin.User, error) {
			return admin.User{
				ID:                 "new-id",
				LoginName:          req.LoginName,
				DisplayName:        req.DisplayName,
				Roles:              req.Roles,
				MustChangePassword: true,
			}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	body := strings.NewReader(`{"login_name":"newuser","display_name":"New User","temporary_password":"TempPass123!","roles":["student"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp struct {
		Data admin.User `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.LoginName != "newuser" {
		t.Errorf("login_name = %q, want newuser", resp.Data.LoginName)
	}
	if !resp.Data.MustChangePassword {
		t.Error("expected must_change_password = true")
	}
}

func TestHandler_CreateUser_DuplicateLogin(t *testing.T) {
	svc := &fakeService{
		createFunc: func(ctx context.Context, orgID, actorID string, req admin.CreateUserRequest) (admin.User, error) {
			return admin.User{}, admin.ErrDuplicateLogin
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	body := strings.NewReader(`{"login_name":"existing","display_name":"Existing","temporary_password":"TempPass123!","roles":["student"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestHandler_UpdateRoles_OK(t *testing.T) {
	svc := &fakeService{
		updateRolesFunc: func(ctx context.Context, orgID, actorID, userID string, req admin.UpdateRolesRequest) error {
			return nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	body := strings.NewReader(`{"roles":["student","teacher"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/u1/roles", body)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "u1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateRoles(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_UpdateRoles_UserNotFound(t *testing.T) {
	svc := &fakeService{
		updateRolesFunc: func(ctx context.Context, orgID, actorID, userID string, req admin.UpdateRolesRequest) error {
			return admin.ErrUserNotFound
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	body := strings.NewReader(`{"roles":["student"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/u1/roles", body)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "u1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateRoles(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandler_ResetPassword_OK(t *testing.T) {
	svc := &fakeService{
		resetPasswordFunc: func(ctx context.Context, orgID, actorID, userID string, req admin.ResetPasswordRequest) error {
			return nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	body := strings.NewReader(`{"temporary_password":"ResetPass123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/u1/reset-password", body)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", "u1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_GetOrganization_OK(t *testing.T) {
	svc := &fakeService{
		getOrgFunc: func(ctx context.Context, orgID string) (admin.Organization, error) {
			return admin.Organization{ID: orgID, Code: "school-a", Name: "Trường Demo"}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/current", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.GetOrganization(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data admin.Organization `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Code != "school-a" {
		t.Errorf("code = %q, want school-a", resp.Data.Code)
	}
}

func TestHandler_UpdateOrganization_OK(t *testing.T) {
	svc := &fakeService{
		updateOrgFunc: func(ctx context.Context, orgID, actorID string, req admin.UpdateOrganizationRequest) (admin.Organization, error) {
			return admin.Organization{ID: orgID, Code: "school-a", Name: req.Name}, nil
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	body := strings.NewReader(`{"name":"Trường Mới"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/current", body)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.UpdateOrganization(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Data admin.Organization `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Name != "Trường Mới" {
		t.Errorf("name = %q, want Trường Mới", resp.Data.Name)
	}
}

func TestHandler_ServiceError(t *testing.T) {
	svc := &fakeService{
		listFunc: func(ctx context.Context, orgID string, opts admin.ListOptions) ([]admin.User, error) {
			return nil, errors.New("boom")
		},
	}
	issuer := auth.NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	h := admin.NewHandler(svc, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithRoles([]string{"admin"}))
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
