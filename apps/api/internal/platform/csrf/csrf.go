// Package csrf provides a minimal double-submit CSRF token helper.
// For the cross-origin demo (Vercel -> Koyeb), refresh cookie is SameSite=None,
// so cookie-backed unsafe endpoints require a CSRF token.
package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	// CookieName is sent to the browser on GET /api/v1/auth/csrf-token.
	CookieName = "vts_csrf"
	// HeaderName is expected on unsafe requests (POST/PUT/PATCH/DELETE).
	HeaderName = "X-CSRF-Token"
)

// Token is a 32-byte random value encoded as hex.
type Token string

// Generate creates a new random CSRF token.
func Generate() (Token, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return Token(hex.EncodeToString(b)), nil
}

// SetCookie writes the CSRF cookie with Secure + SameSite=None for cross-origin demo.
// In same-origin deployments, switch to SameSite=Lax.
func SetCookie(w http.ResponseWriter, token Token) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    string(token),
		Path:     "/",
		Secure:   true,
		HttpOnly: false, // must be readable by JS for double-submit
		SameSite: http.SameSiteNoneMode,
		MaxAge:   86400,
	})
}

// Validate checks the request header against the cookie value.
func Validate(r *http.Request) bool {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	header := r.Header.Get(HeaderName)
	if header == "" {
		return false
	}
	// Constant-time comparison not required for random tokens, but good practice.
	return cookie.Value == header
}
