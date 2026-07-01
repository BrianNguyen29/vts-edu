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
4. **Stage 3 — Typed client (adopted)**: Đã thêm `openapi-fetch` runtime dependency, tạo wrapper `apps/web/src/shared/api/openapi-client.ts` với middleware auth/CSRF, và migrate toàn bộ các helper frontend hiện có sang typed client:
   - `attempts.ts`: `listAssignedAssessments`, `startAttempt`, `getAttempt`, `saveAnswer`, `submitAttempt`.
   - `admin.ts`: `getOrganization`, `updateOrganization`, `listUsers`, `createUser`, `updateUserRoles`, `resetUserPassword`, `listAuditLogs`.
   - `assessments.ts`: `listAssessments`, `createAssessment`, `getAssessment`, `updateAssessment`, `createSection`, `createItem`, `createTarget`, `validateAssessment`, `publishAssessment`, `listQuestions`, `updateSection`, `deleteSection`, `reorderSections`, `updateItem`, `deleteItem`, `reorderItems`, `deleteTarget`, `listPublications`.
   - `academics.ts`: `listClasses`, `listEnrollments` (từ lần migrate trước).
   - `apiClient` vẫn giữ lại làm fallback cho các helper chưa migrate hoặc khi schema chưa sẵn sàng; Huma vẫn tạm hoãn.

## openapi-fetch evaluation (Stage 3)

`openapi-fetch` được đánh giá tích cực cho frontend runtime client:

- **Lợi ích**: type-safe paths/parameters/body/response, nhỏ gọn (~6 kb min), dùng native `fetch`, dễ kết hợp middleware cho bearer token và CSRF.
- **Hạn chế**: middleware chạy theo request; cần đảm bảo `credentials: 'include'` và async CSRF fetch tương thích với flow hiện tại.
- **Quyết định**: migrate toàn bộ frontend helpers sang `openapi-fetch`, bao gồm cả GET và POST/PUT/PATCH/DELETE. CSRF middleware lấy token tự động qua `GET /auth/csrf-token` trước mỗi unsafe request và gửi kèm `X-CSRF-Token`; `credentials: 'include'` được đặt trên mọi request. Existing `apiClient` vẫn giữ lại như fallback.

## Huma evaluation (post-sqlc)

Sau khi tất cả các repository chính đã được sqlc hóa, Huma vẫn chưa được áp dụng vì lý do sau:

- **Lợi ích**: tự động sinh OpenAPI, kiểm tra request/response schema ở runtime, giảm sai lệch giữa spec và code.
- **Chi phí**: phải chuyển toàn bộ chi router/handlers sang Huma operations, thay đổi cách đăng ký route, response envelope, error shape, và middleware ordering. Rủi ro regression cao với các endpoint nhạy cảm (auth cookie, CSRF, refresh rotation).
- **Ngưỡng tái xem xét**: khi số endpoint vượt quá ~20–25 hoặc chi phí cập nhật skeleton thủ công trở nên đáng kể so với một sprint refactor Huma.

## Huma revisit (post-academic/openapi-fetch)

Sau khi hoàn thành batch academics và migrate toàn bộ frontend helpers sang `openapi-fetch`, nhóm đã đo lại chi phí duy trì skeleton thủ công:

- **Kích thước spec hiện tại**: OpenAPI skeleton đang định nghĩa **44 paths** (tính theo `paths:` trong `openapi-skeleton.yaml`), vượt xa ngưỡng 20–25 paths ban đầu.
- **Chi phí thực tế**: mỗi lần thêm/sửa endpoint vẫn phải cập nhật tay cả YAML, sqlc queries, Go handlers/services, và generated TypeScript types. Tuy nhiên, quy trình này hiện đã chạy ổn định; `pnpm api:types` và `pnpm api:sqlc` được gọi trong CI, và `openapi-fetch` cung cấp type-safety ở frontend mà không đòi hỏi Huma.
- **Rủi ro migration**: auth cookie, CSRF double-submit, refresh rotation, và response envelope tùy chỉnh vẫn là những điểm cần mapping cẩn thận khi chuyển sang Huma; chi phí refactor ước tính vẫn cao hơn chi phí duy trì skeleton ít nhất là một sprint.

### Decision

- **Huma runtime migration vẫn tạm hoãn.** OpenAPI skeleton được duy trì thủ công; `openapi-typescript` + `openapi-fetch` đã đáp ứng đủ nhu cầu type-safety ở frontend.
- **Ngưỡng tái xem xét tiếp theo**: kích hoạt lại đánh giá Huma khi xảy ra một trong các điều kiện sau:
  1. Spec drift gây lỗi production hoặc lỗi type generation ≥ 2 lần trong một tháng.
  2. Số paths vượt quá **60** (hiện tại ~44), tức là manual maintenance bắt đầu bùng nổ.
  3. Cần tự động kiểm tra request/response schema ở runtime cho API contract phức tạp hơn (ví dụ: public/external API consumer).
  4. Một sprint dành riêng cho refactor được phê duyệt, và có thể viết test coverage ≥ 80% cho các handlers trước khi chuyển đổi.

Nếu kích hoạt, migration sẽ diễn ra theo feature slice (auth → admin → attempts → assessments → academics) thay vì big-bang, và `Repository` interfaces hiện tại sẽ được giữ nguyên.

### Huma revisit (post-builder-polish / student / gradebook / bulk / hardening)

Sau khi hoàn thành các batch gần nhất (builder polish: duplicate section/item + preview; student experience: `/me/attempts`, `/attempts/{id}/result`; gradebook: results/CSV export; bulk operations: CSV import users, bulk enroll/bulk assign teachers; production hardening: rate limit, request ID + structured logging, audit CSV export, Render smoke), nhóm đã đo lại chi phí duy trì skeleton thủ công:

- **Kích thước spec hiện tại**: OpenAPI skeleton đang định nghĩa **58 paths** (tính theo `paths:` trong `openapi-skeleton.yaml`), tăng từ 44 paths ở lần revisit trước và rất gần ngưỡng 60.
- **Tiến bộ giảm chi phí manual maintenance**:
  - `openapi-typescript` + `openapi-fetch` đã cover toàn bộ frontend helpers; type-safety ở frontend không còn phụ thuộc Huma.
  - CI `generated-code-check` (`pnpm api:types`, `pnpm api:sqlc`, `git diff --exit-code`) bắt drift giữa spec và generated code.
  - Request ID và structured logging đã được thêm mà không cần Huma, giảm một trong các lý do ban đầu để dùng Huma (observability/validation).
- **Rủi ro migration vẫn cao**: auth cookie/CSRF double-submit, refresh rotation, response envelope tùy chỉnh, và rate-limit middleware ordering vẫn đòi hỏi mapping cẩn thận khi chuyển sang Huma operations. Chi phí refactor ước tính vẫn lớn hơn chi phí duy trì skeleton ít nhất một sprint.

### Decision (updated)

- **Huma runtime migration vẫn tạm hoãn.** OpenAPI skeleton được duy trì thủ công; `openapi-typescript` + `openapi-fetch` vẫn đáp ứng đủ nhu cầu type-safety ở frontend.
- **Ngưỡng tái xem xét tiếp theo** (điều chỉnh theo tốc độ tăng paths):
  1. Spec drift gây lỗi production hoặc lỗi type generation ≥ 2 lần trong một tháng.
  2. Số paths vượt quá **60** (hiện tại **58**), tức là manual maintenance sắp bùng nổ.
  3. Cần tự động kiểm tra request/response schema ở runtime cho API contract phức tạp hơn (ví dụ: public/external API consumer).
  4. Một sprint dành riêng cho refactor được phê duyệt, và có thể viết test coverage ≥ 80% cho các handlers trước khi chuyển đổi.

Nếu kích hoạt, migration vẫn theo feature slice, bắt đầu từ handler ít nhạy cảm nhất (academics hoặc gradebook) và để auth cuối cùng.

## Consequences

### Positive

- Không rewrite big-bang; giảm rủi ro regression.
- Repository interfaces là seam rõ ràng, dễ test cả hai implementation.
- OpenAPI skeleton vẫn có giá trị ngay cả trước khi Huma được cài đặt.

### Negative

- OpenAPI skeleton phải được cập nhật thủ công cho đến khi Huma xuất hiện.
- sqlc migration đòi hỏi kỷ luật để không lặp lại business logic trong generated code.
