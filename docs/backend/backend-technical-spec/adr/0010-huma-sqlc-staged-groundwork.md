# ADR-0010: Staged Huma/sqlc/OpenAPI Groundwork

- **Status:** Accepted
- **Date:** 2026-06-30

## Context

Backend đã có các feature slices ổn định (auth, attempts, assessments, admin) với repository interfaces rõ ràng và migrations tuần tự. Việc áp dụng Huma (OpenAPI-first handlers) và sqlc (generated queries) là hữu ích cho dài hạn nhưng không nên rewrite runtime code một lúc vì rủi ro regression cao và làm chậm delivery.

## Decision

Tiếp cận từng giai đoạn:

1. **Hiện tại**: Duy trì OpenAPI skeleton bằng tay (`docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml`) phản ánh đúng API surface đã implement. Các `Repository` interfaces trong từng feature package (`auth`, `attempts`, `assessments`, `admin`) là seam ổn định để sau này thay thế implementation.
2. **Stage 1 — sqlc**: Tạo sqlc generated queries và một repository implementation mới cho từng feature, chạy song song với implementation cũ. Di chuyển từng feature một qua test coverage đầy đủ. Không thay đổi service/handler contracts.
3. **Stage 2 — Huma**: Khi sqlc đã ổn, định nghĩa Huma operations cho từng endpoint hoặc thay thế dần handler wiring. OpenAPI spec sẽ được generate tự động từ Huma và đồng bộ với skeleton.
4. **Stage 3 — Client generation**: Sinh frontend API client/types từ OpenAPI contract sau khi spec ổn định.

## Consequences

### Positive

- Không rewrite big-bang; giảm rủi ro regression.
- Repository interfaces là seam rõ ràng, dễ test cả hai implementation.
- OpenAPI skeleton vẫn có giá trị ngay cả trước khi Huma được cài đặt.

### Negative

- OpenAPI skeleton phải được cập nhật thủ công cho đến khi Huma xuất hiện.
- sqlc migration đòi hỏi kỷ luật để không lặp lại business logic trong generated code.
