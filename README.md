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
| Queue | River in-process (planned, not wired yet) |

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

Core product features implemented incrementally:

- Backend auth (`/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/change-password`, `/me`) with JWT + rotating refresh cookie + CSRF, persisted multi-role memberships, and forced password change.
- Attempt runtime endpoints (`GET /attempts/{id}`, `PUT /attempts/{id}/answers/{item_id}`, `POST /attempts/{id}/submit`) with tenant ownership, request-time expiration, real question prompt/choices snapshots, and synchronous MCQ grading.
- Teacher assessment list (`GET /assessments`) role-gated to teacher/admin.
- Admin organization/user management (`GET/POST /users`, `PUT /users/{id}/roles`, `POST /users/{id}/reset-password`, `GET/PATCH /organizations/current`).
- E2E smoke covering auth roles, change password, attempt grading, teacher assessment list, and admin user/org flow.

Remaining work (academics, full assessment builder, resources, gradebook, advanced grading, OpenAPI client generation, sqlc/Huma migration) is documented in the roadmap and will be built in later phases.
