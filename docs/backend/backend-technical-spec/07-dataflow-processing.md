# 07. Dataflow & Processing Logic

## 1. Generic request flow

```mermaid
sequenceDiagram
    participant Client
    participant Proxy as CDN/Reverse Proxy
    participant MW as Go Middleware
    participant Handler
    participant Service as Application Service
    participant Authz as Authorizer
    participant Repo as Repository/sqlc
    participant DB as PostgreSQL

    Client->>Proxy: HTTPS request
    Proxy->>MW: forwarded request
    MW->>MW: request-id, recover, security, log, rate limit
    MW->>MW: authenticate JWT, build actor context
    MW->>Handler: validated route request
    Handler->>Service: typed command/query
    Service->>Authz: permission + resource scope
    Authz-->>Service: allow/deny
    Service->>Repo: transaction/query
    Repo->>DB: parameterized SQL
    DB-->>Repo: rows/error
    Repo-->>Service: domain/result
    Service-->>Handler: output/error
    Handler-->>Client: JSON or Problem Details
```

## 2. Layer responsibilities

### Middleware

- Request ID.
- Panic recovery.
- Security headers.
- Access logging.
- Body size limit.
- CORS/origin policy.
- Authentication token parsing.
- Coarse rate limiting.

Không làm:

- Query class ownership.
- Tính điểm.
- State transition.

### Handler/controller

- Bind input.
- Structural validation do Huma hỗ trợ.
- Lấy actor từ context.
- Gọi đúng use case.
- Map output.

### Application service

- Business validation.
- Authorization resource-level.
- Transaction boundary.
- Gọi domain methods.
- Repository orchestration.
- Enqueue job transactional.

### Domain

- Invariant.
- State transition.
- Pure calculation.
- Không biết HTTP/SQL.

### Repository

- sqlc calls.
- Map DB row ↔ domain/data model.
- Translate known DB errors.
- Không quyết định permission.

## 3. Authentication flow

```mermaid
sequenceDiagram
    participant C as Client
    participant A as Auth Handler
    participant S as Auth Service
    participant DB as PostgreSQL

    C->>A: login credentials
    A->>S: LoginCommand
    S->>DB: load org/user/membership
    S->>S: verify Argon2id
    S->>DB: insert refresh session
    S->>S: sign short-lived JWT
    S->>DB: append audit/login event
    S-->>A: token pair metadata
    A-->>C: access token + HttpOnly refresh cookie
```

## 4. Publish assessment flow

```text
request
-> authenticate teacher
-> load assessment scoped by org
-> authorize assessment:publish and class assignment
-> acquire assessment row lock
-> require DRAFT
-> validate schedule/settings/targets
-> resolve every fixed item and random-rule pool
-> copy immutable question/version content to snapshots
-> create publication version
-> set assessment status SCHEDULED/OPEN according to server time
-> append audit event
-> store idempotent response
-> commit
```

Nếu một question version không hợp lệ hoặc bị archive theo policy, toàn bộ publish rollback.

## 5. Start attempt flow

```text
client start request + idempotency key
-> verify token
-> load assessment publication
-> verify target/enrollment/time/max attempts
-> resolve or return existing resumable attempt
-> compute effective duration/accommodation
-> select and order snapshot items
-> insert attempt + attempt_items + event
-> persist idempotent response
-> commit
-> return server_time/expires_at/items metadata
```

## 6. Autosave flow

```text
client debounce save
-> PUT answer with expected revision
-> authenticate and verify attempt ownership
-> verify server deadline/status
-> validate payload against snapshot type
-> optimistic update/insert answer
-> return new revision + server saved_at
-> client removes local pending item only after 2xx
```

Client retry cùng expected revision sau network timeout:

- Nếu server chưa commit: update thành công.
- Nếu server đã commit: revision mismatch. API có thể so request hash/answer equality và trả current state, hoặc 409 để client fetch current answer. Khuyến nghị lưu client operation ID nhỏ trong answer event/idempotency scope nếu cần retry trong mạng yếu.

## 7. Submit & auto-grade flow

### Demo: synchronous MCQ/simple grading

```mermaid
sequenceDiagram
    participant C as Client
    participant API
    participant DB

    C->>API: submit
    API->>DB: transaction + lock attempt
    API->>DB: status=SUBMITTED or EXPIRED, append event
    API->>API: auto-grade MCQ/simple items synchronously
    API->>DB: upsert grading results idempotently
    API->>DB: set MANUAL_REVIEW or FINALIZED
    API->>DB: update source grade item/entry if policy
    API->>DB: commit
    API-->>C: submitted with grading_status
```

Điều kiện đồng bộ:

- Tất cả item là MCQ/simple auto-grade.
- Tổng thởi gian grading < timeout request (ví dụ < 2 giây).
- Không có essay cần manual review.

Nếu không đủ điều kiện, enqueue River job và trả `grading_status=QUEUED`.

### Scale: async grading via River

```mermaid
sequenceDiagram
    participant C as Client
    participant API
    participant DB
    participant W as River Worker

    C->>API: submit
    API->>DB: transaction + lock attempt
    API->>DB: status=SUBMITTED/EXPIRED, append event
    API->>DB: insert River job transactionally
    API->>DB: commit
    API-->>C: submitted
    W->>DB: load attempt snapshots/answers
    W->>W: calculate auto scores
    W->>DB: upsert grading results idempotently
    W->>DB: set MANUAL_REVIEW or FINALIZED
    W->>DB: update source grade item/entry if policy
```

## 8. File upload flow

Tách metadata transaction và network upload. Không giữ transaction khi client gửi bytes.

1. Create upload intent.
2. Client PUT trực tiếp storage.
3. Complete callback.
4. API `HEAD` object.
5. Mark processing và enqueue scan.
6. Worker scan/inspect.
7. Mark ready/rejected.

## 9. Grade publish flow

```text
teacher selects grade items
-> authorize on class
-> validate every entry status/review completeness
-> transaction:
     create grade_publication
     mark selected entries published
     enqueue notification fan-out jobs
     append audit
-> commit
-> return count
```

## 10. Error propagation

```text
pgx/sqlc error
-> repository maps known SQLSTATE
-> application maps domain error
-> HTTP layer maps to Problem Details
-> unexpected error logged with request_id
-> client receives generic 500, no internal detail
```

Examples:

| Internal error | HTTP |
|---|---|
| `domain.ErrAttemptExpired` | 409 `ATTEMPT_EXPIRED` |
| `repository.ErrNotFound` | 404 |
| unique violation username | 409 `USERNAME_EXISTS` |
| validation error | 422 |
| context deadline | 503/504 tùy layer |
