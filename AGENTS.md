# Agent Instructions — VTS EDU

Current source of truth for repo operations. The docs in `docs/backend/backend-technical-spec` and `docs/frontend/frontend-technical-spec` are specifications, not runnable commands.

## Repo shape

```text
apps/api          Go API (Render)
apps/web          React 19 + Vite (Vercel)
supabase/         migrations + config.toml
config/           *.env.example (no secrets committed)
docs/             technical specs (read-only reference)
```

## Verified commands

Install:

```bash
pnpm install --frozen-lockfile
```

Frontend:

```bash
pnpm web:dev
pnpm web:typecheck
pnpm web:build
pnpm web:test
```

Backend:

```bash
cd apps/api
go run ./cmd/server

go test ./...
go vet ./...
test -z "$(gofmt -l .)"
```

Root helpers:

```bash
pnpm api:dev    # runs pnpm --filter @vts-edu/api dev
pnpm api:test   # runs pnpm --filter @vts-edu/api test
pnpm check      # web typecheck + build, then Go test/vet/gofmt
```

Local E2E / smoke (Postgres 15 Docker required):

```bash
pnpm e2e:db:start   # start vts-e2e-postgres on port 5434
pnpm e2e:db:migrate # apply supabase/migrations/*.sql
pnpm e2e:smoke      # full API smoke against local API/DB
pnpm e2e:browser    # browser E2E via Playwright Chromium (starts/stops DB, API, and Vite automatically)
pnpm e2e:browser:all # cross-browser matrix (chromium + firefox + webkit) — auto-skips webkit on hosts missing libgtk-4 / libgraphene / libxslt / libevent-2.1 / libopus / libgstallocators
pnpm web:e2e:install:all # one-shot install for chromium + firefox + webkit browser binaries
pnpm e2e:db:stop    # tear down the container
```

Supabase CLI:

```bash
pnpm supabase:version
pnpm supabase:start
pnpm supabase:stop
pnpm supabase:status
pnpm supabase:db:reset
pnpm supabase:db:push:local
pnpm supabase:db:push:remote
pnpm supabase:migration:list
pnpm supabase:link
```

Vercel CLI:

```bash
pnpm vercel:version
pnpm vercel:pull
pnpm vercel:build
pnpm vercel:deploy:preview
```

Docker:

```bash
docker build -t vts-edu-api -f apps/api/Dockerfile apps/api
```

## What is NOT wired yet

- No Makefile.
- No root lint script; use `pnpm check` for bounded validation.
- No `api:migrate`, `api:generate`, or other `api:*` scripts beyond `api:dev`, `api:test`, and `api:sqlc`/`api:types`.
- Huma and River are not installed/wired yet (deferred; see ADR-0010 and ADR-0012).
- Supabase Auth is disabled; auth is backend-owned JWT + rotating opaque refresh cookie.

## Recently implemented

- Auth slice (`internal/features/auth`): login, `/me`, refresh rotation, logout, change-password, JWT access tokens, CSRF double-submit, Argon2id password hashing, persisted multi-role memberships (`membership_roles`), and forced password change flag/claim.
- Attempt runtime slice (`internal/features/attempts`): `GET /attempts/{id}`, `PUT /attempts/{id}/answers/{item_id}`, `POST /attempts/{id}/submit` with tenant-scoped ownership, request-time expiration, real question prompt/choices snapshots, synchronous MCQ grading, and optimistic answer revision.
- Question bank slice (`internal/features/...` schema via migrations): minimal `question_banks`, `questions`, `question_versions`, prompt/choices/answer key snapshots copied into `attempt_items`.
- Teacher assessment list slice (`internal/features/assessments`): `GET /assessments` role-gated to teacher/admin and tenant scoped.
- Admin slice (`internal/features/admin`): `GET/POST /users`, `PUT /users/{user_id}/roles`, `POST /users/{user_id}/reset-password`, `GET/PATCH /organizations/current`, all admin-only and tenant scoped.
- E2E smoke coverage for student attempt grading, teacher role + assessment list, forced password change, admin user/org/audit/bulk/academic management.
- Playwright browser E2E setup (`pnpm e2e:browser`) covering login redirects, teacher builder publish, student attempt/submit, teacher gradebook export, and admin bulk import.
- In-process scheduler groundwork (`internal/platform/scheduler`) with assessment open/close transition job; River deferred.
- Frontend role dashboards: student dashboard (assigned assessments, attempt history, result review), teacher dashboard (classes, assessments, gradebook/export), admin dashboard (org, users, audit logs, CSV import, academic CRUD, bulk ops).
- Frontend pages: login, change-password, assessment builder (duplicate/preview/publish), exam runner.
- TanStack Query server-state layer: query provider, query keys/hooks for attempts/gradebook/assessments/academics, migrated student/teacher/gradebook/review pages.
- Attempt history cursor pagination: backend keyset pagination for `GET /me/attempts`, frontend infinite-query/load-more UI.
- Exam offline resilience MVP: IndexedDB per-attempt/item draft storage, local-save-before-API, pending draft sync on load/online, cleanup after submit, offline banner/status.
- Resources MVP: org-scoped file materials with `LocalProvider` storage seam (server-generated hex keys, path-traversal safe), tenant + role checks, multipart upload (max 10 MiB), publish/archive, bearer-auth download with sanitized `Content-Disposition`. OpenAPI skeleton and `openapi-schema.d.ts` regenerated. Minimal teacher/admin upload UI and student list/download UI at `/app/resources`.
- Production storage adapter: `SupabaseProvider` in `apps/api/internal/platform/storage/supabase.go` (POST/GET/DELETE `/storage/v1/object/{bucket}/{key}` with service-role auth). Opt-in via `RESOURCE_STORAGE_TYPE=supabase`; requires `SUPABASE_URL` + `SUPABASE_SERVICE_ROLE_KEY` + `SUPABASE_STORAGE_BUCKET` (fail-fast on missing). Service role key never in errors/logs/responses. Bucket must be private; download is server-proxy only (`X-Content-Type-Options: nosniff` + content type allowlist via `storage.SanitizeContentType`). Local provider remains the default.
- Non-MCQ foundation + minimal question bank editor: `question_versions.question_type` + `attempt_items.question_type` (CHECK, default `multiple_choice`); 6 new routes under `/question-banks` (list/create bank, list/create question, create/publish version); per-type grading dispatch (essay → `PENDING_REVIEW`, short_answer → exact match against `accepted_answers`, MCQ unchanged); `AttemptResultItem.IsCorrect` is now `*bool` (null for PENDING_REVIEW); new `apps/web/src/pages/question-banks` page for teacher/admin; exam runner and attempt review render per-type UI; PENDING_REVIEW attempts have `score=NULL` in DB; smoke covers mixed MCQ/SA/essay attempt → `PENDING_REVIEW` with `max_score=3.00`. OpenAPI + `openapi-schema.d.ts` regenerated.
- Manual review workflow: new `item_grades` table (`UNIQUE(organization_id, attempt_item_id)`, re-grade allowed); new `apps/api/internal/features/grading` package with 3 routes (`GET /assessments/{id}/review-queue`, `GET /attempts/{id}/review`, `PUT /attempts/{id}/items/{id}/grade` under CSRF); `RecomputeAttemptScore` CTE promotes attempts to `GRADED` only when every essay/short_answer item has a grade; small `AuditLogger` interface in `grading` package with an `admin.GradingAuditAdapter` reusing the existing audit insert (no circular dep); every save writes an `attempt.grade` audit log entry with before/after JSON; `AttemptResultItem` gains nullable `awarded_score` + `feedback` surfaced in the student review view; new `/app/grading` queue + `/app/grading/:attemptId` detail page for teacher/admin with re-grade form; smoke covers the full flow (review-queue → grade essay → grade SA → MCQ 400 → re-grade audit ≥2 → student result GRADED → gradebook updated). OpenAPI + `openapi-schema.d.ts` regenerated.
- Resources UX (P2): class-scoped resources via new `ClassAccessChecker` + `AcademicAccessAdapter` (resources owns the interface, academics owns the implementation); `academics.Repository.IsStudentEnrolled` query for student-side filtering; `GET /resources?context_type=class&context_id=<uuid>` narrows the list; class-scoped create / publish / upload / list / download all 403 for unauthorized callers. `POST /resources/{id}/files` accepts `file` (backward compat), `files[]`, and `files` (multi-part); all entries become ACTIVE (no auto-archive); new `GET /resources/{id}/files` lists ACTIVE files. `GET /resources/{id}/download?file_id=<uuid>&disposition=inline` returns a specific file; `disposition=inline` switches to inline Content-Disposition only for safe preview MIME (image/*, application/pdf, text/{plain,csv,markdown}); `X-Content-Type-Options: nosniff` always set. Frontend `/app/resources` rewritten as grouped list of cards with multi-file XHR upload (concurrency 3, per-file progress bars + live region), inline preview modal (image / pdf / text), and a class filter / class scope selector for managers. No DB migration. OpenAPI + `openapi-schema.d.ts` regenerated.
- Frontend bundle split (P2): all 16 route pages + 3 dashboard panels (`AuditLogsPanel`, `AcademicManagementPanel`, `ClassRosterPanel`) are now `React.lazy` dynamic imports wrapped in a `SuspenseRoute` helper (loading fallback `Đang tải…`); initial chunk dropped from 511.19 kB / 147.04 kB gz to 335.39 kB / 106.56 kB gz (16 page + 3 panel chunks load on demand). `vite.config.ts` switches to `sourcemap: 'hidden'` in production so maps stay on disk for server-side stack trace decoding but the public bundle no longer references them. `.loading-fallback` CSS class added. No new dependencies; no behavior changes. `pnpm web:typecheck` + `pnpm web:build` + `pnpm web:test` (57/57) + `pnpm e2e:browser` (20/20) + `pnpm check` all green.
- Playwright cross-browser matrix (P2): `apps/web/playwright.config.ts` adds `firefox` + `webkit` projects under the `PLAYWRIGHT_BROWSERS=1` env flag; default `pnpm e2e:browser` stays Chromium-only for speed. New `pnpm e2e:browser:all` script (also `pnpm web:e2e:install:all`) probes the WebKit host deps and gracefully falls back to chromium + firefox when libgtk-4 / libgraphene-1.0 / libxslt / libevent-2.1 / libopus / libgstallocators are missing (the system libraries WebKit needs on Linux). `pnpm e2e:browser` (chromium) still 20/20 green. Matrix run: 38/40 pass (chromium 20/20; firefox 18/20 with 2 long-flow tests hitting Firefox-headless GPU/SWGL renderer crashes on the WSL2 host — infrastructure, not test logic). One small critical-flow locator made robust to accumulated test data via `.first()`. `pnpm check` green.
- Notification inbox + best-effort events (slice-15): new `notifications` table + `apps/api/internal/features/notifications` package with `Notifier` seam consumed by grading (`attempt.graded` post-recompute), assessments (`assessment.published` post-publish via `RecipientsResolver`), and resources (`resource.published` post-publish, class-scoped only, via `ClassRecipientsResolver`); notifier failures are swallowed + logged so they never roll back the business tx. 3 endpoints (`GET /me/notifications`, `GET /me/notifications/unread-count`, `POST /me/notifications/{id}/read`) all tenant-isolated and tenant-defaulted to caller; mark-read is idempotent (`COALESCE(read_at, now())`). Frontend `NotificationBell` in app shell with 30s `refetchInterval` polling, accessible dropdown (aria-haspopup/dialog, Escape, click-outside), per-item Mark-Read; hidden when `isRestricted`. `openapi-schema.d.ts` regenerated. Smoke `assertNotificationFlow` covers list + unread-count + assessment.published fan-out + mark-read + idempotency + cross-user ownership (403/404). `pnpm check` + `pnpm e2e:smoke` + `pnpm e2e:browser` (20/20) all green.
- PWA Level 0 installability (manifest only): `apps/web/public/manifest.json` (name/short_name "VTS EDU", `start_url`/`scope`/`id` `/app`, `display: standalone`, theme `#1d4ed8`, vi) + 2 SVG icons (`icon.svg` purpose `any`, `icon-maskable.svg` purpose `maskable`). `apps/web/index.html` gains `<link rel="manifest">`, `<link rel="apple-touch-icon">`, `<meta name="theme-color">`, `application-name`, `apple-mobile-web-app-capable=yes`, `mobile-web-app-capable=yes`; legacy `/vite.svg` favicon replaced with `/icons/icon.svg`. **No service worker**: zero hits for `serviceWorker|workbox|pwa-register` across `apps/web/{src,index.html,vite.config.ts,public}`. No `vite-plugin-pwa`, no Workbox, no API/asset caching, no background sync, no push. `apps/web/package.json` dependency surface unchanged. `pnpm web:typecheck` + `pnpm web:build` (bundle sizes unchanged — 360.50 kB / 114.37 kB gz for `index-*.js`) + `pnpm e2e:browser` (chromium 20/20) + `pnpm check` all green. Exam offline draft resilience stays at the app-layer IndexedDB level (already shipped). SW registration, real PNG icons (192/512), splash/startup images, install-prompt UI, push, background sync deferred.
- Exam heartbeat / deadline hardening (slice-18): `AnswerSaved` returns `server_time` + `expires_at` so the client can recalibrate its countdown without a separate round trip; existing `GET /attempts/{id}` is reused as the 60s heartbeat (no new endpoint) to refresh offset + status + expires_at while the attempt is `IN_PROGRESS` (failures swallowed silently; no auto-submit; inputs/submit still disabled at `timeLeft=0`). Frontend maintains a `serverTimeOffsetRef` calibrated from the initial `getAttempt` and every `saveAnswer` response, and the countdown reads `Date.now() + serverTimeOffset` so a drifted client clock or stale tab cannot extend or steal time. Deadline warning UX: `≤5 min` warning banner (`role="status"`, amber) + `≤1 min` critical banner (`role="alert"`, red, gentle pulse); threshold transitions only (no every-tick `aria-live`); timer chip recolors. OpenAPI `SaveAnswerResponse` extended + `openapi-schema.d.ts` regenerated. Smoke `saveAnswerForAttempt` asserts `server_time` + `expires_at` on every save. `pnpm api:types` + `pnpm check` + `pnpm e2e:smoke` + `pnpm e2e:browser` (23/23) all green.

> **Note on Koyeb artifacts:** files such as `apps/api/koyeb.yaml` are legacy. Render is the current backend deployment target.

## Env / deploy gotchas

- Copy `config/*.env.example` to platform dashboards or local `.env` files. Never commit secrets.
- Vercel must set `VITE_API_BASE_URL` to the absolute Render origin, e.g. `https://<api>.onrender.com/api/v1`.
- Frontend `apiClient` joins `runtimeConfig.apiBaseUrl` with path `/api/v1/...`. Before changing any API URL, verify you are not producing `.../api/v1/api/v1/...`.
- Backend CORS reads `FRONTEND_ORIGINS` comma-separated; no wildcard with credentials.
- Refresh cookie is `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth` for the cross-origin demo.
- Unsafe cookie-backed endpoints require `X-CSRF-Token` header (double-submit with `vts_csrf` cookie).

## Implementation boundaries

- Keep Go backend under `apps/api` with feature-first modular structure.
- Keep frontend under `apps/web/src`.
- Add migrations to `supabase/migrations/` with sequential `000XXX_` names.
- Do not add Supabase Auth; keep backend-owned auth.
- Do not commit `pnpm-lock.yaml` changes unless dependency metadata actually changed.

## Hard safety rules

- No secrets, tokens, or passwords in any committed file.
- No `sudo` or global installs in scripts/docs.
- No production deploys without explicit user confirmation.
- Do not run `supabase db push:remote`, `vercel deploy`, or Git push unless asked.
- Preserve existing technical specs; do not rewrite them as implementation code.
