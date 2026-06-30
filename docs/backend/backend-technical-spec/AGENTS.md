# AGENTS.md — Backend Repository Instructions

## Mission

Implement a cost-efficient Go modular monolith for an LMS and online assessment platform. Preserve data integrity, tenant isolation and exam reliability above feature speed.

## Read first

- `docs/backend/backend-technical-spec/README.md`
- `docs/backend/backend-technical-spec/00-project-scope.md`
- Relevant domain/API specification.
- Relevant ADR.

## Hard constraints

1. PostgreSQL is the source of truth.
2. No Redis, Kafka, MongoDB, microservices or ORM without an accepted ADR.
3. All tenant data access is scoped by `organization_id`.
4. HTTP handlers never call sqlc directly.
5. Question versions and published assessment snapshots are immutable.
6. Attempt submit is transactional and idempotent.
7. Server time controls exam deadlines.
8. Scores use decimal types/strings, never binary floating point.
9. Grade changes are audited and historical records are retained.
10. Never edit generated sqlc/OpenAPI/client files manually.

## Expected commands

```bash
pnpm install
pnpm api:dev
pnpm api:migrate
pnpm api:generate
pnpm api:test
pnpm lint
pnpm test
```

Actual scripts may differ; inspect root `package.json`.

> **Note:** several commands above (`pnpm api:migrate`, `pnpm api:generate`, `pnpm lint`, `pnpm test`, sqlc/Huma/OpenAPI/client generation) are planned/spec-only and not wired yet. Current verification relies on `go test ./...`, `go vet ./...`, and `gofmt`.

## Before modifying code

Provide a short plan listing:

- Modules affected.
- Invariants affected.
- Database changes.
- Authorization rules.
- Tests to add.

## Definition of done

- Code formatted and linted.
- Unit/integration tests added.
- Migration and sqlc generated.
- OpenAPI/client generated.
- Security and tenant isolation reviewed.
- Documentation updated.
