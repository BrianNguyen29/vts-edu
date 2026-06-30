# 07. API Client, Authentication & Token Handling

## 1. Contract generation

Backend sinh OpenAPI 3.1 tại:

```text
apps/api/openapi/openapi.json
```

Frontend generation:

```bash
pnpm --filter @lms/api-client generate
```

Kết quả:

```text
packages/api-client/src/generated/schema.d.ts
```

Không sửa generated schema.

## 2. Client composition

```ts
import createClient from 'openapi-fetch';
import createQueryClient from 'openapi-react-query';
import type { paths } from '@lms/api-client/generated/schema';

export const rawApi = createClient<paths>({
  baseUrl: runtimeConfig.apiBaseUrl,
  credentials: 'include',
});

export const apiQuery = createQueryClient(rawApi);
```

`credentials: 'include'` cần cho refresh cookie khi cross-origin (Vercel → Render). `baseUrl` là absolute API origin trong demo, ví dụ `https://<api>.onrender.com/api/v1`.

CORS phải allowlist chính xác Vercel origin (production và preview); không dùng `*` với credentials. Cookie refresh phải `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth`.

## 3. Authentication model

Backend contract:

- Access JWT sống 10–15 phút, trả trong JSON.
- Refresh token opaque, rotating, nằm trong cookie `HttpOnly`.
- `/auth/refresh` cấp access token mới và rotate refresh token.
- `/me` trả actor/roles/permissions.

Frontend storage:

| Dữ liệu | Nơi lưu |
|---|---|
| Access token | Memory auth store |
| Refresh token | Browser cookie HttpOnly, frontend không đọc |
| Actor/permissions | Memory + Query cache |
| Theme/locale | localStorage |
| Exam pending answer | IndexedDB, không token |

## 4. Bootstrap flow

```text
load runtime config
-> create auth store/API client with absolute baseUrl
-> POST {apiBaseUrl}/auth/refresh with cookie
   -> 200: set access token
      -> GET {apiBaseUrl}/me
      -> status authenticated
   -> 401: status anonymous
   -> network/5xx: bootstrap degraded/error with retry
-> render router
```

Không hiển thị authenticated route trong khi `bootstrapping`.

## 5. Request middleware

```ts
const authMiddleware: Middleware = {
  async onRequest({ request }) {
    const token = authSession.getSnapshot().accessToken;
    const headers = new Headers(request.headers);

    if (token) headers.set('Authorization', `Bearer ${token}`);
    headers.set('X-Request-ID', createRequestId());

    return new Request(request, { headers });
  },
};
```

Không attach Authorization cho signed S3 URL ngoài API origin.

## 6. Cross-tab refresh serialization

Mọi refresh phải đi qua Web Lock chung `vts-auth-refresh` hoặc fallback leader lease qua `localStorage` + `BroadcastChannel` cho browser không hỗ trợ Web Locks.

Tab giữ lock gọi `/auth/refresh`; tab khác chờ signal completion rồi tự refresh tuần tự nếu cần access token riêng.

Response middleware cũ giữ single-flight trong một JS context; cross-tab serialization bổ sung bảo vệ trước race revoke family.

```ts
async function serializedRefresh(): Promise<boolean> {
  if ('locks' in navigator) {
    return navigator.locks.request('vts-auth-refresh', async () => performRefresh());
  }
  // Fallback: leader lease with timeout/fencing token
  return fallbackLeaderLeaseRefresh();
}
```

Response middleware:

```text
401 protected request
  -> nếu request đã retry: fail
  -> cross-tab serialized refresh
      -> success: clone/rebuild request với token mới, retry một lần
      -> fail 401: clear session, broadcast logout, redirect login
      -> network/5xx: return recoverable auth error, không tự logout ngay
```

Refresh endpoint itself không đi qua retry-refresh loop.

## 7. Initial login

```text
submit login form
-> POST /auth/login
-> backend sets refresh cookie and returns access token + user summary
-> set access token in memory
-> GET /me if login response không đủ permissions
-> clear stale cache
-> navigate validated returnTo hoặc role home
```

Nếu `must_change_password`:

- Chỉ cho route `/change-password` và logout.
- Không preload dashboard.

## 8. Logout

```text
POST /auth/logout
-> on success: clear local access token/cache
-> on network fail: set local `logout_pending` marker (non-sensitive), retry revoke in background
-> clear local access token/cache regardless
-> broadcast logout
-> navigate /login
```

Nếu network fail, UI vẫn clear local session; server refresh session có thể còn đến khi expire. Tuy nhiên, `logout_pending` marker phải được kiểm tra trong bootstrap: nếu marker tồn tại, suppress auto-refresh cookie và hiển thị cảnh báo rõ trước khi cho tiếp tục. Điều này ngăn auto-login không mong muốn trên thiết bị dùng chung khi lỗi mạng.

## 9. Cross-tab behavior

```ts
const channel = new BroadcastChannel('lms-auth');
channel.postMessage({ type: 'LOGOUT' });
```

Các event:

- `LOGOUT`.
- `SESSION_REVOKED`.
- `PASSWORD_CHANGED`.

Không gửi token qua channel.

## 10. Route guard

Guard chỉ dùng auth snapshot và actor permissions:

```ts
if (auth.status === 'anonymous') {
  throw redirect(`/login?returnTo=${encodeURIComponent(safePath)}`);
}

if (!hasAllPermissions(actor.permissions, route.requiredPermissions)) {
  throw new Response('Forbidden', { status: 403 });
}
```

## 11. Problem Details mapping

Backend error:

```json
{
  "type": "https://example.local/problems/validation-error",
  "title": "Validation failed",
  "status": 422,
  "code": "VALIDATION_ERROR",
  "detail": "One or more fields are invalid.",
  "request_id": "req_01...",
  "errors": [
    {"field":"body.title","code":"required","message":"title is required"}
  ]
}
```

Frontend normalized:

```ts
interface AppError {
  kind: 'network' | 'auth' | 'forbidden' | 'not-found' | 'validation' |
        'conflict' | 'rate-limit' | 'server' | 'unknown';
  code: string;
  message: string;
  status?: number;
  requestId?: string;
  fieldErrors?: Record<string, string[]>;
  retryable: boolean;
}
```

Không hiển thị raw `detail` nếu message có thể lộ thông tin; dùng map theo `code`, giữ `requestId` cho hỗ trợ.

## 12. Request cancellation

- Query function nhận `signal` và truyền xuống Fetch.
- Search autocomplete cancel request cũ.
- Route unmount không bắt buộc cancel mutation đã gửi.
- Upload có `AbortController` hoặc XHR abort handle.

## 13. Idempotency keys

Endpoints cần key:

- Start attempt.
- Submit attempt.
- Publish assessment.
- Export.
- Import initiation.

Generator:

```ts
export function createIdempotencyKey(scope: string): string {
  return `${scope}:${crypto.randomUUID()}`;
}
```

Key phải được giữ ổn định khi retry cùng user action. Không tạo key mới cho mỗi network retry.

## 14. Optimistic concurrency

Resource edit có version/ETag:

```text
GET resource -> version/etag
PATCH resource + If-Match
412/409 -> show conflict resolution
```

Không âm thầm overwrite. Với assessment builder, cho phép reload latest hoặc copy local changes để người dùng xử lý.

## 15. Upload transport

Flow:

```text
POST file intent to API
-> receive signed upload URL + file ID
-> upload binary directly to object storage
-> POST confirm to API
-> poll/query processing status if virus scan/preview async
```

Upload binary không dùng API bearer token với signed third-party URL.

Progress:

- Dùng XHR wrapper nhỏ nếu cần progress event.
- Abort khi user cancel.
- Không retry file lớn vô hạn.

## 16. Download

```text
request authorized download URL from API
-> receive short-lived signed URL
-> navigate/fetch immediately
```

Không cache signed URL trong localStorage hoặc query cache lâu.

## 17. Retry matrix

| Tình huống | Hành vi |
|---|---|
| Network GET | Retry tối đa 2 với backoff |
| 401 access expired | Refresh single-flight, retry 1 |
| Refresh 401 | Clear session |
| Refresh 5xx/network | Show retryable session error |
| 403 | Không retry |
| 409/412 | Conflict UI |
| 422 | Map field/general errors |
| 429 | Tôn trọng `Retry-After`, không spam |
| 5xx write idempotent | Có thể retry với cùng key |
| 5xx write không idempotent | Không tự retry |

## 18. Deployment topology

### MVP demo (default): cross-origin

```text
https://<app>.vercel.app/                  -> Vercel Hobby SPA
https://<api>.onrender.com/api/v1          -> Render Free Go API
```

Cross-origin requirements:

- Frontend `baseUrl` absolute Render origin.
- `credentials: 'include'`.
- Refresh cookie: `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth`.
- CORS allowlist exact Vercel origins (production + preview); no `*` with credentials.
- CSRF token bắt buộc cho cookie-backed auth endpoints/mutations (double-submit cookie hoặc header).

### Production same-origin/same-site (future)

```text
https://lms.example.com/        -> static SPA
https://lms.example.com/api/v1  -> Go API
```

Cookie policy:

- `HttpOnly; Secure; SameSite=Lax` cho cùng site.
- Nếu thực sự cross-site: `Secure; SameSite=None`, kèm Origin/Referer allowlist, CSRF token cho cookie-auth endpoints và test browser.

Lợi ích same-origin:

- Cookie/CORS đơn giản.
- CSP/connect-src gọn.
- Ít lỗi môi trường.

Nếu khác origin, backend phải cấu hình CORS allowlist cụ thể; không dùng `*` với credentials.
