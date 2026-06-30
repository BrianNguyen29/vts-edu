# Config Environment Files

This directory contains example environment files for the MVP demo stack.

## Files

- `vercel.env.example` — Variables used by the Vercel build / frontend SPA.
- `render.env.example` — Variables injected into the Render Go API service.
- `supabase.env.example` — Variables pointing at the Supabase project.

## CLI setup

The project uses `vercel` and `supabase` CLIs installed as root devDependencies (see root `package.json`).

```bash
# Install everything (pnpm-lock.yaml is committed; CI uses --frozen-lockfile)
pnpm install

# Verify CLIs
pnpm vercel:version
pnpm supabase:version
```

### Vercel

1. `pnpm vercel:pull` — pull environment variables into `apps/web/.env.local`.
2. `pnpm vercel:build` — build the frontend locally.
3. `pnpm vercel:deploy:preview` — deploy a preview (requires `vercel login`).

Do not commit `.vercel/`, `.env.local`, or any token files.

### Supabase

1. `pnpm supabase:start` — start local Supabase stack.
2. `pnpm supabase:db:reset` — reset local DB and apply migrations.
3. `pnpm supabase:db:push:remote` — push migrations to linked project (requires `supabase login` and `supabase link`).

Local Supabase volumes and `.temp/` are ignored by `.gitignore`.

## No secrets in repo

All committed files are examples with placeholder values. Real secrets live only in:

- Vercel project dashboard
- Render Web Service environment variables
- Supabase dashboard
- GitHub repository secrets (for Actions)
- Your local `.env` / `.env.local` files (ignored by git)
