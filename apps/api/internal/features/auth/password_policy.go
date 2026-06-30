package auth

import (
	"errors"
	"strings"
	"unicode"
)

// Password policy errors.
var (
	ErrWeakPassword      = errors.New("password does not meet strength requirements")
	ErrPasswordUnchanged = errors.New("new password must be different from current password")
)

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
