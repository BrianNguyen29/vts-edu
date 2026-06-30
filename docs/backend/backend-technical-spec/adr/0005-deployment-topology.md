# ADR-0005: MVP Demo Deployment Topology

- **Status:** Superseded — the current demo stack uses Render Free for the backend instead of Koyeb.
- **Date:** 2026-06-29
- **Updated:** 2026-06-30

## Context

MVP demo cần zero-cost deployment để validate product với ngườ dùng nhỏ, học sinh thật trong phạm vi hạn chế, hoặc demo nội bộ. Các ràng buộc:

- Không chi phí hạ tầng nếu có thể.
- Không vận hành VPS/VM/mail server.
- Frontend SPA deploy đơn giản từ Git.
- Backend Go có thể chạy như một service/container.
- PostgreSQL và object storage được quản lý.

## Decision

Demo stack:

```text
Frontend : Vercel Hobby (React SPA/PWA-ready)
Backend  : Render Free Go API (một service, một region)
Database : Supabase Free PostgreSQL 15+
Storage  : Supabase Storage (S3-compatible, private bucket)
Queue    : River in-process trong Render service
```

### Chi tiết

- **Vercel Hobby:** build từ Git, preview deployment per branch, 100 GB bandwidth, giới hạn function/serverless không áp dụng vì đây là static SPA.
- **Render Free:** một Docker Web Service với instance type free tier; Go binary chạy HTTP API + River workers + scheduler in-process. Không dùng Render background workers riêng ở demo.
- **Supabase Free:** PostgreSQL 15+ (không phải 18), 500 MB DB, 1 GB egress, 2 GB file storage, Supavisor transaction pooler. Không dùng Supabase Auth; auth do backend tự quản lý.
- **Supabase Storage:** private bucket, signed URL ngắn hạn, CORS allowlist origin Vercel.

### Cross-origin auth

Vercel origin khác Render API origin, do đó:

- Frontend `apiBaseUrl` là absolute URL, ví dụ `https://<api>.onrender.com/api/v1`, cấu hình qua biến môi trường/build-time tùy pipeline.
- Refresh cookie: `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth`.
- CORS allowlist chính xác Vercel origin (production + preview).
- CSRF token bắt buộc cho cookie-backed auth endpoints/mutations (double-submit cookie hoặc header).

## Consequences

### Positive

- Zero cost cho demo/synthetic data.
- Không cần quản lý server OS, reverse proxy, CDN.
- Vercel preview branch hữu ích cho review.
- River in-process đơn giản, không cần queue broker riêng.

### Negative

- Render Free cold start và single-instance limit; không có SLA.
- Supabase Free giới hạn DB 500 MB, connection pool nhỏ, egress 1 GB.
- Cross-origin cookie phức tạp hơn same-origin; cần CSRF và CORS chặt.
- Vercel preview origin động; CORS allowlist cần cập nhật hoặc pattern hợp lý.
- Backup/restore không tự động đầy đủ như paid plan; cần script export định kỳ.

## Scale path

| Trigger | Action |
|---|---|
| User thật/commercial | Vercel Pro + Render paid + Supabase Pro |
| >30 concurrent active | Render paid hoặc thêm replica; tách `cmd/api` và `cmd/worker` |
| DB >400 MB hoặc connection pressure | Supabase Pro; tối ưu pool/query; xem xét PgBouncer/Supavisor |
| Cold start ảnh hưởng UX | Render paid min instance hoặc region gần user |
| Grading/complex async load | Tách worker process, dùng async grading cho essay/aggregate |

## Notes

- Đây là **demo stack**, không phải production SLA cho dữ liệu học sinh quan trọng.
- Không lưu PII nhạy cảm hoặc điểm quan trọng trên demo nếu chưa có backup/restore đáng tin cậy.
- Theo dõi giới hạn free tier: Render hours, Supabase egress/storage, Vercel bandwidth.
