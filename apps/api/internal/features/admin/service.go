package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
	"github.com/jackc/pgx/v5"
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// Service is the admin application service contract.
type Service interface {
	ListUsers(ctx context.Context, orgID string, opts ListOptions) ([]User, *PageInfo, error)
	ListAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, *PageInfo, error)
	CreateUser(ctx context.Context, orgID, actorID string, req CreateUserRequest) (User, error)
	UpdateRoles(ctx context.Context, orgID, actorID, userID string, req UpdateRolesRequest) error
	ResetPassword(ctx context.Context, orgID, actorID, userID string, req ResetPasswordRequest) error
	GetOrganization(ctx context.Context, orgID string) (Organization, error)
	UpdateOrganization(ctx context.Context, orgID, actorID string, req UpdateOrganizationRequest) (Organization, error)
}

type service struct {
	repo Repository
	tm   TransactionManager
}

// NewService creates the concrete admin service.
func NewService(repo Repository, tm TransactionManager) Service {
	return &service{repo: repo, tm: tm}
}

func (s *service) ListUsers(ctx context.Context, orgID string, opts ListOptions) ([]User, *PageInfo, error) {
	queryOpts := opts
	if opts.Limit > 0 {
		queryOpts.Limit = opts.Limit + 1
	}

	users, err := s.repo.ListUsers(ctx, orgID, queryOpts)
	if err != nil {
		return nil, nil, err
	}

	page := &PageInfo{Limit: opts.Limit, Offset: opts.Offset}
	if opts.Limit > 0 {
		if len(users) > opts.Limit {
			page.HasMore = true
			last := users[opts.Limit-1]
			cursor := pagination.Encode(pagination.Cursor{Key: last.LoginName, ID: last.ID})
			page.NextCursor = &cursor
			users = users[:opts.Limit]
		}
	}

	if opts.Count {
		count, err := s.repo.CountUsers(ctx, orgID, opts)
		if err != nil {
			return nil, nil, err
		}
		page.TotalCount = &count
	}

	return users, page, nil
}

func (s *service) ListAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, *PageInfo, error) {
	queryOpts := opts
	if opts.Limit > 0 {
		queryOpts.Limit = opts.Limit + 1
	}

	logs, err := s.repo.ListAuditLogs(ctx, orgID, queryOpts)
	if err != nil {
		return nil, nil, err
	}

	page := &PageInfo{Limit: opts.Limit, Offset: opts.Offset}
	if opts.Limit > 0 {
		if len(logs) > opts.Limit {
			page.HasMore = true
			last := logs[opts.Limit-1]
			cursor := pagination.Encode(pagination.Cursor{Key: last.CreatedAt, ID: last.ID})
			page.NextCursor = &cursor
			logs = logs[:opts.Limit]
		}
	}

	if opts.Count {
		count, err := s.repo.CountAuditLogs(ctx, orgID, opts)
		if err != nil {
			return nil, nil, err
		}
		page.TotalCount = &count
	}

	return logs, page, nil
}

func (s *service) CreateUser(ctx context.Context, orgID, actorID string, req CreateUserRequest) (User, error) {
	loginName := strings.ToLower(strings.TrimSpace(req.LoginName))
	displayName := strings.TrimSpace(req.DisplayName)
	email := strings.TrimSpace(req.Email)

	if loginName == "" || displayName == "" || req.TemporaryPassword == "" {
		return User{}, ErrInvalidInput
	}

	if err := auth.ValidatePasswordStrength(req.TemporaryPassword); err != nil {
		return User{}, err
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

		if err := auth.StorePasswordHistory(ctx, s.repo, tx, u.ID, "", hash); err != nil {
			return err
		}

		after, _ := json.Marshal(map[string]any{
			"id":           u.ID,
			"login_name":   u.LoginName,
			"display_name": u.DisplayName,
			"roles":        u.Roles,
		})
		meta, _ := json.Marshal(map[string]any{
			"temporary_password": true,
		})
		return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
			OrganizationID: orgID,
			ActorUserID:    actorID,
			Action:         "user.create",
			ResourceType:   "user",
			ResourceID:     u.ID,
			AfterJSON:      after,
			MetadataJSON:   meta,
		})
	}); err != nil {
		return User{}, err
	}

	return created, nil
}

func (s *service) UpdateRoles(ctx context.Context, orgID, actorID, userID string, req UpdateRolesRequest) error {
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
		if err := s.repo.RevokeUserSessions(ctx, tx, userID); err != nil {
			return err
		}

		after, _ := json.Marshal(map[string]any{
			"roles": roles,
		})
		meta, _ := json.Marshal(map[string]any{
			"membership_id": membershipID,
		})
		return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
			OrganizationID: orgID,
			ActorUserID:    actorID,
			Action:         "user.update_roles",
			ResourceType:   "user",
			ResourceID:     userID,
			AfterJSON:      after,
			MetadataJSON:   meta,
		})
	})
}

func (s *service) ResetPassword(ctx context.Context, orgID, actorID, userID string, req ResetPasswordRequest) error {
	if req.TemporaryPassword == "" {
		return ErrInvalidInput
	}

	if err := auth.ValidatePasswordStrength(req.TemporaryPassword); err != nil {
		return err
	}

	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		oldHash, err := s.repo.GetLoginPasswordHash(ctx, tx, userID, orgID)
		if err != nil {
			return err
		}

		if err := auth.CheckPasswordHistory(ctx, s.repo, userID, req.TemporaryPassword); err != nil {
			return err
		}

		hash, err := auth.HashPassword(req.TemporaryPassword)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		if err := s.repo.ResetPassword(ctx, tx, userID, orgID, hash); err != nil {
			return err
		}

		if err := auth.StorePasswordHistory(ctx, s.repo, tx, userID, oldHash, hash); err != nil {
			return err
		}

		if err := s.repo.RevokeUserSessions(ctx, tx, userID); err != nil {
			return err
		}

		meta, _ := json.Marshal(map[string]any{
			"admin_initiated": true,
		})
		return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
			OrganizationID: orgID,
			ActorUserID:    actorID,
			Action:         "user.reset_password",
			ResourceType:   "user",
			ResourceID:     userID,
			MetadataJSON:   meta,
		})
	})
}

func (s *service) GetOrganization(ctx context.Context, orgID string) (Organization, error) {
	return s.repo.GetOrganization(ctx, orgID)
}

func (s *service) UpdateOrganization(ctx context.Context, orgID, actorID string, req UpdateOrganizationRequest) (Organization, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Organization{}, ErrInvalidInput
	}

	before, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return Organization{}, err
	}

	if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.UpdateOrganization(ctx, tx, orgID, name); err != nil {
			return err
		}

		beforeJSON, _ := json.Marshal(map[string]any{
			"name": before.Name,
		})
		afterJSON, _ := json.Marshal(map[string]any{
			"name": name,
		})
		return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
			OrganizationID: orgID,
			ActorUserID:    actorID,
			Action:         "organization.update",
			ResourceType:   "organization",
			ResourceID:     orgID,
			BeforeJSON:     beforeJSON,
			AfterJSON:      afterJSON,
		})
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
