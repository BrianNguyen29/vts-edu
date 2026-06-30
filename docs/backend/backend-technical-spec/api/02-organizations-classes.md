# API 02 — Organizations, Academic Terms, Classes & Enrollments

Base path: `/api/v1`

## 1. Endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| GET | `/organizations/current` | Organization hiện tại | — | `{"data":{"id":"...","name":"Trường A","settings":{...}}}` |
| PATCH | `/organizations/current` | Cập nhật settings được phép | `{"name":"Trường A","timezone":"Asia/Ho_Chi_Minh"}` | `{"data":{"id":"...","name":"Trường A"}}` |
| GET | `/academic-terms` | Danh sách học kỳ | — | `{"data":[{"id":"...","name":"HK1 2026","status":"ACTIVE"}]}` |
| POST | `/academic-terms` | Tạo học kỳ | `{"name":"HK1 2026","starts_at":"...","ends_at":"..."}` | `{"data":{"id":"...","status":"DRAFT"}}` |
| PATCH | `/academic-terms/{term_id}` | Cập nhật học kỳ | `{"name":"HK1 2026-2027"}` | `{"data":{"id":"..."}}` |
| GET | `/subjects` | Danh sách môn | — | `{"data":[{"id":"...","code":"MATH","name":"Toán"}]}` |
| POST | `/subjects` | Tạo môn | `{"code":"MATH","name":"Toán"}` | `{"data":{"id":"..."}}` |
| GET | `/courses` | Danh sách course | — | `{"data":[{"id":"...","name":"Toán 8"}]}` |
| POST | `/courses` | Tạo course | `{"subject_id":"...","name":"Toán 8","grade_level":"8"}` | `{"data":{"id":"..."}}` |
| GET | `/classes` | Danh sách lớp actor có quyền xem | — | `{"data":[{"id":"...","name":"8A1","student_count":42}]}` |
| POST | `/classes` | Tạo class section | `{"course_id":"...","academic_term_id":"...","name":"Toán 8A1"}` | `{"data":{"id":"...","status":"ACTIVE"}}` |
| GET | `/classes/{class_id}` | Chi tiết lớp | — | `{"data":{"id":"...","teachers":[...],"student_count":42}}` |
| PATCH | `/classes/{class_id}` | Cập nhật metadata lớp | `{"name":"Toán 8A1 - HK1"}` | `{"data":{"id":"..."}}` |
| POST | `/classes/{class_id}/archive` | Archive lớp | `{}` | `204 No Content` |
| GET | `/classes/{class_id}/teachers` | Giáo viên/trợ giảng | — | `{"data":[{"user_id":"...","role":"TEACHER"}]}` |
| POST | `/classes/{class_id}/teachers` | Gán giáo viên | `{"user_id":"...","role":"TEACHER"}` | `{"data":{"user_id":"...","role":"TEACHER"}}` |
| DELETE | `/classes/{class_id}/teachers/{user_id}` | Gỡ giáo viên | — | `204 No Content` |
| GET | `/classes/{class_id}/enrollments` | Danh sách học sinh | — | `{"data":[{"id":"...","student":{"id":"...","display_name":"..."},"status":"ACTIVE"}]}` |
| POST | `/classes/{class_id}/enrollments` | Ghi danh học sinh | `{"student_user_id":"..."}` | `{"data":{"id":"...","status":"ACTIVE"}}` |
| POST | `/classes/{class_id}/enrollments/bulk` | Ghi danh hàng loạt | `{"student_user_ids":["..."],"dry_run":false}` | `{"data":{"accepted":40,"rejected":2,"errors":[...]}}` |
| PATCH | `/classes/{class_id}/enrollments/{enrollment_id}` | Đổi trạng thái | `{"status":"WITHDRAWN","reason":"transferred"}` | `{"data":{"status":"WITHDRAWN"}}` |
| GET | `/classes/{class_id}/groups` | Nhóm học sinh | — | `{"data":[{"id":"...","name":"Nhóm 1"}]}` |
| POST | `/classes/{class_id}/groups` | Tạo nhóm | `{"name":"Nhóm 1","member_user_ids":["..."]}` | `{"data":{"id":"..."}}` |
| PUT | `/classes/{class_id}/groups/{group_id}/members` | Thay members | `{"member_user_ids":["...","..."]}` | `{"data":{"member_count":2}}` |

## 2. Authorization rules

- `org_admin`: quản lý toàn bộ classes trong organization.
- `teacher`: xem/chỉnh class họ được gán, theo permission cụ thể.
- `teaching_assistant`: quyền hạn hẹp hơn, không mặc định quản lý enrollment.
- `student`: chỉ xem class đang enrollment active.

Cross-tenant ID phải trả 404 để giảm resource enumeration.

## 3. Enrollment invariants

- User phải có active membership trong organization.
- Chỉ user có role/capability student mới được enroll làm học sinh.
- Không tạo hai active enrollment cùng class/student.
- Withdraw không xóa attempt/grade lịch sử.
- Re-enroll tạo record mới hoặc reactivate theo policy; phải giữ timeline rõ.

## 4. Class aggregate notes

Không load toàn bộ 1.000 enrollment vào class entity. Use case dùng repository query theo nhu cầu và batch operations có giới hạn.
