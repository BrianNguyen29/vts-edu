# 15. Observability & Deployment

## 1. Deployment model

MVP demo stack:

```text
Vercel Hobby
  |- React SPA build from Git
  |- /app-config.json static runtime config
  |- SPA fallback for client routes
        |
        | HTTPS cross-origin to Render API
        v
Render Free Go service
  |- /api/v1/* API
  |- /api/v1/auth/* refresh/logout cookie endpoints
```

Production same-origin (future):

```text
Go binary/container
  |- /api/v1/*       API
  |- /app-config.json runtime config
  |- /*              Vite static assets + SPA fallback
```

Hoặc CDN phục vụ static assets, nhưng route fallback và config vẫn phải rõ.

## 2. Build pipeline

```bash
pnpm install --frozen-lockfile
pnpm api:generate
pnpm web:typecheck
pnpm web:lint
pnpm web:test
pnpm web:build
```

Build output:

```text
apps/web/dist/
```

Vercel build command chạy pipeline trên và deploy `dist/`. Không copy dist vào Go image ở demo.

## 3. Runtime config

Vercel phục vụ `/app-config.json` như static asset. `apiBaseUrl` là absolute Render API origin cho demo, ví dụ `https://<api>.onrender.com/api/v1`.

Schema version config phải tương thích app release; invalid config render bootstrap error.

CSP `connect-src` phải bao gồm Render API origin và Supabase Storage host.

## 4. Release identification

Inject:

- `release`/git SHA.
- Build time.
- Environment.

Hiển thị ở support/about screen; gửi trong safe telemetry.

## 5. Frontend telemetry

Event tối thiểu:

| Event | Fields an toàn |
|---|---|
| `route_error` | route template, release, error code |
| `api_error` | operation ID, status, request ID, release |
| `exam_save_latency` | duration bucket, status, attempt hash/anonymous ID nếu policy cho phép |
| `exam_sync_failure` | error category, retry count |
| `bootstrap_failure` | config/auth/network category |
| Web Vitals | metric, value, route group |

Không gửi content answer/PII.

## 6. Error reporting

MVP lựa chọn:

- Console structured logs development.
- Production optional Sentry-compatible service hoặc endpoint nội bộ.
- Sampling và redaction.
- Source maps upload private.

Nếu chưa dùng vendor, vẫn có logger interface để thêm sau không sửa feature code.

## 7. Request correlation

API middleware gửi `X-Request-ID`; response/error lưu request ID. UI error state cho phép copy mã hỗ trợ.

Frontend-generated request ID không chứa user info.

## 8. Health/support screen

`/app/support` optional hiển thị:

- Release.
- Browser/version.
- Online status.
- IndexedDB capability.
- Last safe request ID.
- Không hiển thị token.

## 9. Static caching

- Hashed assets immutable.
- `index.html` no-cache.
- Runtime config no-store/short cache.
- SPA fallback không trả index cho `/api/*` hoặc file missing asset.

## 10. Deployment safety

- Build artifact immutable.
- Database/backend compatibility checked before rollout.
- Frontend should tolerate additive API fields.
- Breaking API contract requires coordinated release/version.
- Keep previous static artifact for rollback.

## 11. Active exam deployment

- Không force refresh client.
- New release notification trì hoãn.
- API backward compatibility cho active attempt trong deployment window.
- Service worker update không activate destructively.

## 12. Environment matrix

| Environment | Hosting | API | Auth cookie | Telemetry | PWA |
|---|---|---|---|---|---|
| Local | Vite dev | localhost proxy | Secure off only local | console | off |
| Test | ephemeral | test config | test config | disabled/captured | off |
| Staging | Vercel preview | Render preview | production-like | enabled sample | optional |
| Demo/Production | Vercel Hobby | Render Free | `Secure; SameSite=None` | redacted/sample | controlled |

Demo không có production SLA; dữ liệu nên synthetic/non-critical cho đến khi nâng cấp.

## 13. Vite dev proxy

```ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: false,
    },
    '/app-config.json': {
      target: 'http://localhost:8080',
    },
  },
}
```

Giữ URL frontend giống production để giảm CORS drift.

## 14. Monitoring alerts

Frontend-derived alerts theo aggregate, không theo một lỗi đơn:

- Tăng mạnh bootstrap failure.
- 401 refresh failure rate bất thường.
- Exam save failure rate.
- JS error spike theo release.
- Web Vitals regression.

## 15. Rollback

```text
rollback static artifact
-> verify app-config compatibility
-> keep API compatible
-> invalidate index.html/CDN if needed
```

Không purge hashed assets đang được active client dùng ngay lập tức.
