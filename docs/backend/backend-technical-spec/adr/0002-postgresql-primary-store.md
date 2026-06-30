# ADR-0002: PostgreSQL as Primary Structured Data Store

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Domain có quan hệ mạnh và cần transaction: enrollment, assessment snapshots, attempts, answers, grades và audit.

## Decision

Dùng PostgreSQL 15+ làm datastore structured duy nhất (Supabase Free PostgreSQL cho demo). File binary ở object storage. Không dùng NoSQL ở MVP.

## Consequences

- Có ACID, constraints, SQL analytics và FTS.
- Giảm vận hành datastore thứ hai.
- JSONB dùng có kiểm soát cho payload linh hoạt.
- Cần quản lý index/query/pool cẩn thận khi scale.
