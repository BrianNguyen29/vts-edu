package admin

import (
	"context"
	"encoding/csv"
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
	ExportAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLogExport, error)
	CreateUser(ctx context.Context, orgID, actorID string, req CreateUserRequest) (User, error)
	UpdateRoles(ctx context.Context, orgID, actorID, userID string, req UpdateRolesRequest) error
	ResetPassword(ctx context.Context, orgID, actorID, userID string, req ResetPasswordRequest) error
	GetOrganization(ctx context.Context, orgID string) (Organization, error)
	UpdateOrganization(ctx context.Context, orgID, actorID string, req UpdateOrganizationRequest) (Organization, error)

	ImportUsers(ctx context.Context, orgID, actorID string, req ImportUsersRequest) (ImportUsersResult, error)
}

const maxBulkRows = 100

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

func (s *service) ExportAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLogExport, error) {
	return s.repo.ExportAuditLogs(ctx, orgID, opts)
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

type userImportRow struct {
	rowNumber    int
	loginName    string
	displayName  string
	email        string
	tempPassword string
	rawRoles     string
}

func parseUserImportCSV(csvText string) ([]userImportRow, error) {
	reader := csv.NewReader(strings.NewReader(csvText))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid csv: %w", err)
	}
	if len(records) == 0 {
		return nil, ErrInvalidInput
	}

	expected := []string{"login_name", "display_name", "email", "temporary_password", "roles"}
	header := records[0]
	if len(header) < len(expected) || !csvHeaderMatches(header, expected) {
		return nil, ErrInvalidInput
	}

	data := records[1:]
	if len(data) == 0 {
		return nil, ErrInvalidInput
	}
	if len(data) > maxBulkRows {
		return nil, ErrInvalidInput
	}

	rows := make([]userImportRow, len(data))
	for i, rec := range data {
		for len(rec) < len(expected) {
			rec = append(rec, "")
		}
		rows[i] = userImportRow{
			rowNumber:    i + 2,
			loginName:    rec[0],
			displayName:  rec[1],
			email:        rec[2],
			tempPassword: rec[3],
			rawRoles:     rec[4],
		}
	}
	return rows, nil
}

func csvHeaderMatches(header, expected []string) bool {
	for i, want := range expected {
		if i >= len(header) {
			return false
		}
		if strings.ToLower(strings.TrimSpace(header[i])) != want {
			return false
		}
	}
	return true
}

func parseRoles(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (s *service) ImportUsers(ctx context.Context, orgID, actorID string, req ImportUsersRequest) (ImportUsersResult, error) {
	rows, err := parseUserImportCSV(req.CSV)
	if err != nil {
		return ImportUsersResult{}, err
	}

	result := ImportUsersResult{
		Total:  len(rows),
		DryRun: req.DryRun,
		Rows:   make([]ImportUserRow, len(rows)),
	}

	createdCount := 0
	failedCount := 0
	for i, row := range rows {
		loginName := strings.ToLower(strings.TrimSpace(row.loginName))
		displayName := strings.TrimSpace(row.displayName)
		email := strings.TrimSpace(row.email)
		tempPassword := row.tempPassword
		roles, roleErr := normalizeRoles(parseRoles(row.rawRoles))

		var rowErr error
		switch {
		case loginName == "" || displayName == "" || tempPassword == "":
			rowErr = ErrInvalidInput
		case roleErr != nil:
			rowErr = roleErr
		default:
			if err := auth.ValidatePasswordStrength(tempPassword); err != nil {
				rowErr = err
			}
		}

		if rowErr == nil {
			exists, err := s.repo.LoginExists(ctx, orgID, loginName)
			if err != nil {
				return result, err
			}
			if exists {
				rowErr = ErrDuplicateLogin
			}
		}

		var createdUserID string
		if rowErr == nil && !req.DryRun {
			hash, err := auth.HashPassword(tempPassword)
			if err != nil {
				return result, fmt.Errorf("hash password: %w", err)
			}

			var created User
			err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
				u, err := s.repo.CreateUser(ctx, tx, orgID, displayName, email, loginName, hash, roles)
				if err != nil {
					return err
				}
				created = u
				return auth.StorePasswordHistory(ctx, s.repo, tx, u.ID, "", hash)
			})
			if err != nil {
				rowErr = err
			} else {
				createdUserID = created.ID
			}
		}

		status := "created"
		if rowErr != nil {
			status = "error"
			failedCount++
		} else if req.DryRun {
			status = "valid"
		} else {
			createdCount++
		}

		errMsg := ""
		if rowErr != nil {
			errMsg = rowErr.Error()
		}
		result.Rows[i] = ImportUserRow{
			RowNumber: row.rowNumber,
			UserID:    createdUserID,
			LoginName: loginName,
			Status:    status,
			Error:     errMsg,
		}
	}

	result.Created = createdCount
	result.Failed = failedCount

	if !req.DryRun {
		after, _ := json.Marshal(map[string]any{
			"created": createdCount,
			"failed":  failedCount,
			"total":   len(rows),
		})
		meta, _ := json.Marshal(map[string]any{
			"dry_run": false,
		})
		if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			return s.repo.InsertAuditLog(ctx, tx, AuditLogParams{
				OrganizationID: orgID,
				ActorUserID:    actorID,
				Action:         "user.import",
				ResourceType:   "organization",
				ResourceID:     orgID,
				AfterJSON:      after,
				MetadataJSON:   meta,
			})
		}); err != nil {
			return result, fmt.Errorf("audit log: %w", err)
		}
	}

	return result, nil
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
