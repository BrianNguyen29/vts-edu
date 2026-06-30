# Local E2E / Smoke Run Guide

This guide verifies the full auth → attempt runtime flow locally without deploying.

## Prerequisites

- Node.js 22+ and pnpm 9+
- Docker (for Supabase local stack)
- Go 1.25.0+ (for backend)

## 1. Install dependencies

```bash
pnpm install
```

## 2. Prepare the database

### Option A — Direct PostgreSQL container (recommended for smoke checks)

This path only needs a Postgres 15 container and avoids waiting for the full Supabase local stack.

```bash
pnpm e2e:db:start   # starts vts-e2e-postgres on localhost:5434
pnpm e2e:db:migrate # applies supabase/migrations/*.sql and verifies the demo attempt
```

### Option B — Full Supabase CLI stack

```bash
pnpm supabase:start
pnpm supabase:db:reset
```

This applies all migrations and seeds the demo user/attempt. Useful when you need Supabase Storage/Auth services, but it can fail on non-DB service health in constrained environments. If `supabase:start` hangs, use Option A.

## 3. Run the backend

Copy and fill `config/render.env.example` as `apps/api/.env`, then:

```bash
cd apps/api
# minimum required env:
#   DATABASE_URL=<Postgres connection string>
#   JWT_SIGNING_KEY=<256-bit random key>
#   REFRESH_TOKEN_KEY=<256-bit random key>
#   FRONTEND_ORIGINS=http://localhost:5173
#   DB_MAX_OPEN_CONNS=5
#   DB_MAX_IDLE_CONNS=2
#   ACCESS_TOKEN_TTL=15m
#   REFRESH_TOKEN_TTL=7d
go run ./cmd/server
```

For the direct Postgres path use `DATABASE_URL=postgres://postgres:postgres@localhost:5434/postgres`.

The API listens on `http://localhost:8080`.

## 4. Run the frontend

Copy and fill `config/vercel.env.example` as `apps/web/.env.local`, then:

```bash
# apps/web/.env.local
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

```bash
pnpm web:dev
```

The SPA is served on `http://localhost:5173`.

## 5. Smoke the demo flow

### Automated API smoke

```bash
pnpm e2e:smoke
```

This starts the Postgres container, applies migrations, runs the Go API, and exercises `/readyz` → login → `/me` → get attempt → save answer → submit → role seeds → forced password change → teacher assessment list → admin user/org management. Cleanup is handled automatically.

### Manual browser smoke

1. Open `http://localhost:5173`.
2. Log in with one of the seeded demo credentials:
   - Student: organization `school-a`, username `hs001`, password `Password123!`.
   - Teacher: organization `school-a`, username `gv001`, password `Password123!` (forced password change on first login).
   - Admin: organization `school-a`, username `admin001`, password `Password123!` (forced password change on first login).
3. From the dashboard, click **Thi thử demo**.
4. The exam runner loads attempt `00000000-0000-4000-8000-000000000001`.
5. Save answers and submit.

Stop the E2E database when done:

```bash
pnpm e2e:db:stop
```

## Troubleshooting

- **CORS errors**: ensure `FRONTEND_ORIGINS` includes `http://localhost:5173` exactly, and `VITE_API_BASE_URL` ends with `/api/v1` (not `/api/v1/`).
- **CSRF 403 on save/submit**: the frontend fetches `/api/v1/auth/csrf-token` before unsafe requests. Verify the API is reachable and the browser can read the `vts_csrf` cookie.
- **API URL duplication**: `apiClient` joins `VITE_API_BASE_URL` with `/api/v1/...`. Do not set the base URL to `http://localhost:8080/api/v1/api/v1`.
- **Database connection refused**: check `pnpm supabase:status` for the local DATABASE_URL and confirm `DB_SKIP` is not set.
