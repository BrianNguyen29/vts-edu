-- 000021_notifications.sql
-- Notification inbox MVP per slice-15. One row per (org, recipient,
-- event) with an optional JSON metadata payload. Best-effort inserts
-- are issued from the business transaction's caller after the
-- transaction commits, so a notification failure does not roll the
-- caller back.

CREATE TABLE notifications (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id     uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  recipient_user_id   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  event_type          text NOT NULL,
  title               text NOT NULL,
  body                text NOT NULL DEFAULT '',
  metadata_json       jsonb NOT NULL DEFAULT '{}'::jsonb,
  is_read             boolean NOT NULL DEFAULT false,
  read_at             timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now()
);

-- Tenant isolation + inbox pagination.
CREATE INDEX idx_notifications_inbox
  ON notifications (organization_id, recipient_user_id, created_at DESC, id DESC);

-- Cheap unread counter for the bell badge.
CREATE INDEX idx_notifications_unread
  ON notifications (organization_id, recipient_user_id, is_read)
  WHERE is_read = false;
