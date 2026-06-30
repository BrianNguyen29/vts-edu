# 13. AI Coding Agent Guide

## 1. Mandatory reading order

Trước khi code, agent phải đọc:

1. `README.md`
2. `00-project-scope.md`
3. `02-system-architecture.md`
4. `03-folder-structure.md`
5. Tài liệu domain/API liên quan.
6. `09-security.md` nếu endpoint có auth/data nhạy cảm.
7. ADR liên quan.

## 2. Source of truth precedence

Khi tài liệu mâu thuẫn:

1. Non-negotiable invariants trong `00-project-scope.md`.
2. ADR đã accepted.
3. Domain/API spec.
4. Existing tests.
5. Existing code.

Agent không tự ý “đơn giản hóa” invariant để làm code chạy.

## 3. Rules for code generation

- Không tạo microservice.
- Không thêm Redis/Kafka/MongoDB/ORM nếu task không có ADR được duyệt.
- Không truy cập sqlc từ HTTP handler.
- Không bind request vào DB model.
- Không sửa generated files.
- Không dùng float cho score.
- Không dùng client time để quyết định deadline.
- Không tin organization ID từ body.
- Không trả raw internal error.
- Không log token/password/answer content.
- Không hard-delete historical academic records.

## 4. Feature implementation workflow

```text
1. Restate feature and invariants.
2. Identify module owner.
3. Identify DB changes.
4. Write/modify migration.
5. Write sqlc queries.
6. Generate sqlc code.
7. Implement domain/application logic.
8. Implement repository adapter.
9. Implement HTTP input/output/routes.
10. Add authorization.
11. Add audit/job side effects.
12. Add unit + integration + API tests.
13. Regenerate OpenAPI/client.
14. Run full checks.
```

## 5. Required response from agent before coding

Agent nên đưa plan ngắn:

```markdown
### Affected modules
- attempts
- grading

### Invariants
- submit idempotent
- server time authoritative

### Data changes
- new unique index ...

### Tests
- concurrent duplicate submit
- expired attempt
```

Nếu task có ambiguity làm thay đổi schema/invariant, agent phải dừng và hỏi thay vì tự đoán.

## 6. SQL rules

- Mọi tenant query có `organization_id`.
- Parameterized SQL.
- List query có limit/cursor.
- Index đề xuất phải giải thích query pattern.
- Không dùng `SELECT *` trong query production.
- Check rows affected cho update/delete.
- Lock chỉ khi cần và transaction ngắn.

## 7. Go rules

- `context.Context` là parameter đầu cho I/O functions.
- Wrap errors có context bằng `%w`.
- Không panic cho user input.
- Constructor validate required dependencies.
- Clock injectable cho time-dependent logic.
- Keep packages cohesive.
- Interfaces defined near consumer.
- Avoid global mutable state.

## 8. API rules

- Huma operation có stable operation ID.
- Request/response typed.
- Problem Details errors.
- Permission và resource scope explicit.
- Idempotency key cho operations đã liệt kê.
- Decimal trong JSON là string.
- Không leak existence cross-tenant.

## 9. Test rules

Mỗi bug fix phải có regression test.

Feature liên quan transaction/concurrency cần integration test PostgreSQL thật. Mock test không đủ.

Agent không được xóa/skip test chỉ để CI xanh.

## 10. Generated code commands

Tên script cuối cùng tùy repository, nhưng contract dự kiến:

```bash
pnpm api:sqlc-generate
pnpm api:openapi-generate
pnpm api:client-generate
pnpm api:test
pnpm test
```

Hoặc Makefile tương đương:

```bash
make generate
make test
make lint
```

## 11. Completion checklist

- [ ] Migration safe.
- [ ] Tenant scoping.
- [ ] Authorization.
- [ ] Validation.
- [ ] Error mapping.
- [ ] Audit.
- [ ] Idempotency/concurrency.
- [ ] Unit tests.
- [ ] Integration tests.
- [ ] OpenAPI/client regenerated.
- [ ] No secret/PII logging.
- [ ] Docs updated.

## 12. No-go examples

Sai:

```go
// handler directly queries DB
row, _ := h.Queries.GetAttempt(ctx, id)
```

Đúng:

```go
out, err := h.GetAttempt.Execute(ctx, actor, application.GetAttemptInput{AttemptID: id})
```

Sai:

```sql
SELECT * FROM attempts WHERE id = $1;
```

Đúng:

```sql
SELECT id, status, student_user_id, expires_at, version
FROM attempts
WHERE organization_id = $1 AND id = $2;
```
