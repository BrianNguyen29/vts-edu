# 06. API Design Conventions

## 1. Base URL & versioning

```text
/api/v1
```

Không đưa version vào media type ở MVP.

## 2. Resource naming

- Dùng plural nouns.
- Không dùng verbs trừ action/state transition rõ ràng.
- Nested route chỉ khi child phụ thuộc context parent.

Ví dụ:

```text
GET  /api/v1/classes/{class_id}/enrollments
POST /api/v1/attempts/{attempt_id}/submit
POST /api/v1/assessments/{assessment_id}/publish
```

## 3. Headers

| Header | Yêu cầu |
|---|---|
| `Authorization: Bearer <access-token>` | Protected API |
| `Content-Type: application/json` | JSON write |
| `Idempotency-Key` | Start attempt, submit, publish, export |
| `If-Match` | Optional optimistic concurrency cho resource edit |
| `X-Request-ID` | Client có thể gửi; server validate/generate |

## 4. Envelope

### Single resource

```json
{
  "data": {
    "id": "019...",
    "name": "Toán 8A"
  }
}
```

### Collection

```json
{
  "data": [],
  "page": {
    "next_cursor": "opaque-token",
    "has_more": false
  }
}
```

Không bọc thêm `success: true`.

## 5. Error format

Dùng `application/problem+json` theo Problem Details:

```json
{
  "type": "https://example.local/problems/validation-error",
  "title": "Validation failed",
  "status": 422,
  "code": "VALIDATION_ERROR",
  "detail": "One or more fields are invalid.",
  "instance": "/api/v1/questions",
  "request_id": "req_01...",
  "errors": [
    {
      "field": "body.title",
      "code": "required",
      "message": "title is required"
    }
  ]
}
```

## 6. Status codes

| Code | Dùng cho |
|---:|---|
| 200 | Read/update/action thành công |
| 201 | Create thành công |
| 202 | Async job đã nhận |
| 204 | Delete/archive không cần body |
| 400 | Malformed request |
| 401 | Thiếu/invalid authentication |
| 403 | Authenticated nhưng không có quyền |
| 404 | Không tìm thấy hoặc cố ý che resource cross-tenant |
| 409 | Conflict/state transition/version/idempotency mismatch |
| 412 | `If-Match` mismatch |
| 422 | Structural/business validation |
| 429 | Rate limit |
| 500 | Unexpected error |
| 503 | Dependency unavailable/readiness failure |

## 7. Pagination

Ưu tiên cursor pagination:

```text
GET /api/v1/questions?limit=50&cursor=...
```

- `limit` mặc định 20, tối đa 100.
- Cursor opaque, signed hoặc base64url payload không nhạy cảm.
- Sort ổn định gồm tie-breaker `id`.

Offset pagination chỉ dùng bảng quản trị nhỏ hoặc export không tương tác.

## 8. Filtering & sorting

```text
GET /questions?status=PUBLISHED&type=SINGLE_CHOICE&tag=algebra&sort=-created_at
```

- Whitelist filter/sort fields.
- Không cho client truyền raw SQL-like expression.

## 9. Date/time

- RFC 3339 UTC, ví dụ `2026-06-29T10:30:00Z`.
- Client timezone chỉ dùng hiển thị.
- Request schedule có thể kèm timezone ở UI, backend nhận UTC đã chuẩn hóa.

## 10. Identifiers

- UUID serialized dạng canonical string.
- Client không được tự đặt organization/user ID trừ import flow được kiểm soát.

## 11. PATCH semantics

Dùng JSON Merge Patch hoặc explicit update DTO. MVP khuyến nghị explicit update DTO để tránh ambiguity:

```json
{
  "title": "Đề giữa kỳ",
  "duration_minutes": 45
}
```

Không dùng generic map update.

## 12. Idempotency

Bắt buộc cho:

- `POST /attempts`.
- `POST /attempts/{id}/submit`.
- `POST /assessments/{id}/publish`.
- Export generation.
- Bulk user import initiation.

Hành vi:

- Cùng key + cùng request hash: trả response cũ.
- Cùng key + request khác: 409 `IDEMPOTENCY_KEY_REUSED`.
- Key hết hạn sau thời gian cấu hình.

## 13. Bulk endpoints

Bulk request phải có giới hạn và partial failure model rõ.

```json
{
  "items": [
    {"external_id": "HS001", "username": "hs001", "display_name": "..."}
  ],
  "dry_run": true
}
```

Response:

```json
{
  "data": {
    "accepted": 1,
    "rejected": 0,
    "errors": []
  }
}
```

## 14. OpenAPI

- OpenAPI 3.1 generated từ Huma.
- Operation ID ổn định: `auth.login`, `attempts.submit`.
- Mọi endpoint có tag, security, request schema, response schema và error responses.
- CI fail nếu generated spec thay đổi mà client types chưa regenerate.
