package auth

import (
	"testing"
	"time"
)

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("Password123!")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := VerifyPassword(hash, "Password123!")
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}

	ok, err = VerifyPassword(hash, "WrongPassword")
	if err != nil {
		t.Fatalf("VerifyPassword returned error for wrong password: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestVerifyPassword_DemoSeedHash(t *testing.T) {
	seedHash := "$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE"
	ok, err := VerifyPassword(seedHash, "Password123!")
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected demo seed password to verify")
	}
}

func TestTokenIssuer(t *testing.T) {
	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)

	token, exp, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if exp.Before(time.Now()) {
		t.Fatal("expected expiration in the future")
	}

	claims, err := issuer.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}
	if claims.Subject != "user-id" {
		t.Errorf("subject = %q, want %q", claims.Subject, "user-id")
	}
	if claims.OrgID != "org-id" {
		t.Errorf("org = %q, want %q", claims.OrgID, "org-id")
	}
	if claims.SessionID != "session-id" {
		t.Errorf("sid = %q, want %q", claims.SessionID, "session-id")
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "student" {
		t.Errorf("roles = %v, want [student]", claims.Roles)
	}
	if claims.AuthVersion != 1 {
		t.Errorf("av = %d, want 1", claims.AuthVersion)
	}
}

func TestTokenIssuer_InvalidToken(t *testing.T) {
	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)

	if _, err := issuer.ValidateAccessToken("not-a-valid-token"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestTokenIssuer_WrongSigningKey(t *testing.T) {
	issuer := NewTokenIssuer("test-signing-key-minimum-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", nil, 1)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	other := NewTokenIssuer("different-signing-key-32-bytes-long", "test-issuer", "test-audience", 15*time.Minute)
	if _, err := other.ValidateAccessToken(token); err == nil {
		t.Fatal("expected error when validating with wrong signing key")
	}
}
