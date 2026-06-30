# ADR-0005 — Native Fetch with OpenAPI-generated Client

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Backend sinh OpenAPI 3.1. Dự án cần type-safe client và auth/error middleware, nhưng muốn giảm dependency.

## Decision

Dùng `openapi-typescript`, `openapi-fetch` và `openapi-react-query` trên native Fetch. Không dùng Axios trong MVP.

## Consequences

- Type contract trực tiếp từ API.
- Runtime nhỏ và AbortSignal native.
- Upload progress dùng XHR adapter riêng nếu cần.
- Middleware refresh/error cần implement cẩn thận.
