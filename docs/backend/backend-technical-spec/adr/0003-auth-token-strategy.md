# ADR-0003: Short-lived JWT Access Token + Rotating Opaque Refresh Token

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

Frontend là SPA. Cần API authorization rõ, không lưu long-lived credential trong localStorage và có khả năng revoke session.

## Decision

- Access token là JWT 10–15 phút, gửi bằng `Authorization` header, lưu memory phía SPA.
- Refresh token là opaque random token trong HttpOnly/Secure/SameSite cookie.
- DB lưu refresh token hash.
- Rotate refresh token mỗi lần dùng và phát hiện reuse theo token family.
- MVP demo cross-origin (Vercel → Render): cookie cần `SameSite=None; Secure` và CSRF token cho cookie-backed endpoints.

## Consequences

### Positive

- Access API nhanh và standard.
- Refresh có server-side revocation.
- Giảm exposure của long-lived token với JavaScript.

### Negative

- Flow phức tạp hơn opaque session thuần.
- Access token đã phát hành không revoke tức thì; giảm bằng TTL ngắn và auth version/policy cho thao tác nhạy cảm.
- Refresh endpoint cần origin/cookie security.
