# 14. Frontend Security Specification

## 1. Security boundary

Frontend hỗ trợ security UX nhưng không phải authorization boundary. Backend luôn kiểm tra:

- Authentication.
- Permission.
- Tenant/resource scope.
- State transition.
- Validation.

## 2. Token security

- Access JWT chỉ memory.
- Refresh cookie `HttpOnly; Secure; SameSite` do backend đặt.
- Không token trong URL, localStorage, logs hoặc error monitoring.
- Logout clear cache và memory.
- Refresh reuse/revocation từ backend dẫn tới session termination.

## 3. CSRF

Access-token protected API dùng Authorization header. Refresh/logout dựa cookie cần backend CSRF policy phù hợp:

- SameSite.
- Origin/Referer validation.
- **CSRF token bắt buộc cho cookie-backed endpoints/mutations khi cross-origin (Vercel → Render).** Sử dụng double-submit cookie hoặc CSRF header.

Frontend gửi `credentials: include`, không bypass policy, và attach CSRF token theo backend requirement cho refresh/logout/state-changing cookie endpoints.

## 4. XSS prevention

- React escape text mặc định.
- Không render raw HTML trừ renderer được kiểm soát.
- DOMPurify sanitize defense-in-depth.
- Backend cũng sanitize/validate canonical rich content.
- Không cho teacher content chèn script, iframe tùy ý, event handler hoặc unsafe URL scheme.

## 5. Content Security Policy

Production CSP mục tiêu:

```text
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline';  # giảm/nonce nếu khả thi
img-src 'self' data: blob: <supabase-storage-host>;
font-src 'self';
connect-src 'self' <render-api-origin> <supabase-storage-host>;
frame-ancestors 'none';
base-uri 'self';
form-action 'self';
```

Nếu KaTeX/style yêu cầu, cấu hình cụ thể; không mở `*`.

## 6. Open redirect

`returnTo` phải:

- Relative same-origin path.
- Không bắt đầu `//`.
- Không chứa protocol.
- Optionally allowlist route prefixes.

## 7. File security

- Không render SVG/HTML upload trực tiếp.
- Preview PDF qua viewer/sandbox phù hợp.
- Signed URL short-lived, không persist.
- Object URL được revoke.
- Client MIME check chỉ là UX; không coi là security.

## 8. Download and CSV

- CSV từ backend có thể chứa công thức; backend phải mitigate CSV injection.
- Frontend không tự dựng export nhạy cảm từ data cache nếu backend có audited export endpoint.
- Export action cần permission, confirm và download receipt/status.

## 9. Clickjacking

Backend/static server đặt `frame-ancestors 'none'` hoặc allowlist nếu LTI phase sau. Không cho app bị embed tùy ý.

## 10. Browser storage privacy

| Storage | Cho phép |
|---|---|
| localStorage | theme, locale, sidebar preference |
| sessionStorage | non-sensitive transient return state nếu cần |
| IndexedDB | exam pending answers, scoped user/attempt |
| Cookie readable JS | Không dùng cho auth secret |

Shared-device logout phải cleanup user-scoped local data theo policy.

## 11. Sensitive telemetry

Redaction bắt buộc:

- URL query token.
- Authorization header.
- Form values password.
- Essay/answer content.
- Grade comments nếu có PII.
- Signed URLs.

Error monitoring chỉ gửi stack, release, route template, request ID và safe context.

## 12. Supply-chain security

pnpm 11:

- `minimumReleaseAge` giữ mặc định hoặc policy chặt hơn.
- `allowBuilds` allowlist package cần install script.
- Lockfile committed.
- `--frozen-lockfile` CI.
- Dependabot/Renovate PR, không auto-merge critical packages.
- Dependency audit và license review định kỳ.

Ví dụ:

```yaml
# pnpm-workspace.yaml
packages:
  - apps/*
  - packages/*

minimumReleaseAge: 1440
allowBuilds:
  esbuild: true
  '@swc/core': false
```

Danh sách thực tế theo dependencies; không copy mù.

## 13. Environment/config security

- `VITE_*` đều public; không đặt secret.
- API secret, S3 secret, private key chỉ backend.
- Source maps production không public hoặc upload private vào monitoring.

## 14. Permission UX

- Hide action khi không có permission.
- Nếu URL trực tiếp, guard 403.
- Nếu backend trả 403 sau permission stale, update actor permissions/query và hiển thị forbidden.
- Không expose data trong disabled tooltip nếu user không được xem.

## 15. Exam integrity UX

- Không quảng cáo “chống gian lận tuyệt đối”.
- Visibility/device signals là audit signals, không kết luận.
- Không thu webcam/sinh trắc học trong MVP.
- Không chặn phím hoặc browser controls theo cách gây mất accessibility.

## 16. Security testing

- XSS rich text/link.
- Token absent from storage/logs.
- Open redirect.
- Route permission bypass UX.
- Signed URL leakage.
- Cross-user IndexedDB cleanup.
- CSRF behavior refresh/logout.
- Dependency/build script policy.
