package notifications

import "context"

// NotifierAdapter bridges the notifications.Service (which implements
// Notifier via Notify/NotifyMany) into the smaller, per-package
// Notifier interfaces used by grading, assessments, and resources.
//
// The notifications package owns the adapter (no inbound dependency
// on the other packages) so the dependency direction stays one-way.
type NotifierAdapter struct {
	Svc Service
}

// Notify satisfies the per-package Notifier interface.
func (a *NotifierAdapter) Notify(
	ctx context.Context,
	orgID, recipientID, eventType, title, body string,
	metadata map[string]any,
) {
	_ = a.Svc.Notify(ctx, NewNotificationInput{
		OrgID:       orgID,
		RecipientID: recipientID,
		EventType:   eventType,
		Title:       title,
		Body:        body,
		Metadata:    metadata,
	})
}

// NotifyMany satisfies the per-package Notifier interface.
func (a *NotifierAdapter) NotifyMany(
	ctx context.Context,
	orgID, eventType, title, body string,
	recipientIDs []string,
	metadata map[string]any,
) {
	_ = a.Svc.NotifyMany(ctx, orgID, eventType, title, body, recipientIDs, metadata)
}
