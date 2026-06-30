package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LoginIdentity is the data required to authenticate a login attempt.
type LoginIdentity struct {
	UserID       string
	MembershipID string
	OrgID        string
	Username     string
	PasswordHash string
	AuthVersion  int64
}

// ActorInfo is the public actor information returned by GET /api/v1/me.
type ActorInfo struct {
	UserID   string
	OrgID    string
	Username string
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
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new auth repository backed by a pgx connection pool.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) FindLoginByCredentials(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
	query := `
		SELECT
			u.id,
			m.id,
			o.id,
			ln.username_normalized,
			ln.password_hash,
			u.auth_version
		FROM membership_login_names ln
		JOIN organizations o ON o.id = ln.organization_id
		JOIN organization_memberships m
			ON m.organization_id = ln.organization_id AND m.user_id = ln.user_id
		JOIN users u ON u.id = ln.user_id
		WHERE lower(o.code) = $1
		  AND lower(ln.username_normalized) = $2
		  AND o.status = 'ACTIVE'
		  AND m.status = 'ACTIVE'
		  AND ln.status = 'ACTIVE'
		LIMIT 1
	`

	var id LoginIdentity
	err := r.pool.QueryRow(ctx, query, orgCode, username).Scan(
		&id.UserID,
		&id.MembershipID,
		&id.OrgID,
		&id.Username,
		&id.PasswordHash,
		&id.AuthVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find login: %w", err)
	}
	return &id, nil
}

func (r *repository) InsertRefreshSession(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error) {
	query := `
		INSERT INTO refresh_sessions (
			user_id,
			membership_id,
			organization_id,
			token_hash,
			family_id,
			auth_version,
			device_metadata_json,
			expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, '{}', $7)
		RETURNING id
	`

	var id string
	err := tx.QueryRow(ctx, query,
		p.UserID,
		p.MembershipID,
		p.OrgID,
		p.TokenHash,
		p.FamilyID,
		p.AuthVersion,
		p.ExpiresAt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert refresh session: %w", err)
	}
	return id, nil
}

func (r *repository) GetActorByUserID(ctx context.Context, userID, orgID string) (*ActorInfo, error) {
	query := `
		SELECT user_id, organization_id, username_normalized
		FROM membership_login_names
		WHERE user_id = $1
		  AND organization_id = $2
		  AND status = 'ACTIVE'
		LIMIT 1
	`

	var a ActorInfo
	err := r.pool.QueryRow(ctx, query, userID, orgID).Scan(&a.UserID, &a.OrgID, &a.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get actor: %w", err)
	}
	return &a, nil
}

func (r *repository) GetRefreshSessionWithContext(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error) {
	query := `
		SELECT
			rs.id,
			rs.user_id,
			rs.membership_id,
			rs.organization_id,
			rs.family_id,
			rs.auth_version,
			rs.expires_at,
			rs.revoked_at,
			rs.replaced_by_token_hash
		FROM refresh_sessions rs
		JOIN organizations o ON o.id = rs.organization_id
		JOIN organization_memberships m ON m.id = rs.membership_id
		JOIN users u ON u.id = rs.user_id
		WHERE rs.token_hash = $1
		  AND o.status = 'ACTIVE'
		  AND m.status = 'ACTIVE'
		  AND u.status = 'ACTIVE'
		  AND u.auth_version = rs.auth_version
		FOR UPDATE
	`

	var s RefreshSession
	err := tx.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID,
		&s.UserID,
		&s.MembershipID,
		&s.OrgID,
		&s.FamilyID,
		&s.AuthVersion,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.ReplacedByTokenHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh session: %w", err)
	}
	return &s, nil
}

func (r *repository) MarkSessionReplaced(ctx context.Context, tx pgx.Tx, sessionID, replacedByTokenHash string) error {
	query := `
		UPDATE refresh_sessions
		SET replaced_by_token_hash = $2
		WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, sessionID, replacedByTokenHash)
	if err != nil {
		return fmt.Errorf("mark session replaced: %w", err)
	}
	return nil
}

func (r *repository) RevokeSession(ctx context.Context, tx pgx.Tx, sessionID string) error {
	query := `
		UPDATE refresh_sessions
		SET revoked_at = now()
		WHERE id = $1
		  AND revoked_at IS NULL
	`
	_, err := tx.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *repository) RevokeFamily(ctx context.Context, tx pgx.Tx, familyID string) error {
	query := `
		UPDATE refresh_sessions
		SET revoked_at = now()
		WHERE family_id = $1
		  AND revoked_at IS NULL
	`
	_, err := tx.Exec(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("revoke family: %w", err)
	}
	return nil
}

func (r *repository) FindRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	query := `
		SELECT
			id,
			user_id,
			membership_id,
			organization_id,
			family_id,
			auth_version,
			expires_at,
			revoked_at,
			replaced_by_token_hash
		FROM refresh_sessions
		WHERE token_hash = $1
		LIMIT 1
	`

	var s RefreshSession
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID,
		&s.UserID,
		&s.MembershipID,
		&s.OrgID,
		&s.FamilyID,
		&s.AuthVersion,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.ReplacedByTokenHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh session: %w", err)
	}
	return &s, nil
}
