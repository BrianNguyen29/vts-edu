# AGENTS.md — Frontend Repository Instructions

## Mission

Implement a cost-efficient React/Vite frontend for an LMS and online assessment platform. Preserve exam data, API contract, accessibility and privacy above visual speed.

## Read first

- `docs/frontend/frontend-technical-spec/README.md`
- `docs/frontend/frontend-technical-spec/00-project-scope.md`
- Relevant route/feature specification.
- Relevant ADR.
- Matching backend API specification.

## Hard constraints

1. Access token is memory-only; refresh token is HttpOnly cookie.
2. Never edit generated OpenAPI client/types.
3. Never define duplicate API DTOs when generated types exist.
4. TanStack Query owns server state.
5. No Redux, Zustand, Axios or major dependency without accepted ADR.
6. Pages/shared UI never perform raw API calls.
7. Server time controls exam deadlines.
8. Exam pending answers are persisted in IndexedDB before local-safe status.
9. Submit/start/publish retries reuse the same idempotency key.
10. Scores remain decimal strings; frontend does not calculate final grade.
11. Permission UI is not a security boundary.
12. Never log/store tokens, answer content, essay text, grade detail or PII unnecessarily.
13. All critical interactions must be keyboard accessible.
14. Every data page needs loading, empty, error and forbidden handling.

## Expected commands

```bash
pnpm install
pnpm api:generate
pnpm web:dev
pnpm web:typecheck
pnpm web:lint
pnpm web:test
pnpm web:e2e
pnpm web:build
```

Inspect repository scripts before assuming exact names.

> **Note:** several commands above (`pnpm api:generate`, `pnpm web:lint`, `pnpm web:test`, `pnpm web:e2e`) are planned/spec-only and not wired yet. Current verification relies on `pnpm web:typecheck` and `pnpm web:build`.

## Before modifying code

Provide a short plan with:

- Routes and slices affected.
- API operations and generated types.
- State owner.
- Permissions.
- Error/loading/empty states.
- Accessibility behavior.
- Tests.

## Definition of Done

- Typecheck/lint/format pass.
- Unit/component/E2E tests appropriate to risk.
- Generated API diff clean.
- No critical a11y regression.
- Responsive smoke test.
- No sensitive storage/logging.
- Documentation updated.
