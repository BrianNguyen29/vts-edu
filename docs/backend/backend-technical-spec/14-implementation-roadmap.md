# 14. Solo Implementation Roadmap

## 1. Guiding strategy

- Làm vertical slice hoàn chỉnh, không xây mọi infrastructure trước.
- Mỗi phase phải có deployable product.
- Bài thi runtime được proof-of-concept sớm vì rủi ro cao nhất.
- Dashboard/AI/gamification làm sau data integrity.

## 2. Phase plan

### Phase 0 — Foundation & proof of concept (2–3 tuần)

- Repository/pnpm workspace.
- Go app skeleton, chi, Huma.
- PostgreSQL, goose, sqlc.
- Structured errors/logging/config.
- CI.
- POC transaction + River.
- POC attempt autosave revision và concurrent submit.

Exit criteria:

- Duplicate submit không tạo job trùng.
- sqlc/OpenAPI generation chạy trong CI.

### Phase 1 — Auth, users, tenancy (3–4 tuần)

- Organization.
- User/membership/roles.
- Login, JWT, refresh rotation.
- Admin user CRUD.
- Audit cơ bản.

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

### Phase 4 — Question bank (3–5 tuần)

- Bank/question/version.
- 6 MVP types.
- Validation/publish.
- Search/filter.

### Phase 5 — Assessment builder (3–4 tuần)

- Assessment/sections/items.
- Settings/targets/accommodation.
- Validate/publish snapshots.

### Phase 6 — Attempt runtime (4–6 tuần)

- Start/resume.
- Item selection/shuffle.
- Save answer/revision.
- Heartbeat/deadline.
- Submit/expire.
- Auto-grade/manual review.
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
