# 13. Performance, Bundle & PWA

## 1. Performance principles

- Đo trước khi tối ưu sâu.
- Giảm JavaScript ban đầu quan trọng hơn micro-optimization component.
- Lazy-load theo route và capability.
- Không poll/render nhiều hơn nghiệp vụ cần.
- Thiết bị học sinh cấu hình thấp là target thật.

## 2. Budget ban đầu

| Metric | Target pilot |
|---|---:|
| Initial app-shell JS gzip | ≤ 250 KB |
| Initial CSS gzip | ≤ 60 KB |
| Lazy route chunk thông thường | ≤ 200 KB gzip |
| Heavy editor/chart chunk | Theo route, không initial |
| LCP dashboard | < 2,5s điều kiện pilot |
| INP | < 200ms mục tiêu |
| CLS | < 0,1 |

Budget được kiểm tra bằng build report; thay đổi lớn phải giải thích.

## 3. Code splitting

Lazy-load:

- Teacher question editor.
- Assessment builder.
- Exam runner.
- Recharts.
- TipTap.
- PDF/document preview.
- Admin audit/import pages.

Không tạo manual chunk quá chi tiết trước khi analyzer cho thấy lợi ích.

## 4. Query performance

- Dùng pagination/cursor.
- `keepPreviousData`/placeholder hợp lý cho list filter.
- Không refetch on window focus trong active exam.
- Dashboard query parallel, không waterfall không cần thiết.
- Prefetch detail khi hover/focus chỉ nếu data nhỏ và network hợp lý.

## 5. Rendering performance

- Component split theo update frequency.
- Không memo tất cả component.
- Dùng stable selectors/hooks.
- Gradebook lớn cân nhắc `@tanstack/react-virtual` sau profiling.
- Long question list dùng pagination trước virtualization.

## 6. Asset performance

- SVG/icon tree-shake.
- Illustration dùng WebP/AVIF khi phù hợp.
- Không embed base64 ảnh lớn.
- Font subset; ưu tiên system font giai đoạn đầu hoặc 1 family có ít weights.
- Static assets hash và cache immutable.

## 7. Rich editor

- Lazy import TipTap.
- Chỉ load extensions cần thiết.
- Không mount editor ở list/preview.
- Preview dùng renderer nhẹ.

## 8. PWA scope

### Phase 1

- Web app responsive.
- Manifest cơ bản.
- HTTPS.
- Không service worker nếu chưa kiểm soát update.

### Phase 2

- Installable PWA.
- Precache app shell/static assets.
- Offline fallback page.
- Notification permission chỉ sau user action và business need.

### Không cam kết

- Full offline course/resource access.
- Offline start exam không có server authorization.
- Background Sync là guarantee.

## 9. Service worker update policy

- New version available -> non-blocking banner.
- Không reload active exam hoặc dirty editor.
- Activate/reload sau terminal attempt hoặc user xác nhận.
- Cache API response nhạy cảm không mặc định.
- `Cache-Control: no-store` cho auth/attempt runtime nếu backend chỉ định.

## 10. IndexedDB vs service worker

Exam durability dùng IndexedDB repository trực tiếp. Service worker chỉ là transport enhancement sau này.

```text
Page -> IndexedDB durable queue -> API sync
```

không phải:

```text
Page -> hope service worker catches request
```

## 11. Cache headers deployment

| Asset | Header |
|---|---|
| `index.html` | `no-cache` hoặc short max-age |
| Hashed JS/CSS | `public, max-age=31536000, immutable` |
| `app-config.json` | `no-store`/short cache |
| Service worker | `no-cache` |
| User API | Backend policy; không CDN public cache |

## 12. Web Vitals and telemetry

Thu thập sampling:

- LCP.
- INP.
- CLS.
- Route transition duration.
- Exam save acknowledgement latency.
- Error count by release.

Không gắn PII hoặc answer content.

## 13. Low-bandwidth behavior

- Skeleton nhẹ, không tải illustration đầu tiên.
- Download resource có size trước khi bắt đầu.
- Upload có progress/cancel.
- Dashboard không cần tất cả chart trước khi usable.
- Exam question snapshot nên được backend/page load hợp lý để tránh request từng câu nếu có thể.

## 14. Bundle review checklist

- Dependency mới có tree shaking không?
- Chỉ dùng một function nhưng import cả library?
- Có thể lazy-load không?
- Locale data có bị bundle toàn bộ không?
- Editor/chart có vào app shell không?
- Source maps có public ngoài ý muốn không?
