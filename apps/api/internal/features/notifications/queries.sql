-- name: InsertNotification :one
INSERT INTO notifications (
  organization_id, recipient_user_id, event_type, title, body, metadata_json
) VALUES (
  $1, $2, $3, $4, $5, $6::jsonb
)
RETURNING id, organization_id, recipient_user_id, event_type, title, body, metadata_json, is_read, read_at, created_at;

-- name: ListNotifications :many
SELECT
  id, organization_id, recipient_user_id, event_type, title, body, metadata_json,
  is_read, read_at, created_at
FROM notifications
WHERE organization_id = $1
  AND recipient_user_id = $2
  AND ($3::timestamptz IS NULL OR created_at < $3::timestamptz)
ORDER BY created_at DESC, id DESC
LIMIT $4;

-- name: CountUnread :one
SELECT COUNT(*)
FROM notifications
WHERE organization_id = $1
  AND recipient_user_id = $2
  AND is_read = false;

-- name: GetNotification :one
SELECT
  id, organization_id, recipient_user_id, event_type, title, body, metadata_json,
  is_read, read_at, created_at
FROM notifications
WHERE id = $1 AND organization_id = $2 AND recipient_user_id = $3;

-- name: MarkRead :one
UPDATE notifications
SET is_read = true,
    read_at = COALESCE(read_at, now())
WHERE id = $1 AND organization_id = $2 AND recipient_user_id = $3
RETURNING id, organization_id, recipient_user_id, event_type, title, body, metadata_json, is_read, read_at, created_at;

-- name: ListClassStudentUserIDs :many
-- Used by the resources publish notifier to expand a class-scoped
-- resource to its currently enrolled students.
SELECT m.user_id
FROM enrollments e
JOIN organization_memberships m ON m.id = e.membership_id
WHERE e.organization_id = $1
  AND e.class_section_id = $2
  AND e.status = 'ACTIVE'
  AND m.status = 'ACTIVE';

-- name: ListAssessmentTargetStudentUserIDs :many
-- Expand the active target classes of an assessment into distinct
-- student user ids for assessment.published notifications.
SELECT DISTINCT m.user_id
FROM assessment_targets t
JOIN enrollments e
  ON e.organization_id = t.organization_id
 AND e.class_section_id = t.class_section_id
 AND e.status = 'ACTIVE'
JOIN organization_memberships m ON m.id = e.membership_id
WHERE t.organization_id = $1
  AND t.assessment_id = $2
  AND t.status = 'ACTIVE'
  AND m.status = 'ACTIVE';
