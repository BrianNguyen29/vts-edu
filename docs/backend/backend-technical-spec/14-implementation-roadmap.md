# 14. Solo Implementation Roadmap

## 1. Guiding strategy

- Làm vertical slice hoàn chỉnh, không xây mọi infrastructure trước.
- Mỗi phase phải có deployable product.
- Bài thi runtime được proof-of-concept sớm vì rủi ro cao nhất.
- Dashboard/AI/gamification làm sau data integrity.

## 2. Phase plan

### Phase 0 — Foundation & proof of concept (2–3 tuần) ✅ Implemented

- Repository/pnpm workspace.
- Go app skeleton, chi.
- PostgreSQL, sequential SQL migrations.
- Structured errors/logging/config.
- CI.
- TxManager + POC transaction boundaries.
- POC attempt autosave revision và concurrent submit.

Exit criteria:

- Duplicate submit không tạo job trùng.
- sqlc/OpenAPI groundwork staged behind Repository interfaces (see ADR-0010).

### Phase 1 — Auth, users, tenancy (3–4 tuần) ✅ Implemented

- Organization.
- User/membership/roles (`membership_roles`).
- Login, JWT, refresh rotation, CSRF.
- Forced password change (`/auth/change-password`).
- Admin user CRUD + org update.
- Audit schema baseline (chưa ghi log).

### Phase 2 — Academic structure (2–3 tuần)

- Terms, subjects, courses, classes.
- Teacher assignment.
- Enrollment/bulk import.
- Authorization class scope.

### Phase 3 — Resources/files (2–3 tuần)

- Upload intent/complete.
- File states.
- Resource CRUD/publish.
- Signed download.
- Basic processing job.

### Phase 4 — Question bank (3–5 tuần) — Minimal version implemented

- Bank/question/version schema ✅
- Snapshot prompt/choices/answer key into `attempt_items` ✅
- 6 MVP types.
- Validation/publish.
- Search/filter.

### Phase 5 — Assessment builder (3–4 tuần)

- Assessment/sections/items.
- Settings/targets/accommodation.
- Validate/publish snapshots.
- Teacher assessment list ✅

### Phase 6 — Attempt runtime (4–6 tuần) — Core implemented

- Start/resume.
- Item selection/shuffle.
- Save answer/revision ✅
- Heartbeat/deadline.
- Submit/expire ✅
- Auto-grade/manual review (MCQ auto-grade ✅)
- Load/concurrency tests.

### Phase 7 — Assignment & gradebook (4–5 tuần)

- Assignment/submission versions/files.
- Feedback/grade.
- Grade items/entries/history.
- Publish/export.

### Phase 8 — Hardening & pilot (3–5 tuần)

- Security negative tests.
- Load tests.
- Backup restore drill.
- Monitoring/alerts.
- Pilot data/import.
- Bug fixing.

## 2.5 Staged Huma/sqlc/OpenAPI plan

- **Current**: Hand-maintained OpenAPI skeleton covers the current API surface in `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`. Repository interfaces in each feature package are the migration seam.
- **Stage 1 — sqlc migration**: Introduce sqlc generated queries side-by-side with existing repository implementations; keep interfaces stable and migrate one feature at a time. No runtime handler/service rewrite.
- **Stage 2 — Huma migration**: Add Huma operation definitions behind existing handlers or incrementally replace handler wiring while preserving response envelopes. OpenAPI generation becomes automatic.
- **Stage 3 — Client generation**: Generate frontend API client/types from the Huma/OpenAPI contract once it stabilizes.

## 3. Effort estimate

| Mức | Thời gian tham khảo |
|---|---|
| Demo functional | 8–12 tuần |
| Pilot hẹp | 5–7 tháng full-time |
| MVP ổn định hơn | 7–10 tháng full-time |
| Part-time | 10–16 tháng |

Ước lượng thay đổi theo kinh nghiệm và độ hoàn thiện UI.

## 4. Cost-control rules

- Một managed PostgreSQL nhỏ trước.
- Một app container.
- Object storage pay-as-you-go.
- Không Redis.
- Không Kubernetes.
- Không separate observability stack lúc đầu; dùng provider logs + structured logs.
- Chỉ tách worker khi queue làm ảnh hưởng API.

## 5. Priority order

```text
Data integrity
> Authorization/security
> Exam reliability
> Grade correctness
> Teacher workflow
> Student UX
> Analytics
> Gamification
> AI
```
