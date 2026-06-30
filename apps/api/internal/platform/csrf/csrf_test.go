package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerate(t *testing.T) {
	tok, err := Generate()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if len(tok) == 0 {
		t.Fatal("token is empty")
	}
}

func TestValidate(t *testing.T) {
	tok, err := Generate()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: string(tok)})
	req.Header.Set(HeaderName, string(tok))

	if !Validate(req) {
		t.Fatal("expected valid csrf token")
	}
}

func TestValidateMissingCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.Header.Set(HeaderName, "token")
	if Validate(req) {
		t.Fatal("expected invalid when cookie missing")
	}
}
