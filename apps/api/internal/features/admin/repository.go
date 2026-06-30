package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	adminsqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/admin/sqlc"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the admin feature.
type Repository interface {
	ListUsers(ctx context.Context, orgID string, opts ListOptions) ([]User, error)
	CountUsers(ctx context.Context, orgID string, opts ListOptions) (int64, error)
	ListAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, error)
	CountAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) (int64, error)
	LoginExists(ctx context.Context, orgID, loginName string) (bool, error)
	CreateUser(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error)
	GetMembershipID(ctx context.Context, orgID, userID string) (string, error)
	ReplaceRoles(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error
	BumpAuthVersion(ctx context.Context, tx pgx.Tx, userID string) error
	GetLoginPasswordHash(ctx context.Context, tx pgx.Tx, userID, orgID string) (string, error)
	ResetPassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error
	RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error
	InsertAuditLog(ctx context.Context, tx pgx.Tx, p AuditLogParams) error
	GetOrganization(ctx context.Context, orgID string) (Organization, error)
	UpdateOrganization(ctx context.Context, tx pgx.Tx, orgID, name string) error

	ListPasswordHistory(ctx context.Context, userID string, limit int) ([]string, error)
	InsertPasswordHistory(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error
	DeleteOldPasswordHistory(ctx context.Context, tx pgx.Tx, userID string, keep int) error
}

type sqlcRepository struct {
	queries *adminsqlc.Queries
}

// NewRepository creates a new admin repository backed by generated sqlc queries.
// It preserves the existing Repository interface.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: adminsqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func decodeUserCursor(cursor string) (string, pgtype.UUID, error) {
	if cursor == "" {
		return "", pgtype.UUID{}, nil
	}
	c, err := pagination.Decode(cursor)
	if err != nil {
		return "", pgtype.UUID{}, err
	}
	id, err := toUUID(c.ID)
	if err != nil {
		return "", pgtype.UUID{}, pagination.ErrInvalidCursor
	}
	return c.Key, id, nil
}

func (r *sqlcRepository) ListUsers(ctx context.Context, orgID string, opts ListOptions) ([]User, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	key, cursorID, err := decodeUserCursor(opts.Cursor)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	rows, err := r.queries.ListUsers(ctx, adminsqlc.ListUsersParams{
		OrganizationID: orgUUID,
		SearchQuery:    opts.Query,
		CursorKey:      key,
		CursorID:       cursorID,
		PageOffset:     int32(opts.Offset),
		PageLimit:      int32(opts.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	users := make([]User, len(rows))
	for i, row := range rows {
		users[i] = User{
			ID:                 row.ID.String(),
			DisplayName:        row.DisplayName.String,
			Email:              row.Email.String,
			LoginName:          row.UsernameNormalized,
			Roles:              toStringSlice(row.ArrayAgg),
			MustChangePassword: row.MustChangePassword,
		}
	}
	return users, nil
}

func (r *sqlcRepository) CountUsers(ctx context.Context, orgID string, opts ListOptions) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}

	count, err := r.queries.CountUsers(ctx, adminsqlc.CountUsersParams{
		OrganizationID: orgUUID,
		SearchQuery:    opts.Query,
	})
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return []string{}
	}
	arr, ok := v.([]interface{})
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func decodeAuditCursor(cursor string) (string, pgtype.UUID, error) {
	if cursor == "" {
		return "", pgtype.UUID{}, nil
	}
	c, err := pagination.Decode(cursor)
	if err != nil {
		return "", pgtype.UUID{}, err
	}
	id, err := toUUID(c.ID)
	if err != nil {
		return "", pgtype.UUID{}, pagination.ErrInvalidCursor
	}
	return c.Key, id, nil
}

func (r *sqlcRepository) ListAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) ([]AuditLog, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	key, cursorID, err := decodeAuditCursor(opts.Cursor)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	rows, err := r.queries.ListAuditLogs(ctx, adminsqlc.ListAuditLogsParams{
		OrganizationID: orgUUID,
		ActionName:     opts.Action,
		ActorUserID:    opts.ActorUserID,
		FromTime:       opts.From,
		ToTime:         opts.To,
		CursorKey:      key,
		CursorID:       cursorID,
		PageOffset:     int32(opts.Offset),
		PageLimit:      int32(opts.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}

	logs := make([]AuditLog, len(rows))
	for i, row := range rows {
		logs[i] = AuditLog{
			ID:           row.ID.String(),
			ActorUserID:  row.ActorUserID.String(),
			Action:       row.Action,
			ResourceType: row.ResourceType.String,
			ResourceID:   row.ResourceID.String(),
			Before:       json.RawMessage(row.BeforeJson),
			After:        json.RawMessage(row.AfterJson),
			Metadata:     json.RawMessage(row.MetadataJson),
			CreatedAt:    row.CreatedAt.Time.Format(time.RFC3339),
		}
	}
	return logs, nil
}

func (r *sqlcRepository) CountAuditLogs(ctx context.Context, orgID string, opts AuditLogListOptions) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}

	count, err := r.queries.CountAuditLogs(ctx, adminsqlc.CountAuditLogsParams{
		OrganizationID: orgUUID,
		ActionName:     opts.Action,
		ActorUserID:    opts.ActorUserID,
		FromTime:       opts.From,
		ToTime:         opts.To,
	})
	if err != nil {
		return 0, fmt.Errorf("count audit logs: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) LoginExists(ctx context.Context, orgID, loginName string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	exists, err := r.queries.LoginExists(ctx, adminsqlc.LoginExistsParams{
		OrganizationID: orgUUID,
		Lower:          loginName,
	})
	if err != nil {
		return false, fmt.Errorf("check login exists: %w", err)
	}
	return exists, nil
}

func (r *sqlcRepository) CreateUser(ctx context.Context, tx pgx.Tx, orgID, displayName, email, loginName, passwordHash string, roles []string) (User, error) {
	q := r.queries.WithTx(tx)

	orgUUID, err := toUUID(orgID)
	if err != nil {
		return User{}, fmt.Errorf("invalid organization id: %w", err)
	}

	userID, err := q.CreateUser(ctx, adminsqlc.CreateUserParams{
		DisplayName: toText(displayName),
		Email:       toText(email),
	})
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	membershipID, err := q.CreateMembership(ctx, adminsqlc.CreateMembershipParams{
		OrganizationID: orgUUID,
		UserID:         userID,
	})
	if err != nil {
		return User{}, fmt.Errorf("insert membership: %w", err)
	}

	if err := q.CreateLoginName(ctx, adminsqlc.CreateLoginNameParams{
		OrganizationID: orgUUID,
		Lower:          loginName,
		UserID:         userID,
		PasswordHash:   toText(passwordHash),
	}); err != nil {
		return User{}, fmt.Errorf("insert login name: %w", err)
	}

	for _, role := range roles {
		if err := q.CreateRole(ctx, adminsqlc.CreateRoleParams{
			MembershipID: membershipID,
			Role:         role,
		}); err != nil {
			return User{}, fmt.Errorf("insert role: %w", err)
		}
	}

	return User{
		ID:                 userID.String(),
		DisplayName:        displayName,
		Email:              email,
		LoginName:          loginName,
		Roles:              roles,
		MustChangePassword: true,
	}, nil
}

func (r *sqlcRepository) GetMembershipID(ctx context.Context, orgID, userID string) (string, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return "", fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}
	id, err := r.queries.GetMembershipID(ctx, adminsqlc.GetMembershipIDParams{
		OrganizationID: orgUUID,
		UserID:         userUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get membership: %w", err)
	}
	return id.String(), nil
}

func (r *sqlcRepository) ReplaceRoles(ctx context.Context, tx pgx.Tx, membershipID string, roles []string) error {
	q := r.queries.WithTx(tx)
	mID, err := toUUID(membershipID)
	if err != nil {
		return fmt.Errorf("invalid membership id: %w", err)
	}
	if err := q.DeleteRoles(ctx, mID); err != nil {
		return fmt.Errorf("delete roles: %w", err)
	}
	for _, role := range roles {
		if err := q.CreateRole(ctx, adminsqlc.CreateRoleParams{
			MembershipID: mID,
			Role:         role,
		}); err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
	}
	return nil
}

func (r *sqlcRepository) BumpAuthVersion(ctx context.Context, tx pgx.Tx, userID string) error {
	userUUID, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	if err := r.queries.WithTx(tx).BumpAuthVersion(ctx, userUUID); err != nil {
		return fmt.Errorf("bump auth version: %w", err)
	}
	return nil
}

func (r *sqlcRepository) GetLoginPasswordHash(ctx context.Context, tx pgx.Tx, userID, orgID string) (string, error) {
	userUUID, err := toUUID(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return "", fmt.Errorf("invalid organization id: %w", err)
	}

	row, err := r.queries.WithTx(tx).GetLoginPasswordHash(ctx, adminsqlc.GetLoginPasswordHashParams{
		UserID:         userUUID,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get login password hash: %w", err)
	}
	return row.String, nil
}

func (r *sqlcRepository) ResetPassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
	q := r.queries.WithTx(tx)
	userUUID, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := q.ResetPassword(ctx, adminsqlc.ResetPasswordParams{
		PasswordHash:   toText(passwordHash),
		UserID:         userUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	if err := q.SetMustChangePassword(ctx, userUUID); err != nil {
		return fmt.Errorf("update user flags: %w", err)
	}
	return nil
}

func (r *sqlcRepository) RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error {
	userUUID, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	if err := r.queries.WithTx(tx).RevokeUserSessions(ctx, userUUID); err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

func (r *sqlcRepository) InsertAuditLog(ctx context.Context, tx pgx.Tx, p AuditLogParams) error {
	orgUUID, err := toUUID(p.OrganizationID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	actorUUID, err := toUUID(p.ActorUserID)
	if err != nil {
		return fmt.Errorf("invalid actor user id: %w", err)
	}
	resourceUUID, err := toUUID(p.ResourceID)
	if err != nil {
		return fmt.Errorf("invalid resource id: %w", err)
	}
	if err := r.queries.WithTx(tx).InsertAuditLog(ctx, adminsqlc.InsertAuditLogParams{
		OrganizationID: orgUUID,
		ActorUserID:    actorUUID,
		Action:         p.Action,
		ResourceType:   toText(p.ResourceType),
		ResourceID:     resourceUUID,
		BeforeJson:     p.BeforeJSON,
		AfterJson:      p.AfterJSON,
		MetadataJson:   p.MetadataJSON,
	}); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (r *sqlcRepository) GetOrganization(ctx context.Context, orgID string) (Organization, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Organization{}, fmt.Errorf("invalid organization id: %w", err)
	}
	row, err := r.queries.GetOrganization(ctx, orgUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrOrganizationNotFound
	}
	if err != nil {
		return Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return Organization{
		ID:   row.ID.String(),
		Code: row.Code,
		Name: row.Name,
	}, nil
}

func (r *sqlcRepository) UpdateOrganization(ctx context.Context, tx pgx.Tx, orgID, name string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).UpdateOrganization(ctx, adminsqlc.UpdateOrganizationParams{
		Name: name,
		ID:   orgUUID,
	})
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	if rows == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

func (r *sqlcRepository) ListPasswordHistory(ctx context.Context, userID string, limit int) ([]string, error) {
	userUUID, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	hashes, err := r.queries.ListPasswordHistory(ctx, adminsqlc.ListPasswordHistoryParams{
		UserID: userUUID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list password history: %w", err)
	}
	if hashes == nil {
		hashes = []string{}
	}
	return hashes, nil
}

func (r *sqlcRepository) InsertPasswordHistory(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error {
	userUUID, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	if err := r.queries.WithTx(tx).InsertPasswordHistory(ctx, adminsqlc.InsertPasswordHistoryParams{
		UserID:       userUUID,
		PasswordHash: passwordHash,
	}); err != nil {
		return fmt.Errorf("insert password history: %w", err)
	}
	return nil
}

func (r *sqlcRepository) DeleteOldPasswordHistory(ctx context.Context, tx pgx.Tx, userID string, keep int) error {
	userUUID, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	if err := r.queries.WithTx(tx).DeleteOldPasswordHistory(ctx, adminsqlc.DeleteOldPasswordHistoryParams{
		UserID: userUUID,
		Offset: int32(keep),
	}); err != nil {
		return fmt.Errorf("delete old password history: %w", err)
	}
	return nil
}
