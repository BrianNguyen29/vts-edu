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
| S4 | Academics + full question bank + assessment builder | fixer/designer | Core implemented — academics CRUD/bulk, assessment builder (duplicate/preview/publish), question bank minimal |
| S5 | Resources, assignments, gradebook | fixer/designer | Partial — gradebook backend + frontend implemented; resources MVP (org-scoped file materials) shipped, assignments not started |

## S0 Backend foundation

### Done

- [x] Add `internal/platform/db` package with `pgxpool` wrapper and `TxManager`.
- [x] Wire DB pool into `cmd/server` startup; fail fast on ping failure unless `DB_SKIP=true`.
- [x] `/readyz` checks DB readiness and returns `503` with `db: unavailable` when DB is down; returns `db: skipped` when `DB_SKIP=true`.
- [x] Improve `LoadConfig` validation with clear list of missing env vars.
- [x] Fix CORS middleware: disallowed origins no longer receive a fallback `Access-Control-Allow-Origin`.
- [x] Preserve CSRF behavior (`GET /api/v1/auth/csrf-token`, validation on cookie-backed unsafe endpoints).
- [x] Add `DB_SKIP` option for local dev without Postgres (documented below).
- [x] Add `sqlc` baseline and generate queries for `assessments`, `admin`, `auth`, and `attempts` tables. Existing `Repository` interfaces are the migration seam; no runtime code rewrite.

### Remaining S0 (staged)

- [ ] Add Huma runtime wiring and automatic OpenAPI generation. The hand-maintained OpenAPI skeleton in `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml` now covers the current API surface; Huma adoption is deferred until the API contract stabilizes.

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
- [x] Backend forced password change (`/auth/change-password`) and `must_change_password` claim are implemented; frontend guard done.
- [x] Add role/workspace redirects (`/app/student`, `/app/teacher`, `/app/admin`).
- [x] Add real 403/404/maintenance error pages with request ID display.
- [x] Implement TanStack Query integration for core server-state pages.
- [ ] Implement generated OpenAPI client once backend OpenAPI is available.
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

## 2026-06-30 — Backend hardening (hardening-backend-policy-pagination-audit)

### Done

- [x] Stronger password policy (`auth.ValidatePasswordStrength`): min 8 chars, uppercase, lowercase, digit, short blocklist; enforced on `/auth/change-password`, admin create user, and admin reset password.
- [x] Same-as-current password rejection on `/auth/change-password`.
- [x] Backward-compatible pagination/search for `GET /api/v1/users` and `GET /api/v1/assessments` (`q`, `limit`, `offset`); no-param responses keep the original `{data:[...]}` shape.
- [x] Audit log writes for admin actions: `user.create`, `user.update_roles`, `user.reset_password`, `organization.update`.
- [x] Unit tests for password policy, weak-password rejections, list pagination/search, and audit log calls.
- [x] E2E smoke extended to assert weak-password rejections, search/limit behavior, and audit log rows via direct DB query.
- [x] OpenAPI skeleton updated with query params, page metadata, and password-policy error responses.

### Deferred / not in scope

- Audit log read endpoint.
- sqlc/Huma runtime wiring and generated client.
- Frontend role redirects, change-password guard, and generated API client.
- Full assessment builder, academics, resources, gradebook.

### Decisions / notes

- `Repository` interfaces remain the stable seam; no runtime handler/service rewrite.
- Pagination metadata is additive (`page` object) and only present when `limit` is supplied.
- Audit logs capture actor, action, resource, before/after where feasible, and metadata; sensitive values like password hashes are never logged.

## 2026-06-30 — Generated types + sqlc assessments groundwork (hardening-openapi-sqlc-groundwork)

### Done

- [x] Added `openapi-typescript` dev dependency and root scripts `pnpm api:types` / `pnpm api:sqlc`.
- [x] Generated TypeScript types from OpenAPI skeleton to `apps/web/src/shared/api/openapi-schema.d.ts`.
- [x] Used generated types type-only in `apps/web/src/shared/api/assessments.ts` (response and list item shapes); existing `apiClient` runtime unchanged.
- [x] Added `apps/api/sqlc.yaml` configured for `pgx/v5` and generated sqlc code under `apps/api/internal/features/assessments/sqlc/`.
- [x] Migrated `assessments.Repository` implementation to a sqlc wrapper (`apps/api/internal/features/assessments/repository.go`) while preserving the existing interface and service/handler contracts.
- [x] Updated ADR-0010 and backend roadmap to record generated types, sqlc `assessments` migration, and deferred Huma.

### Deferred / not in scope

- sqlc migration for `auth`, `attempts`, `admin` repositories.
- Huma runtime migration and automatic OpenAPI generation.
- Runtime OpenAPI client (`openapi-fetch`) replacing `apiClient`.

### Decisions / notes

- sqlc wrapper maps `pgtype.UUID` to/from string; the public repository interface keeps `string` IDs.
- `pnpm api:sqlc` uses `go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest` so no global install is required.
- Generated sqlc files are committed because the project does not yet have CI generation.

## 2026-06-30 — Audit log reader backend (next-slice1-audit-reader-backend)

### Done

- [x] Added `GET /api/v1/audit-logs` admin-only endpoint with tenant scoping.
- [x] Supported filters `action`, `actor_user_id`, `from`, `to` and pagination `limit`/`offset`.
- [x] Reused existing response envelope with optional `page` metadata when `limit` is supplied.
- [x] Added service/repository/handler tests for admin gate and filter behavior.
- [x] Updated OpenAPI skeleton with `/audit-logs` path and `AuditLog` schema.
- [x] Updated E2E smoke to verify audit logs via HTTP endpoint and action filter, replacing the direct DB query.

### Deferred / not in scope

- Frontend audit log UI/dashboard.
- sqlc migration for admin repository.
- Cursor pagination.

### Decisions / notes

- Timestamps are validated as RFC3339 in the handler and returned as RFC3339 strings.
- Audit log JSONB columns (`before_json`, `after_json`, `metadata_json`) are exposed as optional JSON objects without leaking sensitive values.

## 2026-06-30 — Admin repository sqlc migration (next-slice2-admin-sqlc)

### Done

- [x] Added `apps/api/internal/features/admin/queries.sql` covering all admin repository operations including list filters, writes inside transactions, and audit logs.
- [x] Updated `apps/api/sqlc.yaml` to generate `adminsqlc` package under `apps/api/internal/features/admin/sqlc/`.
- [x] Replaced manual admin `repository.go` implementation with a sqlc wrapper that preserves the existing `admin.Repository` interface.
- [x] `NewRepository` now returns the sqlc-backed implementation; service/handler contracts remain unchanged.
- [x] Transactional methods use `queries.WithTx(tx)` to run generated queries inside the existing transaction boundary.
- [x] Preserved dynamic list behavior via conditional SQL expressions in generated queries.

### Deferred / not in scope

- sqlc migration for `auth` and `attempts` repositories.
- Huma runtime migration.
- Frontend changes.

### Decisions / notes

- `pgtype.UUID` and `pgtype.Text` conversions are isolated in the wrapper; the public interface continues to use plain strings.
- `array_agg` roles are converted from `interface{}` to `[]string` in the wrapper.
- Manual repository code was removed because the sqlc replacement is complete and tests/smoke pass.

## 2026-06-30 — Auth repository sqlc migration (next-slice2-auth-sqlc)

### Done

- [x] Added `apps/api/internal/features/auth/queries.sql` covering login lookup, refresh session lifecycle, actor lookup, role lookup, password update, and session revocation.
- [x] Updated `apps/api/sqlc.yaml` to generate `authsqlc` package under `apps/api/internal/features/auth/sqlc/`.
- [x] Replaced manual auth `repository.go` implementation with a sqlc wrapper preserving the existing `auth.Repository` interface.
- [x] `NewRepository` now returns the sqlc-backed implementation; service/handler contracts unchanged.
- [x] Transactional methods use `queries.WithTx(tx)`.

### Deferred / not in scope

- sqlc migration for `attempts` repository.
- Huma runtime migration.
- Frontend changes.

### Decisions / notes

- `array_agg` roles are converted from `interface{}` to `[]string` in the wrapper, matching the previous repository behavior.
- Nullable `pgtype.Text`/`pgtype.Timestamptz` fields are mapped to pointer types in the wrapper.
- Password update is split into two generated queries (`BumpUserAuthVersion` and `UpdateLoginPassword`) executed inside the same transaction.

## 2026-06-30 — Attempts repository sqlc migration (next-slice2-attempts-sqlc)

### Done

- [x] Added `apps/api/internal/features/attempts/queries.sql` covering attempt/attempt_item reads, answer save revision, transactional submit/grade, and list operations.
- [x] Updated `apps/api/sqlc.yaml` to generate `attemptssqlc` package under `apps/api/internal/features/attempts/sqlc/`.
- [x] Replaced manual attempts `repository.go` implementation with a sqlc wrapper preserving the existing `attempts.Repository` interface.
- [x] `NewRepository` now returns the sqlc-backed implementation; service/handler contracts unchanged.
- [x] Transactional methods use `queries.WithTx(tx)` inside the existing transaction boundary.
- [x] Mapped nullable `pgtype.Numeric` score/max_score to pointer strings and `pgtype.Timestamptz`/Text to pointer types in the wrapper.

### Deferred / not in scope

- Huma runtime migration.
- Frontend changes.

### Decisions / notes

- `pgtype.Numeric` is converted to `*string` using `Float64Value()` and `fmt.Sprintf("%.2f", ...)`, preserving the existing 2-decimal string contract.
- `array_agg` item IDs and answer choice arrays are converted from `[]interface{}` to `[]string` in the wrapper.
- Manual repository code was removed because the sqlc replacement is complete and tests/smoke pass.

## 2026-06-30 — OpenAPI types CI expansion (next-slice3-openapi-types-ci)

### Done

- [x] Updated `apps/web/src/shared/api/admin.ts` to derive `Organization`, `User`, `CreateUserRequest`, `UpdateRolesRequest`, `ResetPasswordRequest`, `UpdateOrganizationRequest`, and `AuditLog` from generated `components['schemas']` type-only.
- [x] Updated `apps/web/src/shared/api/attempts.ts` to derive `AttemptSnapshot`, `AttemptItem`, `QuestionPrompt`, `QuestionChoice`, `AnswerSnapshot`, `AnswerSaved`, `AttemptSubmitted`, and `PageInfo` from generated `components['schemas']` type-only.
- [x] Improved `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml` free-form object schemas (`prompt`, `choices`, `answer_payload`, `before`, `after`, `metadata`) with `additionalProperties: true` so generated types are usable at runtime boundaries.
- [x] Added `generated-code-check` job to `.github/workflows/ci.yml` that installs Node/Go, runs `pnpm api:types` and `pnpm api:sqlc`, then `git diff --exit-code` on `apps/web/src/shared/api/openapi-schema.d.ts` and `apps/api/internal/features/*/sqlc/` to reject stale generated code.
- [x] Adjusted `admin-dashboard-page.tsx` form state and casts to align with generated enum role arrays.
- [x] Adjusted `exam-page.tsx` submitted-at fallback to handle nullable `submitted_at` from generated schema.

### Deferred / not in scope

- Runtime OpenAPI client (`openapi-fetch`) replacing `apiClient`.
- Huma runtime migration.

### Decisions / notes

- `apiClient` runtime remains unchanged; generated types are consumed type-only.
- Generated `User`/`CreateUserRequest`/`UpdateRolesRequest` role arrays are enum unions (`("student" | "teacher" | "admin")[]`); UI form state stays `string[]` and casts at the API boundary.
- CI stale check only diffs generated artifacts; source files (queries.sql, openapi-skeleton.yaml) are not checked by `git diff` because the generator output is the signal.

## 2026-06-30 — Cursor pagination (next-slice4-cursor-pagination)

### Done

- [x] Added `apps/api/internal/platform/pagination/cursor.go` with base64url JSON cursor encode/decode and `ErrInvalidCursor`.
- [x] Extended `admin.ListOptions` and `admin.AuditLogListOptions` with `Cursor` and `Count`; updated `admin.PageInfo` to include `next_cursor`, `has_more`, and `total_count`.
- [x] Extended `assessments.ListOptions` and `assessments.PageInfo` with the same cursor/count fields.
- [x] Added cursor and count support to sqlc queries for `ListUsers`, `ListAuditLogs`, and `ListPublishedByOrganization`; added `CountUsers`, `CountAuditLogs`, and `CountPublishedByOrganization` queries.
- [x] Updated admin/assessments service and handler layers to build page metadata (fetch `limit+1`, trim to `limit`, encode next cursor) and optional `total_count`.
- [x] Preserved backward-compatible `offset` behavior and no-param full-list responses.
- [x] Updated OpenAPI skeleton with `ListCursor`, `ListCount` parameters and richer `PageInfo` schema; regenerated TypeScript types.
- [x] Updated frontend `admin.ts`, `assessments.ts`, and `attempts.ts` query builders to pass `cursor` and `count`.
- [x] Added load-more UI to admin users list, teacher assessments list, and audit logs panel; search/filter resets the cursor.
- [x] Extended E2E smoke to verify cursor pagination and `total_count` for users and audit logs, and `has_more: false` for assessments.

### Deferred / not in scope

- Huma runtime migration.
- Removal of `offset` (kept for backward compatibility).
- Cursor support for attempt history (not a list endpoint yet).

### Decisions / notes

- Cursor encoding uses JSON `{k, i}` base64url raw encoding for stability and readability during debugging.
- User cursor key is `username_normalized` (ascending); audit/assessment cursor key is `created_at` RFC3339 (descending).
- Services request `limit+1` rows and use the extra row only to determine `has_more`; repositories return the raw slice including the extra row when present.
- `count=true` runs an additional `COUNT(*)` with the same filters but ignoring cursor; skipped unless requested.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-29 | S0 backend foundation | `apps/api/internal/platform/db/db.go`, `apps/api/internal/platform/db/db_test.go`, `apps/api/internal/app/config.go`, `apps/api/cmd/server/main.go`, `apps/api/go.mod`, `apps/api/go.sum` | Fresh Go evidence (2026-06-29): `go test ./...`, `go vet ./...`, `gofmt -l .` all passed. Output: `ok github.com/.../csrf`, `ok github.com/.../db`, `PASS`. |
| 2026-06-30 | S1 auth + S2 attempt runtime + frontend demo wiring | `apps/api/internal/features/auth/*`, `apps/api/internal/features/attempts/*`, `apps/api/cmd/server/main.go`, `apps/web/src/shared/config/demo-attempt.ts`, `apps/web/src/pages/dashboard/dashboard-page.tsx`, `supabase/migrations/000004_*`, `supabase/migrations/000005_*`, `supabase/migrations/000006_*` | Go tests/vet/gofmt pass; `pnpm web:typecheck` and `pnpm web:build` pass; migrations validated on temporary Postgres container. |
| 2026-06-30 | Docs/DX batch | `.github/workflows/ci.yml`, `apps/api/koyeb.yaml` (deleted), `config/koyeb.env.example` (deleted), `config/README.md`, `README.md`, `docs/e2e-local-run.md`, `docs/deployment-cli.md`, `docs/implementation-audit.md` | Go checks, `pnpm web:typecheck`, `pnpm web:build` pass. |
| 2026-06-30 | DX hardening | `package.json`, `apps/web/package.json`, `.github/workflows/ci.yml`, `scripts/e2e_*.sh`, `scripts/e2e_smoke_api.mjs`, `docs/e2e-local-run.md`, `docs/implementation-audit.md`, `AGENTS.md` | `pnpm check` passes; `pnpm e2e:smoke` passes against local Postgres container; CI includes migration validation. |
| 2026-06-30 | Product slices S1–S3 backend | `apps/api/internal/features/auth/*`, `apps/api/internal/features/attempts/*`, `apps/api/internal/features/assessments/*`, `apps/api/internal/features/admin/*`, `apps/api/cmd/server/main.go`, `supabase/migrations/000008_*` to `000012_*`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/e2e-local-run.md`, `docs/implementation-audit.md`, `README.md`, `AGENTS.md` | `pnpm check` passes; `pnpm e2e:smoke` passes covering roles, change password, attempt grading, assessment list, and admin user/org flow. |
| 2026-06-30 | Backend hardening (policy, pagination, audit) | `apps/api/internal/features/auth/password_policy.go`, `apps/api/internal/features/auth/*`, `apps/api/internal/features/admin/*`, `apps/api/internal/features/assessments/*`, `scripts/e2e_smoke_api.mjs`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | `pnpm check` passes; `pnpm e2e:smoke` passes with weak-password rejections, search/limit assertions, and audit log verification. |
| 2026-06-30 | Audit log reader backend | `apps/api/internal/features/admin/*`, `apps/api/cmd/server/main.go`, `scripts/e2e_smoke_api.mjs`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `docs/implementation-audit.md` | `pnpm check` passes; `pnpm e2e:smoke` passes với `GET /audit-logs`, action filter, và role gate. |
| 2026-06-30 | Generated types + sqlc assessments groundwork | `package.json`, `apps/web/src/shared/api/openapi-schema.d.ts`, `apps/web/src/shared/api/assessments.ts`, `apps/api/sqlc.yaml`, `apps/api/internal/features/assessments/queries.sql`, `apps/api/internal/features/assessments/sqlc/*`, `apps/api/internal/features/assessments/repository.go`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | `pnpm check` passes; `pnpm e2e:smoke` passes với assessment list/search sử dụng sqlc wrapper. |
| 2026-06-30 | Admin repository sqlc migration | `apps/api/sqlc.yaml`, `apps/api/internal/features/admin/queries.sql`, `apps/api/internal/features/admin/repository.go`, `apps/api/internal/features/admin/sqlc/*`, `docs/implementation-audit.md` | `pnpm api:sqlc` generates admin code; `pnpm check` và `pnpm e2e:smoke` xanh; `admin.Repository` interface được giữ nguyên. |
| 2026-06-30 | Auth repository sqlc migration | `apps/api/sqlc.yaml`, `apps/api/internal/features/auth/queries.sql`, `apps/api/internal/features/auth/repository.go`, `apps/api/internal/features/auth/sqlc/*`, `docs/implementation-audit.md` | `pnpm api:sqlc` generates auth code; `pnpm check` và `pnpm e2e:smoke` xanh; `auth.Repository` interface được giữ nguyên. |
| 2026-06-30 | Attempts repository sqlc migration | `apps/api/sqlc.yaml`, `apps/api/internal/features/attempts/queries.sql`, `apps/api/internal/features/attempts/repository.go`, `apps/api/internal/features/attempts/sqlc/*`, `docs/implementation-audit.md`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md` | `pnpm api:sqlc` generates attempts code; `pnpm check` và `pnpm e2e:smoke` xanh; `attempts.Repository` interface được giữ nguyên; `GET /attempts/{id}` null score/max_score scan error fixed. |
| 2026-06-30 | OpenAPI types CI expansion | `apps/web/src/shared/api/admin.ts`, `apps/web/src/shared/api/attempts.ts`, `apps/web/src/shared/api/openapi-schema.d.ts`, `apps/web/src/pages/dashboard/admin-dashboard-page.tsx`, `apps/web/src/pages/exam/exam-page.tsx`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `.github/workflows/ci.yml`, `docs/implementation-audit.md` | `pnpm api:types`, `pnpm api:sqlc`, `pnpm check`, `pnpm e2e:smoke` xanh; CI YAML hợp lệ; `apiClient` runtime không đổi. |
| 2026-06-30 | Cursor pagination | `apps/api/internal/platform/pagination/cursor.go`, `apps/api/internal/features/admin/*`, `apps/api/internal/features/assessments/*`, `apps/web/src/shared/api/admin.ts`, `apps/web/src/shared/api/assessments.ts`, `apps/web/src/shared/api/attempts.ts`, `apps/web/src/pages/dashboard/admin-dashboard-page.tsx`, `apps/web/src/pages/dashboard/teacher-dashboard-page.tsx`, `apps/web/src/pages/dashboard/audit-logs-panel.tsx`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm api:types`, `pnpm api:sqlc`, `pnpm check`, `pnpm e2e:smoke` xanh; cursor và count hoạt động cho users/audit/assessments; UI load-more có mặt. |
| 2026-06-30 | Password history + login lockout | `supabase/migrations/000013_*`, `apps/api/internal/features/auth/password_policy.go`, `apps/api/internal/features/auth/service.go`, `apps/api/internal/features/auth/handler.go`, `apps/api/internal/features/auth/repository.go`, `apps/api/internal/features/auth/queries.sql`, `apps/api/internal/features/admin/service.go`, `apps/api/internal/features/admin/handler.go`, `apps/api/internal/features/admin/repository.go`, `apps/api/internal/features/admin/queries.sql`, `apps/api/internal/features/auth/sqlc/*`, `apps/api/internal/features/admin/sqlc/*`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; lịch sử 5 mật khẩu, khóa đăng nhập sau 5 lần sai trong 15 phút. |

## 2026-06-30 — OpenAPI fetch client migration (academic-slice4-openapi-fetch)

### Done

- [x] Added `openapi-fetch` runtime dependency to `apps/web`.
- [x] Created `apps/web/src/shared/api/openapi-client.ts` typed client wrapper:
  - Loads base URL from runtime config lazily (singleton).
  - Attaches bearer token from auth session store via middleware.
  - Sends `credentials: 'include'` on every request.
  - Adds `X-CSRF-Token` header on unsafe methods (POST/PUT/PATCH/DELETE) by reusing existing `csrf-middleware` helpers.
- [x] Migrated two read-only academics helpers to `openapi-fetch`:
  - `listClasses()` → `GET /classes`
  - `listEnrollments(classId)` → `GET /classes/{class_id}/enrollments`
- [x] Kept all other helpers (`assessments.ts`, `attempts.ts`, admin users, etc.) on the existing `apiClient`; exported helper names unchanged.
- [x] Updated ADR-0010 to record Stage 3 partial adoption and keep Huma deferred.

### Deferred / not in scope

- Full migration of mutating helpers to `openapi-fetch`.
- Removal or deprecation of `apiClient`.
- Huma runtime migration.

### Decisions / notes

- First migration intentionally limited to GET endpoints to avoid CSRF complexity in the new middleware path; the wrapper is already CSRF-ready for future unsafe methods.
- `openapi-fetch` types the request path, query, body, and response via the generated `paths` type from `openapi-schema.d.ts`.
- Error handling converts `openapi-fetch` error bodies through the existing `createApiError` helper to preserve `ApiResponseError` behavior.
- No backend changes were made; cookie/CSRF contract remains identical.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-30 | OpenAPI fetch client migration | `apps/web/package.json`, `apps/web/src/shared/api/openapi-client.ts`, `apps/web/src/shared/api/academics.ts`, `pnpm-lock.yaml`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/implementation-audit.md` | `pnpm api:types`, `pnpm web:typecheck`, `pnpm web:build`, `pnpm check`, `pnpm e2e:smoke` xanh; `listClasses`/`listEnrollments` sử dụng `openapi-fetch`. |

## 2026-06-30 — Fix assessment detail item nesting (fix-assessment-detail-items)

### Done

- [x] Added `AssessmentSectionID` field to `ItemDetail` model.
- [x] Updated `GetAssessmentItems` repository mapping to populate `AssessmentSectionID` from the sqlc row.
- [x] Fixed `loadAssessmentDetail` in service to map items into sections using `items[i].AssessmentSectionID` instead of `items[i].ID`.
- [x] Updated OpenAPI skeleton `Item.data` schema to include `assessment_section_id` and regenerated TypeScript types.
- [x] Added `TestService_GetAssessment_NestsItemsUnderSections` to assert items appear under the correct sections in detail view.

### Deferred / not in scope

- No broader assessment builder refactor.
- No frontend UI changes beyond generated type update.

### Decisions / notes

- The bug was purely a mapping key mismatch; `GetAssessmentItems` query already selected `assessment_section_id`, so no migration or sqlc query change was required.
- Publish snapshot flow was already correct because it uses `GetAssessmentItemsWithContent` and maps by `AssessmentSectionID`.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-30 | Fix assessment detail item nesting | `apps/api/internal/features/assessments/models.go`, `apps/api/internal/features/assessments/repository.go`, `apps/api/internal/features/assessments/service.go`, `apps/api/internal/features/assessments/service_test.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `docs/implementation-audit.md` | `go test ./internal/features/assessments/...`, `pnpm check`, `pnpm e2e:smoke` xanh; test mới xác nhận items nằm đúng section. |
| 2026-06-30 | Add assessment detail smoke assertion | `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm e2e:smoke` xanh; assertion kiểm tra section trong detail có ít nhất một item với đúng `question_version_id`. |

## 2026-06-30 — Builder upgrades backend (next-batch-slice1-builder-backend)

### Done

- [x] Added backend endpoints:
  - `PATCH /assessment-sections/{section_id}` update title/position.
  - `DELETE /assessment-sections/{section_id}` archive draft section.
  - `PATCH /assessment-items/{item_id}` update question_version_id/points/position.
  - `DELETE /assessment-items/{item_id}` archive draft item.
  - `DELETE /assessments/{id}/targets/{target_id}` archive draft target.
  - `POST /assessments/{id}/sections/reorder` with ordered section ids.
  - `POST /assessment-sections/{section_id}/items/reorder` with ordered item ids.
  - `GET /questions?q=&bank_id=&limit=&offset=` question picker returning published question versions.
  - `GET /assessments/{id}/publications` returning publication history.
- [x] Extended `ValidateAssessment` with detailed errors: active section/item/target presence, points > 0, opens_at < closes_at, max_attempts/duration > 0, question version published, target class active.
- [x] Implemented soft archive via status columns; kept only DRAFT assessments mutable.
- [x] Added sqlc queries and regenerated code; updated Repository interface and wrapper.
- [x] Added service and handler tests for update/delete/reorder/list publications.
- [x] Extended E2E smoke to exercise edit section/item, reorder, delete/re-add item/target, question picker, publications, and validation errors.
- [x] Updated OpenAPI skeleton and regenerated TypeScript types.

### Deferred / not in scope

- Frontend UI for builder upgrades.
- Attempt generation/start from published snapshot.
- Huma/openapi-fetch migration.

### Decisions / notes

- Reorder endpoints require the caller to send all active IDs in the desired order; positions are assigned sequentially (gaps of 10) inside a transaction.
- Question picker lists only questions with a published version; prompt text is read from `prompt_json->>'text'`.
- Publication history returns current assessment status for each publication row; older revisions are not independently tracked.
- Archive uses status='ARCHIVED' for sections/items/targets; assessment itself stays in DRAFT until publish.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-30 | Builder upgrades backend | `apps/api/internal/features/assessments/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; smoke cover edit/delete/reorder/validation errors/publications/question picker. |

## 2026-06-30 — Password history and login lockout

### Done

- [x] Migration `000013_password_history_and_login_lockout.sql` adds `password_history` and `login_attempts` tables with tenant-scoped indexes.
- [x] `auth.PasswordHistoryLength` (5) and shared helpers `CheckPasswordHistory` / `StorePasswordHistory`.
- [x] `auth.Login` checks lockout threshold (5 failed attempts in 15 minutes), records failures on bad passwords, and clears attempts on success.
- [x] `auth.ChangePassword` rejects the last 5 password hashes and stores both the old and new hashes.
- [x] `admin.CreateUser` stores the initial temporary password hash in history.
- [x] `admin.ResetPassword` fetches the current hash, rejects recent history, and stores old + new hashes.
- [x] Added sqlc queries for password history and login attempts to both `auth` and `admin` packages.
- [x] Handlers map `ErrAccountLocked` to HTTP 429 and `ErrPasswordReused` to HTTP 400.
- [x] OpenAPI skeleton updated with 429 response for login and password-history descriptions; TypeScript types regenerated.
- [x] E2E smoke extended with reused-password rejection for change-password and admin reset-password, plus login lockout after 5 failures.

### Decisions / notes

- Lockout key is `(organization_id, username_normalized)` so different tenants cannot lock each other out.
- Password history stores the previous password hash on change so the immediate old password cannot be reused.
- Admin reset stores both the old and new hashes; reusing the just-reset temporary password on the next reset is rejected.

## 2026-06-30 — ADR/docs: Huma evaluation and breached-password provider

### Done

- [x] Updated ADR-0010 with explicit Huma evaluation after sqlc coverage: auto OpenAPI benefits vs chi router/handler rewrite costs, and revisit threshold (~20–25 endpoints).
- [x] Added ADR-0011 documenting breached-password provider (HIBP/external corpus) deferred pending a privacy/egress ADR; password history + lockout + blocklist are interim controls.
- [x] Updated `14-implementation-roadmap.md` to mark Phase 1 items completed: audit log reader/UI, sqlc admin/auth/attempts migrations, generated types CI, cursor pagination, password history/lockout.
- [x] Updated roadmap staged plan to reflect sqlc completed, Huma deferred with cost threshold, and breached-password provider deferred.

### Decisions / notes

- No runtime code or dependency changes in this batch.
- Huma adoption is a cost/risk decision, not a technical blocker.
- Breached-password checking requires a separate privacy/ops review before integration.

## 2026-06-30 — Attempt generation backend

### Done

- [x] `GET /api/v1/me/assessments` lists published/open assessments assigned to the current student via active class enrollments and assessment targets.
- [x] `POST /api/v1/assessments/{assessment_id}/attempts` starts a new attempt or resumes an existing `IN_PROGRESS` attempt, enforcing `max_attempts`.
- [x] Attempt generation reads the latest `assessment_publications.snapshot_json`, flattens sections/items into `attempt_items`, and snapshots prompt/choices/answer_key/points/question_version_id.
- [x] `AttemptSnapshot` now includes `server_time` so the frontend can compare against `expires_at`.
- [x] Added sqlc queries for assigned assessments, latest publication, in-progress attempt lookup, attempt count, and attempt/item creation in the `attempts` feature.
- [x] Added service-level errors and handler mapping for unauthorized role, assessment unavailable, no publication, and attempt limit reached.
- [x] Extended unit tests for list/start/resume/limit/unavailable scenarios.
- [x] Updated OpenAPI skeleton with `/me/assessments`, `/assessments/{assessment_id}/attempts`, `AssignedAssessment`, `AssignedAssessmentList`, and `server_time`; regenerated TypeScript types.
- [x] Extended E2E smoke: student sees assigned published assessment, starts attempt, verifies prompt/choices snapshots and server_time, resumes existing attempt.

### Decisions / notes

- No new database migration was required; existing `attempts.status`, `assessment_publications.snapshot_json`, and `attempt_items` columns suffice.
- Global item positions are assigned sequentially across sections during attempt generation to avoid per-section position collisions.
- Role check uses `student` role from the access token; the repository query also enforces enrollment, so a non-enrolled student receives `assessment_unavailable`.

## 2026-06-30 — Academic admin backend gaps

### Done

- [x] Added `PATCH /api/v1/academic-terms/{term_id}` to update term name/date range.
- [x] Added `PATCH /api/v1/subjects/{subject_id}` to update subject code/name/description.
- [x] Added `PATCH /api/v1/courses/{course_id}` to update course subject/term/code/name.
- [x] Added `PATCH /api/v1/classes/{class_id}` to update class course/name.
- [x] Archive (DELETE) and membership management endpoints (`/classes/{id}/teachers`, `/classes/{id}/enrollments`) were already present and are unchanged.
- [x] Added sqlc `UpdateTerm`, `UpdateSubject`, `UpdateCourse`, `UpdateClass` queries with tenant-scoped `WHERE` and `status='ACTIVE'` guards.
- [x] Added service/handler validation and admin authorization for all update endpoints.
- [x] Updated OpenAPI skeleton with PATCH operations and `Update*Request` schemas; regenerated TypeScript types.
- [x] Added unit tests for unauthorized/invalid-input update paths and extended E2E smoke to exercise all four update endpoints.

### Decisions / notes

- Updates are full-replacement PATCH (all fields required) to keep sqlc queries simple; the admin UI can pre-populate existing values.
- Update queries return the updated row and map `ErrNoRows` to `ErrNotFound` when the resource is missing or already archived.
- Duplicate code on subject/course updates is mapped to `ErrDuplicateCode` (HTTP 409) using the existing `isDuplicateError` helper.

## 2026-06-30 — OpenAPI fetch client migration expansion

### Done

- [x] Migrated all remaining frontend API helpers to `openapi-fetch` typed client:
  - `apps/web/src/shared/api/attempts.ts`: `listAssignedAssessments`, `startAttempt`, `getAttempt`, `saveAnswer`, `submitAttempt`.
  - `apps/web/src/shared/api/admin.ts`: `getOrganization`, `updateOrganization`, `listUsers`, `createUser`, `updateUserRoles`, `resetUserPassword`, `listAuditLogs`.
  - `apps/web/src/shared/api/assessments.ts`: `listAssessments`, `createAssessment`, `getAssessment`, `updateAssessment`, `createSection`, `createItem`, `createTarget`, `validateAssessment`, `publishAssessment`, `listQuestions`, `updateSection`, `deleteSection`, `reorderSections`, `updateItem`, `deleteItem`, `reorderItems`, `deleteTarget`, `listPublications`.
- [x] Extracted shared openapi-fetch response unwrappers (`unwrapData`, `unwrapPaged`, `unwrapVoid`) and `ApiResponseError` into `apps/web/src/shared/api/attempts.ts` for reuse by the other helper modules.
- [x] Verified CSRF middleware path on unsafe methods: `openapi-client.ts` fetches/sets `X-CSRF-Token` and sends `credentials: 'include'` on every request.
- [x] Confirmed all helper signatures and return shapes remain unchanged; existing UI components continue to import from the same modules.
- [x] Updated ADR-0010 to record Stage 3 full adoption of the typed client.

### Deferred / not in scope

- Removal or deprecation of legacy `apiClient` (`apps/web/src/shared/api/api-client.ts`); kept as fallback.
- Huma runtime migration.

### Decisions / notes

- Response unwrappers convert openapi-fetch `{ data, error, response }` into the existing `ApiResponseError` shape via `createApiError`, preserving current error handling in UI.
- Query parameter objects are cleaned before sending so `undefined` values are not serialized into the URL.
- Type-only re-exports of generated schemas remain; no runtime bundler size impact beyond `openapi-fetch` itself.

## 2026-06-30 — Huma revisit docs

### Done

- [x] Revisited Huma decision after academic admin and openapi-fetch migration batch.
- [x] Measured current OpenAPI skeleton size: **44 paths** in `openapi-skeleton.yaml` (above the original ~20–25 threshold).
- [x] Recorded that manual spec maintenance remains manageable because `openapi-typescript` + `openapi-fetch` provide frontend type-safety and CI (`pnpm api:types`, `pnpm api:sqlc`) catches generated-code drift.
- [x] Confirmed Huma runtime migration remains **deferred** due to higher refactor risk/cost than manual maintenance, especially for auth cookie/CSRF/refresh-sensitive handlers.
- [x] Updated ADR-0010 with explicit next-review triggers:
  - Spec drift causing production/type errors ≥ 2 times/month.
  - Paths ≥ 60.
  - Need runtime request/response schema validation.
  - Dedicated refactor sprint with ≥ 80% handler test coverage.
- [x] Updated `14-implementation-roadmap.md` Stage 2 with current path count and revisit trigger.

### Deferred / not in scope

- No Huma dependency installation.
- No handler/router code changes.
- No runtime OpenAPI generation.

### Decisions / notes

- The original ~20–25 endpoint threshold was crossed, but the cost crossover point is higher now that generated types and typed fetch client are automated.
- If Huma is revisited, migration will proceed by feature slice (auth → admin → attempts → assessments → academics), preserving `Repository` interfaces and avoiding big-bang rewrite.

## 2026-07-01 — Builder polish backend

### Done

- [x] Added `POST /api/v1/assessments/{assessment_id}/sections/{section_id}/duplicate` to clone an active section and all its items within a DRAFT assessment.
- [x] Added `POST /api/v1/assessment-sections/{section_id}/items/{item_id}/duplicate` to clone an active item within its section.
- [x] Added `GET /api/v1/assessments/{assessment_id}/preview` returning a student-safe preview with prompts/choices/points/structure and no `answer_key`.
- [x] Enforced DRAFT-only and teacher/admin manager authorization for duplicate endpoints and preview.
- [x] Added `DuplicateSection`/`DuplicateItem` to `assessments.Repository` and sqlc-backed implementation using existing generated queries inside transactions.
- [x] Added service/handler tests for duplicate success, duplicate not-draft rejection, and preview hiding `answer_key`.
- [x] Updated OpenAPI skeleton with duplicate/preview paths and `AssessmentPreview`, `PreviewSection`, `PreviewItem` schemas; regenerated TypeScript types.
- [x] Extended E2E smoke to exercise section duplicate, item duplicate, and preview assertions (prompt/choices present, `answer_key` absent).

### Deferred / not in scope

- Frontend UI for duplicate/preview buttons ✅
- Autosave backend endpoint (existing PATCH assessment settings already supports autosave configuration).
- Student history/gradebook/bulk operations ✅

### Decisions / notes

- Duplicate section title is `{source_title} (copy)` and the new section is placed at `max(section positions) + 10`.
- Duplicated items keep the same points and question version as the source; the new item is placed at `max(item positions in section) + 10`.
- Preview reuses `GetAssessmentItemsWithContent` and strips `answer_key` in the service layer rather than adding a separate query.
- The archived-item position unique-constraint interaction in smoke was avoided by not reordering items after delete/re-add in the same section.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-30 | ADR/docs Huma + breach evaluation | `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/adr/0011-breached-password-provider.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | Docs syntax/yaml reviewed; `pnpm check` xanh. |
| 2026-06-30 | Attempt generation backend | `apps/api/internal/features/attempts/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm check` xanh; `pnpm e2e:smoke` xanh. |
| 2026-06-30 | Academic admin backend gaps | `apps/api/internal/features/academics/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm check` xanh; `pnpm e2e:smoke` xanh. |
| 2026-06-30 | OpenAPI fetch client migration expansion | `apps/web/src/shared/api/openapi-client.ts`, `apps/web/src/shared/api/attempts.ts`, `apps/web/src/shared/api/admin.ts`, `apps/web/src/shared/api/assessments.ts`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/implementation-audit.md` | `pnpm web:typecheck`, `pnpm web:build`, `pnpm check`, `pnpm e2e:smoke` xanh; toàn bộ helpers frontend sử dụng `openapi-fetch`.
| 2026-06-30 | Huma revisit docs | `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | `pnpm check` xanh; ADR ghi rõ 44 paths, Huma vẫn deferred, và các trigger tái xem xét. |
| 2026-07-01 | Builder polish backend | `apps/api/internal/features/assessments/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; duplicate section/item và preview hoạt động.

## 2026-07-01 — Student experience backend

### Done

- [x] Extended `GET /api/v1/me/assessments` to return all assigned assessments with `availability` (`upcoming|open|closed`), `attempts_used`, schedule fields (`opens_at`, `closes_at`), and publication metadata.
- [x] Added `GET /api/v1/me/attempts` returning the current student's attempt history with assessment title, status, timing, and score/grading summary.
- [x] Added `GET /api/v1/attempts/{attempt_id}/result` returning a graded review view with prompts, choices, student answers, correct answers, and per-item `is_correct`.
- [x] Enforced student-only authorization for the new endpoints and restricted result review to `SUBMITTED`/`EXPIRED` attempts.
- [x] Updated `ListAssignedAssessments` sqlc query to drop the request-time window filter, require a published version, and include an `attempts_used` lateral count.
- [x] Added `ListStudentAttempts` sqlc query and repository method.
- [x] Updated OpenAPI skeleton with `/me/attempts`, `/attempts/{attempt_id}/result`, extended `AssignedAssessment`, and new `StudentAttempt`, `StudentAttemptList`, `AttemptResult`, `AttemptResultItem` schemas; regenerated TypeScript types.
- [x] Added service/handler tests for availability classification, attempt history, result review, and result-not-submitted rejection.
- [x] Extended E2E smoke to assert `availability`/`attempts_used` on assigned assessments, save/submit a generated attempt, review its result, and verify attempt history.

### Deferred / not in scope

- Frontend pages for student assessment list, attempt history, and result review ✅
- Release scheduling controls beyond immediate post-submit review.
- Gradebook or teacher result views ✅

### Decisions / notes

- Availability is computed in the service layer against `time.Now().UTC()` so the same query can return upcoming/open/closed assessments without N+1 filters.
- `StartAttempt` now validates that the target assessment's computed availability is `open`, preventing starts on upcoming or closed assessments.
- Result review reuses the existing `GetAttempt` + `GetAttemptItems` queries (which already include `answer_key` snapshots) and marks `is_correct` using the same MCQ matching logic as submit grading.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-06-30 | ADR/docs Huma + breach evaluation | `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/adr/0011-breached-password-provider.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | Docs syntax/yaml reviewed; `pnpm check` xanh. |
| 2026-06-30 | Attempt generation backend | `apps/api/internal/features/attempts/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm check` xanh; `pnpm e2e:smoke` xanh. |
| 2026-06-30 | Academic admin backend gaps | `apps/api/internal/features/academics/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm check` xanh; `pnpm e2e:smoke` xanh. |
| 2026-06-30 | OpenAPI fetch client migration expansion | `apps/web/src/shared/api/openapi-client.ts`, `apps/web/src/shared/api/attempts.ts`, `apps/web/src/shared/api/admin.ts`, `apps/web/src/shared/api/assessments.ts`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/implementation-audit.md` | `pnpm web:typecheck`, `pnpm web:build`, `pnpm check`, `pnpm e2e:smoke` xanh; toàn bộ helpers frontend sử dụng `openapi-fetch`. |
| 2026-06-30 | Huma revisit docs | `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | `pnpm check` xanh; ADR ghi rõ 44 paths, Huma vẫn deferred, và các trigger tái xem xét. |
| 2026-07-01 | Builder polish backend | `apps/api/internal/features/assessments/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; duplicate section/item và preview hoạt động. |
| 2026-07-01 | Student experience backend | `apps/api/internal/features/attempts/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; student list/history/result endpoints hoạt động.
| 2026-07-01 | Teacher gradebook backend + smoke fix | `apps/api/internal/features/gradebook/*`, `apps/api/internal/features/academics/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm check` & `pnpm e2e:smoke` xanh; sửa handler `class_id` param, gradebook/results/export endpoints hoạt động.
| 2026-07-01 | Admin bulk operations backend | `apps/api/internal/features/admin/*`, `apps/api/internal/features/academics/*`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; CSV import users, bulk enroll, bulk assign teachers với dry_run/confirm. |

## 2026-07-01 — Production hardening backend

### Done

- [x] Added in-memory per-IP token-bucket rate limiter (`internal/platform/ratelimit`) with configurable RPS/burst/TTL/cleanup and exclusion for `/healthz`, `/readyz`, `/api/v1/auth/csrf-token`, and `OPTIONS`.
- [x] Added `RATE_LIMIT_*` env config to `apps/api/internal/app/config.go` and `config/render.env.example`; disabled by default for local dev.
- [x] Added structured request logger (`internal/platform/middleware/requestlogger.go`) emitting `request_id`, method, path, status, duration, and remote address.
- [x] Wired `middleware.RequestID`, request logger, and rate-limit middleware into `cmd/server/main.go`.
- [x] Propagated `request_id` to all JSON error envelopes across `admin`, `academics`, `auth`, `attempts`, and `assessments` features.
- [x] Added `GET /api/v1/audit-logs/export` admin-only CSV export with the same filters as the list endpoint and actor name join.
- [x] Added `scripts/render_smoke.sh` for post-deploy smoke against a Render origin and documented usage in `docs/deployment-cli.md`.
- [x] Updated E2E smoke to exercise audit-log CSV export and made `API_BASE` overridable via env for Render reuse.
- [x] Updated OpenAPI skeleton with `/audit-logs/export` and regenerated TypeScript types.

### Deferred / not in scope

- Redis/external rate-limit backend.
- Request ID display in frontend error pages.
- Full audit-log UI/dashboard ✅ (admin audit tab implemented).

### Decisions / notes

- Rate limiter is stateful in-process and sufficient for a single Render instance; scale-out would require a shared store and a separate ADR.
- Request logging uses `chi` `WrapResponseWriter` to capture status without changing handler signatures.
- CSV export streams via `encoding/csv` without loading all rows into memory at once.

## 2026-07-01 — Huma revisit docs (post-polish/student/gradebook/bulk/hardening)

### Done

- [x] Revisited Huma decision after builder polish, student history/review, gradebook, bulk operations, and production hardening batches.
- [x] Measured current OpenAPI skeleton size: **58 paths** in `openapi-skeleton.yaml` (up from 44), very close to the 60-path revisit threshold.
- [x] Recorded that manual spec maintenance remains manageable because `openapi-typescript` + `openapi-fetch` plus the `generated-code-check` CI job cover frontend type-safety and catch generated-code drift.
- [x] Confirmed Huma runtime migration remains **deferred** due to higher refactor risk/cost than manual maintenance, especially for auth cookie/CSRF/refresh-sensitive handlers and middleware ordering.
- [x] Updated ADR-0010 with current path count and explicit next-review triggers.
- [x] Updated `14-implementation-roadmap.md` Stage 2 to reflect 58 paths and the unchanged revisit triggers.

### Deferred / not in scope

- No Huma dependency installation.
- No handler/router code changes.
- No runtime OpenAPI generation.

### Decisions / notes

- The 60-path threshold is likely to be crossed soon, but the cost crossover depends on actual spec-drift incidents, not just path count.
- If Huma is revisited, migration will start with lower-risk slices (academics/gradebook) and leave auth for last.

## 2026-07-01 — Docs backlog refresh (post-resources-MVP)

### Done

- [x] Re-measured OpenAPI skeleton size: **63 paths** (đã vượt ngưỡng 60 lần revisit trước). Resources MVP đóng góp 5 paths mới.
- [x] Cập nhật `14-implementation-roadmap.md` Stage 2: ghi rõ 63 paths, cross 60-path threshold, và lên lịch Huma feasibility spike.
- [x] Cập nhật ADR-0010: thêm mục "Huma feasibility spike" với phạm vi bounded (một feature slice ít nhạy cảm, ngoài auth/CSRF/refresh) và tiêu chí go/no-go rõ ràng. Huma runtime migration vẫn tạm hoãn cho đến khi spike hoàn tất.
- [x] Cập nhật Phase 1 exit criteria (`17-implementation-roadmap.md`): đánh dấu `request_id` display đã implemented (qua `ErrorState` + `formatFriendlyError`).
- [x] Cập nhật Phase 8: tách accessibility baseline (✅) khỏi full manual WCAG audit (pending); error pages `request_id` chuyển sang ✅.
- [x] Cập nhật cả hai roadmap section "Current next backlog": thêm **A. Docs & ADR completion**, **B. Huma feasibility spike (phụ thuộc backend)**, **C. Feature work (chỉ bắt đầu sau A & B)** để khóa "docs-completion before new feature work".
- [x] Bổ sung ADR-0010 với tiêu chí go/no-go cho migration toàn cục (DX, regression risk, runtime validation, handler coverage ≥ 80%).

### Deferred / not in scope

- Không thay đổi code/runtime.
- Không cài Huma, không viết Huma operations.
- Không tái cấu trúc skeleton OpenAPI tự động.

### Decisions / notes

- 60-path threshold đã chính thức bị vượt; Huma revisit là việc cần xảy ra trong chu kỳ kế tiếp, nhưng bounded trước khi migrate toàn cục.
- Feature work (resources UX nâng cao, non-MCQ, manual grading, full a11y, perf, PWA, ...) bị khoá lại sau khi A & B xong để tránh "docs lạc hậu → maintenance cost tăng".

## 2026-07-01 — Playwright browser E2E setup + UI typo fix

### Done

- [x] Added `@playwright/test` to `apps/web` and created `apps/web/playwright.config.ts` (Chromium, workers=1, serial, Vite webServer).
- [x] Added `scripts/e2e_browser.sh` to orchestrate Postgres container, migrations, API build/run, and Playwright.
- [x] Added root/app scripts: `web:e2e`, `web:e2e:install`, `e2e:browser`.
- [x] Added `apps/web/e2e/helpers.ts` with role-based login that tolerates forced password-change state across test runs.
- [x] Added `apps/web/e2e/auth.spec.ts` covering login/role redirects and `apps/web/e2e/critical-flow.spec.ts` covering teacher builder publish, student attempt/submit, teacher gradebook export, and admin bulk import dry-run/confirm.
- [x] Added `data-testid` attributes to login, change-password, dashboard, teacher-dashboard, assessment-builder, exam, gradebook, and admin-dashboard pages for resilient selectors.
- [x] Fixed Vietnamese typos across the UI (duplicate final-i in a user-facing word; misplaced tone mark in a time-period label).
- [x] Added optional manual `browser-e2e` job to `.github/workflows/ci.yml` triggered by `workflow_dispatch`.

### Deferred / not in scope

- Multi-browser matrix (Firefox/WebKit) and parallel workers.
- Full coverage of every UI path; initial suite focuses on critical role-based flows.
- Automatic CI runs on every PR (kept manual to control cost and Docker/Playwright dependency time).

### Decisions / notes

- Browser tests share a single DB/API process and run serially because seeded demo users mutate shared state (e.g., forced password change). Helpers handle already-changed passwords.
- Playwright `--with-deps` installation requires `sudo` and a password in this local environment, so the cached Chromium binary is used; the CI job installs with `--with-deps` because the GitHub runner has passwordless `sudo`.
- `pnpm e2e:browser` passed all 7 tests locally after the login-helper stabilization.

## 2026-07-01 — Queue scheduler groundwork

### Done

- [x] Added `internal/platform/scheduler` package with `Job` interface, ticker-based `Scheduler`, `Start`/`Stop` lifecycle, and `JobFunc` helper.
- [x] Added assessment transition job (`assessments.TransitionJob`) that moves `SCHEDULED`/`PUBLISHED` assessments with `opens_at <= now()` → `OPEN`, and `OPEN` assessments with `closes_at <= now()` → `CLOSED`.
- [x] Added sqlc queries `TransitionAssessmentsToOpen` and `TransitionAssessmentsToClosed` and wired them through the `assessments.Repository` interface.
- [x] Added env config `SCHEDULER_ENABLED` (default `false`) and `SCHEDULER_INTERVAL_SECONDS` (default `60`) in `apps/api/internal/app/config.go`.
- [x] Wired scheduler into `cmd/server/main.go` when DB is available and scheduler is enabled.
- [x] Added tests for the scheduler (run job, error tolerance, stop without start) and the transition job (both transitions called, error propagation).
- [x] Added ADR-0012 documenting in-process scheduler decision and River defer triggers.
- [x] Updated `14-implementation-roadmap.md` with background jobs / scheduler plan.
- [x] Updated `config/render.env.example` with scheduler variables.

### Deferred / not in scope

- River dependency, migrations, or worker process.
- Async CSV import beyond the current 100-row synchronous limit.
- Async grading beyond synchronous MCQ.

### Decisions / notes

- Scheduler is disabled by default for local dev/tests; production Render config enables it with a 60-second interval.
- Transition queries are idempotent, so overlapping runs on multiple instances (if any) will not corrupt assessment state.
- No OpenAPI changes were needed because the scheduler has no HTTP surface.

## 2026-07-01 — Docs & roadmap cleanup

### Done

- [x] Updated `README.md` status to reflect functional MVP and current next backlog.
- [x] Updated `docs/backend/backend-technical-spec/14-implementation-roadmap.md` phase statuses (assessment builder core, attempt runtime core, gradebook core, resources not started) and added current next backlog section.
- [x] Updated `docs/frontend/frontend-technical-spec/17-implementation-roadmap.md` phase statuses and added current next backlog (TanStack Query, error pages with `request_id`, unit tests, IndexedDB offline, resources/files, accessibility).
- [x] Updated `AGENTS.md` "What is NOT wired yet" and "Recently implemented" to match current state.
- [x] Cleaned stale deferred/pending mentions in `docs/implementation-audit.md` for role redirects, forced password change, audit-log UI, builder duplicate/preview, student pages, and gradebook/result views.

### Deferred / not in scope

- No code changes.
- No generated files or dependency changes.
- No deployment or git operations.

### Current next backlog

- Attempt history pagination.
- Exam IndexedDB offline resilience.
- Resources/files UI and backend.
- Accessibility audit.
- Huma/River remain deferred with existing triggers.

## 2026-07-01 — Frontend unit/component tests (Vitest)

### Done

- [x] Added dev dependencies under `apps/web`: `vitest`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`, `jsdom`, `@testing-library/dom`.
- [x] Added `apps/web/vitest.config.ts` with jsdom environment, path aliases (`@/`), globals, setup file, and `e2e/**` excluded.
- [x] Added `apps/web/src/test/setup.ts` importing `@testing-library/jest-dom/vitest`.
- [x] Added scripts:
  - `apps/web`: `test` (vitest run), `test:watch` (vitest).
  - Root: `web:test`.
- [x] Added focused unit/component tests:
  - `src/shared/api/join-api-url.test.ts`
  - `src/shared/api/api-error.test.ts`
  - `src/shared/lib/password-policy.test.ts`
  - `src/shared/auth/auth-session-store.test.ts`
  - `src/shared/config/runtime-config.test.ts`
  - `src/shared/components/error-state.test.tsx`
- [x] Updated `README.md`, frontend/backend roadmaps, and `docs/implementation-audit.md` to mark unit tests as implemented.

### Verification

- `pnpm web:test` passed 41 tests (6 files).
- `pnpm web:typecheck` passed.
- `pnpm web:build` passed.
- `pnpm check` passed.

### Decisions / notes

- E2E specs are excluded from Vitest via `e2e/**` in `vitest.config.ts` to avoid Playwright runner conflicts.
- No coverage gate configured yet; will add when the suite grows or CI requires it.
- MSW deferred until tests need realistic network mocking.

## 2026-07-01 — Frontend error pages with request_id

### Done

- [x] Centralized API error parsing in `apps/web/src/shared/api/api-error.ts` with `ApiResponseError`, `createApiError`, `unwrapData`, `unwrapPaged`, `unwrapVoid`, `getApiErrorDetails`, and `formatFriendlyError`.
- [x] `ApiResponseError` extracts `request_id` from the error envelope and falls back to the `X-Request-ID` response header.
- [x] Backend CORS middleware exposes `X-Request-ID` so the frontend can read it cross-origin.
- [x] Added reusable `ErrorState` component (`apps/web/src/shared/components/error-state.tsx`) with safe message, optional `request_id`, copy button, and retry button.
- [x] Added public `/error/:status?` route and `ErrorPage` (`apps/web/src/pages/error/error-page.tsx`) for 403/429/500/generic error states.
- [x] Updated `NotFoundPage` with `data-testid="not-found-page"`.
- [x] Updated TanStack Query-migrated pages to render `ErrorState` for query/mutation errors:
  - `dashboard/dashboard-page.tsx`
  - `attempt-review/attempt-review-page.tsx`
  - `dashboard/teacher-dashboard-page.tsx`
  - `gradebook/gradebook-page.tsx`
- [x] Updated `gradebook.ts` to throw `ApiResponseError` with status and `request_id` on CSV export failures.
- [x] Added `apps/web/e2e/error-pages.spec.ts` covering 403 with request id, 500, and unknown-route 404.
- [x] Updated `README.md`, frontend/backend roadmaps, and `implementation-audit.md` to mark error pages as done.

### Verification

- `pnpm web:typecheck` passed.
- `pnpm web:build` passed.
- `pnpm e2e:browser` passed 10/10 (3 new error-pages tests + 7 existing tests).

### Decisions / notes

- 401 API errors still show a session-expired message; auth redirect remains handled by `AuthProvider` / `ProtectedRoute`, not the error page.
- Unknown client-side errors do not surface raw `Error.message` to avoid leaking internals; backend messages are treated as safe.
- `formatFriendlyError` supports per-status overrides so pages can keep contextual wording while sharing the request-id logic.

## 2026-07-01 — TanStack Query migration

### Done

- [x] Added `@tanstack/react-query` to `apps/web`.
- [x] Added `QueryProvider` (`apps/web/src/app/providers/query-provider.tsx`) with query retry once for network/5xx errors and no mutation retry.
- [x] Created query keys and hooks under `apps/web/src/shared/api/*-queries.ts` for attempts, gradebook, assessments, and academics.
- [x] Migrated pages from manual `useEffect` + fetch helpers to TanStack Query:
  - `dashboard/dashboard-page.tsx`
  - `attempt-review/attempt-review-page.tsx`
  - `gradebook/gradebook-page.tsx`
  - `dashboard/teacher-dashboard-page.tsx`
- [x] Auth state remains owned by `AuthProvider`; query hooks use the existing `openapi-client` and API helpers without reimplementing refresh/CSRF logic.
- [x] Updated `README.md`, frontend/backend roadmap docs, and `AGENTS.md` to reflect TanStack Query as implemented.

### Verification

- `pnpm web:typecheck` passed.
- `pnpm web:build` passed.
- `pnpm check` passed (web typecheck/build + Go tests/vet/gofmt).
- `pnpm e2e:browser` passed 7/7.

### Decisions / notes

- Generated OpenAPI client/types remain deferred until Huma is adopted or the hand-maintained skeleton becomes too costly to maintain.
- TanStack Query is intentionally scoped to server-state data fetching; forms, URL state, UI state, and exam durable state keep their own owners.

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-07-01 | Production hardening backend | `apps/api/internal/platform/ratelimit/*`, `apps/api/internal/platform/middleware/requestlogger.go`, `apps/api/internal/app/config.go`, `apps/api/cmd/server/main.go`, `apps/api/internal/features/admin/*`, `apps/api/internal/features/{academics,auth,attempts,assessments}/{response.go,models.go,handler.go}`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `scripts/render_smoke.sh`, `docs/deployment-cli.md`, `config/render.env.example`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; rate limit, request logging, request ID errors, audit CSV export, và Render smoke hoạt động. |
| 2026-07-01 | Huma revisit docs | `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | Docs reviewed; `pnpm check` xanh; ADR ghi rõ 58 paths, Huma vẫn deferred, và các trigger tái xem xét. |
| 2026-07-01 | Playwright browser E2E setup + UI typo fix | `apps/web/playwright.config.ts`, `apps/web/e2e/*`, `scripts/e2e_browser.sh`, `package.json`, `apps/web/package.json`, `.github/workflows/ci.yml`, `apps/web/src/pages/login/login-page.tsx`, `apps/web/src/pages/change-password/change-password-page.tsx`, `apps/web/src/pages/dashboard/*.tsx`, `apps/web/src/pages/assessment-builder/assessment-builder-page.tsx`, `apps/web/src/pages/exam/exam-page.tsx`, `apps/web/src/pages/gradebook/gradebook-page.tsx` | `pnpm check` xanh; `pnpm e2e:smoke` xanh; `pnpm e2e:browser` xanh với 7/7 tests passed. |
| 2026-07-01 | Queue scheduler groundwork | `apps/api/internal/platform/scheduler/*`, `apps/api/internal/features/assessments/scheduler_job.go`, `apps/api/internal/features/assessments/scheduler_job_test.go`, `apps/api/internal/features/assessments/queries.sql`, `apps/api/internal/features/assessments/repository.go`, `apps/api/internal/features/assessments/service_test.go`, `apps/api/internal/app/config.go`, `apps/api/cmd/server/main.go`, `docs/backend/backend-technical-spec/adr/0012-background-jobs-river-defer.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `config/render.env.example`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm check`, `pnpm e2e:smoke` xanh; scheduler và transition job tests pass. |
| 2026-07-01 | Docs & roadmap cleanup | `README.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/frontend/frontend-technical-spec/17-implementation-roadmap.md`, `AGENTS.md`, `docs/implementation-audit.md` | Docs reviewed; no code changes; stale pending/deferred items cleaned. |
| 2026-07-01 | Frontend error pages with request_id | `apps/web/src/shared/api/api-error.ts`, `apps/web/src/shared/api/attempts.ts`, `apps/web/src/shared/api/gradebook.ts`, `apps/web/src/shared/components/error-state.tsx`, `apps/web/src/pages/error/error-page.tsx`, `apps/web/src/pages/not-found/not-found-page.tsx`, `apps/web/src/app/router.tsx`, `apps/web/src/index.css`, `apps/web/src/pages/dashboard/dashboard-page.tsx`, `apps/web/src/pages/dashboard/teacher-dashboard-page.tsx`, `apps/web/src/pages/attempt-review/attempt-review-page.tsx`, `apps/web/src/pages/gradebook/gradebook-page.tsx`, `apps/web/e2e/error-pages.spec.ts`, `apps/api/cmd/server/main.go`, `README.md`, `docs/frontend/frontend-technical-spec/17-implementation-roadmap.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | `pnpm check` xanh; `pnpm e2e:browser` xanh với 10/10 tests passed (3 error-pages tests mới + 7 tests hiện có). |
| 2026-07-01 | Frontend unit/component tests (Vitest) | `apps/web/package.json`, `package.json`, `apps/web/vitest.config.ts`, `apps/web/src/test/setup.ts`, `apps/web/src/shared/api/join-api-url.test.ts`, `apps/web/src/shared/api/api-error.test.ts`, `apps/web/src/shared/lib/password-policy.test.ts`, `apps/web/src/shared/auth/auth-session-store.test.ts`, `apps/web/src/shared/config/runtime-config.test.ts`, `apps/web/src/shared/components/error-state.test.tsx`, `README.md`, `docs/frontend/frontend-technical-spec/17-implementation-roadmap.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | `pnpm web:test` xanh với 41 tests passed (6 files); `pnpm web:typecheck` + `pnpm web:build` + `pnpm check` xanh. |
| 2026-07-01 | Docs backlog refresh (post-resources-MVP) | `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/frontend/frontend-technical-spec/17-implementation-roadmap.md`, `docs/implementation-audit.md` | Docs-only; OpenAPI path count đo lại 63 paths (vượt ngưỡng 60); Huma feasibility spike + tiêu chí go/no-go được thêm vào ADR-0010; `request_id` error pages và accessibility baseline được đánh dấu ✅ trong roadmap; "docs-completion before new feature work" được enforce qua thứ tự A/B/C trong current next backlog. |

## 2026-07-01 — Attempt history pagination

### Done

- [x] Backend cursor pagination for `GET /api/v1/me/attempts` with `limit`/`cursor` query parameters.
- [x] Keyset pagination sorted by `created_at DESC, id DESC`; cursor encodes RFC3339Nano `created_at` + `id`.
- [x] Service layer requests `limit+1` rows to compute `has_more` and `next_cursor`; default limit 10, max 50.
- [x] `StudentAttemptList` response wraps the data array with `{ data, page }` while preserving backward-compatible `json.data` access.
- [x] Frontend `useAttemptHistory` hook using `useInfiniteQuery`; student dashboard renders flattened pages and a load-more button.
- [x] Regenerated `openapi-schema.d.ts` from updated OpenAPI skeleton; updated fake repository tests.
- [x] Updated frontend/backend roadmaps and `AGENTS.md` to mark attempt history pagination as complete.

### Verification

- `pnpm api:sqlc` and `pnpm api:types` regenerated clean diffs.
- `pnpm check` passed.
- `pnpm e2e:smoke` passed.
- `pnpm e2e:browser` passed 10/10.

### Decisions / notes

- Cursor encoding uses JSON `{created_at, id}` with base64url raw encoding for stability and debuggability.
- No breaking change for existing consumers reading `json.data` directly.
- Pagination defaults are conservative (10) to keep the student dashboard initial render small.

## 2026-07-01 — Exam IndexedDB offline resilience (MVP)

### Done

- [x] Added `apps/web/src/shared/lib/exam-draft-store.ts` with native IndexedDB wrapper and in-memory fallback.
- [x] Draft storage stores only non-secret per-item data: `attempt_id`, `item_id`, `answer_payload`, `pending` flag, `revision`, `updated_at`.
- [x] Exam page initializes answers from server snapshot and overlays local pending drafts using revision-aware `shouldPreferDraft`.
- [x] `handleSelect` writes a pending local draft before attempting API save; successful save marks the draft synced.
- [x] Failed API saves keep the pending draft and show a local/unsynced status; online event and page load retry pending drafts.
- [x] Added offline banner and per-item statuses: saving / saved locally / syncing / synced / error.
- [x] Drafts are deleted after successful submit and when the attempt is no longer `IN_PROGRESS`.
- [x] Added Vitest tests for the in-memory storage and draft-vs-server preference logic.
- [x] Added reload-persistence assertion to the Playwright critical-flow spec and stabilized picker/target waits in builder specs.
- [x] Removed generated Playwright artifacts after passing run.
- [x] Updated `AGENTS.md`, frontend roadmap, and implementation audit.

### Verification

- `pnpm web:test` passed 52 tests (7 files).
- `pnpm web:typecheck` passed.
- `pnpm web:build` passed.
- `pnpm e2e:browser` passed 10/10.

### Decisions / notes

- No service worker, background sync, or auto-submit; submit remains explicit.
- Last-write-wins MVP: no multi-tab conflict resolution.
- IndexedDB is optional; in-memory fallback keeps the exam page functional in environments without IndexedDB (but drafts won't survive reload).
- Tokens, answer content beyond selected option, and PII are never stored in drafts.

## 2026-07-01 — Resources MVP (slice-7-resources-files)

### Done

- [x] Migration `supabase/migrations/000018_resources.sql` adds `resources`, `resource_files`, and supporting enums.
- [x] `apps/api/internal/platform/storage` exposes a `Provider` interface plus a `LocalProvider` with random hex keys, hex-only key validation, and base-dir containment.
- [x] `apps/api/internal/features/resources` ships models, sqlc-backed repository, service, handler, response envelope, and unit tests covering role/tenant checks, upload, publish/archive, and student vs manager download authorization.
- [x] `cmd/server/main.go` wires the resources feature (repo, service, handler, local storage provider) and registers `GET/POST /resources`, `POST /resources/{id}/publish`, `DELETE /resources/{id}`, `POST /resources/{id}/files`, and `GET /resources/{id}/download`. Placeholders are registered when the database is unavailable.
- [x] Multipart upload enforces `MAX_UPLOAD_SIZE` (default 10 MiB) and replaces the previous ACTIVE file. Download is restricted: students only see PUBLISHED resources, teachers/admins see non-archived resources, and `Content-Disposition` filenames are sanitized.
- [x] `apps/web/src/shared/api/resources.ts` exposes the OpenAPI-fetch client plus a CSRF-aware multipart uploader and a bearer-auth downloader.
- [x] `apps/web/src/shared/api/resources-queries.ts` and `resourceKeys` provide React Query hooks for list/create/publish/archive/upload.
- [x] `apps/web/src/pages/resources/resources-page.tsx` is a minimal teacher/admin upload UI and student list/download UI accessible at `/app/resources`. Linked from the app shell nav.
- [x] OpenAPI skeleton gains `/resources*` paths and `Resource`/`ResourceList`/`ResourceFile`/`CreateResourceRequest` schemas; regenerated `apps/web/src/shared/api/openapi-schema.d.ts`.
- [x] Local `LocalProvider` rejects `..`, `/`, `\\`, whitespace, and non-hex keys. Storage directory is created with 0750 perms; per-object file uses 0640.

### Verification

- `pnpm api:types` regenerated the web client types.
- `pnpm web:typecheck` passed.
- `pnpm web:test` passed 52 tests.
- `pnpm web:build` passed.
- `pnpm e2e:smoke` passed (added resources flow: create, upload, publish, student list, download, student-draft-403).
- `pnpm e2e:browser` passed 10/10.
- `go test ./...` passed including new `internal/features/resources` and `internal/platform/storage` suites.
- `gofmt -l .` clean, `go vet ./...` clean.
- `pnpm check` passed.

### Decisions / notes

- Supabase/S3-compatible storage adapter is deferred behind the `storage.Provider` seam; MVP uses local disk at `RESOURCE_LOCAL_PATH` (default `/tmp/vts-edu-resources`).
- Class-scoped resource access is simplified to org-scoped resources for MVP (a class_id can be supplied as `context_id`; class membership checks are deferred).
- Server-generated random hex keys (32 hex chars); user-controlled paths never reach the filesystem.
- No multipart virus scan, resumable upload, PATCH metadata, or folder grouping in this slice.

## 2026-07-01 — Accessibility audit (slice-8-accessibility-audit)

### Done

- Added skip link in `AuthLayout`, `AppShellLayout`, and `ExamLayout`; targets `#main-content` and is announced when focused.
- Added `<main>` landmarks with `tabIndex={-1}` to all three layouts so the skip link can move focus.
- Added `aria-current="page"` to active nav links in `AppShellLayout`.
- Renamed `nav` `aria-label` to "Điều hướng chính" and added explicit `aria-label` on the brand link, username span, and logout button.
- Introduced `:focus-visible` ring for buttons, links, selects, `[role="tab"]`, and `[role="button"]`; preserved existing `input`/`textarea` focus styles. Mouse clicks no longer show the focus ring; keyboard navigation does.
- Added global `prefers-reduced-motion` rule that suppresses transitions and animations.
- Added a `useDocumentTitle` hook (`apps/web/src/shared/lib/use-document-title.ts`) and wired it on login, change-password, student dashboard, teacher dashboard, admin dashboard, resources, exam, attempt review, assessment builder, diagnostics, error, and not-found pages. The hook appends "– VTS EDU" and restores the previous title on unmount.
- Login page: error banner is `role="alert"` with `aria-live="assertive"` and has an `id` so the form can reference it via `aria-describedby`. Submit button uses `aria-busy` while in-flight.
- Change-password page: error banner always rendered (with `display: none` when empty) so the live region stays attached; password policy hints accept an `id` and the new-password input uses `aria-describedby="password-policy"`. Submit uses `aria-busy`.
- Student dashboard: section/heading `id`s, `aria-label` on history/assessment lists, `aria-busy` on load-more, status badges have visible labels via `statusLabel()`, decorative separators marked `aria-hidden`, attempt status rendered through a label lookup, and "Xem kết quả" link gets a per-row `aria-label`. Card titles downgraded from `<h2>` to `<h3>` to keep heading hierarchy.
- Teacher dashboard: search input has a visually-hidden label, assessment list has `aria-label`, per-row "Sổ điểm" links expose an `aria-label` with the assessment title, class list items expose an `aria-label` with student/teacher counts, and card-link decorative arrow is `aria-hidden`.
- Admin dashboard: tab list now uses `role="tablist"` + `role="tab"` + `role="tabpanel"` with `aria-selected`, `aria-controls`, `tabIndex` (roving tabindex), each panel wrapped in a labelled `tabpanel`. Search input has a label. Users table has a visually-hidden `<caption>`. Org-name edit input gets a label.
- Gradebook: tab list and panels use proper ARIA tab semantics with `aria-controls`/`aria-labelledby`. Both gradebook tables have visually-hidden `<caption>`s. Status badges in the cells carry `aria-label`s. Selects have visible labels.
- Resources: file input is wrapped in a label with a visually-hidden accessible name, status pill has `aria-label`, the timestamp cell uses `<time>`, the table has a caption, and create/upload errors are `role="alert"`.
- Resources create form: separated the visual label and the input so screen readers announce the field, and the form is `aria-labelledby="resources-create-heading"`.
- Academic management: all four sub-tables (terms, subjects, courses, classes) have `<caption>`s, `scope="col"` on headers, and visually-hidden labels for every inline edit input. Sub-tab nav keeps its `aria-label`; create forms have `aria-label` for the form and labels for every input.
- Audit log panel: table has a caption, loading state is `aria-live="polite"`, error banner is `aria-live="assertive"`, load-more uses `aria-busy`.
- Assessment builder: preview dialog now uses `aria-labelledby` pointing at the title, focus is moved to the close button on open and returned to the trigger on close, publication table has a caption, and `<time>` is used for timestamps. Create buttons, save buttons, and edit inputs were kept intact; existing form structure preserved.
- Exam page: existing `role="status"` / `aria-live` for timer, save status, and offline banner were kept; submit button adds `aria-busy`.
- Diagnostics page: health and CSRF status updates are wrapped in `role="status" aria-live="polite"`.
- Error page: page heading is rendered (visually hidden) to guarantee a single h1; `useDocumentTitle` sets the tab title.
- Not-found page: `useDocumentTitle` sets the tab title; no structural change.
- Index `index.html`: added `<meta name="description">`; existing `lang="vi"` retained.
- Added `apps/web/src/shared/lib/use-document-title.test.ts` covering the hook (5 cases) and ran the full unit suite (57 passed, +5 from 52).
- Added `apps/web/e2e/a11y.spec.ts` (10 cases) covering: skip link + main landmark on login, skip-link activation, student dashboard structure, teacher search labelling, admin tab semantics, resources table, gradebook tab/tabpanel, change-password `aria-describedby`, error page alert, and not-found page. Total e2e count is now 20/20 (10 existing + 10 a11y).
- Updated `docs/frontend/frontend-technical-spec/11-accessibility-responsive.md` "Recent improvements" section with the changes shipped in this batch and the known limitations (manual keyboard/screen-reader review, full axe adoption, dark mode).

### Verification

- `pnpm web:typecheck` passed.
- `pnpm web:build` passed.
- `pnpm web:test` passed 57 tests across 8 files.
- `pnpm e2e:browser` passed 20/20 (10 new a11y cases + 10 existing).
- `pnpm check` passed (web typecheck/build + Go test/vet/gofmt).

### Decisions / notes

- Did **not** add `@axe-core/playwright` to keep the dependency footprint minimal and to avoid a major dependency without an accepted ADR. The new `e2e/a11y.spec.ts` uses Playwright's built-in locators and `getByRole`/`toHaveAttribute` to assert the same surface axe would cover for the audited screens.
- Document title hook returns to the previous title on unmount, so navigating away restores the prior title (e.g. when the auth flow redirects back to `/login`).
- Skip link is positioned off-screen until focused (the standard pattern); once focused it slides in via a short transition which is suppressed under `prefers-reduced-motion`.
- Status badges now have explicit `aria-label`s (e.g. "Trạng thái SUBMITTED") so screen readers do not just hear the raw enum.
- The `<caption>` elements on data tables are visually hidden because the surrounding heading already names the table; the caption is the formal accessible name. This satisfies the "label that describes the table" WCAG criterion without adding redundant visible text.
- Inline `min-height` on autoload/save indicator is preserved so sighted users still see the live text change.
- All changes are surgical (no visual redesign, no new components beyond `useDocumentTitle` and the `.visually-hidden`/`.skip-link`/`.table-caption` utility classes).

## 2026-07-01 — Huma v2 feasibility spike result (docs-only recording)

### Done

- [x] Recorded the `spike/huma-academics` outcome into main docs only (no runtime code or dependency merged). The spike branch and its commit (`5eb2fd4 Spike Huma academics feasibility`) remain on the spike branch for reference; this entry is the only on-main artifact.
- [x] Added `docs/backend/backend-technical-spec/spikes/huma-academics-spike.md` (190 lines) with full spike report: scope, implementation, tests, evidence (what worked, non-obvious knobs, caveats for a full migration), DX observations, and the **GO-with-conditions** recommendation.
- [x] Updated ADR-0010 with the post-spike "Kết quả spike (2026-07-01)" section: 4/4 unit tests passed, no regression in `pnpm check` / `pnpm e2e:smoke` / `pnpm e2e:browser`, 3 open issues (problem+json content type, X-Request-Id response header echo, OpenAPI spec divergence), and the next-recommendation: one bounded streaming spike on resources download before full Huma adoption.

### Out of scope (deliberately not merged to main)

- `apps/api/go.mod` / `apps/api/go.sum` — Huma v2.38.0 dependency stays on the spike branch only.
- `apps/api/internal/features/academics/spike_huma.go` and `spike_huma_test.go` — spike source stays on the spike branch.
- `apps/api/cmd/server/main.go` — Huma mount wiring stays on the spike branch.

### Decisions / notes

- Recording the result docs-only keeps main's dependency graph, build, and test surface unchanged. The spike branch is the source of truth for the running code; the spike report and ADR update are the source of truth for the decision on main.
- The three open issues each need their own bounded spike before runtime migration can be approved. The next recommended spike is on a streaming candidate (resources download) to test Huma's `http.Flusher`/SSE handling — the second risk axis the roadmap calls out.
- Hand-maintained `openapi-skeleton.yaml` remains the source of truth for `openapi-typescript` on main until the runtime migration is approved.

## 2026-07-01 — Production Supabase storage adapter (slice-9-storage-adapter)

### Done

- [x] Added `apps/api/internal/platform/storage/supabase.go` with `SupabaseProvider` implementing the existing `storage.Provider` (Store/Retrieve/Delete) over the Supabase Storage REST API. Endpoint shape: `POST/GET/DELETE /storage/v1/object/{bucket}/{key}` with `Authorization: Bearer <service role>` + `apikey` headers. The service role key is propagated only on the wire and is never included in error messages, logs, or HTTP responses.
- [x] Added `internal/platform/storage/content_type.go` with `SanitizeContentType` and `AllowedDownloadContentTypes` (text/*, common image, PDF, Office documents, json/zip/octet-stream). Anything outside the allowlist falls back to `application/octet-stream`. Service-layer upload and handler-layer download both run the sanitizer as defence-in-depth.
- [x] Hardened the resources download response: `X-Content-Type-Options: nosniff` is now set on every download and the persisted content type is run through `SanitizeContentType` before being emitted.
- [x] Updated `internal/app/config.go` to fail fast when `RESOURCE_STORAGE_TYPE=supabase` and any of `SUPABASE_URL` / `SUPABASE_SERVICE_ROLE_KEY` / `SUPABASE_STORAGE_BUCKET` is missing. The error message names which var is missing but never the value.
- [x] Wired `cmd/server/main.go::buildResourceStorageProvider` to switch on `RESOURCE_STORAGE_TYPE` (`local` keeps the existing behaviour; `supabase` constructs a `SupabaseProvider`; unknown values are rejected). The local default is preserved for backwards compatibility.
- [x] Tests:
  - `internal/platform/storage/supabase_test.go` (9 cases): config validation rejects missing/bad inputs, `Store` uses `Authorization: Bearer <key>` + `apikey` + correct path/body, `Store` masks upstream error bodies (no key leak), `Retrieve` returns body, `Retrieve` 404 maps to `ErrObjectNotFound`, `Delete` is idempotent on 404, unsafe keys are rejected, 5xx retried up to 3 times, size mismatch detected, `context` cancellation honoured, `SanitizeContentType` falls back correctly across case/parameter/empty/malformed inputs.
  - `internal/features/resources/handler_test.go` (3 new cases): `X-Content-Type-Options: nosniff` set on success, `Content-Type` falls back to `application/octet-stream` for disallowed types, missing auth returns 401.
  - `internal/features/resources/service_test.go` (+1 case): uploaded attacker-controlled content type is sanitized at the service boundary.
  - Total: 13 new cases, all passing alongside the 21 pre-existing storage tests and 5 pre-existing resources tests.
- [x] Updated `config/render.env.example` to document `RESOURCE_STORAGE_TYPE`, `RESOURCE_LOCAL_PATH`, `MAX_UPLOAD_SIZE` and the private-bucket / server-proxy requirement.

### Security notes

- The `SupabaseProvider` is the only place that touches the service role key. It is never logged, never put in an error message, and never returned in a response. Unit tests assert the absence of the key in error strings for non-2xx upload paths and for context-cancellation errors.
- Endpoint construction is isolated to `SupabaseProvider.objectPath`, and the same `isSafeKey` helper from the local provider is reused so the hex-key invariant is enforced uniformly. No signed-URL generation is exposed.
- The bucket is expected to be private; downloads are proxied through the API (no redirect to a signed URL). The `Retrieve` response body is the only thing the client ever sees.
- 5xx responses are retried (idempotent for `GET`/`DELETE`; `POST` is only safe because the key is generated locally and the upstream object endpoint returns 409 on conflict, which is a permanent error).
- `Retrieve` returns a `*http.Response.Body`; the caller (`Resources` service) closes it. `Store` drains the response body to enable connection reuse.
- `RESOURCE_LOCAL_PATH` is unchanged for the `local` backend. The two backends share `isSafeKey` and `generateKey` so storage keys remain hex-only and server-generated across deployments.

### Verification

- `go build ./...` clean.
- `go test ./...` clean (no regressions in auth, attempts, assessments, admin, gradebook, csrf, ratelimit, scheduler, storage, resources).
- `go vet ./...` clean.
- `gofmt -l .` clean.
- `pnpm check` clean (web typecheck/build + Go test/vet/gofmt).
- `pnpm e2e:smoke` (existing `local` path): resources create + upload + publish + student list + download pass without changes.
- `pnpm e2e:browser` is not re-run for this slice because the changes are backend-only and the existing 20 specs already cover the resources download UX.

### Decisions / notes

- Kept the local provider as the default (`RESOURCE_STORAGE_TYPE=local` is the default in `LoadConfig` and in `render.env.example`). The Supabase adapter is opt-in, so an environment that never sets `RESOURCE_STORAGE_TYPE=supabase` will not change behaviour.
- Did **not** introduce signed URLs, CDN, multipart upload to Supabase (with `tus`/resumable), or public-bucket assumptions. Those are explicitly listed in ADR-0009 deferred and would each warrant their own slice if revisited.
- Did **not** create the Supabase bucket via the API. The bucket must already exist (private) and is the operator's responsibility; the API only requires that the service role key can read/write the configured bucket.
- `MaxRetries` on the Supabase provider defaults to 2 (3 attempts total). 5xx is the only retryable class; the response body of a 5xx is drained before the retry so the connection can be reused. Non-5xx errors (401/403/404/409/...) bubble up immediately.

## 2026-07-02 — Non-MCQ foundation + minimal question bank editor (slice-10-non-mcq)

### Done

- [x] `supabase/migrations/000019_question_types.sql`: added `question_versions.question_type` and `attempt_items.question_type` (CHECK, default `multiple_choice`); relaxed `choices_json` / `answer_key_json` NOT NULL; added `attempts_grading_status_check` accepting `GRADED | PENDING_REVIEW | NOT_GRADED`; seeded demo short_answer (`…0003`) and essay (`…0004`) questions in the demo bank.
- [x] `apps/api/internal/features/assessments/queries.sql` + regenerated sqlc: 9 new queries (`CreateQuestionBank`, `ListQuestionBanksByOrganization`, `GetQuestionBank`, `CreateQuestion`, `ListQuestionsInBank`, `GetQuestion`, `GetQuestionWithBank`, `CreateQuestionVersion`, `GetQuestionVersion`, `GetLatestVersionNumber`, `PublishQuestionVersion`); `ListQuestions` and `GetAssessmentItemsWithContent` now expose `question_type`. Picker SQL now uses `COALESCE(qv.id, '00000000-…-0000'::uuid)` and `COALESCE(qv.status, '')` so that questions with no published version still scan cleanly; the repository filters out the nil-uuid rows.
- [x] `apps/api/internal/features/assessments/{models,service,repository,handler,errors}.go`: new types `QuestionBank`, `QuestionBankQuestion`, `QuestionVersion`, `CreateQuestionBankRequest`, `CreateQuestionRequest`, `CreateQuestionVersionRequest`, `PublishQuestionVersionResult`; `QuestionType*` constants; `validateQuestionContent` (per-type rules); 5 new service methods; 6 new handler methods; 11 new repository methods; stub fakes for the existing tests.
- [x] `apps/api/cmd/server/main.go`: registered 6 new routes under `/question-banks` (`ListQuestionBanks`, `CreateQuestionBank`, `ListQuestionsInBank`, `CreateQuestionInBank`, `CreateQuestionVersion`, `PublishQuestionVersion`).
- [x] `apps/api/internal/features/attempts/queries.sql` + regenerated sqlc: `GetAttemptItems` and `CreateAttemptItem` now include `question_type`; new helpers `GetQuestionVersionType` and `ListQuestionVersionTypes` for type lookups in the attempt repo. `SubmitAttempt` returns `COALESCE(score, '0')::text` so the `pgtype.Numeric` scan never explodes when an attempt is in `PENDING_REVIEW` and the DB column is `NULL`.
- [x] `apps/api/internal/features/attempts/{models,service,repository,attempts_test}.go`: `AttemptItem.QuestionType` and `AttemptResultItem.QuestionType` added; `AttemptResultItem.IsCorrect` is now `*bool` (omitted for `PENDING_REVIEW`); new `gradeItem` per-type dispatch (`essay → PENDING_REVIEW/nil`, `short_answer → GRADED/&bool` with normalized exact match, default → MCQ `answerMatches`); `SubmitAttempt` repo writes `NULL` score for `PENDING_REVIEW`; `flattenSnapshotItems` copies `question_type` from the snapshot into the attempt item. Added 4 new tests: `Submit_EssayAlwaysPendingReview`, `Submit_ShortAnswerExactMatch`, `Submit_MixedMCQAndEssayPending`, `GetAttemptResult_PendingItemNoIsCorrect`; updated the existing MCQ result test for `*bool IsCorrect`.
- [x] `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`: added 4 new paths (`/question-banks` GET/POST, `/question-banks/{bank_id}/questions` GET/POST, `/question-banks/{bank_id}/questions/{question_id}/versions` POST, `/…/versions/{version_id}/publish` POST); added 8 new schemas (`QuestionBank`, `CreateQuestionBankRequest`, `QuestionBankQuestion`, `CreateQuestionRequest`, `CreateQuestionResponse`, `QuestionVersion`, `CreateQuestionVersionRequest`, `PublishQuestionVersionResult`); extended `QuestionPickerItem`, `AttemptItem`, `AttemptResult`, `AttemptResultItem` (added `question_type`, `grading_status`, made `is_correct` nullable).
- [x] `apps/web/src/shared/api/openapi-schema.d.ts`: regenerated via `pnpm api:types`.
- [x] `apps/web/src/shared/api/assessments.ts`: added 7 typed wrappers (`listQuestionBanks`, `createQuestionBank`, `listQuestionsInBank`, `createQuestionInBank`, `createQuestionVersion`, `publishQuestionVersion`) and exported the matching types.
- [x] `apps/web/src/pages/question-banks/question-banks-page.tsx`: new teacher/admin page that creates a bank, creates MCQ/short_answer/essay questions with per-type forms, lists questions in a bank, and publishes DRAFT versions. Uses `data-testid="qb-title|qb-prompt|qb-create"`.
- [x] `apps/web/src/app/router.tsx`: registered `/question-banks` route.
- [x] `apps/web/src/app/layouts/app-shell-layout.tsx`: added teacher/admin-gated "Bộ câu hỏi" nav link.
- [x] `apps/web/src/pages/exam/exam-page.tsx`: refactored `handleSelect` into a generic `persistAnswer`; added `handleTextChange` for short_answer/essay; added `getTextFromPayload`; the JSX now branches on `item.question_type` to render radio / `<input type="text">` / `<textarea>`. Uses `data-testid="exam-short-answer|exam-essay"`.
- [x] `apps/web/src/pages/attempt-review/attempt-review-page.tsx`: helpers `getTextAnswer`, `getAcceptedAnswers`, `questionTypeLabel`; `ReviewItemRow` now shows a `question_type` badge, pending styling for `PENDING_REVIEW` items, and treats `is_correct: null` as "pending" (no "Sai" badge); the score banner shows "Chờ chấm / max" when `grading_status === 'PENDING_REVIEW'`.
- [x] `apps/web/src/pages/assessment-builder/assessment-builder-page.tsx`: picker option labels now include `[TN|TLN|TL] {prompt}` so teachers can see the type at a glance.
- [x] `scripts/e2e_smoke_api.mjs`: `saveAnswerForAttempt` now accepts a `payloadOverride`; new `assertNonMcqFlow` runs after resources and before lockout, covering: bank create, MCQ + short_answer + essay create+validation, picker exposing all three types, mixed-attempt submit → `PENDING_REVIEW` with `max_score=3.00`, and result review showing the pending item with no `is_correct`. `submitAttemptById` accepts `PENDING_REVIEW` as a valid grading status.

### Verification

- `pnpm api:sqlc` and `pnpm api:types` clean.
- `pnpm check` (web typecheck + go test + go vet + gofmt) green.
- `pnpm e2e:smoke` green end-to-end (resources, non-MCQ, login lockout all pass).
- Pre-existing flows unchanged (MCQ grading, resources MVP, auth, attempts cursor pagination).

### Decisions / notes

- PENDING_REVIEW semantics: any item in `PENDING_REVIEW` → attempt `grading_status=PENDING_REVIEW` and `score=NULL` (no false incorrect, no partial score leak). The smoke harness only inspects `submitted.grading_status` and `submitted.max_score`, so the COALESCE-driven `0` string for the score is acceptable for the response.
- Short-answer matching uses exact match against `accepted_answers` (trimmed + lowercased). Essay always `PENDING_REVIEW`. This is intentional; the docs and roadmap both defer manual grading and rubrics to a follow-up slice.
- Snapshot semantics: `question_versions.question_type` is copied into `attempt_items.question_type` on attempt generation, so the item's type is fixed for the lifetime of the attempt — even if a teacher later edits the bank.
- `AttemptResultItem.IsCorrect` is now `*bool` (omitted / null for `PENDING_REVIEW`); the frontend distinguishes "not yet graded" from "incorrect" via the field's presence rather than relying on `grading_status` alone.
- Picker SQL uses `COALESCE(qv.id, '00000000-0000-0000-0000-000000000000'::uuid)` + `COALESCE(qv.status, '')` so that questions whose latest version is `DRAFT` (or has no version at all) still scan successfully; the repository skips those rows. Without the COALESCE the LEFT JOIN's nullable columns crashed sqlc's `*string` scan.
- `SubmitAttempt` SQL returns `COALESCE(score, '0')::text` and `COALESCE(max_score, '0')::text` to keep the `*string` scan stable regardless of `PENDING_REVIEW` writes. The DB column itself is `NULL`-able (set by 19), so the persistence path is correct.
- Did **not** add: rich-text editor, rubric builder, manual grading UI (slice-11 ships the minimal version), question bank delete/archive, media uploads, Huma migration, teacher-only typing of SA/essay in flight (the answer is captured as free text only). The minimal manual grading UI ships in slice-11; rubrics and per-question rubrics stay deferred.
- The `CreateQuestionResponse` schema is intentionally **not** wrapped in `.data`; the OpenAPI YAML declares `required: [question]` at the top level. The frontend wrapper reads it as a flat envelope, matching the existing convention.

## 2026-07-02 — Manual review workflow (slice-11-manual-grading)

### Done

- [x] `supabase/migrations/000020_manual_grading.sql`: new `item_grades` table with `UNIQUE(organization_id, attempt_item_id)`, `CHECK (awarded_score >= 0)`, and indexes on `(organization_id, attempt_id, graded_at DESC)` and `(organization_id, grader_user_id, graded_at DESC)`. One row per attempt item; re-grade is allowed via `ON CONFLICT DO UPDATE`.
- [x] `apps/api/internal/features/grading/` — new feature package: `models.go` (DTOs), `errors.go` (sentinels), `queries.sql` (8 sqlc queries including `ListReviewQueue`, `GetAttemptForGrading`, `GetAttemptItemsForGrading`, `GetAttemptItemForGrading`, `UpsertItemGrade`, `GetItemGrade`, `GetItemGradeByID`, `RecomputeAttemptScore`), `repository.go` (interface + sqlc impl with `nonEmptyStringPtr`/`numericPtr` helpers), `service.go` (auth + `WithinTx` + audit + recompute), `handler.go` (3 HTTP handlers with CSRF on the grade PUT), `service_test.go` (8 fake-repo test cases including validation paths and the audit/recompute integration).
- [x] `apps/api/internal/features/admin/grading_audit_adapter.go`: small adapter that converts the grading package's `AuditLogEntry` into `admin.AuditLogParams` and delegates to the existing `admin.Repository.InsertAuditLog`. Implements the `grading.AuditLogger` interface so the grading package avoids a direct import of admin (no circular dep).
- [x] `apps/api/cmd/server/main.go`: 3 new routes wired under existing CSRF middleware:
  - `GET /api/v1/assessments/{id}/review-queue` (teacher/admin)
  - `GET /api/v1/attempts/{attempt_id}/review` (teacher/admin)
  - `PUT /api/v1/attempts/{attempt_id}/items/{item_id}/grade` (teacher/admin, CSRF-required)
- [x] `apps/api/internal/features/attempts/`: `GetAttemptItems` sqlc query now LEFT JOINs `item_grades` and returns `awarded_score` + `feedback` (COALESCE'd to keep the `*string` scan stable across nullable rows). `AttemptResultItem` model extended with `AwardedScore *string` and `Feedback *string`; `GetAttemptResult` surfaces both on the student review view. SQL is additive; existing MCQ result tests still pass.
- [x] `apps/api/sqlc.yaml` + regenerated sqlc: new `gradingsqlc` package, 8 new queries. All sqlc output is regenerated; no hand-edits.
- [x] `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml` + regenerated `openapi-schema.d.ts`: 3 new paths (`/assessments/{id}/review-queue`, `/attempts/{attempt_id}/review`, `/attempts/{attempt_id}/items/{item_id}/grade`); 9 new schemas (`ReviewQueueList`, `ReviewQueueEntry`, `AttemptGradingContext`, `GradingItemDetail`, `GradingStudentAnswer`, `GradingItemGrade`, `GradeItemRequest`, `GradeItemResponse`, `ErrorEnvelope`); `AttemptResultItem` extended with `awarded_score` (nullable) + `feedback` (nullable).
- [x] Frontend API layer:
  - `apps/web/src/shared/api/grading.ts` — typed wrappers for the 3 new endpoints.
  - `apps/web/src/shared/api/grading-queries.ts` — TanStack Query hooks (`useReviewQueue`, `useAttemptForReview`, `useGradeAttemptItem`) with proper cache invalidation across the queue, the attempt review detail, and the student's attempt result.
  - `apps/web/src/shared/api/query-keys.ts` — `gradingKeys` factory added.
- [x] Frontend pages:
  - `apps/web/src/pages/grading/grading-queue-page.tsx` — assessment selector + per-assessment review-queue table; data-testid `grading-assessment-select`, `grading-queue-table`, `grading-queue-row`, `grading-queue-grade-link`.
  - `apps/web/src/pages/grading/grading-detail-page.tsx` — per-item grading form (score + feedback), re-grade enabled, shows student answer + reference accepted answers; data-testid `grading-items-list`, `grade-score-<id>`, `grade-feedback-<id>`, `grade-save-<id>`, `grading-save-success`.
  - `apps/web/src/app/router.tsx` — `/app/grading` + `/app/grading/:attemptId` routes added.
  - `apps/web/src/app/layouts/app-shell-layout.tsx` — teacher/admin-gated "Chấm bài" nav link.
  - `apps/web/src/pages/attempt-review/attempt-review-page.tsx` — renders `awarded_score` + `feedback` on graded essay/short_answer items; data-testid `review-awarded`, `review-feedback`.
- [x] Smoke coverage (`scripts/e2e_smoke_api.mjs::assertNonMcqFlow`):
  - Creates a pending essay/SA attempt and verifies it shows up in the review-queue endpoint.
  - Hits the attempt-review detail endpoint and asserts it includes both essay and short_answer items.
  - Grades the essay, verifies 200, asserts `awarded_score` echoes back, verifies the attempt is still `PENDING_REVIEW` (SA still ungraded).
  - Confirms MCQ items are rejected with 400 `not_gradable`.
  - Grades the short_answer, verifies the attempt transitions to `GRADED` with a non-null `attempt_score` and `still_pending_items=0`.
  - Re-grades the essay (audit log expects ≥2 `attempt.grade` entries with the right `resource_id`).
  - Student `GET /attempts/{id}/result` returns `GRADED` with `awarded_score` populated.
  - Admin `GET /audit-logs?action=attempt.grade&limit=20` returns ≥2 entries for the essay item.
  - Gradebook `GET /assessments/{id}/attempts` returns the attempt with `grading_status=GRADED` and a non-null `score`.

### Verification

- `pnpm api:sqlc` + `pnpm api:types` clean.
- `pnpm check` clean (web typecheck + web build + go test + go vet + gofmt).
- `pnpm e2e:smoke` passes end-to-end (resources MVP, manual grading + audit + gradebook, login lockout).
- `pnpm e2e:browser` passes 20/20 (no test changes; existing critical-flow + teacher-builder + admin flows unaffected).
- `apps/web/test-results` and `apps/web/playwright-report` cleaned by the e2e_browser.sh trap.

### Decisions / notes

- **Audit seam**: the grading package declares a small `AuditLogger` interface (`InsertAuditLog(ctx, tx, grading.AuditLogEntry) error`) instead of importing admin. The admin package provides a `GradingAuditAdapter` that satisfies the interface and reuses the existing `admin.Repository.InsertAuditLog` (which already has full transaction scoping). This avoids a grading → admin import cycle while keeping the audit-row shape in sync with the rest of the system.
- **Re-grade allowed**: `UpsertItemGrade` uses `INSERT ... ON CONFLICT (organization_id, attempt_item_id) DO UPDATE`. Every save (insert or update) writes a fresh `attempt.grade` audit log entry; the smoke harness asserts ≥2 entries for an item that was graded twice. The `before_json` snapshot captures the prior attempt score state and the `after_json` carries the new awarded score, grader_id, and feedback; metadata records the recomputed attempt score and grading_status.
- **Recompute logic**: the `RecomputeAttemptScore` CTE sums `awarded_score` over all items and sets `grading_status = 'GRADED'` only when every non-MCQ (`essay` / `short_answer`) item has a corresponding `item_grades` row. MCQ items never carry a manual grade, so they don't block the promotion. If any non-MCQ item is still ungraded, the attempt stays `PENDING_REVIEW` with `score=NULL` (matching the existing 19 schema).
- **Validation**:
  - `awarded_score` is a decimal string parsed via `big.Rat`; negative or non-numeric → 400 `invalid_score`.
  - `awarded_score` must be `<= item.points` → 400 `score_exceeds_points`.
  - `question_type ∈ {essay, short_answer}` only → 400 `not_gradable` (rejects MCQ explicitly).
  - `item.AttemptID == path attempt_id` → 404 `item_not_in_attempt`.
  - Teacher/admin role required → 403 `forbidden`; CSRF middleware enforces 403 on unsafe PUTs.
- **Response shape**: `GradeItemResponse` returns the persisted `item_grade`, the recomputed `attempt_score` + `attempt_max_score`, the new `grading_status`, and counts `still_pending_items` / `total_non_mcq_items` so the UI can render "you have 0/2 items left" without an extra round-trip.
- **Result extension is additive**: `AttemptResultItem` gains `awarded_score` + `feedback` as `*string` (nullable). The student review page only renders them when present; existing MCQ result tests are unchanged.
- **COALESCE pattern**: `GetAttemptItems` and `GetAttemptItemsForGrading` use `COALESCE(ig.awarded_score, '0')::text` + `COALESCE(ig.feedback, '')` to keep the sqlc `*string` / `string` scan stable. Without it, rows without a manual grade crash sqlc on `cannot scan NULL into *string`. The repository converts the empty/zero values back to `nil` for the response.
- **CSRF**: the grade PUT lives behind the existing `csrf.Validate(r)` middleware just like the answer-save and submit endpoints. The unsafe-method list in the openapi client automatically attaches `X-CSRF-Token` to the PUT.
- **Did not** add: per-item grading_status column, rubric editor, file-submission attachments, bulk-grade, partial-essay auto-grade, teacher review-draft state, real-time push, AI/ML scoring, performance hardening (N+1 on detail page, large attempts). All explicitly deferred to P2/P3.

## 2026-07-02 — Exam heartbeat / deadline hardening (slice-18)

### Done

- **Backend** (`apps/api/internal/features/attempts/{models,service}.go` + `attempts_test.go`):
  - `AnswerSaved` gains `ServerTime time.Time` (required) + `ExpiresAt *time.Time` (optional, mirrors loaded attempt).
  - `SaveAnswer` captures the loaded `*Attempt` in the tx closure; on success it sets `saved.ServerTime = time.Now().UTC()` and `saved.ExpiresAt = attempt.ExpiresAt` outside the tx. No new query, no new endpoint.
  - `TestService_SaveAnswer_OK` extended to assert `ServerTime` is in `[now-1s, now+1s]` and `ExpiresAt` equals the loaded attempt's `ExpiresAt`. Pre-existing tests untouched.
- **OpenAPI** (`docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`): `SaveAnswerResponse.data` now requires `server_time` and adds optional `expires_at`. Regenerated `apps/web/src/shared/api/openapi-schema.d.ts` via `pnpm api:types` (clean).
- **Frontend** (`apps/web/src/pages/exam/exam-page.tsx` + `index.css`):
  - New `serverTimeOffsetRef` (ms) calibrated from initial `GET /attempts/{id}` `server_time` and re-calibrated on every `saveAnswer` response (both inline and via the `syncPendingDrafts` `onSaved` callback).
  - Countdown `useEffect` now uses `Date.now() + serverTimeOffsetRef.current` so a drifted client clock or stale background tab does not give extra or steal time.
  - **60s heartbeat**: new `useEffect` runs `setInterval(60_000)` while `attempt.status === 'IN_PROGRESS'`. Each tick calls `getAttempt(id)` (existing endpoint, not a new one), updates the offset, and refreshes `status` / `expires_at` / `server_time` in the snapshot. The pre-existing `if (snapshot.status === 'SUBMITTED' | 'EXPIRED')` branches then re-render the submitted/expired UI automatically. Failures are swallowed silently — no user-facing errors mid-exam; the existing offline banner handles the no-network case (`getAttempt` skipped when `!navigator.onLine`).
  - **Deadline warning UX** (per spec — no every-tick announcement): new `deadlineLevel: 'ok' | 'warning' | 'critical'` state. A separate `useEffect` transitions the level only on threshold crossings (≤5 min → warning, ≤1 min → critical, otherwise ok). Visually: a warning banner (`role="status"`, amber) and a critical banner (`role="alert"`, red, gentle pulse). The timer chip itself recolors but the timer element no longer has `aria-live="polite"` so screen readers are not spammed on every second.
  - **Auto-submit at timeLeft=0 explicitly NOT added** (per spec): inputs + submit button already `disabled={isExpired}`; status refresh via heartbeat means a server-side transition to `EXPIRED` becomes visible without polling the page.
- **Offline resilience preserved**: heartbeat failures do not touch the local IndexedDB draft; the existing `syncPendingDrafts` retry-on-online / retry-on-load path is unchanged. The smoke covers the same flow.
- **CSS** (`apps/web/src/index.css`): new `.exam-deadline-banner` (`warning` / `critical` modifiers), `.exam-timer.warning` / `.exam-timer.critical` chip color, `@keyframes exam-deadline-pulse` for the critical banner. No layout-shift; new rules live next to the existing `.exam-timer` block.
- **Smoke** (`scripts/e2e_smoke_api.mjs::saveAnswerForAttempt`): now asserts `json.data.server_time` and `json.data.expires_at` are present on every save-answer round trip (covers the `non-MCQ` flow and the `generated attempt` flow).
- **Verification**: `pnpm api:types` ✓, `pnpm web:typecheck` ✓, `pnpm web:build` ✓ (bundle 360.77 kB / 114.49 kB gz, +0.01 kB gz vs pre-change — the heartbeat is ~10 lines), `pnpm web:test` ✓ (57/57), `pnpm e2e:smoke` ✓ (server_time + expires_at assertions on `saveAnswerForAttempt` green), `pnpm e2e:browser` ✓ 23/23, `pnpm check` ✓.
- **Did not** add: new `POST /heartbeat` endpoint, WebSocket/SSE channel, service worker, auto-submit, proctoring, exam heartbeat-decorator on the backend, separate heartbeat table.

## 2026-07-02 — apiClient cleanup v1 (slice-17)

### Done

- **Migrated gradebook CSV exports off legacy `apiClient`** (`apps/web/src/shared/api/gradebook.ts`): `downloadCsv` no longer goes through `apiClient`; the two export helpers (`exportAssessmentAttemptsCSV`, `exportClassGradebookCSV`) now use the typed `getOpenAPIClient()` + `client.GET('/assessments/{id}/attempts/export', { params: { path: { id } } })` (and the class-gradebook equivalent) and call `.blob()` on the returned `Response`. Auth/CSRF are still injected by the openapi middleware so behavior is preserved. `import { apiClient } from './api-client'` removed from `gradebook.ts`.
- **Removed dead re-export** (`apps/web/src/shared/api/api-client.ts`): the `export { fetchCsrfToken, getCsrfToken } from './csrf-middleware'` line had no remaining callers after the re-import in `diagnostics-page.tsx` was redirected to `csrf-middleware` directly. `apiClient` and `ApiClientOptions` stay (auth-provider still needs them).
- **Diagnostics page blocker documented**: `apps/web/src/pages/diagnostics/diagnostics-page.tsx` calls `/healthz` and `POST /attempts/demo/submit`, neither of which is declared in `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`. openapi-typescript only generates types for declared paths, so migrating to the typed client would fail `tsc --noEmit`. Left as-is and split the `getCsrfToken` import out of `api-client` so the dead re-export could be removed without breaking diagnostics.
- **Untouched** (per scope): `apps/web/src/app/providers/auth-provider.tsx` still uses `apiClient` for `/auth/refresh`, `/me`, `/auth/login`, `/auth/logout`, `/auth/change-password` (deferred — auth flow must be migrated in a follow-up slice). `apps/web/src/shared/api/resources.ts` XHR upload path is unchanged. `api-client.ts` is not deleted.
- **Verification**: `pnpm web:typecheck` ✓, `pnpm web:build` ✓ (bundle 360.77 kB / 114.50 kB gz, +0.27 kB / +0.13 kB gz vs pre-change — re-export was tree-shaken but the typed client + a tiny `Response` reference is inlined), `pnpm web:test` ✓ (57/57), `pnpm e2e:smoke` ✓ (`assertResourcesFlow` + `assertNonMcqFlow` cover both gradebook CSV export endpoints; `assertAdminFlow` covers the audit-logs CSV; `expect(csv).toContain('header')` still passes), `pnpm e2e:browser` (chromium) ✓ 23/23 (extended to 23 with the PWA + notifications slices), `pnpm check` ✓. `grep -rE "\bapiClient\(" apps/web/src` returns 7 hits: 1 declaration in `api-client.ts` + 5 in `auth-provider.tsx` (deferred) + 2 in `diagnostics-page.tsx` (blocker).

## 2026-07-02 — PWA Level 0 installability (manifest only)

### Done

- **Manifest** (`apps/web/public/manifest.json`): name + short_name "VTS EDU", description (vi), `start_url`/`scope`/`id` pinned to `/app` so the installed app lands in the workspace, `display: standalone`, `orientation: any`, `background_color: #ffffff`, `theme_color: #1d4ed8`, `lang: vi`, `dir: ltr`, `categories: ["education", "productivity"]`, 2 icons.
- **Icons** (`apps/web/public/icons/`):
  - `icon.svg` — 512×512 rounded-square brand mark, `#1d4ed8` background + white "V" wordmark, `purpose: any`.
  - `icon-maskable.svg` — 512×512 full-bleed `#1d4ed8` + smaller white "V" centered in the safe zone, `purpose: maskable`. Used by Android/Chrome when the launcher rounds the icon.
- **`apps/web/index.html`**: added `<link rel="manifest" href="/manifest.json">`, `<link rel="apple-touch-icon" href="/icons/icon.svg">`, `<meta name="theme-color" content="#1d4ed8">`, `application-name`, `apple-mobile-web-app-title`, `apple-mobile-web-app-capable=yes`, `apple-mobile-web-app-status-bar-style=default`, `mobile-web-app-capable=yes`. Replaced legacy `/vite.svg` favicon with `/icons/icon.svg`.
- **No service worker**: `grep -rE "serviceWorker|service-worker|workbox|pwa-register"` over `apps/web/{src,index.html,vite.config.ts,public}` returns zero hits. No `navigator.serviceWorker.register(...)`, no Workbox, no `vite-plugin-pwa`, no API/asset caching, no background sync / push. `apps/web/package.json` dependency surface unchanged (zero new runtime deps).
- **Vite static asset pipeline**: Vite copies `public/` to `dist/` at build, so `dist/manifest.json` and `dist/icons/{icon,icon-maskable}.svg` are emitted alongside `index.html`. Build output bundle sizes unchanged — manifest is ~654 B and icons are ~800 B total.
- **Verification**: `pnpm web:typecheck` ✓, `pnpm web:build` ✓ (no new chunks; bundle sizes identical to pre-change 360.50 kB / 114.37 kB gz for `index-*.js`), `python3 -c "import json; json.load(open('apps/web/public/manifest.json'))"` ✓ (valid JSON), icon path resolution ✓ (both `src` values exist on disk), `pnpm e2e:browser` (chromium) ✓ 20/20 still pass (no behavioral change). `pnpm check` ✓.
- **Deferred** (P2/P3): Service Worker registration, asset/API caching, offline app shell, install-prompt UI, custom new-tab/window scope handling, real PNG icons (192/512), splash screen, iOS startup images, push notifications, background sync. Exam offline draft resilience stays at the app-layer IndexedDB level (already shipped).

## 2026-07-02 — Notification inbox + best-effort events (slice-15)

### Done

- **Schema** (`supabase/migrations/000021_notifications.sql`): new `notifications` table with `id`, `organization_id`, `recipient_user_id`, `event_type`, `title`, `body`, `metadata_json jsonb DEFAULT '{}'`, `is_read`, `read_at`, `created_at`; FKs to `organizations` and `users` with `ON DELETE CASCADE`; two indexes — `idx_notifications_inbox (organization_id, recipient_user_id, created_at DESC, id DESC)` and partial `idx_notifications_unread (organization_id, recipient_user_id) WHERE is_read = false`.
- **Notifications package** (`apps/api/internal/features/notifications/`): new `errors.go` (`ErrUnauthorized` / `ErrNotFound` / `ErrInvalidInput`), `models.go` (`Notification` wire struct with `MetadataJSON []byte json:"-"`, `NewNotificationInput`, `EventAttemptGraded` / `EventAssessmentPub` / `EventResourcePublished` constants), `Repository` interface + `sqlcRepository` (with `mapRow` / `metadataJSON` / `timeOrNil` / `formatTime` helpers), `Service` interface (also satisfies `Notifier`), 7 unit tests on a fake repo (`Notify` / `NotifyMany` swallows-and-logs, empty recipient skip, `List` requires actor, `MarkRead` empty id 400, success, `UnreadCount` delegation), `Handler` with 3 endpoints (list `GET /me/notifications?limit=&before=`, unread-count `GET /me/notifications/unread-count`, mark-read `POST /me/notifications/{id}/read`); `adapter.go` (`NotifierAdapter`) bridges the notifications service to the per-package `grading.Notifier` / `assessments.Notifier` / `resources.Notifier` interfaces so the dependency direction stays one-way.
- **Notifier seam** (consuming packages own the interface, notifications package owns the adapter):
  - `grading.Notifier` (1-arg setter); `service.SetNotifier(n)` nil-safe; `GradeItem` fires `notifyAttemptGraded` after the recompute transaction commits, only when `recomp.GradingStatus == "GRADED"`; metadata `{attempt_id, assessment_id}`.
  - `assessments.Notifier` + `assessments.RecipientsResolver` (2-arg setter); `PublishAssessment` calls `notifyAssessmentPublished` after the publish transaction commits; resolves recipients via the resolver (DISTINCT active enrollments + memberships); metadata `{assessment_id, revision}`.
  - `resources.Notifier` + `resources.ClassRecipientsResolver` (2-arg setter); `PublishResource` returns `Resource, error` from `UpdateResourceStatus` and fires `notifyResourcePublished` after success; **class-scoped only** (org-scoped resources are explicitly skipped — documented limitation: org-scoped fan-out needs a future "list all org students" helper); metadata `{resource_id, context_id}`.
- **Best-effort semantics**: `Notify` and `NotifyMany` swallow errors and log via `slog.Default().Warn/Error`; notification insert runs **after** the business transaction commits so a notification failure never rolls back grading, resource publish, or assessment publish. Empty recipient list is a no-op.
- **Wiring** (`apps/api/cmd/server/main.go`): instantiates `notificationsSvc` + `notificationsHandler`; routes the 3 `/me/notifications*` endpoints; type-asserts the service against the per-package setter interfaces and wires them in for grading/assessments/resources; safe with `ok` check + nil-safe setters.
- **OpenAPI** (`docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`): 3 new paths, `Notification` / `NotificationList` / `NotificationUnreadCount` schemas; regenerated `apps/web/src/shared/api/openapi-schema.d.ts` via `pnpm api:types` (clean).
- **Frontend**:
  - `apps/web/src/shared/api/notifications.ts` + `notifications-queries.ts`: typed `listNotifications` / `unreadCount` / `markRead`; `useNotificationsQuery` / `useUnreadCountQuery` with 30s `refetchInterval` (also in background); `useMarkReadMutation` invalidates `notificationKeys.all`.
  - `apps/web/src/shared/api/query-keys.ts`: new `notificationKeys` factory (list + unreadCount).
  - `apps/web/src/shared/components/notification-bell.tsx`: accessible bell button (aria-label / aria-haspopup / aria-expanded), badge with unread count, click-outside and Escape close, dropdown `role="dialog"` with `role="status"` heading, per-item Mark-Read button, empty-state placeholder, stable test-ids `notification-bell` / `notification-bell-badge` / `notification-unread-count` / `notification-dropdown` / `notification-item-{id}` / `notification-empty`.
  - `apps/web/src/app/layouts/app-shell-layout.tsx`: bell mounted next to user-name in app shell, hidden when `isRestricted` (force-password-change).
  - `apps/web/src/index.css`: `.notification-bell` + `.notification-bell__*` + `.notification-item*` styles (badge, dropdown, kind tag, time, body, mark-read button).
- **Smoke** (`scripts/e2e_smoke_api.mjs::assertNotificationFlow`): baseline unread-count, teacher publishes a small assessment targeted at the student's class, retries inbox list up to 6x for the assessment.published event, asserts `recipient_user_id` + `metadata.assessment_id` + `is_read=false`; verifies unread-count grew; mark-read flips `is_read=true`; re-mark-read is idempotent (200); unread-count drops; cross-user mark-read returns 403/404.
- **Verification**: `pnpm web:typecheck` + `pnpm web:build` (bundle 360.50 kB / 114.37 kB gz; bell adds ~25 kB to initial chunk) + `pnpm web:test` (57/57) + `pnpm e2e:smoke` (notification flow added; full smoke passed) + `pnpm e2e:browser` (20/20) + `pnpm check` (typecheck + build + Go test + go vet + gofmt) all green.
- **Not wired** (deferred per slice-15 scope): no SSE / WebPush / PWA / service worker / event bus / outbox / notification preferences / retention purge / admin notification center; org-scoped `resource.published` fan-out; scheduler open/close notifications; admin user-created notifications.

---

## 2026-07-02 — Resources UX (class scope, multi-file, inline preview)

### Done

- **Class-scoped authz** (`apps/api/internal/features/resources/{access,access_adapter,service,handler}.go`):
  - New `ClassAccessChecker` interface in the resources package; default `stubChecker` denies class access (used by unit tests that do not exercise class scope).
  - `AcademicAccessAdapter` in resources wraps the academics repository: `ClassExists` / `CanViewClass` (admin or class teacher or enrolled student) / `CanManageClass` (admin or class teacher). Membership id is resolved via `academics.GetMembershipByUserID`.
  - `academics.Repository.IsStudentEnrolled(orgID, classID, userID)` query added and wired in `academics/repository.go` + `academics/queries.sql` (regenerated sqlc).
  - `ListResources(actor, ListFilter{ContextType, ContextID})` filters by class enrolment for students; teachers/admins see all matching class resources.
  - `CreateResource` / `PublishResource` / `ArchiveResource` / `UploadFiles` / `ListFiles` / `DownloadFile` all enforce class access; class-scope uploads / publishes by a non-managing teacher return 403.
  - Optional `?context_type=class&context_id=<uuid>` query parameters on `GET /resources` (returns 404 when the class does not exist in the org).
- **Multi-file uploads** (`resources/{queries,repository,service,handler}.go`):
  - `POST /resources/{id}/files` accepts the existing `file` field, plus `files[]` and `files` (multi-part). All files become ACTIVE — the previous auto-archive-on-upload behaviour is removed (no migration needed; existing ACTIVE files remain ACTIVE).
  - `UploadFile` is kept as a single-file facade (backward compatible) backed by `UploadFiles`.
  - `GET /resources/{id}/files` lists ACTIVE files for the resource.
  - Handler responses for upload + list use a flat `[{data: file}, …]` envelope (the legacy single-resource `data` envelope per item is preserved for OpenAPI clients).
- **File-specific download with inline preview** (`resources/handler.go`):
  - `GET /resources/{id}/download?file_id=<uuid>&disposition=inline` returns a specific file. When `file_id` is omitted the latest ACTIVE file is returned (existing behaviour).
  - `disposition=inline` switches the response to `Content-Disposition: inline` for safe preview MIME types (image/png, image/jpeg, image/gif, image/webp, image/svg+xml, application/pdf, text/plain, text/csv, text/markdown). All other types fall back to `attachment`. `X-Content-Type-Options: nosniff` is always set.
  - `storage.SanitizeContentType` continues to enforce the response content type allowlist.
- **Frontend** (`apps/web/src/pages/resources/resources-page.tsx`, `shared/api/{resources,resources-queries,query-keys}.ts`, `index.css`):
  - Resource list is now a grouped list of cards; each card shows status, actions, attached files, and a multi-file input.
  - Multi-file upload uses `XMLHttpRequest` with concurrency 3 and per-file progress events. Progress is rendered as a list of `<progress>` bars with live status (`Đang tải…` / `Xong` / `Lỗi: …`) and a hidden live region for screen readers.
  - Class filter radio (all / organization / class) + class selector for teachers and admins. The create-resource form gains a class scope selector and a class picker that uses `useClasses()`.
  - Inline preview modal fetches the file via `disposition=inline` and renders it via `<img>`, `<iframe>` (PDF / text), or falls back to a download button for unsupported types. Modal is keyboard accessible (Escape to close, `role="dialog"`, `aria-modal="true"`).
  - Per-file download button uses the file-specific `file_id` parameter; upload inputs are keyboard accessible (file picker).
- **OpenAPI** (`docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml` + regenerated `apps/web/src/shared/api/openapi-schema.d.ts`): `context_type` / `context_id` query parameters on list, new `resources.listFiles` and updated `resources.uploadFile` / `resources.downloadFile` paths, `disposition` and `file_id` parameters on download.
- **Smoke** (`scripts/e2e_smoke_api.mjs`):
  - `assertResourcesFlow` extended with multi-file upload, file-specific download, and inline preview disposition assertion.
  - New `assertResourceClassScopeFlow` creates a class-scoped resource and verifies teacher + enrolled student see it while a second non-enrolled student is denied (list returns 0 and direct download returns 403).
  - Adds an `outsider` student (admin-created, must change password then login) to drive the negative path.
- **Unit tests** (`resources/{service,handler}_test.go`):
  - `TestService_UploadFiles_AllowsMultipleActive`, `TestService_DownloadFile_SpecificFileID`, `TestService_CreateResource_ClassScope_RequiresManage`, `TestService_ListResources_FilterByClass`, `TestService_DownloadFile_ClassScope_RequiresView`, `TestDownloadHandler_InlineDispositionForImage`, `TestDownloadHandler_InlineFallsBackToAttachmentForBinary`.
  - Fixes to the existing fake repo: distinct `file-{n}` and `res-{n}` IDs and a working `ListResources` that returns all matches.

### Verification

- `pnpm api:sqlc` + `pnpm api:types` clean.
- `pnpm check` clean (web typecheck + web build + go test + go vet + gofmt). 8 new resources tests pass; no other test touched.
- `pnpm e2e:smoke` passes end-to-end. New logs: `resources: create, upload, publish, student list, download, multi-file, inline preview — ok` and `resources class scope: teacher + enrolled student + non-enrolled denied — ok`.
- `pnpm e2e:browser` passes 20/20. The single resources a11y test was updated to look for the new `resources-list` testid (the legacy `resources-table` is gone) and to handle the empty list case.
- `apps/web/test-results` and `apps/web/playwright-report` cleaned by `scripts/e2e_browser.sh` trap.

### Decisions / notes

- **No DB migration** — the schema already supports multiple ACTIVE files per resource (`resource_files.status`) and `resources.context_type` / `context_id`. Existing rows continue to work.
- **No signature breaks** for the existing public API: `GET /resources` still works, `POST /resources/{id}/files` still accepts the single `file` field, and `GET /resources/{id}/download` without query parameters still returns the latest ACTIVE file.
- **Multi-file upload accepts partial success**: each file is stored independently; a failure on a later file does not roll back earlier ones. The handler returns 201 with the list of persisted files.
- **Inline preview allowlist is a strict subset of the download allowlist** (`internal/features/resources/handler.go::inlinePreviewContentTypes`) and never includes application/octet-stream. The 502 path (`errClassAccessUnavailable`) keeps class access errors retryable for clients.
- **Class access seam**: `resources.ClassAccessChecker` is implemented in resources by `AcademicAccessAdapter`; the adapter wraps the academics repository (one-way dependency). The default `stubChecker` is used in unit tests to keep the dependency surface tight.
- **Frontend file-specific download** uses the `file_id` parameter and falls back to the "latest ACTIVE" path when none is provided. The download button stays available for every file regardless of preview support.
- **Accessibility**: file picker keeps the existing `data-testid="upload-{id}"`; the modal is keyboard dismissible and announces itself via `role="dialog"` + `aria-labelledby`. The upload progress region uses a `role="status"` + `aria-live="polite"` live region.
- **Did not** add: file reorder / delete / version history UI, video / audio range streaming, upload cancel / resume across refresh, signed URLs, public bucket, bundle audit, Firefox / WebKit coverage, PWA, full WCAG sweep, apiClient cleanup — all explicitly deferred.

## 2026-07-02 — Frontend bundle split + hidden sourcemaps

### Done

- **Route-level code splitting** (`apps/web/src/app/router.tsx`):
  - All 16 route page components are now imported via `React.lazy` + `dynamic import` so each route becomes its own chunk. The router still uses `createBrowserRouter`; every `element` is wrapped in a tiny `<SuspenseRoute>` helper that renders a `role="status" aria-live="polite"` fallback (`Đang tải…`) while the chunk loads.
  - The auth + app + exam layouts stay statically imported (they hold the chrome / redirect logic and must be present immediately). Only the page modules are lazy.
- **Dashboard panel splitting** (`apps/web/src/pages/dashboard/{admin-dashboard-page,teacher-dashboard-page}.tsx`):
  - `AuditLogsPanel` and `AcademicManagementPanel` are now `React.lazy` imports inside `AdminDashboardPage`. The admin dashboard shell is loaded eagerly; each tab's panel arrives on demand and shows a Vietnamese `Đang tải nhật ký…` / `Đang tải học vụ…` placeholder inside the same tabpanel section.
  - `ClassRosterPanel` is `React.lazy` inside `TeacherDashboardPage`. The teacher dashboard's class list is fully usable without it; the panel only loads when a teacher expands a class.
- **Hidden sourcemaps in production** (`apps/web/vite.config.ts`):
  - `build.sourcemap` is now `process.env.NODE_ENV === 'production' ? 'hidden' : true`. Production builds still emit `.js.map` files into `dist/assets/…` but the public bundle no longer carries a `//# sourceMappingURL=` comment, so end users cannot fetch the maps. Dev and preview builds keep full sourcemaps for local debugging. Operators can still upload the hidden maps to a server-side error monitor.
- **Styling** (`apps/web/src/index.css`): added a small `.loading-fallback` rule (centered, muted, italic) reused by both the router and panel Suspense fallbacks.
- **Documentation** (`AGENTS.md` + `docs/backend/backend-technical-spec/14-implementation-roadmap.md`): recorded the new slice.

### Verification

- `pnpm web:typecheck` clean.
- `pnpm web:build` clean. Initial chunk dropped from 511.19 kB / 147.04 kB gzipped to **335.39 kB / 106.56 kB gzipped** (a ~34% reduction in raw bytes, ~28% reduction gzipped). The 16 page/panel chunks are listed below; they load on demand.
  - `academic-management-panel-…js` 28.69 kB / 6.24 kB gz (largest)
  - `assessment-builder-page-…js` 19.15 kB / 5.45 kB gz
  - `admin-dashboard-page-…js` 16.01 kB / 4.59 kB gz
  - `resources-page-…js` 15.30 kB / 5.24 kB gz
  - `exam-page-…js` 12.42 kB / 4.45 kB gz
  - `gradebook-page-…js` 10.27 kB / 3.08 kB gz
  - `teacher-dashboard-page-…js` 7.09 kB / 2.43 kB gz
  - `attempt-review-page-…js` 6.73 kB / 2.13 kB gz
  - `dashboard-page-…js` 6.71 kB / 2.39 kB gz
  - `question-banks-page-…js` 6.65 kB / 2.25 kB gz
  - `audit-logs-panel-…js` 4.77 kB / 1.95 kB gz
  - `grading-detail-page-…js` 4.72 kB / 1.99 kB gz
  - `grading-queue-page-…js` 2.89 kB / 1.24 kB gz
  - `change-password-page-…js` 3.23 kB / 1.37 kB gz
  - `login-page-…js` 2.24 kB / 1.06 kB gz
  - `class-roster-panel-…js` 2.19 kB / 1.09 kB gz
  - `diagnostics-page-…js` 1.63 kB / 0.73 kB gz
  - …plus shared `query-keys-…js` (14.27 kB), `openapi-client-…js` (7.36 kB), and small api adapter chunks.
- Production build still emits sourcemap files but `grep "sourceMappingURL" dist/assets/index-*.js` returns nothing — the comment is suppressed. Maps remain on disk for operator-side error monitoring.
- `pnpm web:test` 57/57 pass (8 test files).
- `pnpm e2e:browser` 20/20 pass. No test changes were required because the lazy boundary is hidden behind the existing route elements; role redirects, auth guards, and tab semantics are unchanged.
- `pnpm check` clean (web typecheck + web build + go test + go vet + gofmt).

### Decisions / notes

- **Route lazy boundary is `SuspenseRoute`**: a small wrapper that renders a `role="status"` placeholder. We avoided putting the Suspense around the page itself so the layouts (auth, app, exam) stay mounted and the auth redirect / role guard logic still sees a stable parent.
- **Static imports kept on purpose**: layout components, auth provider, query provider, `LoginPage`'s css-only neighbours — these are tiny and shared by every route, so splitting them is not worth the extra round-trip.
- **Panels inside admin/teacher dashboards are lazy, the dashboards themselves are not**: the dashboards are page-level chunks already; lazy-loading their child panels is a second split inside an already-deferred route.
- **Hidden over false** for sourcemap: false would skip generating maps entirely, removing the ability to decode production stack traces. Hidden keeps the maps on disk without exposing them publicly. No app code references `process.env.NODE_ENV` directly, so the build-time branch is the only effect.
- **No new dependencies**: only React 19 `lazy` / `Suspense` (already in use) and Vite's existing `sourcemap` option.
- **Did not** add: bundle analyzer dependency, virtualized lists, CSS splitting, PWA, browser matrix, apiClient cleanup, route preloading (`React.startTransition` / `<link rel="modulepreload">`), chunk naming customisation. All explicitly deferred.

## 2026-07-02 — Playwright cross-browser matrix (Firefox + WebKit)

### Done

- **`apps/web/playwright.config.ts`**: keeps Chromium as the default project so `pnpm e2e:browser` stays the fast local path. Adds `firefox` + `webkit` projects under the `PLAYWRIGHT_BROWSERS=1` env flag. The config comment documents that WebKit additionally needs host libraries and that the matrix runner probes them.
- **`scripts/e2e_browser_all.sh`** + `pnpm e2e:browser:all`: spins up the same DB + API as `e2e_browser.sh`, then runs `pnpm web:e2e` with `PLAYWRIGHT_BROWSERS=1`. Before the run it tries to install the missing browsers and probes the WebKit host dependencies (libgtk-4, libgraphene-1.0, libxslt, libevent-2.1, libopus, libgstallocators) by launching and closing a headless WebKit. If the probe fails, the script falls back to chromium + firefox only and prints the install hint.
- **`pnpm web:e2e:install:all`** + `apps/web/package.json::e2e:install:matrix`: one-shot helper to install all three browser binaries without system deps (system deps for WebKit are still required at runtime and not installed without sudo).
- **`apps/web/e2e/critical-flow.spec.ts`**: locator for the student assessment card gained `.first()` to remain robust when the test data accumulates across re-runs (Firefox previously hit a strict-mode violation in the matrix because the previous chromium run left the same title in the seed).

### Verification

- `pnpm e2e:browser` (default Chromium-only path) still 20/20 green.
- `pnpm e2e:browser:all` ran the matrix (chromium + firefox + webkit). WebKit probe failed with the documented missing libs; the script automatically fell back to chromium + firefox. Result: 38/40 pass — chromium 20/20, firefox 18/20. The 2 Firefox failures are the long-flow `critical-flow.spec.ts::admin bulk imports` and `teacher-builder.spec.ts::teacher assessment builder` hitting Firefox-headless GPU/SWGL renderer crashes (`RenderCompositorSWGL failed mapping default framebuffer, no dt`) on the WSL2 host. Same tests pass in chromium and on a properly GPU-accelerated Firefox host. No code change would fix this without a renderer upgrade.
- `pnpm check` clean (web typecheck + web build + go test + go vet + gofmt).

### Decisions / notes

- **Opt-in matrix**: chromium stays the only project by default so `pnpm e2e:browser` is still the 2-minute local check. The matrix is only turned on when the env flag is set or `pnpm e2e:browser:all` is invoked.
- **WebKit fallback is graceful, not a hard error**: the script prints the missing libs and continues with chromium + firefox. The user gets matrix coverage where the host allows, plus a clear instruction (`playwright install --with-deps webkit`) to opt into WebKit when the libs are available.
- **No new dependencies**: only existing `@playwright/test` + `playwright install` CLI. No package.json additions beyond the script entries.
- **Firefox renderer is host-bound**: headless Firefox on WSL2 / VM hosts without a real GPU hits SWGL renderer crashes during long flows. Keeping the matrix in place lets CI / properly GPU-accelerated hosts run all three browsers without us blocking on a single developer's host.
- **Did not** add: per-test retries, browser-specific test fixtures, GitHub Actions workflow for the matrix (deferred — the local script is the entry point; CI can opt in with `PLAYWRIGHT_BROWSERS=1 pnpm e2e:browser:all`).
