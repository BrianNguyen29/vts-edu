# 08. Core Functions

## 1. Authentication

### 1.1 Password hashing

- Argon2id.
- Salt ngẫu nhiên per password.
- Encoded hash chứa parameters để nâng cost sau này.
- Parameters benchmark trên production-like hardware; mục tiêu hash đủ chậm để chống brute force nhưng không gây DoS.
- Verify path chống timing leak ở mức thư viện/flow hợp lý.
- Rehash khi parameters cũ thấp hơn policy.

Interface:

```go
type PasswordHasher interface {
    Hash(password string) (encoded string, err error)
    Verify(password, encoded string) (match bool, needsRehash bool, err error)
}
```

### 1.2 Access token

- JWT ký bằng asymmetric key nếu có khả năng nhiều service/verifier về sau; EdDSA/ES256 hoặc RS256 tùy thư viện/ops.
- Với một service solo, HS256 có thể vận hành nhưng key sharing/rotation kém hơn. Khuyến nghị EdDSA nếu deployment tooling quản lý key tốt.
- TTL 10–15 phút.
- Validate issuer, audience, expiration, algorithm và key ID.
- Access token lưu trong memory của SPA, không localStorage.

### 1.3 Refresh token

- 32 bytes random trở lên.
- Base64url encode.
- Chỉ raw token trong HttpOnly cookie.
- DB lưu hash SHA-256/HMAC của token.
- Rotate mỗi lần refresh.
- Có token family để phát hiện reuse.
- Revoke khi logout/reset password/suspend.

## 2. Authorization

### 2.1 Actor context

```go
type Actor struct {
    UserID         uuid.UUID
    OrganizationID uuid.UUID
    SessionID      uuid.UUID
    Roles          []string
    AuthVersion    int64
}
```

### 2.2 Authorizer

```go
type Authorizer interface {
    Require(ctx context.Context, actor Actor, permission Permission, resource ResourceScope) error
}
```

`ResourceScope` có thể chứa class ID, owner ID, organization ID và state cần kiểm tra.

Không chỉ dựa vào JWT role. Repository query phải scope org, service phải kiểm tra assignment/enrollment.

## 3. Validation

### Structural validation

- Required fields.
- Type/format.
- String length.
- Enum value.
- Array size.
- UUID/RFC3339.

Thực hiện ở Huma/input type.

### Business validation

- Opens before closes.
- Teacher belongs to class.
- Question answer key hợp lệ.
- Score within max.
- State transition allowed.
- Attempt count.

Thực hiện ở application/domain.

### Sanitization

Rich text:

- Backend lưu canonical structured content.
- Sanitize allowlist HTML ở write hoặc render pipeline.
- Không tin frontend sanitize.
- Link scheme allowlist.

## 4. Global error handling

### Domain error taxonomy

```go
var (
    ErrNotFound        = errors.New("not found")
    ErrForbidden       = errors.New("forbidden")
    ErrConflict        = errors.New("conflict")
    ErrInvalidState    = errors.New("invalid state")
    ErrValidation      = errors.New("validation")
)
```

Dùng typed errors để giữ code/context:

```go
type AppError struct {
    Code       string
    Message    string
    Status     int
    FieldErrors []FieldError
    Cause      error
}
```

Không trả `Cause` ra client.

### Panic recovery

- Recovery middleware bắt panic.
- Log stack + request ID.
- Trả generic 500.
- Không recover panic rồi tiếp tục transaction không rõ trạng thái.

## 5. Transaction helper

```go
type TxManager interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context, q *db.Queries) error) error
}
```

Yêu cầu:

- Rollback khi error/panic.
- Context cancellation.
- Isolation mặc định Read Committed; nâng isolation/lock theo use case.
- Không nested transaction giả; nếu cần truyền `DBTX`/queries trong context/use case explicit.

## 6. Idempotency

Middleware/service phối hợp:

1. Validate key format/length.
2. Hash key và request body canonical hash.
3. Trong transaction, insert claim row hoặc lock existing row.
4. Nếu completed cùng request: replay status/body.
5. Nếu same key khác request: conflict.
6. Nếu processing: 409/425 hoặc wait ngắn theo policy.
7. Business operation và response persistence cùng transaction.

Không dùng idempotency cho GET.

## 7. Optimistic concurrency

Dùng `version` hoặc `revision`:

```sql
UPDATE assessments
SET title = $1, version = version + 1
WHERE id = $2 AND organization_id = $3 AND version = $4;
```

0 rows → 409/412.

## 8. Grading functions

### Single choice

```text
correct selected ID == answer key -> max score
else -> 0
```

### Multiple choice

MVP nên hỗ trợ policy rõ:

- `ALL_OR_NOTHING` trước.
- Partial credit chỉ thêm khi có test ma trận đầy đủ.

### Numeric

```text
abs(answer - target) <= absolute_tolerance
OR relative tolerance policy
```

Dùng decimal/rational logic phù hợp, không float tùy tiện cho điểm.

### Essay

- Mark `MANUAL_REVIEW_REQUIRED`.
- Không tự động final score.

Mọi grader phải pure/idempotent trên snapshot + answer.

## 9. Time handling

Inject clock:

```go
type Clock interface {
    Now() time.Time
}
```

- Tests dùng fake clock.
- Không gọi `time.Now()` rải rác trong domain/application.
- DB timestamps quan trọng có thể dùng application clock thống nhất hoặc DB `clock_timestamp()` theo policy; tránh trộn không kiểm soát.

## 10. Rate limiting

MVP single instance:

- In-memory token bucket/sliding window.
- Login theo IP + identifier.
- Refresh theo session/IP.
- Upload intent, export, AI future endpoints giới hạn riêng.

Khi multi-instance, chuyển rate-limit state sang Redis hoặc edge provider.

Rate limit không phải authorization.

## 11. Request logging

Log fields:

```text
request_id
method
route_template
status
latency_ms
user_id (nếu auth)
organization_id
remote_ip_prefix/hash
user_agent_hash
error_code
```

Không log body mặc định.

## 12. Configuration

- Environment variables.
- Parse thành typed struct lúc startup.
- Fail fast nếu missing/invalid.
- Secret không có default production.
- `.env` chỉ development; không commit secret.

Ví dụ:

```text
APP_ENV
HTTP_ADDR
DATABASE_URL
JWT_ISSUER
JWT_AUDIENCE
JWT_PRIVATE_KEY_FILE / secret reference
S3_ENDPOINT
S3_BUCKET
S3_ACCESS_KEY_ID
S3_SECRET_ACCESS_KEY
```
