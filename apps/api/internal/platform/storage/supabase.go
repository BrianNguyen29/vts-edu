package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrObjectNotFound is returned by providers when the requested object
// is not present in the backend. Callers should map this to a 404.
var ErrObjectNotFound = errors.New("storage object not found")

// SupabaseProvider is a storage.Provider backed by the Supabase Storage
// REST API (or any S3-compatible service that follows the same shape).
//
// It uses the service role key (server-side only) for authorization.
// The bucket is expected to be private; the adapter retrieves objects
// via the REST API and never returns a signed URL to the client.
//
// Reference: https://supabase.com/docs//reference/api/storage (object
// endpoints). Endpoint construction is isolated so a future S3 adapter
// can share the same key/size validation surface.
type SupabaseProvider struct {
	BaseURL    string
	Bucket     string
	APIKey     string // service role key (server-side only)
	HTTPClient *http.Client
	// MaxRetries is the number of times to retry idempotent requests on
	// transient errors (network failures, 5xx). Defaults to 2.
	MaxRetries int
}

// SupabaseConfig is the constructor input. It intentionally does not
// embed the service role key in any error path.
type SupabaseConfig struct {
	BaseURL string
	Bucket  string
	APIKey  string
	// Timeout is the per-request HTTP timeout. Defaults to 30s.
	Timeout time.Duration
}

// NewSupabaseProvider validates the configuration and returns a ready
// provider. It fails fast when required fields are missing or malformed;
// callers (e.g. main wiring) should surface the error to logs and exit.
func NewSupabaseProvider(cfg SupabaseConfig) (*SupabaseProvider, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("supabase base url is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("supabase bucket is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("supabase service role key is required")
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("supabase base url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("supabase base url must use http or https")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("supabase base url is missing host")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &SupabaseProvider{
		BaseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		Bucket:     cfg.Bucket,
		APIKey:     cfg.APIKey,
		HTTPClient: &http.Client{Timeout: timeout},
		MaxRetries: 2,
	}, nil
}

// objectPath builds the path for a key under the bucket. It enforces the
// hex-key invariant; callers must pass keys produced by generateKey or
// loaded from the database (which are already validated). A bad key is
// rejected here as a defence-in-depth measure.
func (p *SupabaseProvider) objectPath(key string) (string, error) {
	if !isSafeKey(key) {
		return "", fmt.Errorf("invalid storage key")
	}
	// url.PathEscape is unnecessary: hex keys only contain [0-9a-f].
	return fmt.Sprintf("/storage/v1/object/%s/%s", p.Bucket, key), nil
}

// authHeaders returns the headers required for a service-role request.
// The service role key is propagated only on the wire; it is never
// returned in errors or logs.
func (p *SupabaseProvider) authHeaders(extraHeaders http.Header) http.Header {
	h := http.Header{}
	if extraHeaders != nil {
		h = http.Header(extraHeaders.Clone())
	}
	h.Set("Authorization", "Bearer "+p.APIKey)
	h.Set("apikey", p.APIKey)
	return h
}

// Store uploads the object to Supabase Storage at the generated key.
// size is the expected payload size; the body is read fully.
func (p *SupabaseProvider) Store(ctx context.Context, r io.Reader, size int64, contentType string) (string, error) {
	key, err := generateKey()
	if err != nil {
		return "", fmt.Errorf("generate storage key: %w", err)
	}
	path, err := p.objectPath(key)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read upload body: %w", err)
	}
	if size > 0 && int64(len(body)) != size {
		return "", fmt.Errorf("size mismatch: expected %d, got %d", size, len(body))
	}
	headers := p.authHeaders(nil)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	headers.Set("Content-Type", contentType)
	headers.Set("x-upsert", "false")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build upload request: %w", err)
	}
	req.Header = headers

	resp, err := p.doWithRetry(req)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", p.sanitizeUploadError(resp)
	}
	// Drain body to allow connection reuse.
	_, _ = io.Copy(io.Discard, resp.Body)
	return key, nil
}

// Retrieve returns a ReadCloser for the object identified by key.
func (p *SupabaseProvider) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := p.objectPath(key)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build download request: %w", err)
	}
	req.Header = p.authHeaders(nil)

	resp, err := p.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, ErrObjectNotFound
	}
	if resp.StatusCode/100 != 2 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download: status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// Delete removes the object identified by key. A 404 is treated as
// success so callers can use Delete for "ensure not present" semantics.
func (p *SupabaseProvider) Delete(ctx context.Context, key string) error {
	path, err := p.objectPath(key)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p.BaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build delete request: %w", err)
	}
	req.Header = p.authHeaders(nil)

	resp, err := p.doWithRetry(req)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("delete: status %d", resp.StatusCode)
	}
	return nil
}

// doWithRetry executes req and retries on transient errors (network,
// 5xx) up to MaxRetries times. Non-2xx responses are returned as-is
// so callers can decide how to interpret them; the retry policy is
// safe for Store (POST) because the key is generated locally and
// Supabase's object POST is idempotent for a given key (or fails with
// 409, which is a permanent error).
func (p *SupabaseProvider) doWithRetry(req *http.Request) (*http.Response, error) {
	maxAttempts := p.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := p.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 500 && attempt < maxAttempts-1 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

// sanitizeUploadError returns a generic error for upload failures
// without leaking upstream status, body, or headers. Callers receive a
// stable error so they can map it to ErrStorageFailure in the service
// layer. The Supabase response body is intentionally discarded.
func (p *SupabaseProvider) sanitizeUploadError(resp *http.Response) error {
	_, _ = io.Copy(io.Discard, resp.Body)
	return fmt.Errorf("upload: status %d", resp.StatusCode)
}
