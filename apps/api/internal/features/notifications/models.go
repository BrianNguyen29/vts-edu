package notifications

import "context"

// Event types are the supported notification triggers. They mirror
// the business events the caller decides to fire from and keep the
// frontend switch small and stable.
const (
	EventAttemptGraded     = "attempt.graded"
	EventAssessmentPub     = "assessment.published"
	EventResourcePublished = "resource.published"
)

// Notification is the in-package row shape returned to handlers and
// downstream callers. JSON tags mirror the wire shape.
type Notification struct {
	ID           string  `json:"id"`
	OrgID        string  `json:"organization_id"`
	RecipientID  string  `json:"recipient_user_id"`
	EventType    string  `json:"event_type"`
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	MetadataJSON []byte  `json:"-"`
	IsRead       bool    `json:"is_read"`
	ReadAt       *string `json:"read_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// NewNotificationInput is the contract the business code uses to
// enqueue a notification. Recipients may be empty; the notifier
// ignores the call silently if no one to notify.
type NewNotificationInput struct {
	OrgID       string
	RecipientID string
	EventType   string
	Title       string
	Body        string
	Metadata    map[string]any
}

// Notifier is the small seam the rest of the codebase depends on so
// the notifications package stays leaf-level. The interface accepts
// already-resolved recipient ids; for class-scoped events the caller
// is expected to expand the list once.
type Notifier interface {
	Notify(ctx context.Context, input NewNotificationInput) error
	NotifyMany(ctx context.Context, orgID, eventType, title, body string, recipientIDs []string, metadata map[string]any) error
}
