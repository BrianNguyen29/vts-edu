# ADR-0012: Background Jobs — In-Process Scheduler and River Defer

- **Status:** Accepted
- **Date:** 2026-07-01

## Context

Hệ thống cần một số tác vụ nền tự động, ví dụ:

- Mở/đóng bài kiểm tra theo lịch (`opens_at` / `closes_at`).
- Trong tương lai: import CSV lớn, chấm điểm bất đồng bộ cho các loại câu hỏi phức tạp hơn MCQ.

River là một lựa chọn khả thi với `pgx/v5` nhưng đòi hỏi migrations riêng, worker process, retry/dead-letter policy, và operational complexity chưa cần thiết ở giai đoạn hiện tại.

## Decision

1. **Hiện tại**: Sử dụng scheduler in-process nhẹ trong `apps/api/internal/platform/scheduler`:
   - Giao diện `Job` đơn giản (`Name`, `Run`).
   - `Scheduler` chạy các job theo khoảng thờI gian cố định bằng `time.Ticker`.
   - Bật/tắt và tần suất qua biến môi trường `SCHEDULER_ENABLED` và `SCHEDULER_INTERVAL_SECONDS`.
   - Job đầu tiên: `assessment-transition` chuyển trạng thái assessment `SCHEDULED`/`PUBLISHED` → `OPEN` khi `opens_at <= now()`, và `OPEN` → `CLOSED` khi `closes_at <= now()`.
2. **River tạm hoãn**: Không thêm dependency River trong slice này.
3. **Giữ nguyên giới hạn hiện tại**:
   - Import CSV đồng bộ với giới hạn 100 dòng.
   - Chấm điểm MCQ đồng bộ trong request submit.

## Consequences

- **Ưu điểm**: Không thêm dependency/migration mới; scheduler đủ cho mở/đóng bài kiểm tra theo lịch; dễ test và debug.
- **Nhược điểm**: Scheduler chạy trong process server; nếu scale-out nhiều instance, các job có thể chạy trùng lặp. Các transition query là idempotent (`UPDATE … WHERE …`) nên trùng lặp không gây lỗi trạng thái, chỉ tạo thêm log/dòng affected.
- **Rủi ro**: Instance crash trong khoảng giữa hai tick có thể làm trễ transition vài chục giây. Điều này chấp nhận được với lịch mở/đóng bài kiểm tra (không yêu cầu millisecond precision).

## Triggers for River adoption

Kích hoạt lại đánh giá River khi xảy ra một trong các điều kiện sau:

1. Cần durability/retry cho job (ví dụ: chấm điểm bất đồng bộ, import CSV lớn).
2. Scale-out nhiều instance và cần tránh duplicate job execution.
3. Cần schedule chính xác hơn hoặc cron-like expression.
4. Một sprint dành riêng cho infrastructure được phê duyệt, bao gồm migration, worker process, monitoring, và dead-letter handling.

## Related decisions

- ADR-0010 (Huma/sqlc staged groundwork): giữ nguyên phong cách tách rõ interface và wrapper.
- ADR-0011 (breached-password provider): tương tự, tạm hoãn external integration cho đến khi có privacy/ops review.
