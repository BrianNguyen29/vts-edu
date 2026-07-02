# 14. Solo Implementation Roadmap

## 1. Guiding strategy

- Làm vertical slice hoàn chỉnh, không xây mọi infrastructure trước.
- Mỗi phase phải có deployable product.
- Bài thi runtime được proof-of-concept sớm vì rủi ro cao nhất.
- Dashboard/AI/gamification làm sau data integrity.

## 2. Phase plan

### Phase 0 — Foundation & proof of concept (2–3 tuần) ✅ Implemented

- Repository/pnpm workspace.
- Go app skeleton, chi.
- PostgreSQL, sequential SQL migrations.
- Structured errors/logging/config.
- CI.
- TxManager + POC transaction boundaries.
- POC attempt autosave revision và concurrent submit.

Exit criteria:

- Duplicate submit không tạo job trùng.
- sqlc/OpenAPI groundwork staged behind Repository interfaces (see ADR-0010).

### Phase 1 — Auth, users, tenancy (3–4 tuần) ✅ Implemented

- Organization.
- User/membership/roles (`membership_roles`).
- Login, JWT, refresh rotation, CSRF.
- Password policy (min 8, mixed case, digit, blocklist).
- Password history (5 hashes) and login lockout (5 failures / 15 min).
- Forced password change (`/auth/change-password`).
- Admin user CRUD + org update.
- Backward-compatible pagination/search for users and assessments.
- Cursor pagination + optional `count` for users, assessments, and audit logs.
- Audit log writes and admin audit-log reader UI for admin actions (create user, update roles, reset password, update org).
- sqlc migration for `assessments`, `admin`, `auth`, and `attempts` repositories, preserving `Repository` interfaces.
- Generated OpenAPI TypeScript types CI check (`pnpm api:types`, `pnpm api:sqlc`).

### Phase 2 — Academic structure (2–3 tuần) ✅ Core implemented

- Terms, subjects, courses, classes (CRUD + archive).
- Teacher assignment (single + bulk).
- Enrollment/bulk import (single + bulk).
- Authorization class scope (teacher sees assigned classes; student access gated via enrollment).

### Phase 3 — Resources/files (2–3 tuần) — MVP implemented (local storage)

- Upload (multipart) and storage seam (local provider; S3/Supabase adapter deferred). ✅
- File states (`ACTIVE` / `ARCHIVED`); latest upload replaces the previous active file. ✅
- Resource CRUD/publish (DRAFT → PUBLISHED → ARCHIVED) with tenant + role checks. ✅
- Bearer-auth download; `Content-Disposition` filename is sanitized; size cap (`MAX_UPLOAD_SIZE`, default 10 MiB). ✅
- Storage keys are server-generated random hex; user-controlled paths never reach the filesystem. ✅
- Basic processing job — deferred.

### Phase 4 — Question bank (3–5 tuần) — Minimal version implemented

- Bank/question/version schema ✅
- Snapshot prompt/choices/answer key into `attempt_items` ✅
- 6 MVP types — deferred; currently MCQ only.
- Validation/publish — deferred until more question types are added.
- Search/filter ✅ (basic picker for builder).

### Phase 5 — Assessment builder (3–4 tuần) — Core implemented

- Assessment/sections/items ✅
- Settings/targets/accommodation ✅ (settings + schedule + targets)
- Validate/publish snapshots ✅
- Duplicate section/item ✅
- Preview ✅
- Teacher assessment list ✅

### Phase 6 — Attempt runtime (4–6 tuần) — Core implemented

- Start/resume ✅
- Save answer/revision ✅
- Submit/expire ✅
- Auto-grade/manual review (MCQ auto-grade ✅)
- Heartbeat/deadline — partial (request-time expiry only; no client heartbeat).
- Load/concurrency tests — deferred.

### Phase 7 — Assignment & gradebook (4–5 tuần) — Core implemented

- Assignment/submission versions/files — deferred.
- Feedback/grade — deferred (MCQ auto-graded; manual review not built).
- Grade items/entries/history ✅
- Publish/export ✅
- Teacher gradebook grid + CSV export ✅

### Phase 8 — Hardening & pilot (3–5 tuần)

- Security negative tests.
- Load tests.
- Backup restore drill.
- Monitoring/alerts.
- Pilot data/import.
- Bug fixing.

## 2.5 Staged Huma/sqlc/OpenAPI plan

- **Current**: Hand-maintained OpenAPI skeleton covers the current API surface in `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`. TypeScript types are generated from it into `apps/web/src/shared/api/openapi-schema.d.ts` using `openapi-typescript` (type-only; no runtime client).
- **Stage 1 — sqlc migration (completed)**: `assessments`, `admin`, `auth`, and `attempts` repositories migrated to sqlc generated queries while keeping `Repository` interfaces stable. No runtime handler/service rewrite.
- **Stage 2 — Huma migration (deferred)**: Add Huma operation definitions behind existing handlers or incrementally replace handler wiring while preserving response envelopes. OpenAPI generation becomes automatic. As of the latest batch (resources MVP + query state + accessibility polish), the manual skeleton covers **63 paths** (up from 58, crossing the original 60-path threshold). `openapi-typescript` + `openapi-fetch` plus the `generated-code-check` CI job continue to keep manual maintenance manageable. Huma remains deferred because migration risk/cost still outweighs manual maintenance, especially for auth cookie/CSRF/refresh-sensitive handlers and middleware ordering. A bounded Huma feasibility spike (no runtime migration) is queued in the next-backlog section to revisit the decision before the spec drifts further. Revisit trigger: spec drift incidents ≥ 2/month, paths ≥ 60 (already reached — see backlog), need runtime contract validation, or a dedicated refactor sprint with ≥ 80% handler test coverage.
- **Stage 3 — Client generation**: Generate frontend API client/types from the Huma/OpenAPI contract once it stabilizes.
- **Breached-password provider (deferred)**: HIBP/external corpus integration deferred pending a privacy/egress ADR; password history + lockout + blocklist implemented as interim controls.

## 2.6 Background jobs / scheduler plan

- **Current**: In-process scheduler in `apps/api/internal/platform/scheduler` runs lightweight `Job` implementations on a fixed interval. First job: `assessment-transition` opens/closes assessments based on `opens_at`/`closes_at`. Controlled by `SCHEDULER_ENABLED` and `SCHEDULER_INTERVAL_SECONDS` (default disabled; enable on Render).
- **River (deferred)**: River/pgvector-style durable queue is not adopted yet. The in-process scheduler is sufficient for scheduled assessment transitions and avoids extra migrations, worker processes, and operational complexity.
- **Large CSV import**: Remains synchronous with a 100-row cap until a durable queue is justified.
- **Async grading**: MCQ grading stays synchronous. Non-MCQ / manual-review async grading is deferred until those question types are implemented.
- **Triggers for River adoption**: need durability/retry (e.g., large CSV, async grading), multi-instance scale-out requiring duplicate-job prevention, cron-like scheduling, or an approved infrastructure sprint covering migration, worker process, monitoring, and dead-letter handling.

## 2.7 Current next backlog (not started)

> **Thứ tự bắt buộc**: (1) hoàn tất cập nhật docs/ADRs còn stale, (2) chạy Huma feasibility spike và quyết định go/no-go, (3) mới triển khai feature mới. Mục tiêu là khóa "docs-completion before new feature work".

**A. Docs & ADR completion (ưu tiên cao nhất, chạy trước feature mới)**

- Cập nhật các roadmap/backend/frontend/ADR còn stale (path count, resources MVP, accessibility baseline, error pages `request_id`).
- Viết ADR mới hoặc cập nhật ADR-0010 với quyết định Huma sau spike.

**B. Huma feasibility spike (sau khi docs xong, trước feature mới)**

- Bounded spike: triển khai Huma trên **một feature ít nhạy cảm** (academics hoặc gradebook), giữ auth/CSRF/refresh ngoài phạm vi, so sánh DX và regression risk với skeleton thủ công. Không migrate runtime toàn cục.
- Quyết định: go (migrate theo slice) hoặc no-go (tiếp tục skeleton thủ công + tăng cường `generated-code-check`).

**C. Feature work (chỉ bắt đầu sau khi A & B xong)**

- Resources/files: signed download (URL hết hạn ngắn) + inline preview + class-scoped resources + upload progress + resume. **Production storage adapter đã ship 2026-07-01** (`SupabaseProvider` với server-proxy download, `X-Content-Type-Options: nosniff`, content type allowlist). Local provider vẫn là default; Supabase bật qua `RESOURCE_STORAGE_TYPE=supabase` + 3 biến `SUPABASE_*`. **Class scope + multi-file + inline preview đã ship 2026-07-02** (slice-12): `ClassAccessChecker` + `AcademicAccessAdapter`, `academics.Repository.IsStudentEnrolled`, `GET /resources?context_type=class&context_id=<uuid>`, multi-file `POST /resources/{id}/files` (accept `file` + `files[]` + `files`, không auto-archive), `GET /resources/{id}/files`, `GET /resources/{id}/download?file_id=<uuid>&disposition=inline` (inline chỉ cho image/pdf/text, nosniff luôn set), frontend `/app/resources` rewrite grouped list cards + XHR upload (concurrency 3, per-file progress + live region) + inline preview modal + class filter / class scope selector cho managers. Signed URL/CDN/multipart resumable vẫn deferred.
- Non-MCQ question types và manual review workflow (rubric, file submission, teacher feedback). **Foundation đã ship 2026-07-02** (slice-10): `multiple_choice | short_answer | essay`, per-type grading dispatch, PENDING_REVIEW semantics với score=NULL, minimal question bank editor (`/question-banks` với create/list bank, create question, publish version), exam runner + attempt review render per-type. **Manual review workflow đã ship 2026-07-02** (slice-11): `item_grades` table với UNIQUE(organization_id, attempt_item_id) và re-grade allowed, 3 routes (`/assessments/{id}/review-queue`, `/attempts/{id}/review`, `PUT /attempts/{id}/items/{id}/grade` dưới CSRF), `RecomputeAttemptScore` CTE promote GRADED khi tất cả essay/SA đã chấm, audit `attempt.grade` qua `grading.AuditLogger` + `admin.GradingAuditAdapter` (không circular), `/app/grading` queue + detail page. Rubric editor, file submission attachments, bulk-grade, AI scoring vẫn deferred.
- Accessibility full audit (WCAG 2.1 AA, axe-core CI, focus management cho builder/exam/admin, keyboard regression suite).
- Performance, cross-browser (Firefox/WebKit), notifications, installable PWA. **Bundle split đã ship 2026-07-02** (slice-13): 16 route page + 3 dashboard panel `React.lazy` imports qua `SuspenseRoute`, initial chunk giảm từ 511.19 kB / 147.04 kB gz xuống 335.39 kB / 106.56 kB gz, production sourcemap chuyển sang `hidden` (giữ `.map` trên disk nhưng bỏ `//# sourceMappingURL=` trong bundle public). Lazy boundary giữ layouts / auth provider / query provider tĩnh để auth redirect / role guard hoạt động ngay từ first paint. CSS `.loading-fallback` mới. No new deps; behavior không đổi. **Cross-browser matrix đã ship 2026-07-02** (slice-14): `playwright.config.ts` thêm `firefox` + `webkit` projects dưới `PLAYWRIGHT_BROWSERS=1`, default `pnpm e2e:browser` giữ chromium-only, mới `pnpm e2e:browser:all` probe WebKit host deps (libgtk-4 / libgraphene-1.0 / libxslt / libevent-2.1 / libopus / libgstallocators) và gracefully fallback chromium + firefox khi thiếu. Critical-flow locator thêm `.first()` cho robust với accumulated test data. Matrix run 38/40 pass (chromium 20/20, firefox 18/20 — 2 long-flow tests gặp Firefox-headless GPU/SWGL renderer crash trên WSL2 host, infrastructure không phải test logic). `pnpm check` xanh. **Notifications inbox + best-effort events đã ship 2026-07-02** (slice-15): `notifications` table + `internal/features/notifications` package với `Notifier` seam consumed by grading (`attempt.graded` post-recompute), assessments (`assessment.published` post-publish via `RecipientsResolver`), and resources (`resource.published` post-publish, class-scoped only, via `ClassRecipientsResolver`); notifier failures swallow + log để never roll back business tx. 3 endpoints (`GET /me/notifications`, `GET /me/notifications/unread-count`, `POST /me/notifications/{id}/read`) tenant-isolated, mark-read idempotent (`COALESCE(read_at, now())`). Frontend `NotificationBell` trong app shell, 30s `refetchInterval` polling, accessible dropdown (aria-haspopup/dialog, Escape, click-outside), per-item Mark-Read; hidden khi `isRestricted`. OpenAPI + `openapi-schema.d.ts` regenerated. Smoke `assertNotificationFlow` covers list + unread-count + assessment.published fan-out + mark-read + idempotency + cross-user ownership (403/404). `pnpm check` + `pnpm e2e:smoke` + `pnpm e2e:browser` (20/20) all green. Bundle analyzer dependency / virtualized lists / CSS splitting / PWA / GitHub Actions matrix workflow / SSE / WebPush / outbox / notification preferences / retention purge / admin notification center vẫn deferred.
- Generated OpenAPI client vẫn là `openapi-typescript` + `openapi-fetch` cho đến khi có quyết định Huma ở mục B.
- `apiClient` legacy cleanup sau khi tất cả helper đã migrate sang `openapi-fetch`.

## 3. Effort estimate

| Mức | Thời gian tham khảo |
|---|---|
| Demo functional | 8–12 tuần |
| Pilot hẹp | 5–7 tháng full-time |
| MVP ổn định hơn | 7–10 tháng full-time |
| Part-time | 10–16 tháng |

Ước lượng thay đổi theo kinh nghiệm và độ hoàn thiện UI.

## 4. Cost-control rules

- Một managed PostgreSQL nhỏ trước.
- Một app container.
- Object storage pay-as-you-go.
- Không Redis.
- Không Kubernetes.
- Không separate observability stack lúc đầu; dùng provider logs + structured logs.
- Chỉ tách worker khi queue làm ảnh hưởng API.

## 5. Priority order

```text
Data integrity
> Authorization/security
> Exam reliability
> Grade correctness
> Teacher workflow
> Student UX
> Analytics
> Gamification
> AI
```
