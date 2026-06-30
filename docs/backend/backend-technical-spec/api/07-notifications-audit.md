# API 07 — Notifications, Audit & System Operations

Base path: `/api/v1`

## 1. Notification endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| GET | `/me/notifications` | Danh sách notification | — | `{"data":[{"id":"...","type":"GRADE_PUBLISHED","read_at":null}]}` |
| GET | `/me/notifications/unread-count` | Số chưa đọc | — | `{"data":{"count":5}}` |
| POST | `/me/notifications/{notification_id}/read` | Đánh dấu đã đọc | `{}` | `204 No Content` |
| POST | `/me/notifications/read-all` | Đọc tất cả | `{}` | `204 No Content` |
| GET | `/me/notification-preferences` | Cấu hình nhận | — | `{"data":{"email_due_reminder":true}}` |
| PUT | `/me/notification-preferences` | Cập nhật | `{"email_due_reminder":false}` | `{"data":{"email_due_reminder":false}}` |

## 2. Audit endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| GET | `/audit-logs` | Search audit theo permission | — | `{"data":[{"action":"grade.override","actor_user_id":"...","created_at":"..."}]}` |
| GET | `/audit-logs/{audit_id}` | Chi tiết redacted | — | `{"data":{"id":"...","metadata":{...}}}` |
| GET | `/attempts/{attempt_id}/timeline` | Timeline attempt chuyên biệt (owned by attempts module) | — | `{"data":[{"event":"ANSWER_SAVED","at":"...","metadata":{"attempt_item_id":"..."}}]}` |

## 3. Health endpoints

| Method | URL | Description | Response |
|---|---|---|---|
| GET | `/health/live` | Process alive; không query dependency nặng | `200 {"status":"ok"}` |
| GET | `/health/ready` | DB và dependency thiết yếu sẵn sàng | `200/503` |
| GET | `/version` | Build metadata không nhạy cảm | `{"version":"...","commit":"..."}` |

Health endpoint có thể nằm ngoài `/api/v1` tùy reverse proxy convention.

## 4. Notification model

Notification được tạo từ domain event/job, không tạo trực tiếp từ handler UI trừ admin announcement.

```text
notification
  type
  title/template key
  actor/resource references
  metadata (minimal)
  created_at

notification_recipient
  notification_id
  user_id
  channel
  delivery_status
  read_at
```

## 5. Audit access

- Chỉ role có `audit:view`.
- Audit response redacted.
- Không cho filter tùy ý trên raw JSON.
- Export audit là async job và phải audit chính hành động export.
