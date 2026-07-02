package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

type fakeRepo struct {
	inserted   []Notification
	listResult []Notification
	listErr    error
	count      int
	countErr   error
	marked     *Notification
	markErr    error
	classIDs   []string
	targetIDs  []string
}

func (f *fakeRepo) Insert(ctx context.Context, orgID, recipientID, eventType, title, body string, metadata map[string]any) (Notification, error) {
	n := Notification{
		ID:           "n-1",
		OrgID:        orgID,
		RecipientID:  recipientID,
		EventType:    eventType,
		Title:        title,
		Body:         body,
		MetadataJSON: []byte("{}"),
		CreatedAt:    "2026-07-02T00:00:00Z",
	}
	f.inserted = append(f.inserted, n)
	return n, nil
}

func (f *fakeRepo) List(ctx context.Context, orgID, userID string, before *time.Time, limit int) ([]Notification, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeRepo) CountUnread(ctx context.Context, orgID, userID string) (int, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	return f.count, nil
}

func (f *fakeRepo) Get(ctx context.Context, orgID, userID, id string) (Notification, error) {
	return Notification{}, nil
}

func (f *fakeRepo) MarkRead(ctx context.Context, orgID, userID, id string) (Notification, error) {
	if f.markErr != nil {
		return Notification{}, f.markErr
	}
	if f.marked == nil {
		return Notification{ID: id, OrgID: orgID, RecipientID: userID, IsRead: true}, nil
	}
	return *f.marked, nil
}

func (f *fakeRepo) ListClassStudentUserIDs(ctx context.Context, orgID, classID string) ([]string, error) {
	return f.classIDs, nil
}

func (f *fakeRepo) ListAssessmentTargetStudentUserIDs(ctx context.Context, orgID, assessmentID string) ([]string, error) {
	return f.targetIDs, nil
}

func newActor() auth.Actor {
	return auth.Actor{OrgID: "00000000-0000-4000-8000-000000000001", UserID: "00000000-0000-4000-8000-000000000002"}
}

func TestService_Notify_InsertsRow(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	if err := svc.Notify(context.Background(), NewNotificationInput{
		OrgID: "org", RecipientID: "user", EventType: EventAttemptGraded,
		Title: "Đã chấm", Body: "x", Metadata: map[string]any{"attempt_id": "a"},
	}); err != nil {
		t.Fatalf("notify: %v", err)
	}
	if len(repo.inserted) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(repo.inserted))
	}
}

func TestService_NotifyMany_SwallowesErrors(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	if err := svc.NotifyMany(context.Background(), "org", EventAssessmentPub, "title", "body", []string{"u1", "", "u2"}, nil); err != nil {
		t.Fatalf("notifymany: %v", err)
	}
	if len(repo.inserted) != 2 {
		t.Fatalf("expected 2 inserts (empty id skipped), got %d", len(repo.inserted))
	}
}

func TestService_Notify_BadInputSwallowed(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	if err := svc.Notify(context.Background(), NewNotificationInput{}); err != nil {
		t.Fatalf("notify: %v", err)
	}
	if len(repo.inserted) != 0 {
		t.Fatalf("expected no insert, got %d", len(repo.inserted))
	}
}

func TestService_List_RequiresActor(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	if _, err := svc.List(context.Background(), auth.Actor{}, nil, 0); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestService_MarkRead_RequiresID(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	if _, err := svc.MarkRead(context.Background(), newActor(), ""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestService_MarkRead_Success(t *testing.T) {
	repo := &fakeRepo{
		marked: &Notification{ID: "n", IsRead: true, CreatedAt: "2026-07-02T00:00:00Z"},
	}
	svc := NewService(repo)
	got, err := svc.MarkRead(context.Background(), newActor(), "n")
	if err != nil {
		t.Fatalf("markread: %v", err)
	}
	if !got.IsRead {
		t.Fatalf("expected is_read=true")
	}
}

func TestService_UnreadCount_DelegatesToRepo(t *testing.T) {
	repo := &fakeRepo{count: 5}
	svc := NewService(repo)
	n, err := svc.UnreadCount(context.Background(), newActor())
	if err != nil {
		t.Fatalf("unread: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5, got %d", n)
	}
}
