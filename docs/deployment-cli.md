# Deployment CLI Setup

This project uses **Vercel CLI** and **Supabase CLI** as local devDependencies. No global install or sudo is required.

## Install

```bash
pnpm install
```

> pnpm-lock.yaml is committed. CI and Vercel builds use `--frozen-lockfile`. If you intentionally update dependencies, run `pnpm install` locally and commit the lockfile change.

## Vercel CLI

Vercel CLI manages the frontend SPA in `apps/web`.

### Available scripts

| Script | Command | Notes |
|---|---|---|
| `pnpm vercel:version` | `vercel --version` | Verify CLI |
| `pnpm vercel:pull` | `vercel env pull .env.local` | Pull env vars (requires login) |
| `pnpm vercel:build` | `vercel build` | Local production build |
| `pnpm vercel:deploy:preview` | `vercel deploy` | Preview deployment (requires login) |

### Login

```bash
pnpm exec vercel login
```

This writes `~/.vercel/` (not in repo). Never commit `.vercel/`.

### Link project

```bash
cd apps/web
pnpm exec vercel link
```

## Supabase CLI

Supabase CLI manages local Postgres/Storage and remote migrations.

### Available scripts

| Script | Command | Notes |
|---|---|---|
| `pnpm supabase:version` | `supabase --version` | Verify CLI |
| `pnpm supabase:start` | `supabase start` | Start local stack |
| `pnpm supabase:stop` | `supabase stop` | Stop local stack |
| `pnpm supabase:status` | `supabase status` | Show local service URLs |
| `pnpm supabase:db:reset` | `supabase db reset` | Reset local DB and apply migrations |
| `pnpm supabase:db:push:local` | `supabase db push --local` | Push to local DB |
| `pnpm supabase:db:push:remote` | `supabase db push` | Push to linked remote project |
| `pnpm supabase:migration:list` | `supabase migration list` | List migrations |
| `pnpm supabase:link` | `supabase link` | Link to remote project |

### Local workflow

```bash
pnpm supabase:start
pnpm supabase:db:reset
```

### Login / link

```bash
pnpm exec supabase login
pnpm exec supabase link
```

## Authentication

- Do not commit `.vercel/`, `.env.local`, Supabase access tokens, or service keys.
- Use platform dashboards and GitHub repository secrets for CI/CD.

## Render backend

Render manages the Go API as a Docker Web Service.

### Create the Render Web Service

1. In the Render dashboard, create a new **Web Service** and link this repository.
2. Select **Docker** runtime.
3. Set **Root Directory** to `apps/api`.
4. Use `Dockerfile` as the Dockerfile path and `.` as the Docker context.
5. Add the environment variables from `config/render.env.example` (fill in real values in the dashboard).
6. Set the health-check path to `/readyz`.
7. The service listens on `PORT=8080`.

For a blueprint starting point, see `apps/api/render.yaml`.

## Local smoke test

Before deploying, run the full auth → attempt flow locally. See [`docs/e2e-local-run.md`](./e2e-local-run.md) for step-by-step instructions using Supabase local, the Go API, and the Vite dev server.

## Backend deployment

The Go API is deployed to **Render Free**, not Vercel. See `apps/api/render.yaml` and `config/render.env.example`.

## Post-deploy smoke test

After a Render deploy finishes, run the same smoke suite against the live origin:

```bash
API_BASE=https://<your-api>.onrender.com ./scripts/render_smoke.sh
```

The script uses `scripts/e2e_smoke_api.mjs` with `API_BASE` set to the Render origin, so it exercises auth, attempts, assessment builder, admin, academics, gradebook, bulk operations, and audit logs against the deployed service without starting a local database or building the binary.

Requirements:

- The origin must expose `/readyz` and the demo seed data used by the smoke suite.
- Do not run this against a production instance with real user data; it mutates state (creates users, classes, assessments, attempts) using the seeded demo accounts.
