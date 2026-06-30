# 01. Tech Stack

## 1. Tóm tắt lựa chọn

| Lớp | Công nghệ | Trạng thái MVP | Lý do |
|---|---|:---:|---|
| Language | Go 1.25.x | Bắt buộc | Binary gọn, concurrency tốt, vận hành đơn giản |
| HTTP Router | `go-chi/chi/v5` | Bắt buộc | Chuẩn `net/http`, nhẹ, composable |
| API/OpenAPI | Huma + chi adapter | Bắt buộc | Sinh OpenAPI 3.1 và validation từ type |
| SQL Database | PostgreSQL 15+ | Bắt buộc | Transaction, constraints, JSONB, FTS, RLS, mature tooling; Supabase Free cho demo |
| NoSQL | Không dùng | Không | Không có use case biện minh thêm datastore |
| Driver | `pgx/v5` + `pgxpool` | Bắt buộc | PostgreSQL-native, hiệu năng và feature tốt |
| Query generation | `sqlc` | Bắt buộc | Type-safe SQL, minh bạch hơn ORM |
| ORM/ODM | Không dùng | Không | Tránh hidden query và abstraction không cần thiết |
| Migration | `pressly/goose` | Bắt buộc | SQL migration tuần tự, dễ nhúng vào binary/CI |
| Background queue | River | Bắt buộc | Queue transactional trên PostgreSQL, không cần Redis |
| Distributed cache | Không dùng ở MVP | Không | Giảm hạ tầng; chỉ thêm khi profiling chứng minh |
| Local cache | Bounded in-process cache, tùy module | Hạn chế | Chỉ cache reference data không quan trọng |
| Object storage | S3-compatible / Supabase Storage | Bắt buộc | File private, signed URL, dễ đổi nhà cung cấp; Supabase Storage cho demo |
| Authentication | JWT access + opaque refresh | Bắt buộc | API rõ, access ngắn hạn, refresh có thể revoke |
| Password hashing | Argon2id | Bắt buộc | Thuật toán phù hợp cho password hashing |
| Logging | `log/slog` JSON | Bắt buộc | Standard library, structured logging |
| Metrics/tracing | OpenTelemetry tùy cấu hình | Từng bước | Không khóa vendor, bật khi cần |
| Frontend package manager | pnpm workspace | Bắt buộc phía JS | Nhanh, tiết kiệm disk, workspace rõ ràng |

## 2. Version policy

- **Go:** pin major/minor `1.25` trong `go.mod`; CI sử dụng patch mới nhất thuộc nhánh được hỗ trợ.
- **PostgreSQL:** chạy major stable `15+`; luôn cập nhật minor mới nhất của major. Supabase Free hiện cung cấp PostgreSQL 15+.
- Dependency Go dùng version cụ thể trong `go.mod` và `go.sum`.
- Không dùng floating tag như `latest` trong production image.
- Renovate/Dependabot chỉ tạo PR; không tự merge dependency quan trọng.

## 3. Go + chi

`chi` được chọn vì:

- Tương thích `net/http`.
- Middleware chuẩn, dễ test bằng `httptest`.
- Route group và sub-router phù hợp module.
- Không ép application vào custom context/response API.
- Dễ thay hoặc giảm dependency nếu cần.

Không chọn framework lớn hơn chỉ để có CRUD nhanh, vì phần khó của dự án là transaction và domain invariant, không phải route registration.

## 4. Huma

Huma chịu trách nhiệm:

- OpenAPI 3.1 generation.
- JSON Schema cho request/response.
- Parse và structural validation.
- API docs trong development/staging.
- Chuẩn hóa error response.

Huma **không** chịu trách nhiệm:

- Business validation.
- Authorization resource-level.
- Transaction boundary.
- Repository query.

## 5. PostgreSQL

PostgreSQL là datastore duy nhất cho structured data vì cần:

- ACID transaction cho submit, grade và queue enqueue.
- Foreign keys và check constraints.
- Partial/unique indexes.
- `NUMERIC` cho điểm.
- JSONB cho payload có cấu trúc biến đổi vừa phải.
- Full-text search cho giai đoạn đầu.
- Row-Level Security như lớp phòng vệ bổ sung khi cần.

### Không dùng NoSQL

Không thêm MongoDB/Document DB vì:

- Domain có quan hệ mạnh.
- Attempt, answer, grade cần transaction và constraint.
- JSONB đủ cho answer payload hoặc question config linh hoạt.
- Thêm datastore làm tăng backup, migration, monitoring và consistency cost.

## 6. pgx + sqlc

### pgx

Sử dụng `pgxpool.Pool` làm connection pool. Không dùng `database/sql` trừ khi dependency bắt buộc.

### sqlc

SQL là nguồn sự thật cho query. sqlc sinh:

- Query methods.
- Input parameter structs.
- Result structs.
- Interface `DBTX` phù hợp transaction.

Quy tắc:

- Không sửa file generated.
- Query đặt theo module.
- Query list phải có pagination.
- Write query quan trọng phải kiểm tra số row affected.
- Query tenant phải có `organization_id` trong điều kiện.

## 7. Caching strategy

### MVP

Không có Redis. Sử dụng:

- Browser/HTTP caching cho static asset.
- ETag/Last-Modified cho resource phù hợp.
- In-process bounded cache cho reference data ít thay đổi, nếu profiling cho thấy cần.
- Materialized view hoặc precomputed table cho report chậm.

### Không cache

Không cache:

- Attempt state.
- Current answers.
- Final grade.
- Authorization quyết định nhạy cảm.
- Refresh token/session source of truth.

### Điều kiện thêm Redis

Thêm Redis khi ít nhất một điều đúng:

- Nhiều API replicas cần distributed rate limit.
- Có hot read workload mà Postgres/index không giải quyết được.
- Có real-time presence/pub-sub không phù hợp PostgreSQL.
- Load test chỉ ra cache đem lại lợi ích đo được.

## 8. Background jobs: River

River dùng cùng PostgreSQL để:

- Enqueue trong cùng transaction với business data.
- Retry có kiểm soát.
- Scheduled jobs.
- Không cần Redis/RabbitMQ.

Use case:

- Auto-grade sau submit.
- Auto-expire attempt.
- Gửi email.
- Tạo export.
- Quét file.
- Sinh preview.
- Recalculate analytics không đồng bộ.

## 9. JWT library

Khuyến nghị `github.com/golang-jwt/jwt/v5`.

Yêu cầu validation:

- Chỉ chấp nhận algorithm cấu hình cố định.
- Validate `iss`, `aud`, `exp`, `nbf` nếu có.
- Không tin `alg` từ token mà không whitelist.
- Có key ID (`kid`) để hỗ trợ rotation.
- Access token 10–15 phút.
- Refresh token là random opaque token, không phải JWT dài hạn.

## 10. Storage

Interface nội bộ:

```go
type ObjectStore interface {
    Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
    PresignPut(ctx context.Context, key string, expires time.Duration) (string, error)
    PresignGet(ctx context.Context, key string, expires time.Duration) (string, error)
    Head(ctx context.Context, key string) (ObjectMeta, error)
    Delete(ctx context.Context, key string) error
}
```

Development dùng MinIO. Production/demo dùng Supabase Storage hoặc nhà cung cấp S3-compatible phù hợp vị trí dữ liệu và chi phí. Supabase Storage cần cấu hình CORS allowlist origin Vercel.

## 11. Các công nghệ cố ý không chọn

| Công nghệ | Lý do chưa dùng |
|---|---|
| GORM | Query phức tạp khó kiểm soát; sqlc phù hợp hơn |
| MongoDB | Không cần datastore thứ hai |
| Redis | Chưa có distributed cache/lock use case bắt buộc |
| Kafka | Quá nặng cho solo MVP |
| GraphQL | Tăng surface area; REST/OpenAPI đủ |
| gRPC public API | Browser client và debugging REST thuận tiện hơn |
| Elasticsearch/OpenSearch | PostgreSQL FTS đủ ở quy mô đầu |
| Kubernetes | Chi phí vận hành không phù hợp solo |

## 12. Nguồn chính

Xem [REFERENCES.md](REFERENCES.md) cho tài liệu chính thức về Go, PostgreSQL, chi, Huma, pgx, sqlc, goose, River, JWT và OWASP.
