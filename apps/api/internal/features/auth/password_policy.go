package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
)

// Password policy errors.
var (
	ErrWeakPassword      = errors.New("password does not meet strength requirements")
	ErrPasswordUnchanged = errors.New("new password must be different from current password")
	ErrPasswordReused    = errors.New("new password matches a recently used password")
)

// PasswordHistoryLength is the number of previous password hashes retained
// and checked when a user sets a new password.
const PasswordHistoryLength = 5

// PasswordHistoryProvider returns recent password hashes for a user.
type PasswordHistoryProvider interface {
	ListPasswordHistory(ctx context.Context, userID string, limit int) ([]string, error)
}

// PasswordHistoryWriter persists password history rows and trims old entries.
type PasswordHistoryWriter interface {
	InsertPasswordHistory(ctx context.Context, tx pgx.Tx, userID, passwordHash string) error
	DeleteOldPasswordHistory(ctx context.Context, tx pgx.Tx, userID string, keep int) error
}

// blockedPasswords contains obviously weak values that should never be accepted.
var blockedPasswords = map[string]bool{
	"password":     true,
	"password123":  true,
	"12345678":     true,
	"qwertyuiop":   true,
	"11111111":     true,
	"password123!": true,
}

// ValidatePasswordStrength checks a plaintext password against the project
// password policy. It returns a domain error suitable for returning to the
// caller without leaking the password.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	lower := strings.ToLower(password)
	if blockedPasswords[lower] {
		return ErrWeakPassword
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}

	return nil
}

// CheckPasswordHistory verifies the plaintext password does not match any of
// the recent password hashes stored for the user.
func CheckPasswordHistory(ctx context.Context, provider PasswordHistoryProvider, userID, password string) error {
	hashes, err := provider.ListPasswordHistory(ctx, userID, PasswordHistoryLength)
	if err != nil {
		return fmt.Errorf("list password history: %w", err)
	}

	for _, h := range hashes {
		ok, err := VerifyPassword(h, password)
		if err != nil {
			return fmt.Errorf("verify history password: %w", err)
		}
		if ok {
			return ErrPasswordReused
		}
	}

	return nil
}

// StorePasswordHistory persists the previous and new password hashes for a user
// and trims the history to PasswordHistoryLength entries. oldHash may be empty
// when creating a brand-new user.
func StorePasswordHistory(ctx context.Context, writer PasswordHistoryWriter, tx pgx.Tx, userID, oldHash, newHash string) error {
	if oldHash != "" {
		if err := writer.InsertPasswordHistory(ctx, tx, userID, oldHash); err != nil {
			return fmt.Errorf("insert old password history: %w", err)
		}
	}
	if err := writer.InsertPasswordHistory(ctx, tx, userID, newHash); err != nil {
		return fmt.Errorf("insert new password history: %w", err)
	}
	if err := writer.DeleteOldPasswordHistory(ctx, tx, userID, PasswordHistoryLength); err != nil {
		return fmt.Errorf("trim password history: %w", err)
	}
	return nil
}
