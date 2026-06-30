# ADR-0011: Breached-Password Provider Deferred

- **Status:** Accepted
- **Date:** 2026-06-30

## Context

Hệ thống auth đã có các biện pháp bảo vệ mật khẩu cơ bản:

- Password policy (tối thiểu 8 ký tự, hỗn hợp hoa/thường/số, blocklist các mật khẩu phổ biến).
- Password history (5 hash gần nhất) để ngăn tái sử dụng.
- Login lockout (5 lần sai trong 15 phút) để giảm brute-force.

Có thể bổ sung kiểm tra mật khẩu bị rò rỉ qua dịch vụ bên ngoài (ví dụ Have I Been Pwned) hoặc corpus nội bộ.

## Decision

Tích hợp breached-password provider (HIBP API, local corpus, v.v.) được **hoãn lại** cho đến khi có ADR riêng về privacy và egress/ops.

Lý do:

- **Privacy**: gửi hash/password (kể cả k-anonymity prefix) ra ngoài tổ chức cần được pháp chế/bảo mật đánh giá, đặc biệt với dữ liệu học sinh.
- **Ops**: self-hosted corpus đòi hỏi storage, cập nhật và quét; chưa có nhu cầu bắt buộc ở giai đoạn MVP.
- **Hiệu quả**: password history + lockout + blocklist đã đủ giảm thiểu rủi ro cơ bản trong ngắn hạn.

## Consequences

### Positive

- Không tăng dependency hoặc network egress.
- Không phải xử lý dữ liệu nhạy cảm với bên thứ ba.
- Triển khai nhanh, không phụ thuộc vào SLA của external service.

### Negative

- Không phát hiện mật khẩu đã bị rò rỉ trong các breach công khai.
- Cần theo dõi compliance/security requirement để tái xem xét.

## Next steps

- Theo dõi yêu cầu bảo mật/compliance khi triển khai pilot.
- Khi cần, so sánh HIBP k-anonymity API với local breach corpus (ví dụ `haveibeenpwned-offline`) trong ADR privacy/ops riêng.
