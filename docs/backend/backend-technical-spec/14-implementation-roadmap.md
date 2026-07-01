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
- **Stage 2 — Huma migration (deferred)**: Add Huma operation definitions behind existing handlers or incrementally replace handler wiring while preserving response envelopes. OpenAPI generation becomes automatic. As of the latest batch (builder polish / student / gradebook / bulk / hardening), the manual skeleton covers **58 paths** (up from ~44), very close to the 60-path threshold. `openapi-typescript` + `openapi-fetch` plus the `generated-code-check` CI job continue to keep manual maintenance manageable. Huma remains deferred because migration risk/cost still outweighs manual maintenance, especially for auth cookie/CSRF/refresh-sensitive handlers and middleware ordering. Revisit trigger: spec drift incidents ≥ 2/month, paths ≥ 60, need runtime contract validation, or a dedicated refactor sprint with ≥ 80% handler test coverage.
- **Stage 3 — Client generation**: Generate frontend API client/types from the Huma/OpenAPI contract once it stabilizes.
- **Breached-password provider (deferred)**: HIBP/external corpus integration deferred pending a privacy/egress ADR; password history + lockout + blocklist implemented as interim controls.

## 2.6 Background jobs / scheduler plan

- **Current**: In-process scheduler in `apps/api/internal/platform/scheduler` runs lightweight `Job` implementations on a fixed interval. First job: `assessment-transition` opens/closes assessments based on `opens_at`/`closes_at`. Controlled by `SCHEDULER_ENABLED` and `SCHEDULER_INTERVAL_SECONDS` (default disabled; enable on Render).
- **River (deferred)**: River/pgvector-style durable queue is not adopted yet. The in-process scheduler is sufficient for scheduled assessment transitions and avoids extra migrations, worker processes, and operational complexity.
- **Large CSV import**: Remains synchronous with a 100-row cap until a durable queue is justified.
- **Async grading**: MCQ grading stays synchronous. Non-MCQ / manual-review async grading is deferred until those question types are implemented.
- **Triggers for River adoption**: need durability/retry (e.g., large CSV, async grading), multi-instance scale-out requiring duplicate-job prevention, cron-like scheduling, or an approved infrastructure sprint covering migration, worker process, monitoring, and dead-letter handling.

## 2.7 Current next backlog (not started)

- Resources/files signed download + processing job (multipart upload + local storage seam are in place).
- Non-MCQ question types and manual review workflow.
- Accessibility audit.
- Generated OpenAPI client (Huma deferred; hand-maintained skeleton still sufficient).

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
