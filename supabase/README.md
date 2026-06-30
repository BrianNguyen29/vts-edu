# Supabase

Baseline SQL migrations for the MVP demo.

Apply order:

1. `000001_extensions.sql` — UUID, updated_at trigger.
2. `000002_identity.sql` — organizations, users, memberships, login names, refresh sessions.
3. `000003_demo_assessments.sql` — minimal assessment/attempt/answer/idempotency/audit tables.

These migrations are intentionally minimal to support the scaffold. The full domain model will expand per `backend/backend-technical-spec/05-database-design.md`.

## Applying migrations

### Local

```bash
psql $DATABASE_URL -f supabase/migrations/000001_extensions.sql
psql $DATABASE_URL -f supabase/migrations/000002_identity.sql
psql $DATABASE_URL -f supabase/migrations/000003_demo_assessments.sql
```

### Supabase

Use the Supabase SQL Editor or CLI:

```bash
supabase db reset   # local
supabase db push    # remote (use with caution)
```

## Notes

- All tenant tables include `organization_id`.
- `refresh_sessions` stores membership/organization context for multi-tenant auth.
- `attempt_answers` uses optimistic revision checks.
- `idempotency_keys` prevents duplicate writes.
