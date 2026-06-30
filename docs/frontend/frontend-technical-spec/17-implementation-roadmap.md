# 17. Frontend Implementation Roadmap — Solo Project

## 1. Nguyên tắc thứ tự

Thứ tự ưu tiên:

```text
foundation
-> auth/session
-> class/resource basics
-> question bank
-> assessment builder
-> exam runtime
-> assignment/gradebook
-> dashboard polish
-> PWA/analytics
```

Không xây dashboard đẹp đầy đủ trước khi luồng thi và điểm hoạt động.

## 2. Phase 0 — Contract & UX foundation (Tuần 1–2)

### Deliverables

- Chốt route map.
- Chốt design tokens cơ bản.
- Vite/React/TS/pnpm workspace.
- OpenAPI generation package.
- App bootstrap/runtime config.
- Query/API/auth store skeleton.
- Public/Auth/App/Exam layouts skeleton.
- Test setup Vitest/MSW/Playwright.

### Exit criteria

- Build static thành công.
- Mock login -> protected route hoạt động.
- Generated API type không sửa tay.
- CI chạy typecheck/lint/test/build.

## 3. Phase 1 — Authentication & app shell (Tuần 3–4)

- Login.
- Refresh bootstrap.
- Logout/cross-tab.
- Change password.
- Permission route guard.
- Sidebar/header responsive.
- Profile/session list.
- Global error mapping.

Exit:

- Token không persistent.
- 401 refresh single-flight test.
- 403/404 states.
- Mobile shell usable.

## 4. Phase 2 — Classes & resources (Tuần 5–7)

- Class list/detail.
- Student list teacher/admin.
- Resource folder/list.
- Direct upload + progress/cancel.
- Signed download/preview.
- URL filters.

Exit:

- Permission states.
- File failure states.
- E2E teacher upload/student view.

## 5. Phase 3 — Question bank (Tuần 8–11)

- Question list/filter.
- Create/edit new version.
- Six MVP question types.
- TipTap/KaTeX editor.
- Preview.
- Version history/status.

Exit:

- API/form mapper tests.
- Rich text sanitized preview.
- Keyboard accessible choice editor.

## 6. Phase 4 — Assessment builder (Tuần 12–14)

- Create/edit draft.
- Sections/items/reorder.
- Settings/schedule.
- Validation summary.
- Preview.
- Publish confirm/conflict.

Exit:

- Dirty navigation guard.
- Version conflict handled.
- Publish không optimistic.

## 7. Phase 5 — Exam runtime (Tuần 15–19)

- Attempt start.
- Exam layout/question renderers.
- IndexedDB schema/repository.
- Durable autosave queue.
- Server clock/timer.
- Offline/reload/resume.
- Submit intent/idempotency.
- Terminal result/status.

Exit:

- Full test matrix critical scenarios.
- No answer loss in forced reload/offline tests.
- Active exam not force-reloaded on app update.

## 8. Phase 6 — Grading, assignments, gradebook (Tuần 20–23)

- Assignment create/list/detail.
- Student submission/upload.
- Teacher review/feedback.
- Manual assessment review.
- Gradebook grid.
- Grade edit/override/publish.
- CSV/export job status.

Exit:

- Decimal strings preserved.
- Grade conflicts/audit reason UX.
- Student only sees published grade.

## 9. Phase 7 — Dashboard & notifications (Tuần 24–25)

- Student actionable dashboard.
- Teacher actionable dashboard.
- Admin summary.
- Notification inbox/unread.
- Basic line/bar progress charts only if API ready.

Exit:

- Dashboard useful without hero/AI/gamification.
- Charts lazy-loaded and accessible alternative.

## 10. Phase 8 — Hardening & pilot (Tuần 26–28)

- Cross-browser E2E.
- A11y audit.
- Bundle/performance audit.
- Error telemetry.
- PWA manifest; service worker only if update policy tested.
- Pilot bug fixes.

## 11. Backlog after pilot

- Rubric advanced.
- Parent portal.
- Attendance.
- QTI import/export UI.
- AI assistant.
- Gamification.
- Advanced analytics.
- Dark mode.
- Full installable PWA/push.

## 12. Milestone risk gates

| Gate | Điều kiện không được bỏ qua |
|---|---|
| Auth complete | Refresh rotation/revoke behavior hiểu rõ |
| Builder complete | Immutable version/snapshot reflected |
| Exam start | IndexedDB + offline POC pass |
| Pilot | Load/security/a11y/E2E critical pass |
| PWA service worker | Active exam update safety pass |

## 13. Solo workload estimate

| Phạm vi | Thời gian full-time ước lượng |
|---|---:|
| Clickable prototype | 4–6 tuần |
| Functional core without hardened exam | 12–16 tuần |
| Pilot narrow | 24–28 tuần |
| Stable MVP | 7–10 tháng |

Ước lượng phụ thuộc mức hoàn thiện backend và thiết kế UI; không phải cam kết.
