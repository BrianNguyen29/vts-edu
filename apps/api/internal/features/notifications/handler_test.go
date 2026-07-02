package notifications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/notifications"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/go-chi/chi/v5"
)

// withChiPathParam attaches a chi route context with the given path
// parameter so chi.URLParam can resolve it in handler unit tests
// that do not run through the live router.
func withChiPathParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// fakeService is the minimal stub needed to drive Handler.MarkRead
// through CSRF. Only the MarkRead path is exercised.
type fakeService struct {
	markRead func(ctx context.Context, actor auth.Actor, id string) (notifications.Notification, error)
}

func (f *fakeService) Notify(ctx context.Context, input notifications.NewNotificationInput) error {
	return nil
}

func (f *fakeService) NotifyMany(
	ctx context.Context,
	orgID, eventType, title, body string,
	recipientIDs []string,
	metadata map[string]any,
) error {
	return nil
}

func (f *fakeService) List(ctx context.Context, actor auth.Actor, before *time.Time, limit int) ([]notifications.Notification, error) {
	return nil, nil
}

func (f *fakeService) UnreadCount(ctx context.Context, actor auth.Actor) (int, error) {
	return 0, nil
}

func (f *fakeService) MarkRead(ctx context.Context, actor auth.Actor, id string) (notifications.Notification, error) {
	if f.markRead == nil {
		return notifications.Notification{}, nil
	}
	return f.markRead(ctx, actor, id)
}

func issueToken(t *testing.T) string {
	t.Helper()
	issuer := auth.NewTokenIssuer(
		"test-signing-key-minimum-32-bytes-long",
		"test-issuer",
		"test-audience",
		15*time.Minute,
	)
	token, _, err := issuer.IssueAccessToken("user-id", "org-id", "session-id", []string{"student"}, 1, false)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return token
}

func newHandler(svc notifications.Service) *notifications.Handler {
	issuer := auth.NewTokenIssuer(
		"test-signing-key-minimum-32-bytes-long",
		"test-issuer",
		"test-audience",
		15*time.Minute,
	)
	return notifications.NewHandler(svc, issuer)
}

func TestHandler_MarkRead_RejectsMissingCSRF(t *testing.T) {
	var called int
	svc := &fakeService{
		markRead: func(ctx context.Context, actor auth.Actor, id string) (notifications.Notification, error) {
			called++
			return notifications.Notification{}, nil
		},
	}
	h := newHandler(svc)

	req := withChiPathParam(
		httptest.NewRequest(http.MethodPost, "/api/v1/me/notifications/abc/read", nil),
		"id", "abc",
	)
	req.Header.Set("Authorization", "Bearer "+issueToken(t))
	// No vts_csrf cookie, no X-CSRF-Token header.
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if called != 0 {
		t.Fatalf("service called %d times, want 0 when CSRF missing", called)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "invalid_csrf" {
		t.Fatalf("error code = %q, want invalid_csrf", body.Error.Code)
	}
}

func TestHandler_MarkRead_RejectsMismatchedCSRF(t *testing.T) {
	var called int
	svc := &fakeService{
		markRead: func(ctx context.Context, actor auth.Actor, id string) (notifications.Notification, error) {
			called++
			return notifications.Notification{}, nil
		},
	}
	h := newHandler(svc)

	req := withChiPathParam(
		httptest.NewRequest(http.MethodPost, "/api/v1/me/notifications/abc/read", nil),
		"id", "abc",
	)
	req.Header.Set("Authorization", "Bearer "+issueToken(t))
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "good-token"})
	req.Header.Set(csrf.HeaderName, "wrong-token")
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if called != 0 {
		t.Fatalf("service called %d times, want 0 on CSRF mismatch", called)
	}
}

func TestHandler_MarkRead_AllowsMatchingCSRF(t *testing.T) {
	called := false
	svc := &fakeService{
		markRead: func(ctx context.Context, actor auth.Actor, id string) (notifications.Notification, error) {
			called = true
			if actor.UserID != "user-id" || actor.OrgID != "org-id" {
				t.Errorf("actor = %+v, want user-id/org-id", actor)
			}
			if id != "abc" {
				t.Errorf("id = %q, want abc", id)
			}
			return notifications.Notification{
				ID:           "abc",
				OrgID:        "org-id",
				RecipientID:  "user-id",
				EventType:    notifications.EventAssessmentPub,
				Title:        "Đề thi mới",
				Body:         "Đã mở",
				MetadataJSON: []byte(`{}`),
				IsRead:       true,
				CreatedAt:    "2026-07-02T00:00:00Z",
			}, nil
		},
	}
	h := newHandler(svc)

	req := withChiPathParam(
		httptest.NewRequest(http.MethodPost, "/api/v1/me/notifications/abc/read", nil),
		"id", "abc",
	)
	req.Header.Set("Authorization", "Bearer "+issueToken(t))
	const tok = "matching-token"
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: tok})
	req.Header.Set(csrf.HeaderName, tok)
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("service was not called with valid CSRF")
	}
	var body struct {
		Data struct {
			ID     string `json:"id"`
			IsRead bool   `json:"is_read"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Data.ID != "abc" || !body.Data.IsRead {
		t.Fatalf("response = %+v, want id=abc is_read=true", body.Data)
	}
}

func TestHandler_MarkRead_RequiresAuthBeforeCSRF(t *testing.T) {
	// No Authorization header → 401 (auth) must short-circuit before
	// CSRF, and the service must not be called.
	var called int
	svc := &fakeService{
		markRead: func(ctx context.Context, actor auth.Actor, id string) (notifications.Notification, error) {
			called++
			return notifications.Notification{}, errors.New("should not be called")
		},
	}
	h := newHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/notifications/abc/read", nil)
	// No Authorization, no CSRF.
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if called != 0 {
		t.Fatalf("service called %d times, want 0 on unauthenticated", called)
	}
}
