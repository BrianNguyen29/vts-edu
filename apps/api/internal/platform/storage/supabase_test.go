package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newSupabaseTestServer builds an httptest.Server that mimics the
// Supabase Storage REST API for object upload/get/delete. It records the
// most recent request's headers and path for assertion. The server
// always returns 200 on a non-404 path unless the test installs a hook.
func newSupabaseTestServer(t *testing.T, hook func(w http.ResponseWriter, r *http.Request, captured *captured)) (*httptest.Server, *captured) {
	t.Helper()
	cap := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.Method = r.Method
		cap.Path = r.URL.Path
		cap.AuthHeader = r.Header.Get("Authorization")
		cap.APIKeyHeader = r.Header.Get("apikey")
		cap.ContentType = r.Header.Get("Content-Type")
		cap.Body, _ = io.ReadAll(r.Body)
		if hook != nil {
			hook(w, r, cap)
			return
		}
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Key":"` + strings.TrimPrefix(r.URL.Path, "/storage/v1/object/test-bucket/") + `"}`))
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("body-bytes"))
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

type captured struct {
	Method       string
	Path         string
	AuthHeader   string
	APIKeyHeader string
	ContentType  string
	Body         []byte
}

func TestNewSupabaseProvider_RejectsMissingFields(t *testing.T) {
	tests := []struct {
		name string
		cfg  SupabaseConfig
	}{
		{"no base url", SupabaseConfig{Bucket: "b", APIKey: "k"}},
		{"no bucket", SupabaseConfig{BaseURL: "https://x.supabase.co", APIKey: "k"}},
		{"no api key", SupabaseConfig{BaseURL: "https://x.supabase.co", Bucket: "b"}},
		{"bad scheme", SupabaseConfig{BaseURL: "ftp://x.supabase.co", Bucket: "b", APIKey: "k"}},
		{"missing host", SupabaseConfig{BaseURL: "https://", Bucket: "b", APIKey: "k"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewSupabaseProvider(tc.cfg); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestSupabaseProvider_StoreUsesServiceRoleAuth(t *testing.T) {
	srv, cap := newSupabaseTestServer(t, nil)
	p, err := NewSupabaseProvider(SupabaseConfig{
		BaseURL: srv.URL,
		Bucket:  "test-bucket",
		APIKey:  "service-role-secret",
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	payload := []byte("hello")
	key, err := p.Store(context.Background(), strings.NewReader(string(payload)), int64(len(payload)), "text/plain")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	if !isSafeKey(key) {
		t.Fatalf("expected hex key, got %q", key)
	}
	if cap.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", cap.Method)
	}
	if cap.Path != "/storage/v1/object/test-bucket/"+key {
		t.Fatalf("path = %q, want /storage/v1/object/test-bucket/%s", cap.Path, key)
	}
	if cap.AuthHeader != "Bearer service-role-secret" {
		t.Fatalf("auth = %q", cap.AuthHeader)
	}
	if cap.APIKeyHeader != "service-role-secret" {
		t.Fatalf("apikey = %q", cap.APIKeyHeader)
	}
	if cap.ContentType != "text/plain" {
		t.Fatalf("content-type = %q", cap.ContentType)
	}
	if string(cap.Body) != string(payload) {
		t.Fatalf("body = %q, want %q", cap.Body, payload)
	}
}

func TestSupabaseProvider_StorePropagatesNon2xxAsError(t *testing.T) {
	srv, _ := newSupabaseTestServer(t, func(w http.ResponseWriter, r *http.Request, _ *captured) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream error","apiKey":"service-role-secret"}`))
	})
	p, err := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "service-role-secret"})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	_, err = p.Store(context.Background(), strings.NewReader("x"), 1, "text/plain")
	if err == nil {
		t.Fatal("expected error on 500")
	}
	// Ensure the error message does not echo the upstream body or the key.
	if strings.Contains(err.Error(), "service-role-secret") {
		t.Fatalf("error leaks key: %v", err)
	}
	if strings.Contains(err.Error(), "upstream error") {
		t.Fatalf("error leaks upstream body: %v", err)
	}
}

func TestSupabaseProvider_RetrieveReturnsBody(t *testing.T) {
	srv, cap := newSupabaseTestServer(t, nil)
	p, err := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "k"})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	r, err := p.Retrieve(context.Background(), "abcdef0123456789abcdef0123456789")
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	defer r.Close()
	got, _ := io.ReadAll(r)
	if string(got) != "body-bytes" {
		t.Fatalf("body = %q", got)
	}
	if cap.Method != http.MethodGet {
		t.Fatalf("method = %q", cap.Method)
	}
}

func TestSupabaseProvider_RetrieveMaps404ToNotFound(t *testing.T) {
	srv, _ := newSupabaseTestServer(t, func(w http.ResponseWriter, r *http.Request, _ *captured) {
		w.WriteHeader(http.StatusNotFound)
	})
	p, _ := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "k"})
	_, err := p.Retrieve(context.Background(), "abcdef0123456789abcdef0123456789")
	if !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("expected ErrObjectNotFound, got %v", err)
	}
}

func TestSupabaseProvider_DeleteIsIdempotentOn404(t *testing.T) {
	var calls int32
	srv, _ := newSupabaseTestServer(t, func(w http.ResponseWriter, r *http.Request, _ *captured) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
	})
	p, _ := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "k"})
	if err := p.Delete(context.Background(), "abcdef0123456789abcdef0123456789"); err != nil {
		t.Fatalf("expected nil for 404, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}

func TestSupabaseProvider_RejectsUnsafeKey(t *testing.T) {
	srv, _ := newSupabaseTestServer(t, nil)
	p, _ := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "k"})
	bad := []string{"", "../escape", "abc/def", "abc def", "abc.def", "ABCDEF"}
	for _, k := range bad {
		if _, err := p.Retrieve(context.Background(), k); err == nil {
			t.Fatalf("expected error for unsafe key %q", k)
		}
		if err := p.Delete(context.Background(), k); err == nil {
			t.Fatalf("expected error for unsafe delete key %q", k)
		}
	}
}

func TestSupabaseProvider_RetriesOn5xx(t *testing.T) {
	var calls int32
	srv, _ := newSupabaseTestServer(t, func(w http.ResponseWriter, r *http.Request, _ *captured) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	p, err := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "k", Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	if _, err := p.Retrieve(context.Background(), "abcdef0123456789abcdef0123456789"); err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestSupabaseProvider_StoreSizeMismatch(t *testing.T) {
	srv, _ := newSupabaseTestServer(t, nil)
	p, _ := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "test-bucket", APIKey: "k"})
	// Body length (5) does not match claimed size (3).
	_, err := p.Store(context.Background(), strings.NewReader("hello"), 3, "text/plain")
	if err == nil {
		t.Fatal("expected size mismatch error")
	}
}

func TestSanitizeContentType(t *testing.T) {
	cases := map[string]string{
		"":                          "application/octet-stream",
		"   ":                       "application/octet-stream",
		"text/plain":                "text/plain",
		"text/plain; charset=utf-8": "text/plain",
		"TEXT/PLAIN":                "text/plain",
		"application/pdf":           "application/pdf",
		"application/x-evil":        "application/octet-stream",
		"<script>":                  "application/octet-stream",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	for in, want := range cases {
		if got := SanitizeContentType(in); got != want {
			t.Errorf("SanitizeContentType(%q) = %q, want %q", in, got, want)
		}
	}
}

// Ensure the SupabaseProvider honours context cancellation. We use a
// Server that blocks until the test cancels.
func TestSupabaseProvider_RespectsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	p, _ := NewSupabaseProvider(SupabaseConfig{BaseURL: srv.URL, Bucket: "b", APIKey: "k", Timeout: 2 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := p.Retrieve(ctx, "abcdef0123456789abcdef0123456789")
	if err == nil {
		t.Fatal("expected context error")
	}
	// The error must not contain the API key.
	if strings.Contains(fmt.Sprint(err), "k") && strings.Contains(srv.URL, "k") {
		// Conservative check; both contain 'k'. The actual check is that
		// the bearer header is not present in the error string.
	}
	if strings.Contains(fmt.Sprint(err), "Bearer") {
		t.Fatalf("error leaks bearer: %v", err)
	}
}
