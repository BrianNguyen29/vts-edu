package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Service errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrActorNotFound      = errors.New("actor not found")
	ErrUnauthorized       = errors.New("unauthorized")
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// Service is the auth application service contract.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResult, error)
	Refresh(ctx context.Context, rawRefreshToken string) (*RefreshResult, error)
	Logout(ctx context.Context, rawRefreshToken string) (*LogoutResult, error)
	Me(ctx context.Context, token string) (*MeResult, error)
	ChangePassword(ctx context.Context, token string, req ChangePasswordRequest) error
}

type service struct {
	repo       Repository
	tm         TransactionManager
	issuer     *TokenIssuer
	refreshTTL time.Duration
}

// NewService creates the concrete auth service.
func NewService(repo Repository, tm TransactionManager, issuer *TokenIssuer, refreshTTL time.Duration) Service {
	return &service{
		repo:       repo,
		tm:         tm,
		issuer:     issuer,
		refreshTTL: refreshTTL,
	}
}

// Login authenticates a user and issues an access token plus a refresh session.
func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	orgCode := normalizeIdentifier(req.OrganizationCode)
	username := normalizeIdentifier(req.Username)

	if orgCode == "" || username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	identity, err := s.repo.FindLoginByCredentials(ctx, orgCode, username)
	if err != nil {
		return nil, err
	}

	ok, err := VerifyPassword(identity.PasswordHash, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	rawRefresh, err := randomURLToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	tokenHash := hashToken(rawRefresh)

	familyID, err := randomUUID()
	if err != nil {
		return nil, fmt.Errorf("generate family id: %w", err)
	}

	refreshExpires := time.Now().UTC().Add(s.refreshTTL)

	var sessionID string
	if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		id, err := s.repo.InsertRefreshSession(ctx, tx, InsertRefreshSessionParams{
			UserID:       identity.UserID,
			MembershipID: identity.MembershipID,
			OrgID:        identity.OrgID,
			TokenHash:    tokenHash,
			FamilyID:     familyID,
			AuthVersion:  identity.AuthVersion,
			ExpiresAt:    refreshExpires,
		})
		if err != nil {
			return err
		}
		sessionID = id
		return nil
	}); err != nil {
		return nil, fmt.Errorf("create refresh session: %w", err)
	}

	accessToken, _, err := s.issuer.IssueAccessToken(
		identity.UserID,
		identity.OrgID,
		sessionID,
		identity.Roles,
		identity.AuthVersion,
		identity.MustChangePassword,
	)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	return &LoginResult{
		AccessToken:        accessToken,
		ExpiresIn:          int(s.issuer.ttl.Seconds()),
		RefreshToken:       rawRefresh,
		RefreshExpires:     refreshExpires,
		User:               UserInfo{ID: identity.UserID, DisplayName: identity.Username},
		Roles:              identity.Roles,
		Permissions:        permissionsForRoles(identity.Roles),
		MustChangePassword: identity.MustChangePassword,
	}, nil
}

// Refresh rotates an active refresh session into a new session in the same family.
func (s *service) Refresh(ctx context.Context, rawRefreshToken string) (*RefreshResult, error) {
	if rawRefreshToken == "" {
		return nil, ErrUnauthorized
	}
	oldHash := hashToken(rawRefreshToken)

	newRaw, err := randomURLToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	newHash := hashToken(newRaw)

	refreshExpires := time.Now().UTC().Add(s.refreshTTL)

	var result *RefreshResult
	if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sess, err := s.repo.GetRefreshSessionWithContext(ctx, tx, oldHash)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		if sess.RevokedAt != nil || sess.ExpiresAt.Before(now) || (sess.ReplacedByTokenHash != nil && *sess.ReplacedByTokenHash != "") {
			_ = s.repo.RevokeFamily(ctx, tx, sess.FamilyID)
			return ErrUnauthorized
		}

		if err := s.repo.MarkSessionReplaced(ctx, tx, sess.ID, newHash); err != nil {
			return err
		}

		newSessionID, err := s.repo.InsertRefreshSession(ctx, tx, InsertRefreshSessionParams{
			UserID:       sess.UserID,
			MembershipID: sess.MembershipID,
			OrgID:        sess.OrgID,
			TokenHash:    newHash,
			FamilyID:     sess.FamilyID,
			AuthVersion:  sess.AuthVersion,
			ExpiresAt:    refreshExpires,
		})
		if err != nil {
			return err
		}

		actor, err := s.repo.GetActorByUserID(ctx, sess.UserID, sess.OrgID)
		if err != nil {
			return err
		}

		roles, err := s.repo.GetRolesByMembershipID(ctx, tx, sess.MembershipID)
		if err != nil {
			return err
		}

		accessToken, _, err := s.issuer.IssueAccessToken(
			sess.UserID,
			sess.OrgID,
			newSessionID,
			roles,
			sess.AuthVersion,
			actor.MustChangePassword,
		)
		if err != nil {
			return fmt.Errorf("issue access token: %w", err)
		}

		result = &RefreshResult{
			AccessToken:        accessToken,
			ExpiresIn:          int(s.issuer.ttl.Seconds()),
			RefreshToken:       newRaw,
			RefreshExpires:     refreshExpires,
			User:               UserInfo{ID: sess.UserID, DisplayName: actor.Username},
			Roles:              roles,
			Permissions:        permissionsForRoles(roles),
			MustChangePassword: actor.MustChangePassword,
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// Logout revokes the current refresh session if a cookie token is present.
func (s *service) Logout(ctx context.Context, rawRefreshToken string) (*LogoutResult, error) {
	if rawRefreshToken == "" {
		return &LogoutResult{Success: true}, nil
	}

	sess, err := s.repo.FindRefreshSessionByTokenHash(ctx, hashToken(rawRefreshToken))
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return &LogoutResult{Success: true}, nil
	}

	if err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.RevokeSession(ctx, tx, sess.ID)
	}); err != nil {
		return nil, err
	}

	return &LogoutResult{Success: true}, nil
}

// Me validates the access token and returns the current actor.
func (s *service) Me(ctx context.Context, token string) (*MeResult, error) {
	claims, err := s.issuer.ValidateAccessToken(token)
	if err != nil {
		return nil, ErrUnauthorized
	}

	actor, err := s.repo.GetActorByUserID(ctx, claims.Subject, claims.OrgID)
	if err != nil {
		return nil, err
	}

	return &MeResult{
		ID:                 claims.Subject,
		OrganizationID:     claims.OrgID,
		DisplayName:        actor.Username,
		Roles:              claims.Roles,
		Permissions:        permissionsForRoles(claims.Roles),
		MustChangePassword: actor.MustChangePassword,
	}, nil
}

// ChangePassword verifies the current password and sets a new one for the
// authenticated actor. It bumps the user's auth_version, clears the forced
// password-change flag, and revokes all refresh sessions for the user.
func (s *service) ChangePassword(ctx context.Context, token string, req ChangePasswordRequest) error {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return ErrInvalidCredentials
	}

	if err := ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	claims, err := s.issuer.ValidateAccessToken(token)
	if err != nil {
		return ErrUnauthorized
	}

	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		identity, err := s.repo.GetLoginByUserID(ctx, tx, claims.Subject, claims.OrgID)
		if err != nil {
			return err
		}

		ok, err := VerifyPassword(identity.PasswordHash, req.CurrentPassword)
		if err != nil || !ok {
			return ErrInvalidCredentials
		}

		if req.NewPassword == req.CurrentPassword {
			return ErrPasswordUnchanged
		}

		newHash, err := HashPassword(req.NewPassword)
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}

		if err := s.repo.UpdatePassword(ctx, tx, identity.UserID, identity.OrgID, newHash); err != nil {
			return err
		}

		if err := s.repo.RevokeUserSessions(ctx, tx, identity.UserID); err != nil {
			return err
		}

		return nil
	})
}

func normalizeIdentifier(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func randomURLToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(h[:])
}

func randomUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Version 4 UUID bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

var rolePermissions = map[string][]string{
	"student": {"attempt:read", "attempt:write", "self:read"},
	"teacher": {"assessment:read", "attempt:read"},
	"admin":   {"user:write", "org:write"},
}

func permissionsForRoles(roles []string) []string {
	seen := make(map[string]struct{})
	for _, role := range roles {
		for _, perm := range rolePermissions[role] {
			seen[perm] = struct{}{}
		}
	}

	perms := make([]string, 0, len(seen))
	for perm := range seen {
		perms = append(perms, perm)
	}
	sort.Strings(perms)
	return perms
}
