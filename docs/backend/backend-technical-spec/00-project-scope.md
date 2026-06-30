# 00. Project Scope & Non-negotiable Invariants

## 1. Product context

Nền tảng phục vụ hai actor chính:

- **Giáo viên:** quản lý lớp, học sinh, học liệu, câu hỏi, đề kiểm tra, bài tập và điểm.
- **Học sinh:** truy cập bằng tài khoản được cấp, xem học liệu, làm bài tập, thi trực tuyến và xem kết quả đã công bố.

Actor mở rộng sau MVP:

- Organization administrator.
- Teaching assistant.
- Parent/guardian.
- System operator.

## 2. MVP backend scope

| Module | MVP | Ghi chú |
|---|:---:|---|
| Authentication & sessions | Có | JWT access token + rotating opaque refresh token |
| Organizations/tenancy | Có, mức nền tảng | Một database, shared schema, `organization_id` |
| Users & roles | Có | Admin, teacher, student |
| Classes & enrollment | Có | Course, class section, enrollment |
| Resources/files | Có | Metadata + S3-compatible storage |
| Question bank | Có | Versioning bắt buộc |
| Assessment builder | Có | Fixed items, shuffle, schedule |
| Online attempts | Có | Autosave, resume, submit, expire |
| Auto/manual grading | Có | Background job + teacher review |
| Assignments/submissions | Có | Text/file submission |
| Gradebook | Có, cơ bản | Grade items, entries, publish, audit |
| Notifications | Có, in-app | Email là job tùy cấu hình |
| Audit log | Có | Dữ liệu nhạy cảm và thao tác quản trị |
| AI assistant | Không | Phase sau |
| Gamification | Không | Phase sau |
| QTI/OneRoster/LTI | Không | Schema phải không cản trở tích hợp sau |

## 3. Out of scope giai đoạn đầu

- Microservices.
- Kubernetes.
- Redis cluster.
- Kafka/NATS/RabbitMQ.
- Search engine riêng.
- Mobile native.
- Webcam proctoring hoặc nhận diện khuôn mặt.
- AI tự động công bố điểm.
- Billing SaaS.
- Parent portal.
- Live video.

## 4. Non-negotiable invariants

### 4.1. Tenant isolation

- Mọi bảng nghiệp vụ phải có `organization_id`, trừ bảng global/reference được liệt kê rõ.
- Mọi query đọc/ghi tài nguyên tenant phải scope theo `organization_id`.
- Không tin `organization_id` từ request body; lấy từ authenticated context hoặc resource đã kiểm tra.

### 4.2. Question immutability

- `questions` là logical identity.
- Nội dung nằm trong `question_versions`.
- Version đã được dùng bởi assessment đã publish không được update tại chỗ.
- Sửa câu hỏi tạo version mới.

### 4.3. Assessment snapshot

- Khi assessment được publish, mỗi item phải được snapshot.
- Attempt chấm theo snapshot, không chấm theo question version hiện tại.
- Publish là thao tác transactional và idempotent.

### 4.4. Attempt integrity

- Server time quyết định start, expiry và submit validity.
- Một attempt không được từ trạng thái terminal quay lại `IN_PROGRESS`.
- `submit` phải idempotent.
- Answer save phải có revision/version để chống ghi đè cũ.
- Client nhận HTTP 2xx mới được coi là server đã xác nhận lưu.

### 4.5. Grade integrity

- Không xóa lịch sử điểm.
- Override phải lưu điểm cũ, điểm mới, actor, lý do và timestamp.
- Chỉ grade đã publish mới hiển thị cho học sinh.
- Dùng `NUMERIC`, không dùng floating point cho điểm.

### 4.6. File integrity

- Binary file không lưu trong PostgreSQL.
- File object là private mặc định.
- Download sử dụng signed URL ngắn hạn hoặc proxy có authorization.
- Không render trực tiếp HTML/SVG do người dùng tải lên.

### 4.7. Auditability

Audit bắt buộc cho:

- Login thất bại đáng chú ý, reset password, revoke session.
- Tạo/khóa/xóa logic tài khoản.
- Publish/unpublish assessment.
- Start/submit/terminate attempt.
- Manual grade và grade override.
- Export dữ liệu học sinh/điểm.
- Thay đổi role hoặc enrollment.

## 5. Success criteria cho pilot

| Chỉ tiêu | Mục tiêu |
|---|---:|
| Người thi đồng thời | 100–300 |
| Autosave p95 | < 800 ms trong điều kiện pilot |
| API thông thường p95 | < 500 ms |
| Mất answer đã xác nhận | 0 |
| RPO | ≤ 5 phút cho dữ liệu chung |
| RTO | ≤ 60 phút |
| Critical/High security issue | 0 trước production |

Các chỉ tiêu là target kiểm thử, không phải cam kết nếu chưa có benchmark hạ tầng thật.
