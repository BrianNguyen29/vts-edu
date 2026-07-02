package resources

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/go-chi/chi/v5"
)

const testJWTSigningKey = "test-signing-key-32-bytes-long!!!"

// newTestHandler wires a Handler with a real TokenIssuer, fake repo,
// and fake storage so the download path can be exercised end-to-end.
// The fake repo's createFile hook forces the persisted content type to
// the value passed in (so we can probe how the handler reacts to an
// attacker-controlled content type that bypasses service-layer
// sanitization).
func newTestHandler(t *testing.T, persistedContentType string) (*Handler, string) {
	t.Helper()
	issuer := auth.NewTokenIssuer(testJWTSigningKey, "vts-edu-api", "vts-edu-web", 15*time.Minute)
	repo := newFakeRepo()
	store := &fakeStorage{storeKey: "abcdef0123456789abcdef0123456789", stored: []byte("hello")}
	svc := NewService(repo, store, 1024)
	teacher := newActor("teacher")
	res, err := svc.CreateResource(context.Background(), teacher, CreateResourceRequest{Title: "Doc", ContextType: ContextTypeOrganization, ContextID: "org-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	repo.createFile = func(ctx context.Context, file ResourceFile) (ResourceFile, error) {
		file.ContentType = persistedContentType
		file.ID = "file-1"
		file.Status = FileStatusActive
		repo.files[file.ResourceID] = []ResourceFile{file}
		return file, nil
	}
	if _, err := svc.UploadFile(context.Background(), teacher, res.ID, "doc.txt", "text/plain", bytes.NewReader([]byte("hello")), 5); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if _, err := svc.PublishResource(context.Background(), teacher, res.ID); err != nil {
		t.Fatalf("publish: %v", err)
	}
	tok, _, err := issuer.IssueAccessToken(teacher.UserID, teacher.OrgID, "session-1", teacher.Roles, 0, false)
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}
	return NewHandler(svc, issuer), tok
}

func doDownload(t *testing.T, h *Handler, tok, id string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/resources/"+id+"/download", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.DownloadFile(rec, r)
	return rec
}

func TestDownloadHandler_SetsNosniffAndSanitizesContentType(t *testing.T) {
	h, tok := newTestHandler(t, "text/plain")
	rec := doDownload(t, h, tok, "res-1")
	if rec.Code != http.StatusOK {
		t.Logf("debug: files = %+v", h.svc) // placeholder for trace
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/plain" {
		t.Fatalf("Content-Type = %q, want text/plain", got)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "hello") {
		t.Fatalf("body = %q", body)
	}
}

func TestDownloadHandler_DisallowedContentTypeFallsBack(t *testing.T) {
	h, tok := newTestHandler(t, "application/x-evil")
	rec := doDownload(t, h, tok, "res-1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("Content-Type = %q, want application/octet-stream", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestDownloadHandler_RejectsMissingAuth(t *testing.T) {
	h, _ := newTestHandler(t, "text/plain")
	rec := doDownload(t, h, "", "res-1")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestDownloadHandler_InlineDispositionForImage(t *testing.T) {
	h, tok := newTestHandler(t, "image/png")
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/resources/res-1/download?disposition=inline", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "res-1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	h.DownloadFile(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	disp := rec.Header().Get("Content-Disposition")
	if !strings.HasPrefix(disp, "inline;") {
		t.Fatalf("expected inline disposition, got %q", disp)
	}
}

func TestDownloadHandler_InlineFallsBackToAttachmentForBinary(t *testing.T) {
	h, tok := newTestHandler(t, "application/octet-stream")
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/resources/res-1/download?disposition=inline", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "res-1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	h.DownloadFile(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	disp := rec.Header().Get("Content-Disposition")
	if !strings.HasPrefix(disp, "attachment;") {
		t.Fatalf("expected attachment disposition for unsafe MIME, got %q", disp)
	}
}
