# ADR-0001: Feature-first Modular Monolith

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Dự án do một người phát triển, domain phức tạp ở transaction, bài thi và điểm. Microservices làm tăng deployment, networking, distributed transaction và observability cost.

## Decision

Dùng một Go codebase dạng feature-first modular monolith, với domain/application/repository/HTTP boundaries nhẹ. MVP chạy một process; có thể tách API/worker process từ cùng codebase.

## Consequences

### Positive

- Dễ debug/deploy.
- Transaction PostgreSQL trực tiếp.
- Ít boilerplate vận hành.
- Vẫn giữ module boundaries.

### Negative

- Cần discipline để tránh coupling giữa modules.
- Một deployment lỗi có thể ảnh hưởng toàn app.
- Scale độc lập theo module bị hạn chế, nhưng có thể tách worker/process trước khi microservice.
