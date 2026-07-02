package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	notificationssqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/notifications/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the notification
// inbox. Each call carries the actor's organization + user id so
// tenant isolation is enforced at the query boundary.
type Repository interface {
	Insert(ctx context.Context, orgID, recipientID, eventType, title, body string, metadata map[string]any) (Notification, error)
	List(ctx context.Context, orgID, userID string, before *time.Time, limit int) ([]Notification, error)
	CountUnread(ctx context.Context, orgID, userID string) (int, error)
	Get(ctx context.Context, orgID, userID, id string) (Notification, error)
	MarkRead(ctx context.Context, orgID, userID, id string) (Notification, error)
	ListClassStudentUserIDs(ctx context.Context, orgID, classID string) ([]string, error)
	ListAssessmentTargetStudentUserIDs(ctx context.Context, orgID, assessmentID string) ([]string, error)
}

type sqlcRepository struct {
	queries *notificationssqlc.Queries
}

// NewRepository creates a notification repository backed by the
// generated sqlc queries.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: notificationssqlc.New(pool)}
}

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func timeOrNil(t pgtype.Timestamptz) *string {
	if t.Valid {
		s := t.Time.UTC().Format(time.RFC3339Nano)
		return &s
	}
	return nil
}

func formatTime(t pgtype.Timestamptz) string {
	if t.Valid {
		return t.Time.UTC().Format(time.RFC3339Nano)
	}
	return ""
}

// metadataJSON serializes a free-form map for the jsonb column. A
// nil map is stored as "{}" so the NOT NULL constraint is satisfied.
func metadataJSON(meta map[string]any) ([]byte, error) {
	if meta == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(meta)
}

func (r *sqlcRepository) Insert(
	ctx context.Context,
	orgID, recipientID, eventType, title, body string,
	metadata map[string]any,
) (Notification, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(recipientID)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid user id: %w", err)
	}
	metaBytes, err := metadataJSON(metadata)
	if err != nil {
		return Notification{}, fmt.Errorf("encode metadata: %w", err)
	}

	row, err := r.queries.InsertNotification(ctx, notificationssqlc.InsertNotificationParams{
		OrganizationID:  orgUUID,
		RecipientUserID: userUUID,
		EventType:       eventType,
		Title:           title,
		Body:            body,
		Column6:         metaBytes,
	})
	if err != nil {
		return Notification{}, fmt.Errorf("insert notification: %w", err)
	}
	return mapRow(row), nil
}

func (r *sqlcRepository) List(
	ctx context.Context, orgID, userID string, before *time.Time, limit int,
) ([]Notification, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var beforeParam pgtype.Timestamptz
	if before != nil {
		beforeParam = pgtype.Timestamptz{Valid: true, Time: *before}
	}

	rows, err := r.queries.ListNotifications(ctx, notificationssqlc.ListNotificationsParams{
		OrganizationID:  orgUUID,
		RecipientUserID: userUUID,
		Column3:         beforeParam,
		Limit:           int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	out := make([]Notification, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func (r *sqlcRepository) CountUnread(ctx context.Context, orgID, userID string) (int, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user id: %w", err)
	}
	n, err := r.queries.CountUnread(ctx, notificationssqlc.CountUnreadParams{
		OrganizationID:  orgUUID,
		RecipientUserID: userUUID,
	})
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return int(n), nil
}

func (r *sqlcRepository) Get(ctx context.Context, orgID, userID, id string) (Notification, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid user id: %w", err)
	}
	idUUID, err := toUUID(id)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid notification id: %w", err)
	}
	row, err := r.queries.GetNotification(ctx, notificationssqlc.GetNotificationParams{
		ID:              idUUID,
		OrganizationID:  orgUUID,
		RecipientUserID: userUUID,
	})
	if err != nil {
		return Notification{}, mapRepoError(err)
	}
	return mapRow(row), nil
}

func (r *sqlcRepository) MarkRead(ctx context.Context, orgID, userID, id string) (Notification, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid user id: %w", err)
	}
	idUUID, err := toUUID(id)
	if err != nil {
		return Notification{}, fmt.Errorf("invalid notification id: %w", err)
	}
	row, err := r.queries.MarkRead(ctx, notificationssqlc.MarkReadParams{
		ID:              idUUID,
		OrganizationID:  orgUUID,
		RecipientUserID: userUUID,
	})
	if err != nil {
		return Notification{}, mapRepoError(err)
	}
	return mapRow(row), nil
}

func (r *sqlcRepository) ListClassStudentUserIDs(ctx context.Context, orgID, classID string) ([]string, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classID)
	if err != nil {
		return nil, fmt.Errorf("invalid class id: %w", err)
	}
	rows, err := r.queries.ListClassStudentUserIDs(ctx, notificationssqlc.ListClassStudentUserIDsParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list class students: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.String())
	}
	return out, nil
}

func (r *sqlcRepository) ListAssessmentTargetStudentUserIDs(ctx context.Context, orgID, assessmentID string) ([]string, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	assessmentUUID, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.ListAssessmentTargetStudentUserIDs(ctx, notificationssqlc.ListAssessmentTargetStudentUserIDsParams{
		OrganizationID: orgUUID,
		AssessmentID:   assessmentUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list assessment target students: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.String())
	}
	return out, nil
}

func mapRow(row notificationssqlc.Notification) Notification {
	// Preserve the raw jsonb bytes so the handler can pass them through
	// without a round-trip. Callers that need a map can decode on
	// demand.
	meta := []byte("{}")
	if len(row.MetadataJson) > 0 {
		meta = row.MetadataJson
	}
	return Notification{
		ID:           row.ID.String(),
		OrgID:        row.OrganizationID.String(),
		RecipientID:  row.RecipientUserID.String(),
		EventType:    row.EventType,
		Title:        row.Title,
		Body:         row.Body,
		MetadataJSON: meta,
		IsRead:       row.IsRead,
		ReadAt:       timeOrNil(row.ReadAt),
		CreatedAt:    formatTime(row.CreatedAt),
	}
}
