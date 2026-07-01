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
pnpm e2e:browser    # browser E2E via Playwright (starts/stops DB, API, and Vite automatically)
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
