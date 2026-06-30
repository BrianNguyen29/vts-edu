# ADR-0010: Staged Huma/sqlc/OpenAPI Groundwork

- **Status:** Accepted
- **Date:** 2026-06-30

## Context

Backend đã có các feature slices ổn định (auth, attempts, assessments, admin) với repository interfaces rõ ràng và migrations tuần tự. Việc áp dụng Huma (OpenAPI-first handlers) và sqlc (generated queries) là hữu ích cho dài hạn nhưng không nên rewrite runtime code một lúc vì rủi ro regression cao và làm chậm delivery.

## Decision

Tiếp cận từng giai đoạn:

1. **Hiện tại**: Duy trì OpenAPI skeleton bằng tay (`docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`) phản ánh đúng API surface đã implement. Sinh TypeScript types từ skeleton (`apps/web/src/shared/api/openapi-schema.d.ts`) dùng `openapi-typescript`, type-only, không thay thế runtime `apiClient`.
2. **Stage 1 — sqlc (hoàn tất)**: Đã migrate `assessments`, `admin`, `auth`, và `attempts` repositories sang sqlc generated queries qua wrapper giữ nguyên các `Repository` interfaces. Không thay đổi service/handler contracts.
3. **Stage 2 — Huma (deferred)**: Huma vẫn tạm hoãn cho đến khi OpenAPI maintenance cost vượt quá chi phí rewrite router/handlers sang Huma, hoặc đến khi API contract ổn định hơn sau khi có thêm endpoints. OpenAPI skeleton vẫn được duy trì thủ công trong giai đoạn này.
4. **Stage 3 — Client generation**: Sinh frontend API client/types từ OpenAPI contract sau khi spec ổn định.

## Huma evaluation (post-sqlc)

Sau khi tất cả các repository chính đã được sqlc hóa, Huma vẫn chưa được áp dụng vì lý do sau:

- **Lợi ích**: tự động sinh OpenAPI, kiểm tra request/response schema ở runtime, giảm sai lệch giữa spec và code.
- **Chi phí**: phải chuyển toàn bộ chi router/handlers sang Huma operations, thay đổi cách đăng ký route, response envelope, error shape, và middleware ordering. Rủi ro regression cao với các endpoint nhạy cảm (auth cookie, CSRF, refresh rotation).
- **Ngưỡng tái xem xét**: khi số endpoint vượt quá ~20–25 hoặc chi phí cập nhật skeleton thủ công trở nên đáng kể so với một sprint refactor Huma.

## Consequences

### Positive

- Không rewrite big-bang; giảm rủi ro regression.
- Repository interfaces là seam rõ ràng, dễ test cả hai implementation.
- OpenAPI skeleton vẫn có giá trị ngay cả trước khi Huma được cài đặt.

### Negative

- OpenAPI skeleton phải được cập nhật thủ công cho đến khi Huma xuất hiện.
- sqlc migration đòi hỏi kỷ luật để không lặp lại business logic trong generated code.
