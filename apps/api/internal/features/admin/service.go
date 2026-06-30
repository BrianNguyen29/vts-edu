package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/jackc/pgx/v5"
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// Service is the admin application service contract.
type Service interface {
	ListUsers(ctx context.Context, orgID string) ([]User, error)
	CreateUser(ctx context.Context, orgID string, req CreateUserRequest) (User, error)
	UpdateRoles(ctx context.Context, orgID, userID string, req UpdateRolesRequest) error
	ResetPassword(ctx context.Context, orgID, userID string, req ResetPasswordRequest) error
	GetOrganization(ctx context.Context, orgID string) (Organization, error)
	UpdateOrganization(ctx context.Context, orgID string, req UpdateOrganizationRequest) (Organization, error)
}

type service struct {
	repo Repository
	tm   TransactionManager
}

// NewService creates the concrete admin service.
func NewService(repo Repository, tm TransactionManager) Service {
	return &service{repo: repo, tm: tm}
}

func (s *service) ListUsers(ctx context.Context, orgID string) ([]User, error) {
	return s.repo.ListUsers(ctx, orgID)
}

func (s *service) CreateUser(ctx context.Context, orgID string, req CreateUserRequest) (User, error) {
	loginName := strings.ToLower(strings.TrimSpace(req.LoginName))
	displayName := strings.TrimSpace(req.DisplayName)
	email := strings.TrimSpace(req.Email)

	if loginName == "" || displayName == "" || req.TemporaryPassword == "" {
		return User{}, ErrInvalidInput
	}

	roles, err := normalizeRoles(req.Roles)
	if err != nil {
		return User{}, err
	}

	exists, err := s.repo.LoginExists(ctx, orgID, loginName)
	if err != nil {
		return User{}, err
	}
	if exists {
		return User{}, ErrDuplicateLogin
	}

	hash, err := auth.HashPassword(req.TemporaryPassword)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	var created User
	if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		u, err := s.repo.CreateUser(ctx, tx, orgID, displayName, email, loginName, hash, roles)
		if err != nil {
			return err
		}
		created = u
		return nil
	}); err != nil {
		return User{}, err
	}

	return created, nil
}

func (s *service) UpdateRoles(ctx context.Context, orgID, userID string, req UpdateRolesRequest) error {
	roles, err := normalizeRoles(req.Roles)
	if err != nil {
		return err
	}

	membershipID, err := s.repo.GetMembershipID(ctx, orgID, userID)
	if err != nil {
		return err
	}

	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.ReplaceRoles(ctx, tx, membershipID, roles); err != nil {
			return err
		}
		if err := s.repo.BumpAuthVersion(ctx, tx, userID); err != nil {
			return err
		}
		return s.repo.RevokeUserSessions(ctx, tx, userID)
	})
}

func (s *service) ResetPassword(ctx context.Context, orgID, userID string, req ResetPasswordRequest) error {
	if req.TemporaryPassword == "" {
		return ErrInvalidInput
	}

	hash, err := auth.HashPassword(req.TemporaryPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.ResetPassword(ctx, tx, userID, orgID, hash); err != nil {
			return err
		}
		return s.repo.RevokeUserSessions(ctx, tx, userID)
	})
}

func (s *service) GetOrganization(ctx context.Context, orgID string) (Organization, error) {
	return s.repo.GetOrganization(ctx, orgID)
}

func (s *service) UpdateOrganization(ctx context.Context, orgID string, req UpdateOrganizationRequest) (Organization, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Organization{}, ErrInvalidInput
	}

	if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.UpdateOrganization(ctx, tx, orgID, name)
	}); err != nil {
		return Organization{}, err
	}

	return s.repo.GetOrganization(ctx, orgID)
}

func normalizeRoles(roles []string) ([]string, error) {
	if len(roles) == 0 {
		return nil, ErrInvalidInput
	}

	allowed := map[string]bool{"student": true, "teacher": true, "admin": true}
	seen := map[string]bool{}
	out := make([]string, 0, len(roles))

	for _, r := range roles {
		role := strings.ToLower(strings.TrimSpace(r))
		if !allowed[role] {
			return nil, ErrInvalidInput
		}
		if seen[role] {
			continue
		}
		seen[role] = true
		out = append(out, role)
	}

	return out, nil
}
