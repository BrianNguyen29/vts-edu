# 09. Security Specification

## 1. Security baseline

Mục tiêu:

- OWASP ASVS Level 2 cho các luồng chính.
- Rà soát OWASP API Security Top 10, đặc biệt Broken Object Level Authorization.
- Principle of least privilege.
- Deny by default.

## 2. Threat model trọng tâm

| Asset | Threat | Control chính |
|---|---|---|
| Student records | Cross-tenant/IDOR | Org-scoped queries + resource auth + tests |
| Attempts/answers | Tampering/loss | Revision, transaction, audit, server time |
| Grades | Unauthorized override | Permission + class scope + history + audit |
| Credentials | Brute force/token theft | Argon2id, rate limit, short JWT, rotating refresh |
| Files | Malware/content injection | Private bucket, type validation, scan, signed URL |
| Exports | Mass data leakage | Async authorized export, expiring link, audit |
| Rich text | Stored XSS | Sanitization/allowlist/CSP |

## 3. Authentication controls

- TLS bắt buộc production.
- Access JWT ngắn hạn.
- Refresh token HttpOnly/Secure/SameSite.
- Token không lưu localStorage/sessionStorage.
- Refresh rotation và reuse detection.
- Password reset token one-time, hashed trong DB, expiry ngắn.
- Suspend/reset password tăng `auth_version` và revoke sessions.
- MFA phase sau cho admin/teacher nếu product yêu cầu.

## 4. Authorization controls

Mỗi endpoint có ID phải kiểm tra:

1. Actor authenticated.
2. Resource cùng organization.
3. Actor có permission.
4. Actor có scope trên class/resource.
5. State cho phép action.

Không dựa vào việc ID là UUID để coi là an toàn.

### Anti-mass assignment

- Input DTO explicit.
- Không bind request trực tiếp vào DB model/domain entity.
- Client không set `organization_id`, `created_by`, `status`, `final_score` tùy ý.

## 5. CSRF and browser security

- Access token gửi bằng `Authorization` header, giảm CSRF cho API protected.
- Refresh/logout dùng cookie: kiểm tra `Origin`/`Referer`, SameSite và chỉ POST.
- **MVP demo cross-origin (Vercel → Render):** refresh cookie phải `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth`; CORS allowlist chính xác Vercel origins; bắt buộc CSRF token (double-submit cookie hoặc header) cho cookie-backed endpoints/mutations.
- Security headers: CSP, `X-Content-Type-Options: nosniff`, frame policy, Referrer-Policy, HSTS tại edge.

## 6. CORS

- Production allowlist origin cụ thể.
- Không `*` với credentials.
- Vercel preview origin động: cần allowlist pattern hoặc cập nhật env; không bật `*` vì credentials.
- Chỉ allow methods/headers cần thiết.
- Development config riêng.

## 7. SQL and database

- Parameterized SQL qua sqlc/pgx.
- DB application user không phải superuser.
- Migration role tách nếu có thể.
- Network access DB private/restricted.
- Backup encrypted.
- Query timeout và connection limits.

## 8. File security

- Upload direct-to-object-storage qua signed URL.
- Random object key.
- Private bucket.
- MIME sniffing/allowlist.
- Max size.
- Malware scan worker nếu production dùng file thật.
- Quarantine trước READY.
- Attachment download cho type nguy hiểm; không inline.
- Không cho path traversal vì client không kiểm soát path.

## 9. Rich content

- Sanitize HTML server-side.
- Không cho script, event handler, iframe tùy ý.
- External image/link policy rõ.
- KaTeX/Math content được parse an toàn, không eval script.

## 10. Rate limits đề xuất ban đầu

| Endpoint group | Limit khởi điểm |
|---|---|
| Login | 5–10/phút mỗi identifier + IP, adaptive cooldown |
| Refresh | 30/phút mỗi session/IP |
| Forgot password | 3/giờ mỗi identifier |
| Upload intent | 20/phút mỗi user |
| Export | 3–5 đang chạy mỗi user/org |
| Answer save | Cao hơn, theo attempt; không limit làm mất bài hợp lệ |

Các số phải load test và cấu hình, không hard-code rải rác.

## 11. Sensitive data handling

- Data minimization.
- Không log password/token/answer body.
- Mask email/phone khi không cần.
- Audit export/download dữ liệu nhạy cảm.
- Retention policy cho login telemetry/audit.
- Dữ liệu học sinh cần quy trình truy cập, sửa, xuất, xóa/ẩn danh theo pháp lý áp dụng.

## 12. Secrets

- Production secrets từ platform secret store/environment injection.
- Key rotation có `kid`.
- Không commit `.env` thật.
- CI secrets tối thiểu quyền.
- Object storage credentials hạn chế bucket/prefix khi khả thi.

## 13. Dependency and supply chain

CI:

```text
go test ./...
go vet ./...
govulncheck ./...
staticcheck ./...
pnpm audit --prod (frontend workspace policy)
container scan
```

- Review dependency mới.
- Pin GitHub Actions by commit SHA nếu security yêu cầu cao.
- Generate SBOM cho release production khi có quy trình.

## 14. Security test cases bắt buộc

- Student đổi `attempt_id` để đọc bài người khác.
- Teacher đổi `class_id` sang lớp không phụ trách.
- User từ org A truy cập resource org B.
- Gửi field ẩn `organization_id`, `status`, `final_score`.
- Refresh token reuse.
- JWT algorithm confusion/expired/wrong audience.
- Stored XSS trong question/assignment feedback.
- CSV formula injection khi export.
- Upload file giả MIME/oversize.
- Duplicate submit và replay idempotency key.
