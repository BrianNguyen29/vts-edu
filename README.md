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
- DevOps/quality: `pnpm check`, `pnpm e2e:smoke`, `pnpm e2e:browser` (Playwright), in-process scheduler for assessment open/close transitions.

Next backlog (not started): generated OpenAPI client, attempt history pagination, exam IndexedDB offline resilience, resources/files, accessibility audit, Huma/River (deferred with triggers).
