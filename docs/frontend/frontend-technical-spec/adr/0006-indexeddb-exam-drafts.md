# ADR-0006 — IndexedDB for Exam Pending Answer Durability

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Mạng học sinh có thể chập chờn. In-memory state hoặc debounced API call có thể mất answer khi reload/crash. localStorage không phù hợp dữ liệu cấu trúc và transaction.

## Decision

Dùng IndexedDB qua package `idb` để lưu attempt metadata, current answer và pending operations. Network sync chạy foreground; service worker không phải cơ chế duy nhất.

## Consequences

- Cần schema migration và cleanup theo user.
- Cần browser integration tests.
- Có thể resume/retry sau reload.
- Không lưu access token trong database local.
