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

## 4. Phase 2 — Classes & resources (Tuần 5–7) 🟡 Mostly shipped

- Class list/detail ✅
- Student list teacher/admin ✅
- Academic terms/subjects/courses/classes admin CRUD ✅
- Resource folder/list ✅ (org-scoped + class-scoped via `?context_type=class&context_id=`)
- Direct upload + progress/cancel ✅ (multi-file XHR with concurrency 3, per-file progress bars, live region)
- Signed download/preview ✅ (bearer-auth download with sanitized `Content-Disposition`, `X-Content-Type-Options: nosniff`, content-type allowlist; inline preview modal for image / pdf / text)
- URL filters ✅

Exit:

- Permission states ✅ (class-scope authz: teacher/admin manages, enrolled student views; non-enrolled denied)
- File failure states ✅ (create/upload error surfaces with friendly message + `request_id`)
- E2E teacher upload/student view ✅ (`pnpm e2e:smoke` covers create + upload + publish + student list + download + multi-file + inline preview + class-scope ownership; `pnpm e2e:browser` chromium covers the `/app/resources` UI)

## 5. Phase 3 — Question bank (Tuần 8–11) 🟡 Minimal + non-MCQ shipped

- Question list/filter ✅ (basic picker for assessment builder; tenant-scoped, published-version-only).
- Create/edit new version — ✅ (auto-creates an initial version on question create; explicit publish-version endpoint before items can reference it).
- Question types — ✅ (MCQ + short_answer + essay; per-type grading dispatch with PENDING_REVIEW for essay; result review surfaces per-type UI).
- TipTap/KaTeX rich editor — 🟡 **Spike GO with caveats** (branch `spike/rich-text-editor`; production rollout deferred). Bundle impact: initial +0.04 kB gz, question-banks route +12.81 kB gz by lazy-splitting TipTap and KaTeX; opt-in cost ≈ 184 kB gz (editor + KaTeX chunks, loaded only when a teacher opts into rich mode and inserts a math formula). 20 new sanitization unit tests; 77/77 web tests; e2e smoke and chromium 23/23 green on the branch. KaTeX-on-Safari unverified on this host (WebKit system libs missing). Follow-up slice required for production: typed `prompt_doc` column, renderer in exam/review/grading, replace `window.prompt` link UI, WebKit-capable CI. Full report: `docs/frontend/frontend-technical-spec/spikes/rich-text-editor-spike.md`.
- Preview ✅ (within builder/assessment preview; per-type renderer in exam runner + attempt review).
- Version history/status — not started (status badge is rendered but full history timeline deferred).

Exit:

- API/form mapper tests — not started.
- Rich text sanitized preview — not started.
- Keyboard accessible choice editor — ✅ (radio groups with arrow-key navigation; focus-visible; ARIA labels on choice keys).

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

## 7. Phase 5 — Exam runtime (Tuần 15–19) 🟡 Core + heartbeat/deadline shipped

- Attempt start ✅
- Exam layout/question renderers ✅
- IndexedDB schema/repository ✅ (MVP: per attempt/item drafts)
- Durable autosave queue ✅ (MVP: local draft before API, retry on online/load)
- Server clock/timer ✅
- Heartbeat / deadline calibration ✅ (`AnswerSaved` returns `server_time` + `expires_at`; existing `GET /attempts/{id}` reused as 60s heartbeat; `serverTimeOffsetRef` keeps countdown in sync with the authoritative server clock; status refresh visible without reload).
- Deadline warning UX ✅ (≤5 min warning banner, ≤1 min critical banner + recolored timer chip; threshold transitions only, no every-tick `aria-live`).
- Offline/reload/resume ✅ (MVP: overlay local pending drafts, survive reload)
- Submit intent/idempotency ✅
- Terminal result/status ✅

Exit:

- Full test matrix critical scenarios — Playwright critical flow covers basic path + reload persistence.
- No answer loss in forced reload/offline tests — MVP covered; advanced conflict resolution deferred.
- Active exam not force-reloaded on app update — not started (service worker deferred; still gated on milestone risk gate `PWA service worker` requiring active-exam update safety proof).

## 8. Phase 6 — Grading, assignments, gradebook (Tuần 20–23) 🟡 Gradebook + manual review implemented

- Assignment create/list/detail — not started (not required by current pilot scope; assignments surface via `/me/assessments` instead).
- Student submission/upload — not started (essay/short_answer are text-based; no file submission in pilot).
- Teacher review/feedback — ✅ (manual review queue + detail page at `/app/grading`; rubric-style `awarded_score` + `feedback` per item; re-grade allowed).
- Manual assessment review — ✅ (`/assessments/{id}/review-queue` + `/attempts/{id}/review` + `PUT /attempts/{id}/items/{id}/grade`; non-MCQ items graded individually; `RecomputeAttemptScore` CTE promotes attempt to `GRADED` once all essay/SA items are scored; audit `attempt.grade` log entry per save with before/after JSON).
- Gradebook grid ✅ (per-class grid + per-assessment attempts; CSV export).
- Grade edit/override/publish — ✅ (re-grade endpoint; audit trail; student-facing result only when attempt is `GRADED`).
- CSV/export ✅ (assessment-attempts and class-gradebook; migrated to typed `openapi-fetch` in apiClient cleanup v1).

Exit:

- Decimal strings preserved ✅
- Grade conflicts/audit reason UX — ✅ (audit log rendered in admin dashboard with `request_id` link; per-grade before/after JSON captured).
- Student only sees published grade — ✅ (student result endpoint returns `awarded_score` + `feedback` only after the attempt reaches `GRADED`; `PENDING_REVIEW` attempts return `score: null` and per-item `is_correct: null` for essay/SA).

## 9. Phase 7 — Dashboard & notifications (Tuần 24–25) ✅ Dashboards + notifications implemented

- Student actionable dashboard ✅
- Teacher actionable dashboard ✅
- Admin summary ✅
- Notification inbox/unread — ✅ (`NotificationBell` in app shell, 30s polling, accessible dropdown, mark-read per item; inbox feeds `attempt.graded` / `assessment.published` / `resource.published` events with best-effort semantics so notification failures never roll back business transactions).
- Basic line/bar progress charts only if API ready — not started.

Exit:

- Dashboard useful without hero/AI/gamification ✅
- Charts lazy-loaded and accessible alternative — not started.

## 10. Phase 8 — Hardening & pilot (Tuần 26–28) 🟡 Mostly shipped; pilot still in progress

- Cross-browser E2E — ✅ Chromium default + opt-in `firefox`/`webkit` matrix via `PLAYWRIGHT_BROWSERS=1` + `pnpm e2e:browser:all`; WebKit host-dep probe gracefully falls back to chromium + firefox on hosts missing libgtk-4 / libgraphene-1.0 / libxslt / libevent-2.1 / libopus / libgstallocators.
- A11y audit — baseline ✅ (focus-visible, ARIA labels, error/`request_id` on states, keyboard accessible critical flows, accessible `NotificationBell`); **axe-core CI gate ✅** (`pnpm e2e:a11y` — `@axe-core/playwright` scanning 8 stable routes with WCAG 2.0/2.1 A+AA + best-practice, `color-contrast` + `target-size` intentionally disabled, intentionally NOT in `pnpm check`); full manual WCAG 2.1 AA audit + focus-management regression suite still pending.
- Bundle/performance audit — ✅ (16 route pages + 3 dashboard panels `React.lazy` via `SuspenseRoute`; initial chunk 511.19 kB / 147.04 kB gz → 335.39 kB / 106.56 kB gz; production sourcemap `hidden` so `.map` files stay on disk for stack-trace decoding without leaking sourceMappingURL in the public bundle). Full bundle analyzer dependency, virtualized lists, CSS splitting still pending.
- Error telemetry — pending.
- Error pages with `request_id` ✅ (toàn bộ `ErrorState` + error pages show `request_id`; copy-to-clipboard included).
- PWA manifest ✅ (`/manifest.json` + 2 SVG icons in `/icons/`, theme `#1d4ed8`, `start_url`/`scope`/`id` `/app`, `display: standalone`, vi); service worker intentionally deferred until active-exam update safety is proven — see milestone gate below.
- apiClient legacy cleanup — ✅ v1 shipped (gradebook CSV exports migrated to typed `openapi-fetch`; `api-client.ts` retains only the auth-provider critical calls; `auth-provider.tsx` migration deferred to a follow-up slice).
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
- Full installable PWA/push (manifest-only ship ở Phase 8; service worker + push + background sync vẫn post-pilot vì milestone gate yêu cầu chứng minh active-exam update safety trước).

## 12. Current next backlog (pre-pilot)

> **Thứ tự bắt buộc**: (1) hoàn tất cập nhật docs/ADRs còn stale, (2) backend chạy Huma feasibility spike và quyết định go/no-go, (3) mới triển khai feature frontend mới. Mục tiêu là khóa "docs-completion before new feature work".

**A. Docs & ADR completion (ưu tiên cao nhất)**

- Theo dõi quyết định Huma từ backend spike trước khi cam kết client generation tự động.
- Refresh router/ADR với Koyeb artifacts đã legacy (Render là deployment target hiện tại — đã ghi trong `AGENTS.md`).

**B. Phụ thuộc backend (Huma feasibility spike)**

- Huma runtime migration hiện vẫn tạm hoãn. Frontend tiếp tục dùng `openapi-typescript` + `openapi-fetch` từ skeleton thủ công cho đến khi có quyết định go/no-go.
- River background-job runtime chưa wire; in-process scheduler tạm đáp ứng assessment open/close transitions.

**C. Feature work (chỉ bắt đầu sau khi A & B xong)**

- Question bank: rich text / TipTap + KaTeX production rollout (spike complete on `spike/rich-text-editor` — see `docs/frontend/frontend-technical-spec/spikes/rich-text-editor-spike.md`; rollout needs typed `prompt_doc` column + renderer migration in exam/review/grading + replace `window.prompt` link UI + WebKit CI), version history UI, per-type question form polish.
- Manual grading UI nâng cấp: rubric editor, teacher feedback templates, file submission review.
- Auth-provider apiClient migration (5 call site còn lại trong `auth-provider.tsx`; auth flow cần adapter đặc biệt cho refresh rotation/queue — deferred).
- Full WCAG 2.1 AA audit, focus management regression suite cho builder/exam/admin; axe-core CI gate đã ship (xem Phase 8).
- Performance hardening: virtualized lists cho long gradebook/inbox, CSS splitting, bundle analyzer dependency.
- Service Worker (sau khi proven safe cho active-exam update flow): cache shell, app-layer IndexedDB draft + SW handoff, push notifications, background sync.
- Dark mode, parent portal, attendance, QTI import/export UI, AI assistant, gamification, advanced analytics — backlog sau pilot (xem §11).

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
