# ADR-0013: Huma Hybrid Architecture — Huma for JSON CRUD, chi for Binary / Streaming / Auth

- **Status:** Accepted
- **Date:** 2026-07-02
- **Supersedes:** §"Huma revisit (post-builder-polish / student / gradebook / bulk / hardening)" of [ADR-0010](./0010-huma-sqlc-staged-groundwork.md) (the 60-path threshold entry). ADR-0010 vẫn giữ context cho Stage 1 (sqlc) và Stage 3 (typed client) nhưng nội dung về runtime Huma adoption được thay bằng ADR này.

## Context

Tại thời điểm 2026-07-02:

- Bounded Huma feasibility spike trên `academics` (xem [`../spikes/huma-academics-spike.md`](../spikes/huma-academics-spike.md)) đã pass với 4/4 unit test, không regression, `{data,error}` envelope được bảo tồn qua Huma v2.38 với cấu hình không mặc định (`CreateHooks = nil`, `Transformers = nil`), auth/CSRF/role checks tích hợp sạch qua helper hiện có. **Verdict: GO có điều kiện** với 3 open issues cần giải quyết trước khi runtime migration toàn cục.
- Huma spike report cũng đề xuất một follow-up spike trên streaming/resources download để kiểm tra Huma với `http.Flusher`/SSE trước khi đề xuất full adoption.
- OpenAPI skeleton tay đang cover **~63 paths** (đã vượt ngưỡng 60 của ADR-0010), tốc độ tăng vẫn tiếp tục theo từng slice (resources MVP, accessibility polish, notifications, slice-15, slice-18…).
- Production đã có **5 loại endpoint nhạy cảm với streaming/binary** đang chạy trên chi:
  1. **Binary download**: `GET /resources/{id}/download?file_id=<uuid>&disposition=inline` (sanitized `Content-Disposition`, `X-Content-Type-Options: nosniff`, content-type allowlist; inline chỉ cho image/pdf/text).
  2. **Multipart upload**: `POST /resources/{id}/files` (accept `file` + `files[]` + `files`, max 10 MiB, path-traversal safe key generation, server-side multipart parsing).
  3. **CSV export streaming**: `GET /assessments/{id}/attempts.csv` + `GET /classes/{id}/gradebook.csv` (large row counts, mỗi row ghi một dòng, không buffer toàn bộ trong bộ nhớ).
  4. **Cookie + CSRF + refresh rotation**: `/api/v1/auth/*` (login, refresh, logout, change-password, CSRF double-submit, refresh cookie `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth`).
  5. **CSRF header echo cho unsafe methods** trên mọi route không phải GET/HEAD/OPTIONS (qua `csrf.Validate(r)`).

Đánh giá kỹ thuật cho thấy việc đẩy tất cả các loại endpoint trên qua Huma operations sẽ phải đối mặt với:

- **Huma v2 không có first-class streaming API.** Huma operations xây dựng response dựa trên `Body` struct + `Status` int; việc ghi incremental bytes cho CSV export lớn hoặc proxy binary download phải dùng `http.Flusher` / `http.ResponseWriter` thủ công, triệt tiêu lợi ích "request/response validation tự động" của Huma.
- **Huma `Body` type cho multipart upload** không hỗ trợ streaming parse; việc tích hợp `mime/multipart` của stdlib buộc handler phải nhận `*http.Request` qua `huma.Context` thay vì qua Huma input struct, làm mất DX chính mà Huma mang lại.
- **Huma default error content type** (`application/problem+json`, RFC 9457) khác production (`application/json` + envelope `{error:{code,message,request_id}}`). Resource download endpoint đã trả về binary stream với `Content-Disposition` riêng; thay đổi content type cho lỗi download sẽ ảnh hưởng browser download flow.
- **Cookie + CSRF middleware** đang được áp dụng ở chi router level; việc nhúng các middleware này vào Huma operations đòi hỏi custom `huma.Operation` config + `CreateHooks` cho từng route, tăng đáng kể boilerplate.
- **Refresh rotation** cần kiểm soát chính xác response headers (`Set-Cookie` với `HttpOnly; Secure; SameSite=None; Path=/api/v1/auth`, max-age) trước khi Huma ghi body. Trong Huma v2.38, việc set response header trong handler gần như phải dùng `huma.Context` hoặc middleware ngoài, không thuận cho route registration theo khai báo.

Một số open issues còn lại từ academics spike cũng cần được giải quyết trước khi runtime migration:

1. `application/problem+json` content type cho error responses (Huma mặc định khác production).
2. `X-Request-Id` response header chưa được echo bởi chi middleware (spike chỉ pass trong context).
3. OpenAPI spec divergence: skeleton tay vs. Huma-generated.

## Decision

Áp dụng **kiến trúc hybrid** cho Huma adoption. Quyết định cụ thể:

### 1. Huma cho JSON CRUD endpoints

- Áp dụng Huma v2 cho các endpoint JSON CRUD ít nhạy cảm: `academics` (terms, classes, enrollments), `gradebook` (results list, CSV export bằng cách buffer nhỏ), `assessments` (list/get/create/update), `attempts` (list/get), `admin` (org + users + roles, trừ reset password), `notifications` (`/me/notifications` family), `question-banks` (banks/questions/versions CRUD).
- Giữ nguyên `Repository` interfaces và `{data,error}` response envelope; chỉ thay lớp route registration và request/response schema validation.
- Migration theo feature slice, thứ tự ưu tiên (từ ít nhạy cảm → nhiều nhạy cảm): `academics` → `gradebook` → `notifications` → `question-banks` → `assessments` → `attempts` → `admin` (trừ reset-password) → `auth` (cuối cùng, có thể không bao giờ migrate).
- Cấu hình Huma per-feature phải disable `$schema` embed (`cfg.CreateHooks = nil`, `cfg.Transformers = nil`) và dùng concrete response types (Status int + Body struct) theo kinh nghiệm từ spike.

### 2. Giữ chi cho các endpoint nhạy cảm với binary / streaming / cookie

Các endpoint sau **không migrate** sang Huma trong giai đoạn này, trừ khi Huma bổ sung first-class streaming API:

- **Binary download**: `GET /resources/{id}/download` (mọi disposition). Lý do: cần `http.ServeContent` / `http.Flusher` + `Content-Disposition` + content-type allowlist + `X-Content-Type-Options: nosniff` từ middleware. Huma body schema không phù hợp với byte stream.
- **Multipart upload**: `POST /resources/{id}/files` (và các upload tương lai). Lý do: cần parse streaming multipart, kiểm soát memory limit (10 MiB), path-traversal safe key generation, và validation từng phần; Huma input struct không khớp với `mime/multipart` streaming parse.
- **CSV export**: `GET /assessments/{id}/attempts.csv`, `GET /classes/{id}/gradebook.csv`. Lý do: ghi incremental bytes, có thể đạt hàng trăm nghìn rows; Huma buffer toàn bộ response trước khi ghi sẽ tốn memory. Khi export dataset nhỏ (< 1 000 rows) có thể buffer và dùng Huma với `Content-Type: text/csv` + `Content-Disposition: attachment`, nhưng spike đó phải chứng minh được với openapi-typescript rằng response là binary blob, không phải JSON.
- **Auth cookie + CSRF + refresh rotation**: toàn bộ `/api/v1/auth/*` (login, refresh, logout, change-password, CSRF token, register organization). Lý do: cần kiểm soát `Set-Cookie` headers chính xác, double-submit CSRF (`X-CSRF-Token` ↔ `vts_csrf` cookie), refresh rotation với invalidation. Huma operation registration che giấu `http.ResponseWriter` ở mức không đủ chi tiết.
- **CSRF header echo cho unsafe methods**: giữ middleware ở chi router level; Huma operations kế thừa qua sub-router mounting.

### 3. Huma KHÔNG dùng cho SSE / WebSocket / Server-Sent Events

- Chưa có endpoint SSE/WS trong production; nếu thêm vào slice tương lai (ví dụ: real-time attempt monitoring), phải chạy trên chi cho đến khi Huma bổ sung first-class streaming API. Đây là ràng buộc kiến trúc, không phải quyết định từng route.

### 4. Điều kiện tiên quyết (preconditions) trước khi runtime migration toàn cục

Cả 3 open issues từ academics spike phải được giải quyết trước khi approve migration slice đầu tiên (`academics`):

1. **`application/problem+json` content type**: chọn một trong hai — (a) update client (`openapi-fetch` wrapper + frontend error handler) để expect `problem+json` cho error responses, hoặc (b) override Huma default error content type registration về `application/json` + envelope `{error:{code,message,request_id}}` của production. Ưu tiên (b) để giữ backward compat với client hiện tại (đã parse `ApiResponseError` dựa trên `application/json`).
2. **X-Request-Id response header echo**: thêm middleware Huma (qua `huma.NewError` hook hoặc custom `OperationConfig.Middlewares`) để đọc `X-Request-Id` từ `chi`-level middleware context (`middleware.RequestID` đã đặt vào context) và set response header. Tương tự `apps/api/internal/platform/middleware/response.go::respondError` hiện tại.
3. **OpenAPI spec divergence**: hai lựa chọn — (a) giữ skeleton tay làm source of truth, dùng Huma-generated spec như evidence only, hoặc (b) absorb Huma-generated spec vào skeleton (Huma sinh ra spec cho slice đã migrate, skeleton bù phần còn lại). Ưu tiên (a) cho đến khi (b) có tool hỗ trợ merge tự động; `generated-code-check` CI sẽ fail nếu hai spec diverge.

Mỗi precondition cần một bounded sub-spike (≤ 1 ngày, có test, có rollback) trước khi migration chính thức bắt đầu. Pre-spike kết quả phải được review bởi người không tham gia spike đó.

### 5. OpenAPI skeleton thủ công tiếp tục là source of truth

- `docs/backend/backend-technical-spec/openapi/openapi-skeleton.yaml` vẫn là source of truth cho `openapi-typescript` cho đến khi (3) trong mục 4 được giải quyết.
- Huma-generated OpenAPI cho mỗi slice là **evidence only**, không commit vào skeleton trong giai đoạn này.
- Khi migration hoàn tất và (3) được giải quyết, skeleton sẽ được regenerate từ Huma (với các bù cho endpoint không migrate: binary download, multipart upload, CSV export, auth cookie).

### 6. Tiêu chí go/no-go cho mỗi slice migration

- *Go* nếu: (a) 3 preconditions ở mục 4 đã pass cho slice đó, (b) handler test coverage ≥ 80% trước khi chuyển, (c) không regression ở smoke + e2e, (d) thời gian thêm endpoint mới trong slice giảm ≥ 30% so với skeleton tay (đo qua hai endpoint mới trong 2 tuần liên tiếp).
- *No-go / pause* nếu: (a) precondition nào fail, (b) regression ở auth/CSRF/middleware ordering, (c) team không cam kết test coverage ≥ 80%.

## Hệ quả

### Tích cực

- **Giảm risk profile**: binary/streaming/auth — những phần khó nhất của Huma adoption — bị cô lập trên chi, nơi đã ổn định với handler hiện tại. Migration tập trung vào JSON CRUD, nơi Huma validation và OpenAPI generation thực sự tỏa sáng.
- **Tận dụng bounded-spike evidence**: spike đã chứng minh `{data,error}` envelope preservable và auth/CSRF/role checks integrate sạch với helper hiện có trên academics — đúng loại endpoint mà ADR này cho phép migrate.
- **Independent failure domains**: nếu một slice JSON migration gặp vấn đề (ví dụ: `assessments` publish flow phức tạp), rollback chỉ ảnh hưởng slice đó. Binary/streaming/auth vẫn chạy trên chi.
- **OpenAPI spec tăng chất lượng theo từng slice migrate**: mỗi slice Huma sinh OpenAPI tự động, giảm dần manual maintenance cho skeleton.
- **Không cần thay đổi routing primitives**: chi router vẫn là backbone; Huma mount như sub-router tại path prefix riêng (theo mẫu spike), `Repository` interfaces giữ nguyên, middleware ordering vẫn ở chi level.

### Tiêu cực

- **Hai routing paradigm cùng tồn tại**: code mới phải biết endpoint nào dùng Huma operations, endpoint nào dùng chi handler. Cần code review checklist và CONTRIBUTING note trong AGENTS.md khi migration bắt đầu.
- **OpenAPI spec có hai nguồn** trong giai đoạn chuyển tiếp: skeleton tay cho endpoint chưa migrate, Huma-generated cho endpoint đã migrate. Phải có CI guard (`generated-code-check` hiện tại) để phát hiện divergence.
- **Huma version pin phải được review kỹ**: Huma v2.x vẫn active, mỗi minor bump có thể đổi `DefaultConfig` (đã thấy ở v2.38: `OpenAPI.Path`/`DocsPath` bị xóa, `DefaultConfig` đổi từ value sang function). Có ADR maintenance cost nhỏ.
- **Không có cải thiện DX cho binary/streaming**: các endpoint download/upload/export vẫn phải viết `http.ResponseWriter` boilerplate như cũ. Nếu team muốn cải thiện phần này, cần đợi Huma bổ sung first-class streaming API hoặc chuyển sang framework khác (Echo / Fiber / native `net/http`).
- **CSRF middleware mapping**: cần viết custom Huma middleware để call `csrf.Validate(r)` từ context hiện có; nếu làm sai sẽ bypass CSRF cho Huma operations.

### Trung hòa

- **Repository interfaces không đổi**: Huma handlers vẫn gọi `Service` hiện tại, chỉ thay lớp route registration. Test coverage hiện tại của services giữ nguyên giá trị.
- **Frontend không cần thay đổi**: `openapi-fetch` typed client + skeleton tay tiếp tục là source of truth cho types. Khi (3) mục 4 được giải quyết, regenerate types một lần.

## Không áp dụng cho

- **WebSocket / SSE / long polling**: chưa có trong production; nếu thêm, chạy trên chi cho đến khi Huma bổ sung streaming API. ADR này không khóa tương lai nhưng hiện tại đặt ràng buộc kiến trúc.
- **gRPC / Connect / protobuf**: không nằm trong roadmap; nếu cần public/external API consumer, sẽ mở ADR riêng.

## Liên kết

- [ADR-0010](./0010-huma-sqlc-staged-groundwork.md) — context sqlc, typed client, và bounded spike plan.
- [`../spikes/huma-academics-spike.md`](../spikes/huma-academics-spike.md) — bounded spike evidence.
- [`../14-implementation-roadmap.md`](../14-implementation-roadmap.md) §2.5 + §2.7 — cập nhật theo quyết định này.
- [`../../../implementation-audit.md`](../../../implementation-audit.md) — entry docs-only ghi nhận ADR này.

## Ghi chú triển khai

- Khi migration bắt đầu, mỗi slice cần một bản cập nhật `apps/api/cmd/server/main.go` để mount Huma sub-router trước các chi routes hiện tại (theo mẫu spike). Sub-router phải kế thừa `middleware.RequestID` + CSRF middleware ở chi level; Huma operations đọc `X-CSRF-Token` từ request header qua `huma.Context` hoặc custom middleware inject `*http.Request` vào context.
- Test pattern: copy mẫu 4 test từ academics spike (`TestHumaSpike_*`) cho mỗi slice mới, bổ sung thêm test cho endpoint phức tạp hơn.
- Rollback plan: xóa sub-router mount + giữ nguyên chi routes. Mỗi slice migration phải có thể rollback trong 1 commit.
