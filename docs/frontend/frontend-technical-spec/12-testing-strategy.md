# 12. Frontend Testing Strategy

## 1. Mục tiêu

Kiểm thử ưu tiên rủi ro, không chạy theo coverage phần trăm đơn thuần. Ưu tiên cao nhất:

1. Exam autosave/resume/submit.
2. Authentication refresh/logout.
3. Permission-aware routing.
4. Assessment builder mapping/version conflicts.
5. Gradebook edit/publish.
6. File upload state.

## 2. Test pyramid

```text
Few E2E critical journeys
    /
Component + API integration tests
  /
Many pure unit tests
```

## 3. Unit tests — Vitest

Test:

- Permission functions.
- Route path builders.
- Problem Details mapping.
- Server clock offset.
- Exam operation coalescing/backoff.
- Zod schemas and mappers.
- Decimal/date formatters.
- Safe redirect validation.

Không test implementation detail như internal `useState`.

## 4. Component tests — Testing Library

Nguyên tắc:

- Query theo role/label/text trước test ID.
- Tương tác bằng `userEvent`.
- Assert điều người dùng thấy/làm được.
- Không shallow render.

Ví dụ:

```ts
render(<LoginForm onSubmit={submitSpy} />);
await user.type(screen.getByLabelText(/tên đăng nhập/i), 'hs001');
await user.type(screen.getByLabelText(/mật khẩu/i), 'secret');
await user.click(screen.getByRole('button', { name: /đăng nhập/i }));
expect(submitSpy).toHaveBeenCalled();
```

## 5. API integration tests — MSW

MSW handlers mô phỏng network contract:

- Success.
- Loading/delay.
- 401 refresh success/failure.
- 403.
- 422 field errors.
- 409/412 conflicts.
- 429 Retry-After.
- 500/network failure.

Handlers dùng response fixtures type-compatible với generated OpenAPI types.

## 6. E2E — Playwright

### Browser matrix

- Chromium mỗi PR.
- WebKit và Firefox trên main/nightly hoặc trước release.
- Mobile viewport cho student flows.

### Critical journeys

| Journey | Steps |
|---|---|
| Student login | Login -> dashboard -> class |
| Student exam | Start -> answer -> reload -> resume -> submit |
| Offline exam | Answer -> offline -> answer -> online -> sync |
| Teacher question | Create -> validate -> publish version |
| Teacher assessment | Build -> schedule -> publish |
| Teacher grading | Open result -> manual grade -> finalize |
| Assignment | Student submit -> teacher feedback -> student view |
| Gradebook | Edit -> conflict/validation -> publish |
| Admin import | Upload -> dry-run -> status/result |

## 7. Auth test strategy

- Playwright storage state có thể dùng cho non-auth-specific tests.
- Auth flow tests vẫn login thật qua API/UI.
- Không commit real credential.
- Mỗi worker có isolated test user/organization hoặc reset database.

## 8. Exam test strategy

### Pure state machine

- Operation enqueue/coalesce.
- Revision monotonicity.
- Backoff scheduling.
- Submit intent state.
- Terminal state transitions.

### IndexedDB integration

Dùng real IndexedDB implementation trong browser test; Node unit có thể dùng fake adapter theo interface, nhưng ít nhất một browser suite phải dùng thật.

### Network scenarios

Playwright route/network control:

- Abort save request.
- Delay response.
- Duplicate response.
- Return conflict.
- Expire token.
- Reload giữa pending save.

## 9. Accessibility tests

- axe on login, dashboard, question editor, exam page, gradebook.
- Keyboard E2E cho modal, navigation, exam choices.
- Snapshot/accessibility tree chỉ dùng chọn lọc.

## 10. Visual regression

Chỉ dùng cho:

- App shell desktop/mobile.
- Exam layout.
- Shared form primitives.

Không screenshot mọi dashboard data state vì dễ flaky. Disable animation và cố định timezone/locale/clock.

## 11. Test data factories

```text
src/test/factories/
├── actor.ts
├── class.ts
├── question.ts
├── assessment.ts
├── attempt.ts
└── problem-details.ts
```

Factory tạo dữ liệu tối thiểu hợp lệ và cho override. Không dùng một JSON fixture khổng lồ cho mọi test.

## 12. Time control

- Unit dùng fake timers/clock adapter.
- E2E ưu tiên backend test clock hoặc seed expiry đủ dài.
- Không phụ thuộc thời gian local máy CI.
- Set timezone/locale rõ trong test config.

## 13. Coverage policy

Không đặt một con số duy nhất cho toàn app. Mục tiêu:

- Core pure logic: branch coverage cao, gần đầy đủ.
- Exam sync/auth: bắt buộc có test lỗi và concurrency.
- Presentational component đơn giản: test chọn lọc.
- Generated code: exclude.

## 14. CI pipeline

```text
pnpm install --frozen-lockfile
-> typecheck
-> lint
-> unit/component tests
-> build
-> OpenAPI generated diff check
-> Playwright Chromium critical suite
```

Main/release:

```text
+ WebKit/Firefox
+ accessibility scan
+ bundle budget
+ dependency audit
```

## 15. Flaky test policy

- Không retry vô hạn.
- Playwright retry tối đa 1 trên CI để thu trace, nhưng flaky vẫn là bug.
- Dùng role locators và auto-wait.
- Không `waitForTimeout` trừ test timer có lý do rõ.
- Quarantine phải có issue và thời hạn.

## 16. Test Definition of Done

- Happy path.
- Loading/empty/error.
- 401/403 nếu protected.
- Validation/server error nếu có form.
- Keyboard behavior.
- Cache invalidation.
- Exam/grade critical edge cases theo feature.
