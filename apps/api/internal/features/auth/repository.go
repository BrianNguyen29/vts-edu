package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	authsqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LoginIdentity is the data required to authenticate a login attempt.
type LoginIdentity struct {
	UserID             string
	MembershipID       string
	OrgID              string
	Username           string
	PasswordHash       string
	AuthVersion        int64
	MustChangePassword bool
	Roles              []string
}

// ActorInfo is the public actor information returned by GET /api/v1/me.
type ActorInfo struct {
	UserID             string
	OrgID              string
	Username           string
	MustChangePassword bool
}

// InsertRefreshSessionParams holds the fields needed to create a refresh session.
type InsertRefreshSessionParams struct {
	UserID       string
	MembershipID string
	OrgID        string
	TokenHash    string
	FamilyID     string
	AuthVersion  int64
	ExpiresAt    time.Time
}

// RefreshSession is a stored refresh session row.
type RefreshSession struct {
	ID                  string
	UserID              string
	MembershipID        string
	OrgID               string
	FamilyID            string
	AuthVersion         int64
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	ReplacedByTokenHash *string
}

// Repository defines the persistence contract for the auth feature.
type Repository interface {
	FindLoginByCredentials(ctx context.Context, orgCode, username string) (*LoginIdentity, error)
	InsertRefreshSession(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error)
	GetActorByUserID(ctx context.Context, userID, orgID string) (*ActorInfo, error)

	GetRefreshSessionWithContext(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error)
	MarkSessionReplaced(ctx context.Context, tx pgx.Tx, sessionID, replacedByTokenHash string) error
	RevokeSession(ctx context.Context, tx pgx.Tx, sessionID string) error
	RevokeFamily(ctx context.Context, tx pgx.Tx, familyID string) error
	FindRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (*RefreshSession, error)

	GetRolesByMembershipID(ctx context.Context, tx pgx.Tx, membershipID string) ([]string, error)
	GetLoginByUserID(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error)
	UpdatePassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error
	RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error

	CountRecentFailedLoginAttempts(ctx context.Context, orgID, username string, window time.Duration) (int64, error)
	RecordFailedLoginAttempt(ctx context.Context, orgID, username string) error
	ClearLoginAttempts(ctx context.Context, orgID, username string) error
	ListPasswordHistory(ctx context.Context, userID string, limit int) ([]string, error)
	InsertPasswordHistory(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error
	DeleteOldPasswordHistory(ctx context.Context, tx pgx.Tx, userID string, keep int) error
}

type sqlcRepository struct {
	queries *authsqlc.Queries
}

// NewRepository creates a new auth repository backed by generated sqlc queries.
// It preserves the existing Repository interface.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: authsqlc.New(pool)}
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

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
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

func (r *sqlcRepository) FindLoginByCredentials(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
	row, err := r.queries.FindLoginByCredentials(ctx, authsqlc.FindLoginByCredentialsParams{
		Lower:   orgCode,
		Lower_2: username,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find login: %w", err)
	}

	roles := toStringSlice(row.ArrayAgg)
	return &LoginIdentity{
		UserID:             row.ID.String(),
		MembershipID:       row.ID_2.String(),
		OrgID:              row.ID_3.String(),
		Username:           row.UsernameNormalized,
		PasswordHash:       row.PasswordHash.String,
		AuthVersion:        row.AuthVersion,
		MustChangePassword: row.MustChangePassword,
		Roles:              roles,
	}, nil
}

func (r *sqlcRepository) InsertRefreshSession(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error) {
	userUUID, err := toUUID(p.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}
	membershipUUID, err := toUUID(p.MembershipID)
	if err != nil {
		return "", fmt.Errorf("invalid membership id: %w", err)
	}
	orgUUID, err := toUUID(p.OrgID)
	if err != nil {
		return "", fmt.Errorf("invalid organization id: %w", err)
	}
	familyUUID, err := toUUID(p.FamilyID)
	if err != nil {
		return "", fmt.Errorf("invalid family id: %w", err)
	}

	id, err := r.queries.WithTx(tx).InsertRefreshSession(ctx, authsqlc.InsertRefreshSessionParams{
		UserID:         userUUID,
		MembershipID:   membershipUUID,
		OrganizationID: orgUUID,
		TokenHash:      p.TokenHash,
		FamilyID:       familyUUID,
		AuthVersion:    p.AuthVersion,
		ExpiresAt:      pgtype.Timestamptz{Time: p.ExpiresAt, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("insert refresh session: %w", err)
	}
	return id.String(), nil
}

func (r *sqlcRepository) GetActorByUserID(ctx context.Context, userID, orgID string) (*ActorInfo, error) {
	userUUID, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	row, err := r.queries.GetActorByUserID(ctx, authsqlc.GetActorByUserIDParams{
		UserID:         userUUID,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get actor: %w", err)
	}

	return &ActorInfo{
		UserID:             row.UserID.String(),
		OrgID:              row.OrganizationID.String(),
		Username:           row.UsernameNormalized,
		MustChangePassword: row.MustChangePassword,
	}, nil
}

func (r *sqlcRepository) GetRefreshSessionWithContext(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error) {
	row, err := r.queries.WithTx(tx).GetRefreshSessionWithContext(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh session: %w", err)
	}
	return toRefreshSession(row), nil
}

func (r *sqlcRepository) MarkSessionReplaced(ctx context.Context, tx pgx.Tx, sessionID, replacedByTokenHash string) error {
	sessionUUID, err := toUUID(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session id: %w", err)
	}
	if err := r.queries.WithTx(tx).MarkSessionReplaced(ctx, authsqlc.MarkSessionReplacedParams{
		ID:                  sessionUUID,
		ReplacedByTokenHash: toText(replacedByTokenHash),
	}); err != nil {
		return fmt.Errorf("mark session replaced: %w", err)
	}
	return nil
}

func (r *sqlcRepository) RevokeSession(ctx context.Context, tx pgx.Tx, sessionID string) error {
	sessionUUID, err := toUUID(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session id: %w", err)
	}
	if err := r.queries.WithTx(tx).RevokeSession(ctx, sessionUUID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *sqlcRepository) RevokeFamily(ctx context.Context, tx pgx.Tx, familyID string) error {
	familyUUID, err := toUUID(familyID)
	if err != nil {
		return fmt.Errorf("invalid family id: %w", err)
	}
	if err := r.queries.WithTx(tx).RevokeFamily(ctx, familyUUID); err != nil {
		return fmt.Errorf("revoke family: %w", err)
	}
	return nil
}

func (r *sqlcRepository) FindRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	row, err := r.queries.FindRefreshSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh session: %w", err)
	}
	return &RefreshSession{
		ID:                  row.ID.String(),
		UserID:              row.UserID.String(),
		MembershipID:        row.MembershipID.String(),
		OrgID:               row.OrganizationID.String(),
		FamilyID:            row.FamilyID.String(),
		AuthVersion:         row.AuthVersion,
		ExpiresAt:           row.ExpiresAt.Time,
		RevokedAt:           tsPtr(row.RevokedAt),
		ReplacedByTokenHash: textPtr(row.ReplacedByTokenHash),
	}, nil
}

func toRefreshSession(row authsqlc.GetRefreshSessionWithContextRow) *RefreshSession {
	return &RefreshSession{
		ID:                  row.ID.String(),
		UserID:              row.UserID.String(),
		MembershipID:        row.MembershipID.String(),
		OrgID:               row.OrganizationID.String(),
		FamilyID:            row.FamilyID.String(),
		AuthVersion:         row.AuthVersion,
		ExpiresAt:           row.ExpiresAt.Time,
		RevokedAt:           tsPtr(row.RevokedAt),
		ReplacedByTokenHash: textPtr(row.ReplacedByTokenHash),
	}
}

func (r *sqlcRepository) GetRolesByMembershipID(ctx context.Context, tx pgx.Tx, membershipID string) ([]string, error) {
	membershipUUID, err := toUUID(membershipID)
	if err != nil {
		return nil, fmt.Errorf("invalid membership id: %w", err)
	}
	roles, err := r.queries.WithTx(tx).GetRolesByMembershipID(ctx, membershipUUID)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}
	if roles == nil {
		roles = []string{}
	}
	return roles, nil
}

func (r *sqlcRepository) GetLoginByUserID(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error) {
	userUUID, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	row, err := r.queries.WithTx(tx).GetLoginByUserID(ctx, authsqlc.GetLoginByUserIDParams{
		ID:   userUUID,
		ID_2: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get login by user id: %w", err)
	}

	roles := toStringSlice(row.ArrayAgg)
	return &LoginIdentity{
		UserID:             row.ID.String(),
		MembershipID:       row.ID_2.String(),
		OrgID:              row.ID_3.String(),
		Username:           row.UsernameNormalized,
		PasswordHash:       row.PasswordHash.String,
		AuthVersion:        row.AuthVersion,
		MustChangePassword: row.MustChangePassword,
		Roles:              roles,
	}, nil
}

func (r *sqlcRepository) UpdatePassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
	q := r.queries.WithTx(tx)

	userUUID, err := toUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}

	if err := q.BumpUserAuthVersion(ctx, userUUID); err != nil {
		return fmt.Errorf("update user auth version: %w", err)
	}

	rows, err := q.UpdateLoginPassword(ctx, authsqlc.UpdateLoginPasswordParams{
		PasswordHash:   toText(passwordHash),
		UserID:         userUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	if rows == 0 {
		return ErrActorNotFound
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

func (r *sqlcRepository) CountRecentFailedLoginAttempts(ctx context.Context, orgID, username string, window time.Duration) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	count, err := r.queries.CountFailedLoginAttempts(ctx, authsqlc.CountFailedLoginAttemptsParams{
		OrganizationID: orgUUID,
		Lower:          username,
		Column3:        pgtype.Interval{Microseconds: window.Microseconds(), Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("count failed login attempts: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) RecordFailedLoginAttempt(ctx context.Context, orgID, username string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	if err := r.queries.InsertFailedLoginAttempt(ctx, authsqlc.InsertFailedLoginAttemptParams{
		OrganizationID: orgUUID,
		Lower:          username,
	}); err != nil {
		return fmt.Errorf("record failed login attempt: %w", err)
	}
	return nil
}

func (r *sqlcRepository) ClearLoginAttempts(ctx context.Context, orgID, username string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	if err := r.queries.ClearLoginAttempts(ctx, authsqlc.ClearLoginAttemptsParams{
		OrganizationID: orgUUID,
		Lower:          username,
	}); err != nil {
		return fmt.Errorf("clear login attempts: %w", err)
	}
	return nil
}

func (r *sqlcRepository) ListPasswordHistory(ctx context.Context, userID string, limit int) ([]string, error) {
	userUUID, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	hashes, err := r.queries.ListPasswordHistory(ctx, authsqlc.ListPasswordHistoryParams{
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
	if err := r.queries.WithTx(tx).InsertPasswordHistory(ctx, authsqlc.InsertPasswordHistoryParams{
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
	if err := r.queries.WithTx(tx).DeleteOldPasswordHistory(ctx, authsqlc.DeleteOldPasswordHistoryParams{
		UserID: userUUID,
		Offset: int32(keep),
	}); err != nil {
		return fmt.Errorf("delete old password history: %w", err)
	}
	return nil
}
