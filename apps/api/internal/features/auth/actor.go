package auth

import (
	"net/http"
	"strings"
)

// Actor holds the authenticated actor extracted from a valid access token.
type Actor struct {
	UserID             string
	OrgID              string
	Roles              []string
	MustChangePassword bool
}

// ActorFromRequest validates the Bearer token in r and returns the actor.
// It is a thin helper intended for other feature handlers to reuse the same
// JWT issuer without importing service internals.
func ActorFromRequest(r *http.Request, issuer *TokenIssuer) (Actor, error) {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return Actor{}, ErrUnauthorized
	}

	claims, err := issuer.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return Actor{}, ErrUnauthorized
	}

	return Actor{
		UserID:             claims.Subject,
		OrgID:              claims.OrgID,
		Roles:              claims.Roles,
		MustChangePassword: claims.MustChangePassword,
	}, nil
}
