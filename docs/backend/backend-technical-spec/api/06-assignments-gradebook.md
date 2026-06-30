# API 06 — Assignments, Submissions & Gradebook

Base path: `/api/v1`

## 1. Assignment endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| GET | `/classes/{class_id}/assignments` | Danh sách bài tập | — | `{"data":[{"id":"...","title":"Bài tập 1","status":"OPEN"}]}` |
| POST | `/classes/{class_id}/assignments` | Tạo bài tập | `{"title":"Bài tập 1","instructions":{"html":"..."},"due_at":"...","submission_type":"FILE_AND_TEXT","max_score":"10.00"}` | `{"data":{"id":"...","status":"DRAFT"}}` |
| GET | `/assignments/{assignment_id}` | Chi tiết | — | `{"data":{"id":"...","title":"Bài tập 1","settings":{...}}}` |
| PATCH | `/assignments/{assignment_id}` | Cập nhật draft | `{"due_at":"...","allow_resubmission":true}` | `{"data":{"id":"...","revision":2}}` |
| POST | `/assignments/{assignment_id}/publish` | Publish | `{}` + idempotency key | `{"data":{"status":"OPEN"}}` |
| POST | `/assignments/{assignment_id}/close` | Đóng | `{}` | `{"data":{"status":"CLOSED"}}` |
| GET | `/assignments/{assignment_id}/submissions` | Giáo viên xem submissions | — | `{"data":[{"id":"...","student":{...},"status":"SUBMITTED"}]}` |
| GET | `/me/assignments` | Bài tập của học sinh | — | `{"data":[{"id":"...","due_at":"...","submission_status":"DRAFT"}]}` |
| POST | `/assignments/{assignment_id}/submissions` | Tạo/lấy draft submission | `{}` | `{"data":{"id":"...","status":"DRAFT"}}` |
| PATCH | `/submissions/{submission_id}` | Cập nhật draft text | `{"text_content":{"html":"..."}}` | `{"data":{"id":"...","revision":3}}` |
| POST | `/submissions/{submission_id}/files` | Gắn file READY | `{"file_id":"..."}` | `{"data":{"file_id":"..."}}` |
| DELETE | `/submissions/{submission_id}/files/{file_id}` | Gỡ file draft | — | `204 No Content` |
| POST | `/submissions/{submission_id}/submit` | Nộp bài | `{}` + idempotency key | `{"data":{"status":"SUBMITTED","submitted_at":"...","late":false}}` |
| POST | `/submissions/{submission_id}/request-resubmission` | Giáo viên yêu cầu làm lại | `{"reason":"Bổ sung lời giải"}` | `{"data":{"status":"RESUBMISSION_REQUESTED"}}` |
| POST | `/submissions/{submission_id}/grade` | Chấm bài (trả GRADED ngay — legacy) | `{"score":"8.50","feedback":{"format":"SANITIZED_HTML","content":"..."},"rubric_scores":[]}` | `{"data":{"status":"GRADED","score":"8.50"}}` |
| PATCH | `/submissions/{submission_id}/grade-draft` | Lưu draft score/feedback | `{"expected_version":2,"score":"8.50","feedback":{"format":"SANITIZED_HTML","content":"..."}}` | `{"data":{"version":3,"saved_at":"..."}}` |
| POST | `/submissions/{submission_id}/finalize-grade` | Finalize và trả điểm | `{"expected_version":3}` + `Idempotency-Key` | `{"data":{"status":"RETURNED","score":"8.50","finalized_at":"..."}}` |

## 2. Gradebook endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| GET | `/classes/{class_id}/gradebook` | Matrix gradebook (server-paginated) | `?page=1&per_page=50&student_cursor=&item_cursor=` | `{"data":{"items":[...],"students":[...],"entries":[...],"page":{...}}}` |
| GET | `/classes/{class_id}/grade-items` | Grade items | — | `{"data":[{"id":"...","name":"Giữa kỳ","max_score":"10.00"}]}` |
| POST | `/classes/{class_id}/grade-items` | Tạo grade item thủ công | `{"name":"Chuyên cần","max_score":"10.00","category_id":null}` | `{"data":{"id":"..."}}` |
| PATCH | `/grade-items/{grade_item_id}` | Cập nhật item | `{"name":"Chuyên cần HK1"}` | `{"data":{"id":"..."}}` |
| PUT | `/grade-items/{grade_item_id}/entries/{student_user_id}` | Ghi/override điểm | `{"score":"9.00","reason":"Bổ sung minh chứng","expected_version":2}` | `{"data":{"final_score":"9.00","version":3}}` |
| POST | `/classes/{class_id}/grade-publications` | Công bố grade items | `{"grade_item_ids":["..."],"student_user_ids":null}` | `{"data":{"publication_id":"...","published_count":42}}` |
| POST | `/classes/{class_id}/grade-unpublish` | Thu hồi hiển thị nếu policy cho phép | `{"grade_item_ids":["..."]}` | `{"data":{"updated":42}}` |
| GET | `/me/grades` | Học sinh xem điểm đã publish | — | `{"data":[{"grade_item":{"name":"Giữa kỳ"},"score":"8.50","max_score":"10.00"}]}` |
| POST | `/classes/{class_id}/gradebook/exports` | Tạo export CSV | `{"format":"CSV"}` | `202 {"data":{"job_id":"..."}}` |
| GET | `/gradebook/exports/{job_id}` | Trạng thái/download | — | `{"data":{"status":"COMPLETED","file_id":"..."}}` |

## 3. Submission rules

- Submission logical record có nhiều versions.
- Submit tạo immutable `submission_version` snapshot của text + files.
- Draft sau submit chỉ tạo khi resubmission được phép/yêu cầu.
- Nộp trễ xác định bằng server time và accommodation/extension.
- File phải `READY` và thuộc actor/org trước khi attach.

## 4. Grade representation

```json
{
  "raw_score": "8.50",
  "override_score": null,
  "final_score": "8.50",
  "max_score": "10.00",
  "status": "DRAFT",
  "version": 2
}
```

Không dùng JSON number cho decimal quan trọng trong API; dùng string decimal để tránh sai số client.

## 5. Grade update transaction

1. Authorize actor trên class/grade item.
2. Lock grade entry hoặc dùng version optimistic check.
3. Validate score trong range hoặc policy cho extra credit.
4. Insert `grade_entry_history` với before/after và reason.
5. Update current grade entry.
6. Append audit log.
7. Optionally enqueue notification sau publish, không phải mỗi draft update.

## 6. Source-linked grade items

`grade_items` có thể trỏ:

- Assessment.
- Assignment.
- Manual item.

Source score update phải idempotent và không overwrite manual override trừ policy explicit.
