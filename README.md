# VTS EDU — LMS & Online Assessment Platform

**Repository:** https://github.com/BrianNguyen29/vts-edu.git

Product-ready core for a cost-efficient LMS with online assessments.

## Stack

| Layer | Service |
|---|---|
| Frontend | Vercel Hobby — React 19 SPA |
| Backend | Render Free — Go API |
| Database | Supabase Free — PostgreSQL 15+ |
| Storage | Supabase Storage |
| Queue | In-process scheduler (implemented); River deferred |
| E2E | Playwright browser tests (`pnpm e2e:browser`) |

See `docs/backend/backend-technical-spec/adr/0005-deployment-topology.md` for the full deployment ADR.

## Repository layout

```text
├── apps/api/                           # Go API
├── apps/web/                           # Vite + React frontend
├── config/                             # Environment example files
├── docs/backend/backend-technical-spec # Backend technical specifications
├── docs/frontend/frontend-technical-spec # Frontend technical specifications
├── docs/                               # Audit/review documents
├── supabase/                           # Baseline SQL migrations
└── .github/workflows/                  # GitHub Actions (backup, CI)
```

## Getting started

1. Copy `config/*.env.example` to your platform dashboards / local `.env` files.
2. See `apps/api/README.md` and `apps/web/README.md` for local development commands.
3. Apply Supabase migrations from `supabase/migrations/`.

## Security

- Do not commit `.env` files or real secrets.
- CSRF token is required for cookie-backed unsafe requests.
- Refresh cookie uses `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth` for the cross-origin demo.

## Status

Current functional MVP state:

- Auth: login/refresh/logout/change-password, JWT access tokens, rotating HttpOnly refresh cookie, CSRF double-submit, persisted multi-role memberships, password history/lockout, forced password change, role-based redirects.
- Attempt runtime: start/get/save/submit with tenant ownership, request-time expiration, question snapshot, synchronous MCQ grading, result review, attempt history.
- Assessment builder: create draft, sections/items/reorder, duplicate section/item, settings/schedule, validation, preview, publish snapshots, teacher assessment list.
- Teacher workspace: assigned classes, assessment list, attempt results, gradebook grid, CSV export.
- Admin workspace: org settings, user CRUD/roles/reset-password, audit log list/export, CSV user import, academic terms/subjects/courses/classes CRUD, bulk teacher assignment/enrollment.
- Student dashboard: assigned assessments, attempt history, result review, exam runner.
- DevOps/quality: `pnpm check`, `pnpm e2e:smoke`, `pnpm e2e:browser` (Playwright, with optional `firefox`/`webkit` matrix via `PLAYWRIGHT_BROWSERS=1` + `pnpm e2e:browser:all`), in-process scheduler for assessment open/close transitions. Optional `pnpm e2e:load` (manual, NOT part of `pnpm check`) for bounded attempt-runtime load + concurrency scenarios.
- Resources/files: org + class-scoped file materials (class-scope, multi-file, inline preview, per-file upload progress) with `LocalProvider` (default) and `SupabaseProvider` (production) storage adapters; bearer-auth download with sanitized `Content-Disposition`, `X-Content-Type-Options: nosniff`, content-type allowlist.
- Notifications: best-effort inbox (`attempt.graded` / `assessment.published` / `resource.published`) consumed by grading, assessments, and resources via a one-way `Notifier` seam; `NotificationBell` in app shell, 30s polling, accessible dropdown.
- Frontend polish: bundle split (lazy-loaded routes + dashboard panels), cross-browser matrix (chromium default + opt-in firefox/webkit with WebKit host-dep probe), PWA Level 0 installability (manifest only, no service worker), accessibility baseline (focus-visible, ARIA, keyboard flows, `request_id` in error states), apiClient cleanup v1 (gradebook CSV exports migrated to typed `openapi-fetch`).

Next backlog (not started): Huma migration (awaiting backend feasibility spike go/no-go), River background-job runtime, rich text / TipTap + KaTeX production rollout (spike complete, see `docs/frontend/frontend-technical-spec/spikes/rich-text-editor-spike.md`; rollout gated on a follow-up slice that adds a typed `prompt_doc` column and wires the renderer into exam/review/grading), full WCAG 2.1 AA audit, dark mode, full installable PWA / push / background sync, auth-provider apiClient migration.

### Optional load / concurrency tests (manual, not in `pnpm check`)

```bash
pnpm e2e:load         # start DB + apply migrations + start API + run load
pnpm e2e:load:direct  # run load against an already-running API on $API_BASE
```

See `scripts/e2e_load.sh` + `scripts/e2e_load_api.mjs`. The harness covers bounded
attempt-runtime scenarios (concurrent saves, concurrent submits + idempotency,
save-after-submit, burst reads) with N = 8–16. It is **not** wired into `pnpm
check` or any CI pipeline — it is a developer manual check, and it expects a
local E2E Postgres on port 5434 (managed by the script).
