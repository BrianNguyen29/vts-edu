# Supabase Configuration

This directory contains Supabase-related configuration and environment guidance for the MVP demo.

## Files

- `../supabase.env.example` — Environment variables pointing at the Supabase project (URL, keys, bucket).
- `../../supabase/config.toml` — Local Supabase CLI configuration (`project_id = "vts_edu_local"`).
- `../../supabase/migrations/` — Baseline SQL migrations.

## CLI setup

Supabase CLI is installed as a root devDependency:

```bash
pnpm supabase:version
```

## Local development

```bash
pnpm supabase:start       # start local Postgres/Storage/Studio
pnpm supabase:status      # show service URLs
pnpm supabase:db:reset    # reset local DB and apply migrations
```

## Remote project

```bash
pnpm exec supabase login
pnpm exec supabase link
pnpm supabase:db:push:remote   # push migrations (use with caution)
```

## Security

- Do not commit `supabase/.temp/`, `supabase/volumes/`, access tokens, or service keys.
- Use the Supabase dashboard and GitHub repository secrets for remote credentials.
