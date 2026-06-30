package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeRepository struct {
	listFunc          func(ctx context.Context, orgID string) ([]User, error)
	loginExistsFunc   func(ctx context.Context, orgID, loginName string) (bool, error)
	createFunc        func(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error)
	membershipFunc    func(ctx context.Context, orgID, userID string) (string, error)
	replaceRolesFunc  func(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error
	bumpAuthFunc      func(ctx context.Context, tx pgx.Tx, userID string) error
	resetPasswordFunc func(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error
	revokeFunc        func(ctx context.Context, tx pgx.Tx, userID string) error
	getOrgFunc        func(ctx context.Context, orgID string) (Organization, error)
	updateOrgFunc     func(ctx context.Context, tx pgx.Tx, orgID, name string) error
}

func (f *fakeRepository) ListUsers(ctx context.Context, orgID string) ([]User, error) {
	return f.listFunc(ctx, orgID)
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

func (f *fakeRepository) ResetPassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
	return f.resetPasswordFunc(ctx, tx, userID, orgID, passwordHash)
}

func (f *fakeRepository) RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error {
	return f.revokeFunc(ctx, tx, userID)
}

func (f *fakeRepository) GetOrganization(ctx context.Context, orgID string) (Organization, error) {
	return f.getOrgFunc(ctx, orgID)
}

func (f *fakeRepository) UpdateOrganization(ctx context.Context, tx pgx.Tx, orgID, name string) error {
	return f.updateOrgFunc(ctx, tx, orgID, name)
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
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
	user, err := svc.CreateUser(context.Background(), "org-id", CreateUserRequest{
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
}

func TestService_CreateUser_DuplicateLogin(t *testing.T) {
	repo := &fakeRepository{
		loginExistsFunc: func(ctx context.Context, orgID, loginName string) (bool, error) {
			return true, nil
		},
	}

	svc := NewService(repo, stubTxManager{})
	_, err := svc.CreateUser(context.Background(), "org-id", CreateUserRequest{
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
	_, err := svc.CreateUser(context.Background(), "org-id", CreateUserRequest{
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
	_, err := svc.CreateUser(context.Background(), "org-id", CreateUserRequest{
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
	if err := svc.UpdateRoles(context.Background(), "org-id", "user-id", UpdateRolesRequest{Roles: []string{"teacher"}}); err != nil {
		t.Fatalf("UpdateRoles failed: %v", err)
	}
	if !replaced || !bumped || !revoked {
		t.Error("expected roles replaced, auth version bumped, and sessions revoked")
	}
}

func TestService_UpdateRoles_UserNotFound(t *testing.T) {
	repo := &fakeRepository{
		membershipFunc: func(ctx context.Context, orgID, userID string) (string, error) {
			return "", ErrUserNotFound
		},
	}
	svc := NewService(repo, stubTxManager{})
	err := svc.UpdateRoles(context.Background(), "org-id", "user-id", UpdateRolesRequest{Roles: []string{"teacher"}})
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
	if err := svc.ResetPassword(context.Background(), "org-id", "user-id", ResetPasswordRequest{TemporaryPassword: "ResetPass123!"}); err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}
	if !reset || !revoked {
		t.Error("expected password reset and sessions revoked")
	}
}

func TestService_ResetPassword_MissingPassword(t *testing.T) {
	svc := NewService(&fakeRepository{}, stubTxManager{})
	err := svc.ResetPassword(context.Background(), "org-id", "user-id", ResetPasswordRequest{TemporaryPassword: ""})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
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
	org, err := svc.UpdateOrganization(context.Background(), "org-id", UpdateOrganizationRequest{Name: "New Name"})
	if err != nil {
		t.Fatalf("UpdateOrganization failed: %v", err)
	}
	if !updated {
		t.Error("expected organization update")
	}
	if org.Name != "New Name" {
		t.Errorf("name = %q, want New Name", org.Name)
	}
}
