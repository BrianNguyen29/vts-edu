package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type fakeRepository struct {
	findFunc                     func(ctx context.Context, orgCode, username string) (*LoginIdentity, error)
	insertFunc                   func(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error)
	actorFunc                    func(ctx context.Context, userID, orgID string) (*ActorInfo, error)
	refreshFunc                  func(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error)
	markReplaced                 func(ctx context.Context, tx pgx.Tx, sessionID, replacedByTokenHash string) error
	revokeSession                func(ctx context.Context, tx pgx.Tx, sessionID string) error
	revokeFamily                 func(ctx context.Context, tx pgx.Tx, familyID string) error
	findByHashFunc               func(ctx context.Context, tokenHash string) (*RefreshSession, error)
	rolesFunc                    func(ctx context.Context, tx pgx.Tx, membershipID string) ([]string, error)
	loginByUserIDFunc            func(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error)
	updatePasswordFunc           func(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error
	revokeUserSessionsFunc       func(ctx context.Context, tx pgx.Tx, userID string) error
	countFailedLoginAttemptsFunc func(ctx context.Context, orgID, username string, window time.Duration) (int64, error)
	recordFailedLoginAttemptFunc func(ctx context.Context, orgID, username string) error
	clearLoginAttemptsFunc       func(ctx context.Context, orgID, username string) error
	listPasswordHistoryFunc      func(ctx context.Context, userID string, limit int) ([]string, error)
	insertPasswordHistoryFunc    func(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error
	deleteOldPasswordHistoryFunc func(ctx context.Context, tx pgx.Tx, userID string, keep int) error
}

func (f *fakeRepository) FindLoginByCredentials(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
	return f.findFunc(ctx, orgCode, username)
}

func (f *fakeRepository) InsertRefreshSession(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error) {
	return f.insertFunc(ctx, tx, p)
}

func (f *fakeRepository) GetActorByUserID(ctx context.Context, userID, orgID string) (*ActorInfo, error) {
	return f.actorFunc(ctx, userID, orgID)
}

func (f *fakeRepository) GetRefreshSessionWithContext(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error) {
	return f.refreshFunc(ctx, tx, tokenHash)
}

func (f *fakeRepository) MarkSessionReplaced(ctx context.Context, tx pgx.Tx, sessionID, replacedByTokenHash string) error {
	return f.markReplaced(ctx, tx, sessionID, replacedByTokenHash)
}

func (f *fakeRepository) RevokeSession(ctx context.Context, tx pgx.Tx, sessionID string) error {
	return f.revokeSession(ctx, tx, sessionID)
}

func (f *fakeRepository) RevokeFamily(ctx context.Context, tx pgx.Tx, familyID string) error {
	return f.revokeFamily(ctx, tx, familyID)
}

func (f *fakeRepository) FindRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	return f.findByHashFunc(ctx, tokenHash)
}

func (f *fakeRepository) GetRolesByMembershipID(ctx context.Context, tx pgx.Tx, membershipID string) ([]string, error) {
	return f.rolesFunc(ctx, tx, membershipID)
}

func (f *fakeRepository) GetLoginByUserID(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error) {
	return f.loginByUserIDFunc(ctx, tx, userID, orgID)
}

func (f *fakeRepository) UpdatePassword(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
	return f.updatePasswordFunc(ctx, tx, userID, orgID, passwordHash)
}

func (f *fakeRepository) RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID string) error {
	return f.revokeUserSessionsFunc(ctx, tx, userID)
}

func (f *fakeRepository) CountRecentFailedLoginAttempts(ctx context.Context, orgID, username string, window time.Duration) (int64, error) {
	if f.countFailedLoginAttemptsFunc != nil {
		return f.countFailedLoginAttemptsFunc(ctx, orgID, username, window)
	}
	return 0, nil
}

func (f *fakeRepository) RecordFailedLoginAttempt(ctx context.Context, orgID, username string) error {
	if f.recordFailedLoginAttemptFunc != nil {
		return f.recordFailedLoginAttemptFunc(ctx, orgID, username)
	}
	return nil
}

func (f *fakeRepository) ClearLoginAttempts(ctx context.Context, orgID, username string) error {
	if f.clearLoginAttemptsFunc != nil {
		return f.clearLoginAttemptsFunc(ctx, orgID, username)
	}
	return nil
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

func TestService_Login_OK(t *testing.T) {
	repo := &fakeRepository{
		findFunc: func(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:             "user-id",
				MembershipID:       "membership-id",
				OrgID:              "org-id",
				Username:           "hs001",
				PasswordHash:       "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE",
				AuthVersion:        1,
				MustChangePassword: false,
				Roles:              []string{"student"},
			}, nil
		},
		insertFunc: func(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error) {
			return "session-id", nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	result, err := svc.Login(context.Background(), LoginRequest{
		OrganizationCode: "school-a",
		Username:         "hs001",
		Password:         "Password123!",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.ExpiresIn != 900 {
		t.Errorf("expires_in = %d, want 900", result.ExpiresIn)
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.User.ID != "user-id" {
		t.Errorf("user id = %q, want user-id", result.User.ID)
	}
	if len(result.Roles) != 1 || result.Roles[0] != "student" {
		t.Errorf("roles = %v, want [student]", result.Roles)
	}
	if result.MustChangePassword {
		t.Error("expected must_change_password = false")
	}
	if len(result.Permissions) == 0 {
		t.Error("expected non-empty permissions for student")
	}
}

func TestService_Login_BadPassword(t *testing.T) {
	repo := &fakeRepository{
		findFunc: func(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
			return &LoginIdentity{
				PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE",
			}, nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{
		OrganizationCode: "school-a",
		Username:         "hs001",
		Password:         "WrongPassword",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_Me_OK(t *testing.T) {
	repo := &fakeRepository{
		actorFunc: func(ctx context.Context, userID, orgID string) (*ActorInfo, error) {
			return &ActorInfo{
				UserID:             userID,
				OrgID:              orgID,
				Username:           "hs001",
				MustChangePassword: true,
			}, nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, false)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)
	result, err := svc.Me(context.Background(), token)
	if err != nil {
		t.Fatalf("Me failed: %v", err)
	}

	if result.ID != "user-id" {
		t.Errorf("id = %q, want user-id", result.ID)
	}
	if result.OrganizationID != "org-id" {
		t.Errorf("organization_id = %q, want org-id", result.OrganizationID)
	}
	if result.DisplayName != "hs001" {
		t.Errorf("display_name = %q, want hs001", result.DisplayName)
	}
	if len(result.Roles) != 1 || result.Roles[0] != "student" {
		t.Errorf("roles = %v, want [student]", result.Roles)
	}
	if !result.MustChangePassword {
		t.Error("expected must_change_password from DB to override token claim")
	}
}

func TestService_Refresh_OK(t *testing.T) {
	oldHash := hashToken("old-refresh-token")
	repo := &fakeRepository{
		refreshFunc: func(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error) {
			if tokenHash != oldHash {
				return nil, ErrUnauthorized
			}
			return &RefreshSession{
				ID:           "old-session-id",
				UserID:       "user-id",
				MembershipID: "membership-id",
				OrgID:        "org-id",
				FamilyID:     "family-id",
				AuthVersion:  1,
				ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
		markReplaced: func(ctx context.Context, tx pgx.Tx, sessionID, replacedByTokenHash string) error {
			return nil
		},
		insertFunc: func(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error) {
			return "new-session-id", nil
		},
		actorFunc: func(ctx context.Context, userID, orgID string) (*ActorInfo, error) {
			return &ActorInfo{UserID: userID, OrgID: orgID, Username: "hs001", MustChangePassword: true}, nil
		},
		rolesFunc: func(ctx context.Context, tx pgx.Tx, membershipID string) ([]string, error) {
			return []string{"teacher"}, nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	result, err := svc.Refresh(context.Background(), "old-refresh-token")
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty new refresh token")
	}
	if result.ExpiresIn != 900 {
		t.Errorf("expires_in = %d, want 900", result.ExpiresIn)
	}
	if result.User.ID != "user-id" {
		t.Errorf("user id = %q, want user-id", result.User.ID)
	}
	if len(result.Roles) != 1 || result.Roles[0] != "teacher" {
		t.Errorf("roles = %v, want [teacher]", result.Roles)
	}
	if !result.MustChangePassword {
		t.Error("expected must_change_password = true from DB")
	}
}

func TestService_Refresh_ReuseDetected(t *testing.T) {
	oldHash := hashToken("stolen-refresh-token")
	replacedHash := "new-hash"
	familyRevoked := false
	repo := &fakeRepository{
		refreshFunc: func(ctx context.Context, tx pgx.Tx, tokenHash string) (*RefreshSession, error) {
			if tokenHash != oldHash {
				return nil, ErrUnauthorized
			}
			return &RefreshSession{
				ID:                  "old-session-id",
				UserID:              "user-id",
				MembershipID:        "membership-id",
				OrgID:               "org-id",
				FamilyID:            "family-id",
				AuthVersion:         1,
				ExpiresAt:           time.Now().Add(7 * 24 * time.Hour),
				ReplacedByTokenHash: &replacedHash,
			}, nil
		},
		revokeFamily: func(ctx context.Context, tx pgx.Tx, familyID string) error {
			if familyID == "family-id" {
				familyRevoked = true
			}
			return nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	_, err := svc.Refresh(context.Background(), "stolen-refresh-token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if !familyRevoked {
		t.Error("expected family to be revoked on reuse")
	}
}

func TestService_Logout_OK(t *testing.T) {
	hash := hashToken("refresh-token")
	revoked := false
	repo := &fakeRepository{
		findByHashFunc: func(ctx context.Context, tokenHash string) (*RefreshSession, error) {
			if tokenHash != hash {
				return nil, nil
			}
			return &RefreshSession{ID: "session-id"}, nil
		},
		revokeSession: func(ctx context.Context, tx pgx.Tx, sessionID string) error {
			if sessionID == "session-id" {
				revoked = true
			}
			return nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	result, err := svc.Logout(context.Background(), "refresh-token")
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if !revoked {
		t.Error("expected session to be revoked")
	}
}

func TestService_Logout_MissingCookie(t *testing.T) {
	repo := &fakeRepository{}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	result, err := svc.Logout(context.Background(), "")
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if !result.Success {
		t.Error("expected success for missing cookie")
	}
}

func TestService_ChangePassword_OK(t *testing.T) {
	seedHash := "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE"
	updated := false
	revoked := false

	repo := &fakeRepository{
		loginByUserIDFunc: func(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:       userID,
				OrgID:        orgID,
				PasswordHash: seedHash,
			}, nil
		},
		updatePasswordFunc: func(ctx context.Context, tx pgx.Tx, userID, orgID, passwordHash string) error {
			updated = true
			return nil
		},
		revokeUserSessionsFunc: func(ctx context.Context, tx pgx.Tx, userID string) error {
			revoked = true
			return nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, true)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)
	if err := svc.ChangePassword(context.Background(), token, ChangePasswordRequest{
		CurrentPassword: "Password123!",
		NewPassword:     "NewPassword123!",
	}); err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}
	if !updated {
		t.Error("expected password to be updated")
	}
	if !revoked {
		t.Error("expected sessions to be revoked")
	}
}

func TestService_ChangePassword_BadCurrent(t *testing.T) {
	seedHash := "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE"
	repo := &fakeRepository{
		loginByUserIDFunc: func(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:       userID,
				OrgID:        orgID,
				PasswordHash: seedHash,
			}, nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, true)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)
	err = svc.ChangePassword(context.Background(), token, ChangePasswordRequest{
		CurrentPassword: "WrongPassword",
		NewPassword:     "NewPassword123!",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_ChangePassword_MissingFields(t *testing.T) {
	repo := &fakeRepository{}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, false)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)
	if err := svc.ChangePassword(context.Background(), token, ChangePasswordRequest{
		CurrentPassword: "",
		NewPassword:     "NewPassword123!",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for missing current password, got %v", err)
	}
}

func TestService_ChangePassword_ReusedPassword(t *testing.T) {
	seedHash := "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE"
	newPass := "NewPassword123!"
	newHash, err := HashPassword(newPass)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	repo := &fakeRepository{
		loginByUserIDFunc: func(ctx context.Context, tx pgx.Tx, userID, orgID string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:       userID,
				OrgID:        orgID,
				PasswordHash: seedHash,
			}, nil
		},
		listPasswordHistoryFunc: func(ctx context.Context, userID string, limit int) ([]string, error) {
			return []string{newHash}, nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, true)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)
	err = svc.ChangePassword(context.Background(), token, ChangePasswordRequest{
		CurrentPassword: "Password123!",
		NewPassword:     newPass,
	})
	if !errors.Is(err, ErrPasswordReused) {
		t.Fatalf("expected ErrPasswordReused, got %v", err)
	}
}

func TestService_Login_AccountLocked(t *testing.T) {
	repo := &fakeRepository{
		findFunc: func(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:       "user-id",
				MembershipID: "membership-id",
				OrgID:        "org-id",
				Username:     "hs001",
				PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE",
			}, nil
		},
		countFailedLoginAttemptsFunc: func(ctx context.Context, orgID, username string, window time.Duration) (int64, error) {
			return 5, nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{
		OrganizationCode: "school-a",
		Username:         "hs001",
		Password:         "Password123!",
	})
	if !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked, got %v", err)
	}
}

func TestService_Login_RecordsFailedAttempt(t *testing.T) {
	recorded := false
	repo := &fakeRepository{
		findFunc: func(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:       "user-id",
				MembershipID: "membership-id",
				OrgID:        "org-id",
				Username:     "hs001",
				PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE",
			}, nil
		},
		recordFailedLoginAttemptFunc: func(ctx context.Context, orgID, username string) error {
			recorded = true
			return nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{
		OrganizationCode: "school-a",
		Username:         "hs001",
		Password:         "WrongPassword123!",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if !recorded {
		t.Error("expected failed login attempt to be recorded")
	}
}

func TestService_Login_ClearsAttemptsOnSuccess(t *testing.T) {
	cleared := false
	repo := &fakeRepository{
		findFunc: func(ctx context.Context, orgCode, username string) (*LoginIdentity, error) {
			return &LoginIdentity{
				UserID:       "user-id",
				MembershipID: "membership-id",
				OrgID:        "org-id",
				Username:     "hs001",
				PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE",
			}, nil
		},
		insertFunc: func(ctx context.Context, tx pgx.Tx, p InsertRefreshSessionParams) (string, error) {
			return "session-id", nil
		},
		clearLoginAttemptsFunc: func(ctx context.Context, orgID, username string) error {
			cleared = true
			return nil
		},
	}

	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	svc := NewService(repo, stubTxManager{}, issuer, 7*24*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{
		OrganizationCode: "school-a",
		Username:         "hs001",
		Password:         "Password123!",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if !cleared {
		t.Error("expected login attempts to be cleared on success")
	}
}
