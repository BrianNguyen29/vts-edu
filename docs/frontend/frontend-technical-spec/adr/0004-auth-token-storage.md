# ADR-0004 — Access Token in Memory, Refresh Token in HttpOnly Cookie

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Backend dùng access JWT ngắn hạn và rotating opaque refresh token. Lưu JWT trong localStorage tăng hậu quả XSS và làm revoke/session handling khó hơn.

## Decision

- Access token chỉ ở memory store.
- Refresh token trong backend-set HttpOnly cookie.
- Reload bootstrap bằng `/auth/refresh` rồi `/me`.
- 401 dùng single-flight refresh và retry một lần.

## Consequences

- Token không survive reload; cần bootstrap request.
- Không đọc refresh token từ JS.
- Same-origin deployment được ưu tiên cho production tương lai.
- **MVP demo cross-origin (Vercel → Render):** cookie cần `SameSite=None; Secure` và CSRF token; CORS allowlist chính xác Vercel origins.
- Cross-tab logout dùng BroadcastChannel, không truyền token.
