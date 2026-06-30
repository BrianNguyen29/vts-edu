# Implementation Audit — VTS EDU

Repo-wide implementation tracking. Append-only; do not delete historical entries.

## Implementation plan

| Phase | Theme | Owner | Status |
|---|---|---|---|
| S0 | Backend foundation (config, DB pool, TxManager, CORS, readyz) | fixer | Foundation complete; sqlc/Huma staged |
| S1 | Auth + users (JWT, refresh cookie, sessions, CSRF, /me, persisted roles, forced password change) | fixer | Implemented |
| S2 | Attempt runtime + question snapshots + grading (get/save/submit, ownership, request-time expiry, MCQ grading) | fixer | Implemented |
| S2.5 | Teacher assessment list | fixer | Implemented |
| S3 | Admin user/org management | fixer | Implemented |
| S4 | Academics + full question bank + assessment builder | fixer/designer | Not started |
| S5 | Resources, assignments, gradebook | fixer/designer | Not started |

## S0 Backend foundation

### Done

- [x] Add `internal/platform/db` package with `pgxpool` wrapper and `TxManager`.
- [x] Wire DB pool into `cmd/server` startup; fail fast on ping failure unless `DB_SKIP=true`.
- [x] `/readyz` checks DB readiness and returns `503` with `db: unavailable` when DB is down; returns `db: skipped` when `DB_SKIP=true`.
- [x] Improve `LoadConfig` validation with clear list of missing env vars.
- [x] Fix CORS middleware: disallowed origins no longer receive a fallback `Access-Control-Allow-Origin`.
- [x] Preserve CSRF behavior (`GET /api/v1/auth/csrf-token`, validation on cookie-backed unsafe endpoints).
- [x] Add `DB_SKIP` option for local dev without Postgres (documented below).

### Remaining S0 (staged)

- [ ] Add `sqlc` baseline and first generated queries for identity/attempt tables. Existing `Repository` interfaces are the migration seam; do not rewrite runtime code.
- [ ] Add Huma/OpenAPI skeleton endpoint definitions. The hand-maintained OpenAPI skeleton in `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml` now covers the current API surface; Huma adoption is deferred until a staged migration is planned.

### Decisions / notes

- `DATABASE_URL`, `JWT_SIGNING_KEY`, and `REFRESH_TOKEN_KEY` are required. Set `DB_SKIP=true` to run the server locally without Postgres; `/readyz` will then report `db: skipped` with HTTP 200.
- CORS allowlist is exact-match only. No wildcard with credentials.
- `pgx/v5` added as a direct dependency; Go toolchain bumped to 1.25.0 by `go mod tidy`.

## 2026-06-29 — Frontend UX scaffold (frontend-ux-plan-001)

### Done

- [x] Normalize API URL joining (`joinApiUrl`) to prevent `/api/v1/api/v1` duplication.
- [x] Update `apiClient` and CSRF token fetch to use the new join helper.
- [x] Attach bearer token from in-memory auth store to API requests.
- [x] Add `react-router-dom` dependency and scaffold React Router routes.
- [x] Add in-memory auth session store/interface (`shared/auth/auth-session-store.ts`).
- [x] Add `AuthProvider` with bootstrap, login, logout, and serialized refresh single-flight.
- [x] Add login page (`/login`) with organization code, username, password.
- [x] Add app shell layout and protected route guard (`/app`).
- [x] Preserve health/CSRF demo as `/diagnostics` dev screen.
- [x] Add exam runner placeholder route/page (`/exam/attempts/:attemptId`).
- [x] Update `docs/implementation-audit.md` with frontend todo statuses.

### Pending / next steps

- [x] Backend S1 auth endpoints (`/auth/login`, `/auth/refresh`, `/auth/logout`, `/me`, `/auth/csrf-token`, `/auth/change-password`) so the login flow can return real tokens.
- [x] Connect login response actor parsing; currently fetches `/me` after login.
- [x] Backend forced password change (`/auth/change-password`) and `must_change_password` claim are implemented; frontend guard pending.
- [ ] Add role/workspace redirects (`/app/student`, `/app/teacher`, `/app/admin`).
- [ ] Add real 403/404/maintenance error pages with request ID display.
- [ ] Implement TanStack Query integration and generated OpenAPI client once backend OpenAPI is available.
- [x] Implement exam runtime MVP: get/save/submit endpoints, dashboard demo link, fixed demo attempt UUID.
- [ ] Implement advanced exam runtime: IndexedDB drafts, answer save queue, server-timer, offline resilience.
- [ ] Add unit/component tests for `joinApiUrl`, auth store, route guards, and login form.
- [ ] Accessibility audit (focus management, ARIA labels, reduced motion) and responsive smoke tests.

### Decisions / notes

- Default dev `apiBaseUrl` set to `/api/v1` so versioned paths are consistent with production (`https://<api>.onrender.com/api/v1`).
- `joinApiUrl` accepts both relative paths (`/healthz`) and legacy absolute paths (`/api/v1/healthz`) and deduplicates the version prefix when the base already ends with `/api/v1`.
- Auth bootstrap treats backend/network failures as anonymous during MVP so the login screen remains usable while backend is incomplete. This will be tightened to `degraded` state once the API is stable.
- Cross-tab refresh serialization uses `navigator.locks` when available with an in-memory fallback; BroadcastChannel logout events are deferred until the backend contract is finalized.
- No backend files, secrets, deploy commands, or git operations were touched in this changeset.

## 2026-06-30 — Docs/DX batch (docs-dx-batch-001)

### Done

- [x] CI Go version fixed to `1.25.0` in `.github/workflows/ci.yml`.
- [x] Removed legacy Koyeb artifacts: `apps/api/koyeb.yaml`, `config/koyeb.env.example`.
- [x] Updated `config/README.md` and `README.md` to reflect Render as the current backend target and River as planned/not wired.
- [x] Added `docs/e2e-local-run.md` with local auth → attempt smoke instructions.
- [x] Linked the local E2E guide from `docs/deployment-cli.md`.

### Deferred / not in scope

- Huma/sqlc/River wiring.
- Role/workspace redirects (`/app/student`, `/app/teacher`, `/app/admin`).
- Forced password change (`/change-password`) and restricted session guard.
- Playwright/Cypress E2E automation.
- Root lint/test scripts and generated OpenAPI/client.

## 2026-06-30 — DX hardening (dx-hardening-001)

### Done

- [x] Add root `pnpm check` script (web typecheck/build + Go test/vet/gofmt).
- [x] Add fallback E2E scripts using a direct PostgreSQL 15 Docker container:
  - `pnpm e2e:db:start`, `pnpm e2e:db:migrate`, `pnpm e2e:db:stop`, `pnpm e2e:smoke`.
- [x] Remove misleading `lint` script from `apps/web/package.json` (ESLint is not installed).
- [x] Update `.github/workflows/ci.yml` with web typecheck/build and Postgres service migration validation.
- [x] Update `docs/e2e-local-run.md` with the official fallback Postgres path and Supabase CLI limitations.

### Deferred / not in scope

- Feature next: timer, submit confirmation, grading, role-based routes.
- Playwright/Cypress E2E automation.
- Root lint script and ESLint dependency setup.
- sqlc/Huma/River wiring.

## 2026-06-30 — Product slices S1–S3 backend (product-slices-backend-001)

### Done

- [x] Persisted multi-role memberships (`membership_roles`) replacing hardcoded `student` role.
- [x] Forced password change backend: `must_change_password` user flag, `pwd_change_required` JWT claim, and `POST /api/v1/auth/change-password`.
- [x] Minimal question bank schema (`question_banks`, `questions`, `question_versions`) and snapshot of prompt/choices/answer key into `attempt_items`.
- [x] Synchronous MCQ grading on attempt submit returning `score`, `max_score`, `grading_status`.
- [x] Teacher assessment list `GET /api/v1/assessments` role-gated to teacher/admin and tenant scoped.
- [x] Admin user/org management: `GET/POST /api/v1/users`, `PUT /api/v1/users/{id}/roles`, `POST /api/v1/users/{id}/reset-password`, `GET/PATCH /api/v1/organizations/current`.
- [x] E2E smoke extended to cover role seeds, forced password change, teacher assessment list, and admin user/org flow.
- [x] Backend OpenAPI skeleton updated to the current API surface.
- [x] ADR-0010 documenting staged Huma/sqlc groundwork behind existing Repository interfaces.

### Deferred / not in scope

- sqlc/Huma runtime wiring and generated client.
- Frontend role redirects, change-password guard, and generated API client.
- Pagination/search, audit logs, password policy.
- Full assessment builder, academics, resources, gradebook.

### Decisions / notes

- Existing `Repository` interfaces in each feature package are the stable seam for future sqlc migration; no runtime code should be rewritten for sqlc alone.
- OpenAPI skeleton is hand-maintained until Huma is adopted; it covers current endpoints sufficiently for frontend client generation planning.
- All admin endpoints require `admin` role; teacher endpoints require `teacher` or `admin`.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-29 | S0 backend foundation | `apps/api/internal/platform/db/db.go`, `apps/api/internal/platform/db/db_test.go`, `apps/api/internal/app/config.go`, `apps/api/cmd/server/main.go`, `apps/api/go.mod`, `apps/api/go.sum` | Fresh Go evidence (2026-06-29): `go test ./...`, `go vet ./...`, `gofmt -l .` all passed. Output: `ok github.com/.../csrf`, `ok github.com/.../db`, `PASS`. |
| 2026-06-30 | S1 auth + S2 attempt runtime + frontend demo wiring | `apps/api/internal/features/auth/*`, `apps/api/internal/features/attempts/*`, `apps/api/cmd/server/main.go`, `apps/web/src/shared/config/demo-attempt.ts`, `apps/web/src/pages/dashboard/dashboard-page.tsx`, `supabase/migrations/000004_*`, `supabase/migrations/000005_*`, `supabase/migrations/000006_*` | Go tests/vet/gofmt pass; `pnpm web:typecheck` and `pnpm web:build` pass; migrations validated on temporary Postgres container. |
| 2026-06-30 | Docs/DX batch | `.github/workflows/ci.yml`, `apps/api/koyeb.yaml` (deleted), `config/koyeb.env.example` (deleted), `config/README.md`, `README.md`, `docs/e2e-local-run.md`, `docs/deployment-cli.md`, `docs/implementation-audit.md` | Go checks, `pnpm web:typecheck`, `pnpm web:build` pass. |
| 2026-06-30 | DX hardening | `package.json`, `apps/web/package.json`, `.github/workflows/ci.yml`, `scripts/e2e_*.sh`, `scripts/e2e_smoke_api.mjs`, `docs/e2e-local-run.md`, `docs/implementation-audit.md`, `AGENTS.md` | `pnpm check` passes; `pnpm e2e:smoke` passes against local Postgres container; CI includes migration validation. |
| 2026-06-30 | Product slices S1–S3 backend | `apps/api/internal/features/auth/*`, `apps/api/internal/features/attempts/*`, `apps/api/internal/features/assessments/*`, `apps/api/internal/features/admin/*`, `apps/api/cmd/server/main.go`, `supabase/migrations/000008_*` to `000012_*`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/e2e-local-run.md`, `docs/implementation-audit.md`, `README.md`, `AGENTS.md` | `pnpm check` passes; `pnpm e2e:smoke` passes covering roles, change password, attempt grading, assessment list, and admin user/org flow. |
