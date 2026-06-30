# ADR-0004: River with PostgreSQL for Background Jobs

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Cần background jobs cho grading, email, file processing và exports. Dự án solo cần tránh vận hành Redis/RabbitMQ.

## Decision

Dùng River trên PostgreSQL. Enqueue job trong cùng transaction với business state khi cần consistency. MVP demo chạy River workers in-process trong Render Free service; scale phase mới tách `cmd/worker`.

## Consequences

### Positive

- Không thêm datastore/broker.
- Transactional enqueue.
- Retry/schedule/worker model phù hợp Go.

### Negative

- Queue chia tài nguyên với primary DB.
- Workload rất lớn có thể cần tách broker/service sau.
- Cần giới hạn worker concurrency và theo dõi DB load.
