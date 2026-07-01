package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	l := New(true, 1, 2, time.Minute, time.Minute)
	defer l.Stop()

	ip := "10.0.0.1"
	if !l.Allow(ip) {
		t.Fatalf("first request should be allowed")
	}
	if !l.Allow(ip) {
		t.Fatalf("second request within burst should be allowed")
	}
	if l.Allow(ip) {
		t.Fatalf("third request should exceed burst")
	}
}

func TestLimiter_Disabled(t *testing.T) {
	l := New(false, 1, 1, time.Minute, time.Minute)
	defer l.Stop()

	for i := 0; i < 10; i++ {
		if !l.Allow("10.0.0.1") {
			t.Fatalf("disabled limiter should allow all")
		}
	}
}

func TestMiddleware_RateLimited(t *testing.T) {
	l := New(true, 1, 1, time.Minute, time.Minute)
	defer l.Stop()

	handler := Middleware(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request = %d, want %d", rec1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestMiddleware_ExcludedPaths(t *testing.T) {
	l := New(true, 0, 0, time.Minute, time.Minute)
	defer l.Stop()

	handler := Middleware(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, path := range []string{"/healthz", "/readyz", "/api/v1/auth/csrf-token"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	if got := ClientIP(req); got != "192.168.1.2" {
		t.Fatalf("ClientIP = %q, want 192.168.1.2", got)
	}
}
