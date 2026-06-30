# 09. Core Logic & Global Utilities

## 1. Runtime configuration

Không bake mọi environment value vào bundle. Cấu hình runtime/API base URL từ biến môi trường/build-time phù hợp pipeline Vercel.

MVP demo (Vercel → Render cross-origin):

```json
{
  "apiBaseUrl": "https://<api>.onrender.com/api/v1",
  "environment": "production",
  "release": "2026.06.29.1",
  "features": {
    "pwaInstall": false,
    "analyticsDashboard": false
  }
}
```

Local dev với proxy vẫn có thể dùng `/api/v1`.

Load trước render từ `/app-config.json` (do Vercel phục vụ static) hoặc inject vào `index.html`. Validate bằng Zod; fail fast nếu invalid.

Không chứa secret trong frontend config.

## 2. Global error handling

### Error categories

- Network offline/timeout.
- Authentication expired.
- Forbidden.
- Not found.
- Validation.
- Conflict/version.
- Rate limit.
- Server.
- Unexpected client error.

### Presentation

| Error | Presentation |
|---|---|
| Field validation | Inline + form summary |
| Widget query | Inline error state + retry |
| Full route | Route error boundary |
| Auth expired | Login redirect/reauth dialog |
| Rate limit | Inline/toast với thời gian retry |
| Unexpected render | Error boundary + support request ID/release |

## 3. Error Boundary

```tsx
<ErrorBoundary
  fallbackRender={({ error, resetErrorBoundary }) => (
    <UnexpectedErrorState error={normalizeUnknownError(error)} onRetry={resetErrorBoundary} />
  )}
>
  {children}
</ErrorBoundary>
```

Exam boundary không được xóa IndexedDB draft khi reset.

## 4. Logging

Frontend logger interface:

```ts
interface FrontendLogger {
  debug(message: string, context?: SafeContext): void;
  info(message: string, context?: SafeContext): void;
  warn(message: string, context?: SafeContext): void;
  error(message: string, error?: unknown, context?: SafeContext): void;
}
```

Safe context có thể gồm:

- release.
- route template.
- request ID.
- feature name.
- browser capability flags.

Không gồm:

- access token.
- answer content.
- essay text.
- password.
- full student profile.
- signed URL.

## 5. Permission utilities

```ts
export function hasPermission(actor: Actor, permission: string): boolean;
export function hasAllPermissions(actor: Actor, permissions: string[]): boolean;
export function hasAnyPermission(actor: Actor, permissions: string[]): boolean;
```

UI wrapper:

```tsx
<PermissionBoundary require="assessment:publish" fallback={null}>
  <PublishAssessmentButton />
</PermissionBoundary>
```

Không gọi permission bằng role hardcode trong component.

## 6. Date/time utilities

Các hàm:

- `formatDateTime(utc, timezone, locale)`.
- `formatRelativeTime`.
- `toUtcSchedule(localValue, timezone)`.
- `getServerNow(offsetMs)`.
- `formatDuration`.

Tất cả nhận timezone explicit khi nghiệp vụ yêu cầu.

## 7. Server clock utility

Exam bootstrap trả `server_time`. Tính offset:

```ts
const midpoint = requestStartedAt + (responseReceivedAt - requestStartedAt) / 2;
const offsetMs = serverTimeMs - midpoint;
```

Clock:

```ts
interface Clock {
  now(): number;
}

class ServerOffsetClock implements Clock {
  now() { return Date.now() + this.offsetMs; }
}
```

Có thể resync định kỳ với heartbeat nhưng không làm timer nhảy khó hiểu; điều chỉnh có giới hạn.

## 8. Decimal display

```ts
interface DecimalDisplayOptions {
  maximumFractionDigits?: number;
  minimumFractionDigits?: number;
}

formatDecimalString(value: string, locale: string, options?: DecimalDisplayOptions): string;
```

Không thực hiện grade calculation. Nếu cần so sánh/format chính xác, dùng decimal library chỉ sau ADR; MVP có thể validate string và dùng backend-computed values.

## 9. ID utilities

- `crypto.randomUUID()` cho client mutation/idempotency IDs.
- Không tự sinh server entity IDs trừ API cho phép.
- Không dùng timestamp đơn thuần làm unique key.

## 10. File utilities

- `formatFileSize`.
- `isPreviewableMime`.
- `downloadSignedUrlImmediately`.
- `createSafeObjectUrl` và revoke.
- `hashFile` chỉ nếu upload protocol cần; chạy worker nếu file lớn.

## 11. i18n

### Locales

- `vi` mặc định.
- `en` chuẩn bị nhưng có thể chưa hoàn thiện toàn bộ MVP.

### Namespaces

```text
common
navigation
auth
classes
resources
questions
assessments
exam
assignments
gradebook
admin
errors
```

Key semantic:

```text
assessment.publish.confirmTitle
exam.save.pending
errors.validation.generic
```

Không dùng câu tiếng Việt làm key.

### Formatting

Date/number dùng Intl/date utility theo locale. Translation interpolation không nhận raw HTML.

## 12. Feature flags

Feature flag runtime chỉ để:

- Tắt route/feature chưa sẵn sàng.
- Rollout UI không critical.

Không dùng flag frontend để bảo vệ API. Backend cũng phải enforce.

## 13. Theme

- Light là mặc định theo ảnh tham khảo.
- Dark mode là optional P1.
- Theme preference localStorage, không nhạy cảm.
- Tránh flash theme bằng inline bootstrap script nhỏ nếu bật dark mode.

## 14. Toast policy

Dùng toast cho:

- Save thành công không cần điều hướng.
- Copy link.
- Background operation status.

Không dùng toast duy nhất cho:

- Form validation.
- Critical submit failure.
- Permission denied.
- Exam save failure kéo dài.

## 15. Confirmation policy

Confirm dialog bắt buộc cho:

- Publish/unpublish assessment.
- Delete/archive có ảnh hưởng.
- Grade override.
- Terminate attempt.
- Leave dirty editor.

Dialog phải mô tả hậu quả cụ thể, không chỉ “Bạn có chắc không?”.

## 16. Browser capability checks

Exam runner kiểm tra:

- IndexedDB available.
- Web Crypto UUID.
- Page Visibility.
- Storage quota/permission nếu cần.

Nếu IndexedDB unavailable, cảnh báo trước khi start bài thi và áp policy tổ chức; không âm thầm chạy chế độ kém an toàn.

## 17. Page title và breadcrumbs

Route handle cung cấp title key/breadcrumb factory. Title format:

```text
{Page} · {Organization} · LMS
```

Không lấy title từ untrusted rich text.
