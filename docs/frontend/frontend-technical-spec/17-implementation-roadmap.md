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

## 2. Phase 0 — Contract & UX foundation (Tuần 1–2) ✅ Implemented

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

## 3. Phase 1 — Authentication & app shell (Tuần 3–4) ✅ Implemented

- Login.
- Refresh bootstrap.
- Logout/cross-tab.
- Change password (backend `POST /auth/change-password` ready).
- Permission route guard (backend returns `roles` + `permissions` + `must_change_password`).
- Sidebar/header responsive.
- Profile/session list.
- Global error mapping.
- Role-based redirects (`/app/student`, `/app/teacher`, `/app/admin`).

Exit:

- Token không persistent.
- 401 refresh single-flight test.
- 403/404 states ✅ (UI surfaces `request_id` qua `ErrorState` + `formatFriendlyError`).
- Mobile shell usable.
- Forced password change redirect works.

## 4. Phase 2 — Classes & resources (Tuần 5–7) 🟡 Partial

- Class list/detail ✅
- Student list teacher/admin ✅
- Academic terms/subjects/courses/classes admin CRUD ✅
- Resource folder/list ✅ (MVP: org-scoped; class-scope deferred)
- Direct upload + progress/cancel — partial (upload via `FormData`; no progress bar yet)
- Signed download/preview — partial (bearer-auth download via `fetch`; no inline preview)
- URL filters ✅

Exit:

- Permission states ✅
- File failure states ✅ (create/upload error surfaces with friendly message + `request_id`)
- E2E teacher upload/student view — partial (smoke API covered; Playwright coverage deferred)

## 5. Phase 3 — Question bank (Tuần 8–11) 🟡 Minimal

- Question list/filter ✅ (basic picker for assessment builder).
- Create/edit new version — not started.
- Six MVP question types — MCQ only; rest deferred.
- TipTap/KaTeX editor — not started.
- Preview ✅ (within builder/assessment preview).
- Version history/status — not started.

Exit:

- API/form mapper tests — not started.
- Rich text sanitized preview — not started.
- Keyboard accessible choice editor — partial.

## 6. Phase 4 — Assessment builder (Tuần 12–14) ✅ Implemented

- Create/edit draft.
- Sections/items/reorder.
- Duplicate section/item.
- Settings/schedule.
- Validation summary.
- Preview.
- Publish confirm/conflict.
- Teacher assessment list (`GET /assessments`) backend ready.

Exit:

- Dirty navigation guard — basic.
- Version conflict handled.
- Publish không optimistic.

## 7. Phase 5 — Exam runtime (Tuần 15–19) 🟡 Core implemented

- Attempt start ✅
- Exam layout/question renderers ✅
- IndexedDB schema/repository ✅ (MVP: per attempt/item drafts)
- Durable autosave queue ✅ (MVP: local draft before API, retry on online/load)
- Server clock/timer ✅
- Offline/reload/resume ✅ (MVP: overlay local pending drafts, survive reload)
- Submit intent/idempotency ✅
- Terminal result/status ✅

Exit:

- Full test matrix critical scenarios — Playwright critical flow covers basic path + reload persistence.
- No answer loss in forced reload/offline tests — MVP covered; advanced conflict resolution deferred.
- Active exam not force-reloaded on app update — not started (service worker deferred).

## 8. Phase 6 — Grading, assignments, gradebook (Tuần 20–23) 🟡 Gradebook implemented

- Assignment create/list/detail — not started.
- Student submission/upload — not started.
- Teacher review/feedback — not started (MCQ auto-graded).
- Manual assessment review — not started.
- Gradebook grid ✅
- Grade edit/override/publish — not started.
- CSV/export ✅

Exit:

- Decimal strings preserved ✅
- Grade conflicts/audit reason UX — not started.
- Student only sees published grade — not started.

## 9. Phase 7 — Dashboard & notifications (Tuần 24–25) ✅ Dashboards implemented

- Student actionable dashboard ✅
- Teacher actionable dashboard ✅
- Admin summary ✅
- Notification inbox/unread — not started.
- Basic line/bar progress charts only if API ready — not started.

Exit:

- Dashboard useful without hero/AI/gamification ✅
- Charts lazy-loaded and accessible alternative — not started.

## 10. Phase 8 — Hardening & pilot (Tuần 26–28) 🟡 Partial

- Cross-browser E2E — Chromium only; Firefox/WebKit deferred.
- A11y audit — baseline implemented (focus-visible, ARIA labels, error/request_id on states, keyboard accessible critical flows); full manual WCAG 2.1 AA audit pending.
- Bundle/performance audit — pending.
- Error telemetry — pending.
- Error pages with `request_id` ✅ (toàn bộ `ErrorState` + error pages show `request_id`; copy-to-clipboard included).
- PWA manifest; service worker only if update policy tested — pending.
- Pilot bug fixes — pending.

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

## 12. Current next backlog (pre-pilot)

> **Thứ tự bắt buộc**: (1) hoàn tất cập nhật docs/ADRs còn stale, (2) backend chạy Huma feasibility spike và quyết định go/no-go, (3) mới triển khai feature frontend mới. Mục tiêu là khóa "docs-completion before new feature work".

**A. Docs & ADR completion (ưu tiên cao nhất)**

- Cập nhật roadmap/ADR còn stale (`request_id` display, accessibility baseline, resources MVP UI, path count).
- Theo dõi quyết định Huma từ backend spike trước khi cam kết client generation tự động.

**B. Phụ thuộc backend (Huma feasibility spike)**

- Huma runtime migration hiện vẫn tạm hoãn. Frontend tiếp tục dùng `openapi-typescript` + `openapi-fetch` từ skeleton thủ công cho đến khi có quyết định go/no-go.

**C. Feature work (chỉ bắt đầu sau khi A & B xong)**

- Resources/files UI: progress bar, resumable upload, inline preview (PDF/image), class-scope, multi-file upload.
- Question bank: editor TipTap/KaTeX, non-MCQ types, version history.
- Manual grading UI: rubric editor, teacher feedback, file submission review.
- Performance, cross-browser (Firefox/WebKit), notifications, installable PWA, dark mode.
- Full WCAG 2.1 AA audit, axe-core CI gate, focus management regression suite cho builder/exam/admin.
- `apiClient` legacy cleanup sau khi tất cả helper đã migrate sang `openapi-fetch`.

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
