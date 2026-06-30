# Frontend Agent Task Template

## Task

[Mô tả feature/bug]

## Specs to read

- `docs/frontend-technical-spec/...`
- `docs/backend-technical-spec/...`
- Relevant ADRs.

## Affected routes

- `...`

## Affected slices

- `pages/...`
- `widgets/...`
- `features/...`
- `entities/...`
- `shared/...`

## API operations

| Operation | Method/path | Generated type |
|---|---|---|
| | | |

## State ownership

| State | Owner |
|---|---|
| Server data | TanStack Query |
| Form | RHF |
| URL filters | Router search params |
| Local UI | React local state |
| Durable pending | IndexedDB if applicable |

## Permissions

- Required: `...`
- Backend remains authoritative.

## Error states

- Loading.
- Empty.
- 401.
- 403.
- 404.
- 409/412.
- 422.
- Network/5xx.

## Accessibility

- Keyboard flow.
- Focus behavior.
- Labels/live regions.
- Mobile behavior.

## Tests

- Unit:
- Component/MSW:
- E2E:

## Definition of Done

- [ ] Typecheck/lint/tests pass.
- [ ] API generated contract used.
- [ ] No sensitive storage/logging.
- [ ] Docs updated.
