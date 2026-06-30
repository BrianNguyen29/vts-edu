# ADR-0003 — TanStack Query Owns Server State

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

App có nhiều list/detail/mutation và trạng thái loading/error/cache. Copy API data vào Redux/Zustand tạo hai nguồn sự thật.

## Decision

TanStack Query 5 quản lý server state. Form, URL, UI và exam durable state có owner riêng.

## Consequences

- Cache/invalidation thống nhất.
- Không cần global state library ở MVP.
- Developer phải thiết kế query key/stale time cẩn thận.
- Query cache không được dùng cho pending exam answer hoặc access token.
