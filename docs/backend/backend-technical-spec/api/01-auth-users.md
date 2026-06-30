# API 01 — Authentication & Users

Base path: `/api/v1`

## 1. Endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| POST | `/auth/login` | Đăng nhập bằng organization code + username/password | `{"organization_code":"school-a","username":"hs001","password":"***"}` | `{"data":{"access_token":"ey...","expires_in":900,"user":{"id":"...","display_name":"Nguyễn A"}}}` + refresh cookie |
| POST | `/auth/refresh` | Rotate refresh token và cấp access token mới | Không có; refresh cookie | `{"data":{"access_token":"ey...","expires_in":900}}` |
| POST | `/auth/logout` | Revoke session hiện tại | `{}` | `204 No Content` |
| POST | `/auth/logout-all` | Revoke toàn bộ refresh sessions của user | `{"current_password":"***"}` hoặc step-up auth | `204 No Content` |
| POST | `/auth/change-password` | Đổi mật khẩu | `{"current_password":"***","new_password":"***"}` | `204 No Content` |
| POST | `/auth/forgot-password` | Yêu cầu reset nếu policy cho phép | `{"organization_code":"school-a","identifier":"teacher@example.com"}` | `202 Accepted` |
| POST | `/auth/reset-password` | Reset bằng one-time token | `{"token":"...","new_password":"***"}` | `204 No Content` |
| GET | `/me` | Thông tin actor hiện tại | — | `{"data":{"id":"...","organization_id":"...","roles":["student"],"permissions":[...]}}` |
| GET | `/me/sessions` | Danh sách session của chính mình | — | `{"data":[{"id":"...","created_at":"...","last_used_at":"...","current":true}]}` |
| DELETE | `/me/sessions/{session_id}` | Revoke session | — | `204 No Content` |
| GET | `/users/{user_id}/sessions` | Admin xem session của user (step-up permission) | — | `{"data":[{"id":"...","created_at":"...","device_metadata":{...}}]}` |
| DELETE | `/users/{user_id}/sessions/{session_id}` | Admin revoke một session | `{"reason":"security_review"}` | `204 No Content` |
| DELETE | `/users/{user_id}/sessions` | Admin revoke toàn bộ session của user | `{"reason":"security_review","revoke_current":false}` | `204 No Content` |
| GET | `/users` | Danh sách user trong organization | — | `{"data":[{"id":"...","username":"hs001","status":"ACTIVE"}],"page":{...}}` |
| POST | `/users` | Tạo user và membership | `{"username":"hs001","display_name":"Nguyễn A","roles":["student"],"temporary_password":"***"}` | `{"data":{"id":"...","status":"ACTIVE","must_change_password":true}}` |
| GET | `/users/{user_id}` | Xem user | — | `{"data":{"id":"...","profile":{...},"roles":[...]}}` |
| PATCH | `/users/{user_id}` | Cập nhật profile/status fields được phép | `{"display_name":"Nguyễn Văn A"}` | `{"data":{"id":"...","display_name":"Nguyễn Văn A"}}` |
| POST | `/users/{user_id}/suspend` | Khóa tài khoản | `{"reason":"left_school"}` | `{"data":{"status":"SUSPENDED"}}` |
| POST | `/users/{user_id}/activate` | Kích hoạt lại | `{}` | `{"data":{"status":"ACTIVE"}}` |
| POST | `/users/{user_id}/reset-password` | Admin reset password tạm | `{"temporary_password":"***","revoke_sessions":true}` | `204 No Content` |
| PUT | `/users/{user_id}/roles` | Thay role trong organization | `{"roles":["teacher"]}` | `{"data":{"roles":["teacher"]}}` |
| POST | `/users/imports` | Bắt đầu import CSV/JSON | `{"file_id":"...","dry_run":true}` | `202 {"data":{"job_id":"...","status":"QUEUED"}}` |
| GET | `/users/imports` | Danh sách import jobs | `?status=FAILED&limit=20` | `{"data":[{"job_id":"...","status":"FAILED","created_at":"..."}],"page":{...}}` |
| GET | `/users/imports/{job_id}` | Trạng thái import | — | `{"data":{"status":"COMPLETED","accepted":42,"rejected":2,"errors_file_id":"..."}}` |
| POST | `/users/imports/{job_id}/confirm` | Xác nhận dry-run import | `{"idempotency_key":"..."}` | `202 {"data":{"job_id":"...","status":"QUEUED"}}` |
| POST | `/users/imports/{job_id}/cancel` | Hủy import chưa chạy | `{}` | `204 No Content` |

## 2. Login processing

1. Normalize organization code và username.
2. Apply rate limit theo IP + organization + identifier hash.
3. Load organization active.
4. Load user/membership active.
5. Verify Argon2id hash bằng constant-time compare path.
6. Nếu cần đổi password, vẫn có thể cấp token với claim/flag hạn chế chỉ cho change-password flow.
7. Tạo refresh session với random 256-bit token; lưu hash, không lưu raw token.
8. Cấp access JWT 10–15 phút.
9. Set refresh cookie `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth` cho cross-origin MVP demo (Vercel → Render). Nếu same-origin/same-site, có thể dùng `SameSite=Lax`. CORS allowlist chính xác Vercel origins.
10. CSRF token bắt buộc cho cookie-backed auth endpoints/mutations khi cross-origin.
11. Ghi audit/login event phù hợp.

## 3. Access JWT claims

```json
{
  "iss": "lms-api",
  "aud": "lms-web",
  "sub": "user-uuid",
  "sid": "refresh-session-uuid",
  "org": "organization-uuid",
  "roles": ["teacher"],
  "av": 3,
  "jti": "token-uuid",
  "iat": 1782727200,
  "exp": 1782728100
}
```

- `av`: auth version; tăng khi reset mật khẩu hoặc revoke-all.
- Permission resource-level vẫn kiểm tra trong backend.
- Không chứa PII, điểm hoặc class list trong JWT.

## 4. Refresh rotation

Trong transaction:

```text
lock refresh session
-> verify not revoked/expired/reused
-> verify membership_id/organization_id/auth_version still active
-> mark current token consumed/replaced
-> insert next refresh token in same family
-> issue new access token
-> commit
```

Refresh session chứa `membership_id`, `organization_id`, `auth_version`, device metadata tối thuể. Refresh phải xác nhận organization và membership vẫn active.

Nếu phát hiện refresh token cũ đã bị dùng lại:

- Revoke toàn bộ token family.
- Ghi security audit.
- Trả 401.

## 5. User creation rules

- Username unique theo organization nếu login dùng organization code.
- Email có thể nullable cho học sinh.
- Temporary password không log và không trả lại sau response create.
- Nếu server sinh password, trả đúng một lần qua channel phù hợp; production nên ưu tiên admin nhập hoặc one-time activation flow.
- Không hard-delete user có enrollment/grade/attempt.
