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
| S4 | Academics + full question bank + assessment builder | fixer/designer | Partial — academics CRUD/bulk, assessment builder core, question bank minimal implemented |
| S5 | Resources, assignments, gradebook | fixer/designer | Partial — gradebook backend implemented; resources/assignments not started |

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

- Frontend UI for duplicate/preview buttons.
- Autosave backend endpoint (existing PATCH assessment settings already supports autosave configuration).
- Student history/gradebook/bulk operations.

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

- Frontend pages for student assessment list, attempt history, and result review.
- Release scheduling controls beyond immediate post-submit review.
- Gradebook or teacher result views.

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
- Full audit-log UI/dashboard.

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

## Change log

| Date | Task | Files changed | Verification |
|---|---|---|---|
| 2026-07-01 | Production hardening backend | `apps/api/internal/platform/ratelimit/*`, `apps/api/internal/platform/middleware/requestlogger.go`, `apps/api/internal/app/config.go`, `apps/api/cmd/server/main.go`, `apps/api/internal/features/admin/*`, `apps/api/internal/features/{academics,auth,attempts,assessments}/{response.go,models.go,handler.go}`, `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`, `apps/web/src/shared/api/openapi-schema.d.ts`, `scripts/e2e_smoke_api.mjs`, `scripts/render_smoke.sh`, `docs/deployment-cli.md`, `config/render.env.example`, `docs/implementation-audit.md` | `pnpm api:sqlc`, `pnpm api:types`, `pnpm check`, `pnpm e2e:smoke` xanh; rate limit, request logging, request ID errors, audit CSV export, và Render smoke hoạt động. |
| 2026-07-01 | Huma revisit docs | `docs/backend/backend-technical-spec/adr/0010-huma-sqlc-staged-groundwork.md`, `docs/backend/backend-technical-spec/14-implementation-roadmap.md`, `docs/implementation-audit.md` | Docs reviewed; `pnpm check` xanh; ADR ghi rõ 58 paths, Huma vẫn deferred, và các trigger tái xem xét. |
