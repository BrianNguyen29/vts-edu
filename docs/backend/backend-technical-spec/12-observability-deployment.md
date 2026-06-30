# 12. Observability, Deployment & Operations

## 1. Structured logging

Dùng `log/slog` JSON production.

Required fields:

```text
timestamp
level
message
service
version
environment
request_id
trace_id (nếu có)
organization_id (nếu có)
user_id (nếu có)
route
status
latency_ms
error_code
```

Redaction:

- Password/token/cookie.
- Authorization header.
- Full request body.
- Essay answer.
- Sensitive profile fields.

## 2. Metrics

MVP có thể expose `/metrics` restricted hoặc dùng provider metrics.

### HTTP

- Request count by route/method/status.
- Duration histogram.
- In-flight requests.

### DB

- Pool total/acquired/idle.
- Acquire latency.
- Query error/slow query count.
- Transaction rollback.

### Domain

- Attempts started/submitted/expired.
- Answer save success/conflict/error.
- Submit latency.
- Grading queue latency/failure.
- Upload processing failure.
- Grade publish count.

Không dùng raw user/class IDs làm metric labels vì cardinality cao.

## 3. Tracing

OpenTelemetry optional ở MVP nhưng code nên có context propagation.

Trace quan trọng:

- Submit transaction.
- Auto-grade worker.
- File processing.
- Export.

Không export sensitive attributes.

## 4. Health checks

### Liveness

- Process event loop/goroutine server alive.
- Không query DB.

### Readiness

- PostgreSQL ping với timeout ngắn.
- Required config loaded.
- Optional storage check không nên làm mỗi request quá nặng.

## 5. Deployment model MVP demo

```text
Vercel Hobby (React SPA)
        |
        | HTTPS REST JSON cross-origin
        v
Render Free Go service
        |
        +--> Supabase Free PostgreSQL 15+
        +--> Supabase Storage
```

- **Vercel Hobby:** static SPA build, preview per branch.
- **Render Free:** một Docker Web Service chạy Go binary; HTTP API + River workers + scheduler in-process.
- **Supabase Free:** PostgreSQL 15+ (500 MB, 1 GB egress), Supavisor transaction pooler, Storage 2 GB.

Giới hạn demo:

- Không production SLA.
- Render Free cold start, single-instance.
- Supabase Free pool nhỏ; max connections app 3–5 để tránh cạn pool.
- Backup không tự động hoàn chỉnh; cần script export định kỳ.

Cho production có học sinh thật/commercial: nâng cấp Vercel Pro + Render paid + Supabase Pro.

## 6. Docker image (optional for local/scale)

Multi-stage (dùng khi cần image thay vì Render native build):

```dockerfile
# build Go static binary
# final distroless/alpine-compatible image as appropriate
# copy migrations + binary
```

Yêu cầu:

- Non-root user.
- Read-only filesystem nếu khả thi.
- Health endpoint.
- Graceful shutdown.
- Không chứa build secrets.

## 7. Graceful shutdown

1. Stop accepting new HTTP requests.
2. Allow in-flight requests trong timeout.
3. Stop taking new River jobs.
4. Wait/cancel workers safely.
5. Close DB pool.

Submit transaction đang chạy không được bị cắt tùy tiện nếu có thể hoàn thành trong grace period.

## 8. Database connection sizing

Không đặt pool lớn mặc định.

Ví dụ demo Supabase Free:

- Supavisor transaction pooler giới hạn connections; app nên dùng pool nhỏ (max 3–5).
- Worker concurrency tính trong cùng pool hoặc pool riêng có budget.
- Query timeout theo class endpoint.

Tổng tất cả replicas phải nhỏ hơn DB max connections có headroom.

## 9. Backup & restore

### Database

- Supabase Free có daily backup theo policy nền tảng; **không nên coi là đủ cho dữ liệu quan trọng**.
- Tự động hóa dump định kỳ ra object storage khác (ví dụ Supabase Storage hoặc local) cho demo.
- PITR chỉ có ở Supabase Pro.
- Retention tối thiểu theo policy.
- Restore test định kỳ vào environment cô lập.

### Object storage

- Supabase Storage versioning/lifecycle nếu chi phí cho phép.
- Không coi DB backup đã bao gồm binary files.
- DB metadata và object backup phải có recovery procedure tương thích.

## 10. Release process

1. CI tests/security checks.
2. Build immutable image tagged commit SHA.
3. Backup/check migration risk.
4. Run migration job một lần.
5. Deploy app.
6. Verify readiness/smoke tests.
7. Monitor error/latency.
8. Roll back app nếu cần; migration phải backward compatible.

## 11. Environments

| Environment | Mục tiêu |
|---|---|
| Local | Docker Postgres/MinIO/Mailpit |
| CI | Ephemeral services |
| Staging | Vercel preview + Render preview + Supabase project riêng, synthetic data |
| Demo/Production | Vercel Hobby/Pro + Render Free/paid + Supabase Free/Pro; real users only when upgraded |

Không copy production PII xuống local. Demo data nên là synthetic/non-critical.

Không copy production PII xuống local.

## 12. Scale triggers

| Signal | Action |
|---|---|
| User thật/commercial hoặc data nhạy cảm | Nâng Vercel Pro + Render paid + Supabase Pro |
| >30 concurrent active | Render paid min instance hoặc thêm replica; tách `cmd/api` và `cmd/worker` |
| DB >400 MB hoặc connection pressure | Supabase Pro; tối ưu pool/query; xem xét PgBouncer/Supavisor |
| API CPU cao, DB ổn | Thêm API replica |
| Queue backlog | Tách/tăng worker |
| Cold start ảnh hưởng UX | Render paid min instance hoặc region gần user |
| Read reports chậm | Index/materialized view/read replica |
| Distributed rate limit cần thiết | Thêm Redis |
| Search PostgreSQL không đạt | Đánh giá OpenSearch sau profiling |
| Async grading load lớn | Tách auto-grade thành async River job |

## 13. Incident runbook tối thiểu

- DB unavailable.
- Object storage unavailable.
- Queue backlog.
- Attempt submit error spike.
- Suspected cross-tenant leak.
- Compromised signing key.
- Failed migration.
- Restore from backup.
