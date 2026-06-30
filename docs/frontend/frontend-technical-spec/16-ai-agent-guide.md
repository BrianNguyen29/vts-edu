# 16. AI Coding Agent Guide

## 1. Mission

Triển khai frontend React/Vite cho LMS và online assessment với ưu tiên:

1. Exam reliability.
2. API contract correctness.
3. Accessibility.
4. Security/privacy.
5. Maintainability cho một developer solo.
6. Performance trên thiết bị phổ thông.

## 2. Read first

Trước khi sửa code, đọc:

- `docs/frontend-technical-spec/README.md`
- `docs/frontend-technical-spec/00-project-scope.md`
- File feature/route liên quan.
- ADR liên quan.
- Backend API specification tương ứng.
- Root `AGENTS.md` và frontend `AGENTS.md`.

## 3. Hard constraints

1. Không lưu access/refresh token trong localStorage/sessionStorage/IndexedDB.
2. Không tự định nghĩa API DTO nếu OpenAPI đã có.
3. Không sửa generated API files.
4. Không gọi raw fetch trong page/shared UI; dùng API layer/feature hooks.
5. Không copy server state vào global store.
6. Không tính final grade phía client.
7. Exam answer pending phải persist IndexedDB trước khi coi là local-safe.
8. Server time kiểm soát deadline.
9. Permission UI không thay backend authorization.
10. Không thêm Redux/Zustand/Axios/dependency lớn nếu chưa có ADR.
11. Không log token, answer, essay, grade detail hoặc PII.
12. Mọi feature phải có loading/empty/error/forbidden states.
13. Mọi interaction chính phải dùng keyboard.

## 4. Expected commands

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

Tên script thực tế có thể khác; inspect root/package files.

## 5. Before modifying code

Agent phải đưa plan ngắn:

- Routes/pages affected.
- Feature/entity/shared slices affected.
- API operations used.
- State owner: query/form/local/URL/IndexedDB.
- Permission requirements.
- Error states.
- Tests to add/update.
- Accessibility considerations.

## 6. Implementation procedure

```text
read specs
-> inspect existing public APIs
-> inspect generated OpenAPI types
-> write/update tests for risky logic
-> implement smallest coherent slice
-> run typecheck/lint/tests
-> review bundle/dependency impact
-> update docs
```

## 7. Component rules

- Prefer small composable components.
- Shared UI is domain-agnostic.
- Feature UI owns actions/use cases.
- Entity UI presents domain data.
- Page composes; does not contain raw transport logic.
- Use semantic HTML before ARIA.
- Avoid prop drilling only when real; do not add context prematurely.

## 8. API rules

- Use operation/path generated types.
- Map Problem Details to `AppError`.
- Pass AbortSignal.
- Do not auto-retry non-idempotent writes.
- Stable idempotency key across retry.
- Explicit invalidation/update after mutation.

## 9. Form rules

- Zod schema + RHF.
- Server field errors mapped.
- Dirty navigation handled.
- No silent data discard.
- Do not put translated text inside schema if schema reused outside UI; map message keys where appropriate.

## 10. Exam-specific rules

Any change touching exam runtime must include:

- State transition analysis.
- IndexedDB migration impact.
- Offline/reload behavior.
- Revision/idempotency behavior.
- Timer/server clock behavior.
- E2E scenario.

Never simplify exam sync to a debounced fetch without durable queue.

## 11. Generated code

Do not edit:

```text
packages/api-client/src/generated/**
apps/web/src/shared/i18n/generated/**
```

If type mismatch:

1. Check backend OpenAPI.
2. Regenerate.
3. Fix source contract or mapper.
4. Do not cast to `any`.

## 12. Dependency changes

Agent must state:

- Why existing tools/platform cannot solve it.
- Package size/runtime effect.
- Security/install scripts.
- Alternative considered.
- Removal/migration path.

## 13. Definition of Done

- Typecheck pass.
- ESLint/format pass.
- Tests pass.
- OpenAPI generation clean.
- No new a11y critical issue.
- Responsive smoke checked.
- Error/loading/empty states implemented.
- Docs updated.
- No sensitive logging/storage.
