# VTS EDU — LMS & Online Assessment Platform

**Repository:** https://github.com/BrianNguyen29/vts-edu.git

MVP demo scaffold for a cost-efficient LMS with online assessments.

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

Core MVP features implemented incrementally:

- Backend auth flow (`/auth/login`, `/auth/refresh`, `/auth/logout`, `/me`) with JWT + rotating refresh cookie + CSRF.
- Attempt runtime endpoints (`GET /attempts/{id}`, `PUT /attempts/{id}/answers/{item_id}`, `POST /attempts/{id}/submit`) with tenant ownership and request-time expiration.
- Frontend dashboard demo link and exam runner wired to a fixed demo attempt UUID.

Other features (academics, question bank, full assessment builder, advanced grading, resources, gradebook) remain spec-only and will be built in later phases.
