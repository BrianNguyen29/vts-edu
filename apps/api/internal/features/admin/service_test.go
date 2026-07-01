package admin

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/jackc/pgx/v5"
)

type fakeRepository struct {
	listFunc                     func(ctx context.Context, orgID string, opts ListOptions) ([]User, error)
	countUsersFunc               func(ctx context.Context, orgID string, opts ListOptions) (int64, error)
	listAuditFunc                func(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, error)
	countAuditFunc               func(ctx context.Context, orgID string, opts AuditLogListOptions) (int64, error)
	exportAuditFunc              func(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLogExport, error)
	loginExistsFunc              func(ctx context.Context, orgID, loginName string) (bool, error)
	createFunc                   func(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error)
	membershipFunc               func(ctx context.Context, orgID, userID string) (string, error)
	replaceRolesFunc             func(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error
	bumpAuthFunc                 func(ctx context.Context, tx pgx.Tx, userID string) error
	getLoginPasswordHashFunc     func(ctx context.Context, tx pgx.Tx, userID, orgID string) (string, error)
	resetPasswordFunc            func(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error
	revokeFunc                   func(ctx context.Context, tx pgx.Tx, userID string) error
	insertAuditFunc              func(ctx context.Context, tx pgx.Tx, p AuditLogParams) error
	getOrgFunc                   func(ctx context.Context, orgID string) (Organization, error)
	updateOrgFunc                func(ctx context.Context, tx pgx.Tx, orgID, name string) error
	listPasswordHistoryFunc      func(ctx context.Context, userID string, limit int) ([]string, error)
	insertPasswordHistoryFunc    func(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error
	deleteOldPasswordHistoryFunc func(ctx context.Context, tx pgx.Tx, userID string, keep int) error

	auditLogs []AuditLogParams
}

func (f *fakeRepository) ListUsers(ctx context.Context, orgID string, opts ListOptions) ([]User, error) {
	return f.listFunc(ctx, orgID, opts)
}

func (f *fakeRepository) CountUsers(ctx context.Context, orgID string, opts ListOptions) (int64, error) {
	if f.countUsersFunc != nil {
		return f.countUsersFunc(ctx, orgID, opts)
	}
	return 0, nil
}

func (f *fakeRepository) ListAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, error) {
	return f.listAuditFunc(ctx, orgID, opts)
}

func (f *fakeRepository) CountAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) (int64, error) {
	if f.countAuditFunc != nil {
		return f.countAuditFunc(ctx, orgID, opts)
	}
	return 0, nil
}

func (f *fakeRepository) ExportAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLogExport, error) {
	if f.exportAuditFunc != nil {
		return f.exportAuditFunc(ctx, orgID, opts)
	}
	return nil, nil
}

func (f *fakeRepository) LoginExists(ctx context.Context, orgID, loginName string) (bool, error) {
	return f.loginExistsFunc(ctx, orgID, loginName)
}

func (f *fakeRepository) CreateUser(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error) {
	return f.createFunc(ctx, tx, orgID, displayName, email, loginName, passwordHash, roles)
}

func (f *fakeRepository) GetMembershipID(ctx context.Context, orgID, userID string) (string, error) {
	return f.membershipFunc(ctx, orgID, userID)
}

func (f *fakeRepository) ReplaceRoles(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error {
	return f.replaceRolesFunc(ctx, tx, membershipID, roles)
}

func (f *fakeRepository) BumpAuthVersion(ctx context.Context, tx pgx.Tx, userID string) error {
	return f.bumpAuthFunc(ctx, tx, userID)
}

func (f *fakeRepository) GetLoginPasswordHash(ctx context.Context, tx pgx.Tx, userID, orgID string) (string, error) {
	if f.getLoginPasswordHashFunc != nil {
		return f.getLoginPasswordHashFunc(ctx, tx, userID, orgID)
	}
	return "", nil
}

func (f *fakeRepository) ResetPassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
	return f.resetPasswordFunc(ctx, tx, userID, orgID, passwordHash)
}

func (f *fakeRepository) RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error {
	return f.revokeFunc(ctx, tx, userID)
}

func (f *fakeRepository) InsertAuditLog(ctx context.Context, tx pgx.Tx, p AuditLogParams) error {
	if f.insertAuditFunc != nil {
		return f.insertAuditFunc(ctx, tx, p)
	}
	f.auditLogs = append(f.auditLogs, p)
	return nil
}

func (f *fakeRepository) GetOrganization(ctx context.Context, orgID string) (Organization, error) {
	return f.getOrgFunc(ctx, orgID)
}

func (f *fakeRepository) UpdateOrganization(ctx context.Context, tx pgx.Tx, orgID, name string) error {
	return f.updateOrgFunc(ctx, tx, orgID, name)
}

func (f *fakeRepository) ListPasswordHistory(ctx context.Context, userID string, limit int) ([]string, error) {
	if f.listPasswordHistoryFunc != nil {
		return f.listPasswordHistoryFunc(ctx, userID, limit)
	}
	return nil, nil
}

func (f *fakeRepository) InsertPasswordHistory(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error {
	if f.insertPasswordHistoryFunc != nil {
		return f.insertPasswordHistoryFunc(ctx, tx, userID, passwordHash)
	}
	return nil
}

func (f *fakeRepository) DeleteOldPasswordHistory(ctx context.Context, tx pgx.Tx, userID string, keep int) error {
	if f.deleteOldPasswordHistoryFunc != nil {
		return f.deleteOldPasswordHistoryFunc(ctx, tx, userID, keep)
	}
	return nil
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

func TestService_ListUsers(t *testing.T) {
	repo := &fakeRepository{
		listFunc: func(ctx context.Context, orgID string, opts ListOptions) ([]User, error) {
			if opts.Query != "alice" {
				t.Errorf("query = %q, want alice", opts.Query)
			}
			if opts.Limit != 6 {
				t.Errorf("repository limit = %d, want 6", opts.Limit)
			}
			if opts.Offset != 10 {
				t.Errorf("offset = %d, want 10", opts.Offset)
			}
			return []User{{ID: "u1", LoginName: "alice01"}}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	users, page, err := svc.ListUsers(context.Background(), "org-id", ListOptions{Query: "alice", Limit: 5, Offset: 10})
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 || users[0].LoginName != "alice01" {
		t.Errorf("users = %v", users)
	}
	if page == nil || page.Limit != 5 || page.Offset != 10 {
		t.Errorf("page = %+v", page)
	}
}

func TestService_ListAuditLogs(t *testing.T) {
	repo := &fakeRepository{
		listAuditFunc: func(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, error) {
			if opts.Action != "user.create" {
				t.Errorf("action = %q, want user.create", opts.Action)
			}
			if opts.Limit != 11 {
				t.Errorf("repository limit = %d, want 11", opts.Limit)
			}
			return []AuditLog{{ID: "log-1", Action: "user.create"}}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	logs, page, err := svc.ListAuditLogs(context.Background(), "org-id", AuditLogListOptions{Action: "user.create", Limit: 10})
	if err != nil {
		t.Fatalf("ListAuditLogs failed: %v", err)
	}
	if len(logs) != 1 || logs[0].Action != "user.create" {
		t.Errorf("logs = %v", logs)
	}
	if page == nil || page.Limit != 10 {
		t.Errorf("page = %+v", page)
	}
}

func TestService_CreateUser_OK(t *testing.T) {
	created := false
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error) {
			created = true
			return User{ID: "new-id", LoginName: loginName, DisplayName: displayName, Roles: roles, MustChangePassword: true}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	user, err := svc.CreateUser(context.Background(), "org-id", "actor-1", CreateUserRequest{
		LoginName:         "newuser",
		DisplayName:       "New User",
		TemporaryPassword: "TempPass123!",
		Roles:             []string{"student"},
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if !created {
		t.Fatal("expected repository CreateUser to be called")
	}
	if user.LoginName != "newuser" {
		t.Errorf("login_name = %q, want newuser", user.LoginName)
	}
	if len(user.Roles) != 1 || user.Roles[0] != "student" {
		t.Errorf("roles = %v, want [student]", user.Roles)
	}
	if !user.MustChangePassword {
		t.Error("expected must_change_password = true")
	}

	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != "user.create" {
		t.Errorf("expected audit log user.create, got %v", repo.auditLogs)
	}
}

func TestService_CreateUser_WeakPassword(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.CreateUser(context.Background(), "org-id", "actor-1", CreateUserRequest{
		LoginName:         "newuser",
		DisplayName:       "New User",
		TemporaryPassword: "password",
		Roles:             []string{"student"},
	})
	if !errors.Is(err, auth.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestService_CreateUser_DuplicateLogin(t *testing.T) {
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return true, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	_, err := svc.CreateUser(context.Background(), "org-id", "actor-1", CreateUserRequest{
		LoginName:         "existing",
		DisplayName:       "Existing",
		TemporaryPassword: "TempPass123!",
		Roles:             []string{"student"},
	})
	if !errors.Is(err, ErrDuplicateLogin) {
		t.Fatalf("expected ErrDuplicateLogin, got %v", err)
	}
}

func TestService_CreateUser_InvalidInput(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	_, err := svc.CreateUser(context.Background(), "org-id", "actor-1", CreateUserRequest{
		LoginName:         "",
		DisplayName:       "Missing Login",
		TemporaryPassword: "TempPass123!",
		Roles:             []string{"student"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_CreateUser_InvalidRole(t *testing.T) {
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	_, err := svc.CreateUser(context.Background(), "org-id", "actor-1", CreateUserRequest{
		LoginName:         "newuser",
		DisplayName:       "New User",
		TemporaryPassword: "TempPass123!",
		Roles:             []string{"superuser"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_UpdateRoles_OK(t *testing.T) {
	replaced := false
	revoked := false
	bumped := false
	repo := &fakeRepository{
		membershipFunc: func(ctx context.Context, orgID, userID string) (string, error) {
			return "membership-id", nil
		},
		replaceRolesFunc: func(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error {
			replaced = true
			return nil
		},
		bumpAuthFunc: func(ctx context.Context, tx pgx.Tx, userID string) error {
			bumped = true
			return nil
		},
		revokeFunc: func(ctx context.Context, tx pgx.Tx, userID string) error {
			revoked = true
			return nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	if err := svc.UpdateRoles(context.Background(), "org-id", "actor-1", "user-id", UpdateRolesRequest{Roles: []string{"teacher"}}); err != nil {
		t.Fatalf("UpdateRoles failed: %v", err)
	}
	if !replaced || !bumped || !revoked {
		t.Error("expected roles replaced, auth version bumped, and sessions revoked")
	}
	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != "user.update_roles" {
		t.Errorf("expected audit log user.update_roles, got %v", repo.auditLogs)
	}
}

func TestService_UpdateRoles_UserNotFound(t *testing.T) {
	repo := &fakeRepository{
		membershipFunc: func(ctx context.Context, orgID, userID string) (string, error) {
			return "", ErrUserNotFound
		},
	}
	svc := NewService(repo, stubTxManager{})
	err := svc.UpdateRoles(context.Background(), "org-id", "actor-1", "user-id", UpdateRolesRequest{Roles: []string{"teacher"}})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestService_ResetPassword_OK(t *testing.T) {
	reset := false
	revoked := false
	repo := &fakeRepository{
		resetPasswordFunc: func(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
			reset = true
			return nil
		},
		revokeFunc: func(ctx context.Context, tx pgx.Tx, userID string) error {
			revoked = true
			return nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	if err := svc.ResetPassword(context.Background(), "org-id", "actor-1", "user-id", ResetPasswordRequest{TemporaryPassword: "ResetPass123!"}); err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}
	if !reset || !revoked {
		t.Error("expected password reset and sessions revoked")
	}
	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != "user.reset_password" {
		t.Errorf("expected audit log user.reset_password, got %v", repo.auditLogs)
	}
}

func TestService_ResetPassword_MissingPassword(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	err := svc.ResetPassword(context.Background(), "org-id", "actor-1", "user-id", ResetPasswordRequest{TemporaryPassword: ""})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ResetPassword_WeakPassword(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	err := svc.ResetPassword(context.Background(), "org-id", "actor-1", "user-id", ResetPasswordRequest{TemporaryPassword: "password"})
	if !errors.Is(err, auth.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestService_ResetPassword_ReusedPassword(t *testing.T) {
	newPass := "ResetPass123!"
	newHash, err := auth.HashPassword(newPass)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	repo := &fakeRepository{
		listPasswordHistoryFunc: func(ctx context.Context, userID string, limit int) ([]string, error) {
			return []string{newHash}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	err = svc.ResetPassword(context.Background(), "org-id", "actor-1", "user-id", ResetPasswordRequest{TemporaryPassword: newPass})
	if !errors.Is(err, auth.ErrPasswordReused) {
		t.Fatalf("expected ErrPasswordReused, got %v", err)
	}
}

func TestService_CreateUser_StoresPasswordHistory(t *testing.T) {
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error) {
			return User{ID: "new-id", LoginName: loginName, DisplayName: displayName, Roles: roles, MustChangePassword: true}, nil
		},
	}

	inserted := false
	repo.insertPasswordHistoryFunc = func(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error {
		inserted = true
		return nil
	}

	svc := NewService(repo, stubTxManager{})
	_, err := svc.CreateUser(context.Background(), "org-id", "actor-1", CreateUserRequest{
		LoginName:         "newuser",
		DisplayName:       "New User",
		TemporaryPassword: "TempPass123!",
		Roles:             []string{"student"},
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if !inserted {
		t.Error("expected initial password hash to be stored in history")
	}
}

func TestService_UpdateOrganization_OK(t *testing.T) {
	updated := false
	repo := &fakeRepository{
		updateOrgFunc: func(ctx context.Context, tx pgx.Tx, orgID, name string) error {
			updated = true
			return nil
		},
		getOrgFunc: func(ctx context.Context, orgID string) (Organization, error) {
			return Organization{ID: orgID, Code: "school-a", Name: "New Name"}, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	org, err := svc.UpdateOrganization(context.Background(), "org-id", "actor-1", UpdateOrganizationRequest{Name: "New Name"})
	if err != nil {
		t.Fatalf("UpdateOrganization failed: %v", err)
	}
	if !updated {
		t.Error("expected organization update")
	}
	if org.Name != "New Name" {
		t.Errorf("name = %q, want New Name", org.Name)
	}
	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != "organization.update" {
		t.Errorf("expected audit log organization.update, got %v", repo.auditLogs)
	}
}

func TestService_ImportUsers_DryRun(t *testing.T) {
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	csv := "login_name,display_name,email,temporary_password,roles\n" +
		"bulk01,Bulk One,bulk1@example.com,TempPass123!,student\n" +
		"bulk02,Bulk Two,bulk2@example.com,TempPass123!,teacher\n" +
		"bulk03,,bulk3@example.com,TempPass123!,student\n"

	result, err := svc.ImportUsers(context.Background(), "org-id", "actor-1", ImportUsersRequest{CSV: csv, DryRun: true})
	if err != nil {
		t.Fatalf("ImportUsers failed: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("total = %d, want 3", result.Total)
	}
	if result.Created != 0 {
		t.Errorf("created = %d, want 0 on dry run", result.Created)
	}
	if result.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Failed)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(result.Rows))
	}
	if result.Rows[0].Status != "valid" {
		t.Errorf("row 1 status = %q, want valid", result.Rows[0].Status)
	}
	if result.Rows[2].Status != "error" {
		t.Errorf("row 3 status = %q, want error", result.Rows[2].Status)
	}
	if len(repo.auditLogs) != 0 {
		t.Errorf("expected no audit logs on dry run, got %d", len(repo.auditLogs))
	}
}

func TestService_ImportUsers_Confirm(t *testing.T) {
	created := 0
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error) {
			created++
			return User{ID: fmt.Sprintf("user-%d", created), LoginName: loginName, DisplayName: displayName, Roles: roles, MustChangePassword: true}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	csv := "login_name,display_name,email,temporary_password,roles\n" +
		"bulk01,Bulk One,bulk1@example.com,TempPass123!,student\n"

	result, err := svc.ImportUsers(context.Background(), "org-id", "actor-1", ImportUsersRequest{CSV: csv, DryRun: false})
	if err != nil {
		t.Fatalf("ImportUsers failed: %v", err)
	}
	if result.Created != 1 {
		t.Errorf("created = %d, want 1", result.Created)
	}
	if len(result.Rows) != 1 || result.Rows[0].Status != "created" {
		t.Errorf("row status = %q, want created", result.Rows[0].Status)
	}
	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != "user.import" {
		t.Errorf("expected user.import audit log, got %v", repo.auditLogs)
	}
}

func TestService_ImportUsers_TooManyRows(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo, stubTxManager{})
	csv := "login_name,display_name,email,temporary_password,roles\n"
	for i := 0; i <= maxBulkRows; i++ {
		csv += fmt.Sprintf("u%d,User %d,u%d@example.com,TempPass123!,student\n", i, i, i)
	}
	_, err := svc.ImportUsers(context.Background(), "org-id", "actor-1", ImportUsersRequest{CSV: csv})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
