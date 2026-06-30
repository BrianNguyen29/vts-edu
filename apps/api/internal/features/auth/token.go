package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenIssuer creates and validates access tokens.
type TokenIssuer struct {
	signingKey []byte
	issuer     string
	audience   string
	ttl        time.Duration
}

// AccessClaims is the custom JWT claim set for access tokens.
type AccessClaims struct {
	jwt.RegisteredClaims
	OrgID              string   `json:"org"`
	SessionID          string   `json:"sid"`
	Roles              []string `json:"roles"`
	AuthVersion        int64    `json:"av"`
	MustChangePassword bool     `json:"pwd_change_required"`
}

// NewTokenIssuer creates a token issuer. signingKey should be at least 32 bytes
// for HMAC-SHA256.
func NewTokenIssuer(signingKey, issuer, audience string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{
		signingKey: []byte(signingKey),
		issuer:     issuer,
		audience:   audience,
		ttl:        ttl,
	}
}

// IssueAccessToken signs a new access token and returns the token string and
// its expiration time.
func (ti *TokenIssuer) IssueAccessToken(userID, orgID, sessionID string, roles []string, authVersion int64, mustChangePassword bool) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(ti.ttl)

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    ti.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{ti.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        randomJTI(),
		},
		OrgID:              orgID,
		SessionID:          sessionID,
		Roles:              roles,
		AuthVersion:        authVersion,
		MustChangePassword: mustChangePassword,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(ti.signingKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, exp, nil
}

// ValidateAccessToken parses and validates an access token, returning the
// extracted claims.
func (ti *TokenIssuer) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return ti.signingKey, nil
		},
		jwt.WithIssuer(ti.issuer),
		jwt.WithAudience(ti.audience),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func randomJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is extremely unlikely; return a timestamp-based
		// fallback rather than panicking.
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
