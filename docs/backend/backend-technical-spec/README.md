# Backend Technical Specification — LMS & Online Assessment Platform

**Phiên bản:** 1.0  
**Ngày chốt:** 2026-06-29  
**Đối tượng triển khai:** Dự án solo, tối ưu chi phí giai đoạn đầu, có khả năng scale theo tải thực tế  
**Backend:** Go 1.25.x  
**Frontend workspace:** pnpm + React/Vite (được mô tả để thống nhất contract, nhưng tài liệu này tập trung vào backend)

## 1. Mục đích

Thư mục này là nguồn đặc tả kỹ thuật chuẩn cho backend của nền tảng LMS, quản lý lớp học, học liệu, ngân hàng câu hỏi, bài kiểm tra trực tuyến, bài tập và bảng điểm.

Tài liệu được viết để:

- Một lập trình viên solo có thể triển khai theo từng giai đoạn mà không tự tạo thêm hạ tầng không cần thiết.
- AI coding agents có đủ context để tạo code, migration, query, handler và test đúng kiến trúc.
- Các quyết định quan trọng được ghi lại, giảm tình trạng thay đổi kiến trúc tùy tiện.
- Nghiệp vụ bài thi và bảng điểm có invariant rõ ràng, tránh lỗi dữ liệu khó sửa về sau.

## 2. Kiến trúc chốt

```text
React SPA / PWA-ready (Vercel Hobby)
      |
      | HTTPS + REST/JSON
      v
Go Modular Monolith (Render Free)
  |- HTTP API + OpenAPI
  |- Domain/Application services
  |- PostgreSQL repositories
  |- River background workers (in-process)
  |- Scheduler (in-process)
      |
      +--> Supabase Free PostgreSQL 15+
      +--> Supabase Storage (S3-compatible)
```

MVP demo stack: Vercel Hobby + Render Free + Supabase Free. Xem ADR-0005.
```

### Nguyên tắc triển khai

1. **Một codebase backend, một database, một deployment chính** ở MVP demo (Render Free service chạy API + River + scheduler in-process).
2. **Không dùng Redis, Kafka, Elasticsearch, Kubernetes hoặc microservices** trước khi có dữ liệu đo đạc chứng minh nhu cầu.
3. **PostgreSQL là nguồn sự thật** cho người dùng, bài làm, điểm và background jobs.
4. **Mọi thay đổi quan trọng phải transactional và audit được.**
5. **Question version, assessment snapshot và grade history là bắt buộc.**
6. **Server là nguồn thời gian duy nhất** cho bài thi.
7. **Authorization phải kiểm tra cả permission và phạm vi tài nguyên**, không chỉ kiểm tra vai trò.

## 3. Thứ tự đọc khuyến nghị

| Thứ tự | Tài liệu | Mục đích |
|---:|---|---|
| 1 | [00-project-scope.md](00-project-scope.md) | Phạm vi MVP, actor, invariant |
| 2 | [01-tech-stack.md](01-tech-stack.md) | Công nghệ và lý do lựa chọn |
| 3 | [02-system-architecture.md](02-system-architecture.md) | Kiến trúc tổng thể và boundary |
| 4 | [03-folder-structure.md](03-folder-structure.md) | Cấu trúc repository và source code |
| 5 | [04-domain-model.md](04-domain-model.md) | Bounded modules, aggregate và state machine |
| 6 | [05-database-design.md](05-database-design.md) | Thiết kế PostgreSQL, bảng và index |
| 7 | [06-api-conventions.md](06-api-conventions.md) | Quy ước HTTP/API chung |
| 8 | [api/](api/) | Endpoint theo module |
| 9 | [07-dataflow-processing.md](07-dataflow-processing.md) | Luồng dữ liệu và transaction |
| 10 | [08-core-functions.md](08-core-functions.md) | Authentication, validation, errors, idempotency |
| 11 | [09-security.md](09-security.md) | Security controls |
| 12 | [10-background-jobs.md](10-background-jobs.md) | River jobs và retry |
| 13 | [11-testing-strategy.md](11-testing-strategy.md) | Unit/integration/E2E/load/security test |
| 14 | [12-observability-deployment.md](12-observability-deployment.md) | Log, metrics, backup, deployment |
| 15 | [13-ai-agent-guide.md](13-ai-agent-guide.md) | Quy tắc cho AI coding agents |
| 16 | [14-implementation-roadmap.md](14-implementation-roadmap.md) | Lộ trình triển khai solo |

## 4. Danh sách API module

- [Auth & Users](api/01-auth-users.md)
- [Organizations & Classes](api/02-organizations-classes.md)
- [Resources & Files](api/03-resources.md)
- [Question Bank](api/04-question-bank.md)
- [Assessments & Attempts](api/05-assessments-attempts.md)
- [Assignments & Gradebook](api/06-assignments-gradebook.md)
- [Notifications & Audit](api/07-notifications-audit.md)

## 5. Architectural Decision Records

- [ADR-0001: Modular Monolith](adr/0001-modular-monolith.md)
- [ADR-0002: PostgreSQL as Primary Store](adr/0002-postgresql-primary-store.md)
- [ADR-0003: JWT Access + Rotating Refresh Token](adr/0003-auth-token-strategy.md)
- [ADR-0004: River on PostgreSQL](adr/0004-river-queue.md)
- [ADR-0005: MVP Demo Deployment Topology](adr/0005-deployment-topology.md)

## 6. Definition of Done cấp hệ thống

Một feature backend chỉ được coi là hoàn thành khi:

- Có migration hoặc xác nhận không cần migration.
- Có sqlc query và repository adapter.
- Có application service/use case.
- Có authorization ở resource level.
- Có structural validation và business validation.
- Có error mapping theo Problem Details.
- Có audit event nếu tác động dữ liệu nhạy cảm.
- Có unit test cho invariant.
- Có integration test với PostgreSQL thật.
- OpenAPI được cập nhật tự động và client types sinh lại thành công.
- Không tạo dependency mới nếu chưa có lý do ghi trong PR/ADR.

## 7. Tệp dành riêng cho AI agents

AI agent phải đọc [AGENTS.md](AGENTS.md) trước khi sửa code. Template nhận task nằm tại [templates/agent-task-template.md](templates/agent-task-template.md).
