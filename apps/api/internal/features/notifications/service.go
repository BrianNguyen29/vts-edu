package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
)

// Service is the application service contract for the notifications
// inbox. It also implements the Notifier seam so other packages can
// call into it without depending on a specific storage implementation.
type Service interface {
	Notifier
	List(ctx context.Context, actor auth.Actor, before *time.Time, limit int) ([]Notification, error)
	UnreadCount(ctx context.Context, actor auth.Actor) (int, error)
	MarkRead(ctx context.Context, actor auth.Actor, id string) (Notification, error)
}

type service struct {
	repo Repository
}

// NewService creates a notification service backed by the given
// repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Notify inserts a single notification. Failure is logged and
// swallowed: the notification path is best-effort by design and must
// not fail the calling business operation.
func (s *service) Notify(ctx context.Context, input NewNotificationInput) error {
	if input.OrgID == "" || input.RecipientID == "" || input.EventType == "" {
		// Bad input is logged but does not return an error to the
		// caller. The Notify contract is fire-and-forget.
		slog.Default().Warn("notifications.Notify: missing required fields",
			"event_type", input.EventType,
		)
		return nil
	}
	_, err := s.repo.Insert(
		ctx,
		input.OrgID,
		input.RecipientID,
		input.EventType,
		input.Title,
		input.Body,
		input.Metadata,
	)
	if err != nil {
		slog.Default().Error("notifications.Notify: insert failed",
			"event_type", input.EventType,
			"recipient", input.RecipientID,
			"error", err.Error(),
		)
	}
	return nil
}

// NotifyMany inserts one notification per recipient. Same best-effort
// semantics as Notify; per-recipient failures are logged and skipped.
func (s *service) NotifyMany(
	ctx context.Context,
	orgID, eventType, title, body string,
	recipientIDs []string,
	metadata map[string]any,
) error {
	for _, id := range recipientIDs {
		if id == "" {
			continue
		}
		if err := s.Notify(ctx, NewNotificationInput{
			OrgID:       orgID,
			RecipientID: id,
			EventType:   eventType,
			Title:       title,
			Body:        body,
			Metadata:    metadata,
		}); err != nil {
			// Errors are already logged inside Notify; do not abort the
			// remaining recipients.
			continue
		}
	}
	return nil
}

// List returns the inbox for the calling user, newest first.
func (s *service) List(ctx context.Context, actor auth.Actor, before *time.Time, limit int) ([]Notification, error) {
	if actor.OrgID == "" || actor.UserID == "" {
		return nil, ErrUnauthorized
	}
	return s.repo.List(ctx, actor.OrgID, actor.UserID, before, limit)
}

// UnreadCount returns the current unread count for the bell badge.
func (s *service) UnreadCount(ctx context.Context, actor auth.Actor) (int, error) {
	if actor.OrgID == "" || actor.UserID == "" {
		return 0, ErrUnauthorized
	}
	return s.repo.CountUnread(ctx, actor.OrgID, actor.UserID)
}

// MarkRead is idempotent: re-marking a notification that is already
// read keeps it read and returns the same row.
func (s *service) MarkRead(ctx context.Context, actor auth.Actor, id string) (Notification, error) {
	if actor.OrgID == "" || actor.UserID == "" {
		return Notification{}, ErrUnauthorized
	}
	if id == "" {
		return Notification{}, ErrInvalidInput
	}
	return s.repo.MarkRead(ctx, actor.OrgID, actor.UserID, id)
}

// Ensure interface satisfaction.
var (
	_ Service  = (*service)(nil)
	_ Notifier = (*service)(nil)
)

// DecodeMetadata is a small helper for callers that want to surface
// the metadata map (e.g. the frontend).
func DecodeMetadata(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}
	return out, nil
}

// IsNotFound is a small helper for handlers.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
