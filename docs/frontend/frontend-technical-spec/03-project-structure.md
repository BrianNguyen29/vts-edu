# 03. Project Structure & Folder Tree

## 1. Monorepo tổng thể

```text
lms-platform/
├── AGENTS.md
├── README.md
├── package.json
├── pnpm-lock.yaml
├── pnpm-workspace.yaml
├── compose.yaml
├── Makefile
├── .env.example
├── eslint.config.js
├── prettier.config.mjs
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── e2e.yml
│       └── release.yml
│
├── apps/
│   ├── web/                         # React + Vite SPA
│   │   ├── public/
│   │   ├── src/
│   │   ├── tests/
│   │   ├── index.html
│   │   ├── vite.config.ts
│   │   ├── vitest.config.ts
│   │   ├── playwright.config.ts
│   │   ├── tsconfig.json
│   │   └── package.json
│   │
│   └── api/                         # Go backend
│
├── packages/
│   ├── api-client/                  # Generated OpenAPI types + thin client helpers
│   │   ├── src/
│   │   │   ├── generated/
│   │   │   │   └── schema.d.ts
│   │   │   ├── client.ts
│   │   │   └── index.ts
│   │   └── package.json
│   └── config/                      # Shared TS/ESLint config nếu cần
│
└── docs/
    ├── backend-technical-spec/
    └── frontend-technical-spec/
```

`packages/ui` chưa tạo trong MVP vì chỉ có một web app. Khi có admin app độc lập hoặc mobile web khác, mới trích `shared/ui` thành package.

## 2. Cấu trúc `apps/web/src`

```text
src/
├── main.tsx
├── vite-env.d.ts
│
├── app/
│   ├── bootstrap/
│   │   ├── bootstrap-app.ts
│   │   ├── load-runtime-config.ts
│   │   └── bootstrap-error-screen.tsx
│   ├── providers/
│   │   ├── app-providers.tsx
│   │   ├── query-provider.tsx
│   │   ├── auth-provider.tsx
│   │   ├── i18n-provider.tsx
│   │   └── theme-provider.tsx
│   ├── router/
│   │   ├── router.tsx
│   │   ├── route-guards.ts
│   │   ├── route-meta.ts
│   │   └── paths.ts
│   ├── layouts/
│   │   ├── public-layout.tsx
│   │   ├── auth-layout.tsx
│   │   ├── app-shell-layout.tsx
│   │   ├── role-workspace-layout.tsx
│   │   ├── exam-layout.tsx
│   │   └── error-layout.tsx
│   ├── styles/
│   │   ├── index.css
│   │   ├── tokens.css
│   │   └── print.css
│   └── index.ts
│
├── pages/
│   ├── public/
│   ├── auth/
│   ├── common/
│   ├── student/
│   ├── teacher/
│   ├── admin/
│   └── errors/
│
├── widgets/
│   ├── app-sidebar/
│   ├── app-header/
│   ├── notification-center/
│   ├── upcoming-work-panel/
│   ├── class-overview/
│   ├── question-editor/
│   ├── assessment-outline/
│   ├── gradebook-grid/
│   └── exam-navigation/
│
├── features/
│   ├── auth/
│   │   ├── login/
│   │   ├── logout/
│   │   ├── refresh-session/
│   │   └── change-password/
│   ├── users/
│   ├── classes/
│   ├── resources/
│   ├── questions/
│   ├── assessments/
│   ├── attempts/
│   ├── assignments/
│   ├── gradebook/
│   └── notifications/
│
├── entities/
│   ├── user/
│   ├── class/
│   ├── resource/
│   ├── question/
│   ├── assessment/
│   ├── attempt/
│   ├── assignment/
│   ├── grade/
│   └── notification/
│
├── shared/
│   ├── api/
│   │   ├── api-client.ts
│   │   ├── auth-middleware.ts
│   │   ├── error-mapper.ts
│   │   ├── idempotency.ts
│   │   └── request-id.ts
│   ├── auth/
│   │   ├── auth-session-store.ts
│   │   ├── permissions.ts
│   │   └── broadcast.ts
│   ├── config/
│   │   ├── runtime-config.ts
│   │   └── feature-flags.ts
│   ├── db/
│   │   ├── indexed-db.ts
│   │   └── exam-draft-repository.ts
│   ├── errors/
│   │   ├── app-error.ts
│   │   ├── problem-details.ts
│   │   └── error-boundary.tsx
│   ├── hooks/
│   ├── i18n/
│   │   ├── i18n.ts
│   │   ├── resources.ts
│   │   └── locales/
│   │       ├── vi/
│   │       └── en/
│   ├── lib/
│   │   ├── cn.ts
│   │   ├── invariant.ts
│   │   ├── result.ts
│   │   ├── clock.ts
│   │   └── browser.ts
│   ├── logging/
│   ├── time/
│   ├── decimal/
│   ├── file/
│   ├── ui/
│   │   ├── button.tsx
│   │   ├── dialog.tsx
│   │   ├── form.tsx
│   │   ├── data-table/
│   │   ├── page-state/
│   │   └── ...
│   └── validation/
│
└── test/
    ├── setup.ts
    ├── msw/
    │   ├── server.ts
    │   ├── browser.ts
    │   └── handlers/
    ├── fixtures/
    ├── factories/
    └── render.tsx
```

## 3. Route module structure

```text
pages/teacher/assessment-edit/
├── page.tsx
├── route.ts
├── loader.ts
├── error.tsx
├── skeleton.tsx
├── page.test.tsx
└── index.ts
```

- `route.ts`: route object, handle metadata và lazy export.
- `loader.ts`: permission/precondition hoặc `ensureQueryData` tối thiểu.
- `page.tsx`: composition.
- `error.tsx`: route-specific error.
- `skeleton.tsx`: loading fallback.

## 4. Feature slice structure

```text
features/attempts/save-answer/
├── api/
│   ├── mutation-options.ts
│   └── contracts.ts
├── model/
│   ├── use-save-answer.ts
│   ├── answer-sync-machine.ts
│   └── types.ts
├── ui/
│   ├── save-status.tsx
│   └── retry-save-button.tsx
├── lib/
│   ├── canonicalize-answer.ts
│   └── create-client-mutation-id.ts
├── __tests__/
│   ├── answer-sync-machine.test.ts
│   └── use-save-answer.test.tsx
└── index.ts
```

## 5. Entity slice structure

```text
entities/question/
├── api/
│   └── query-options.ts
├── model/
│   ├── question-view-model.ts
│   └── question-type.ts
├── ui/
│   ├── question-card.tsx
│   ├── question-type-badge.tsx
│   └── question-preview.tsx
├── lib/
│   └── format-question-type.ts
└── index.ts
```

Entity không sở hữu workflow tạo/publish; workflow thuộc feature.

## 6. Shared UI structure

Không chia `atoms`, `molecules`, `organisms`. Dùng tên theo khả năng:

```text
shared/ui/
├── button.tsx
├── input.tsx
├── dialog.tsx
├── confirm-dialog.tsx
├── data-table/
├── date-time-field/
├── file-dropzone/
├── form-field/
├── page-header/
├── page-state/
└── permission-boundary/
```

## 7. Naming conventions

| Loại | Quy ước | Ví dụ |
|---|---|---|
| Folder/file | kebab-case | `assessment-builder` |
| React component | PascalCase | `AssessmentBuilder` |
| Hook | `use` + camelCase | `useAssessmentDetail` |
| Query option factory | noun + `QueryOptions` | `classDetailQueryOptions` |
| Mutation hook | `use` + action | `usePublishAssessment` |
| Route path constant | uppercase nested object | `PATHS.TEACHER.ASSESSMENTS` |
| Permission | backend string nguyên bản | `assessment:publish` |
| Test | `.test.ts(x)` | `login-form.test.tsx` |
| E2E | `.spec.ts` | `student-exam.spec.ts` |
| i18n key | semantic namespace | `assessment.publish.confirmTitle` |

## 8. Import aliases

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@api-client/*": ["../../packages/api-client/src/*"]
    }
  }
}
```

Ưu tiên relative import trong cùng slice; alias cho cross-layer imports.

## 9. Barrel export policy

- Mỗi slice có `index.ts` làm public API.
- Không tạo root barrel export tất cả feature/entity.
- Shared UI có thể export theo file cụ thể, không `export *` toàn bộ.
- Generated API chỉ export entrypoints ổn định.

## 10. Generated files

Không sửa tay:

```text
packages/api-client/src/generated/**
apps/web/src/shared/i18n/generated/**   # nếu dùng i18n CLI
```

Nguồn sinh:

```text
Go API -> OpenAPI 3.1 -> openapi-typescript -> api-client generated types
locale files -> i18next tooling -> generated key types (nếu bật)
```

## 11. Test placement

- Test unit gần source nếu chỉ liên quan một module.
- Shared test setup ở `src/test`.
- Playwright E2E ở `apps/web/tests/e2e`.
- Không tạo `__mocks__` global cho API; dùng MSW handlers.
