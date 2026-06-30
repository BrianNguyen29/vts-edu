# API 04 — Question Bank

Base path: `/api/v1`

## 1. Endpoints

| Method | URL | Description | Request Body mẫu | Response mẫu |
|---|---|---|---|---|
| GET | `/question-banks` | Danh sách ngân hàng câu hỏi | — | `{"data":[{"id":"...","name":"Toán 8"}]}` |
| POST | `/question-banks` | Tạo bank | `{"name":"Toán 8","subject_id":"...","visibility":"PRIVATE"}` | `{"data":{"id":"..."}}` |
| GET | `/question-banks/{bank_id}` | Chi tiết bank | — | `{"data":{"id":"...","name":"Toán 8"}}` |
| PATCH | `/question-banks/{bank_id}` | Cập nhật bank | `{"name":"Toán 8 - HK1"}` | `{"data":{"id":"..."}}` |
| GET | `/questions` | Search/filter câu hỏi | — | `{"data":[{"id":"...","current_version":{"type":"SINGLE_CHOICE","status":"DRAFT"}}],"page":{...}}` |
| POST | `/questions` | Tạo logical question + version 1 | `{"bank_id":"...","type":"SINGLE_CHOICE","prompt":{"format":"SANITIZED_HTML","content":"2+2=?"},"choices":[...],"answer_key":{"choice_ids":["c1"]},"max_score":"1.00"}` | `{"data":{"id":"...","current_version_id":"...","version_number":1}}` |
| GET | `/questions/{question_id}` | Chi tiết + version hiện tại | — | `{"data":{"id":"...","current_version":{...}}}` |
| POST | `/questions/{question_id}/versions` | Tạo version mới từ version hiện tại | `{"base_version_id":"...","changes":{...}}` | `{"data":{"version_id":"...","version_number":2,"status":"DRAFT"}}` |
| GET | `/questions/{question_id}/versions` | Lịch sử version | — | `{"data":[{"id":"...","version_number":1,"status":"PUBLISHED"}]}` |
| GET | `/question-versions/{version_id}` | Xem version cụ thể | — | `{"data":{"id":"...","prompt":{...},"answer_key":{...}}}` |
| PATCH | `/question-versions/{version_id}` | Sửa draft version | `{"prompt":{"format":"SANITIZED_HTML","content":"2 + 2 bằng bao nhiêu?"}}` | `{"data":{"id":"...","revision":4}}` |
| POST | `/question-versions/{version_id}/publish` | Publish version | `{}` | `{"data":{"status":"PUBLISHED","published_at":"..."}}` |
| POST | `/questions/{question_id}/archive` | Archive logical question | `{}` | `204 No Content` |
| POST | `/questions/bulk-tags` | Gắn tag hàng loạt | `{"question_ids":["..."],"tag_ids":["..."]}` | `{"data":{"updated":20}}` |
| GET | `/question-tags` | Tags | — | `{"data":[{"id":"...","name":"Hàm số"}]}` |
| POST | `/question-tags` | Tạo tag | `{"name":"Hàm số"}` | `{"data":{"id":"..."}}` |
| POST | `/questions/imports` | Bắt đầu import nội bộ/CSV | `{"file_id":"...","format":"INTERNAL_CSV","dry_run":true}` | `202 {"data":{"job_id":"..."}}` |

## 2. Supported MVP types

```text
SINGLE_CHOICE
MULTIPLE_CHOICE
TRUE_FALSE
SHORT_TEXT
NUMERIC
ESSAY
```

## 3. Canonical rich-content format

Canonical format cho prompt, choices, explanation, feedback, instructions:

```json
{
  "format": "SANITIZED_HTML",
  "content": "<p>2 + 2 = ?</p>",
  "version": 1
}
```

Hoặc nếu chọn AST:

```json
{
  "format": "AST",
  "ast": { ... },
  "version": 1
}
```

Quy tắc:

- Backend quyết định canonical format và chuyển đổi từ frontend editor (TipTap JSON) khi nhận.
- Luôn sanitize HTML theo allowlist trước khi lưu.
- Không render inline SVG/HTML từ user upload.
- Round-trip từ canonical format về renderer phải ổn định.
- Đánh version format để migration.

## 4. Example create question

```json
{
  "bank_id": "019...",
  "type": "MULTIPLE_CHOICE",
  "prompt": {
    "format": "SANITIZED_HTML",
    "content": "Chọn các số nguyên tố"
  },
  "choices": [
    {"id": "c1", "content": "2"},
    {"id": "c2", "content": "4"},
    {"id": "c3", "content": "5"}
  ],
  "answer_key": {
    "choice_ids": ["c1", "c3"]
  },
  "scoring": {
    "max_score": "2.00",
    "partial_credit": "ALL_OR_NOTHING"
  },
  "metadata": {
    "difficulty": "MEDIUM",
    "estimated_seconds": 60
  }
}
```

## 5. Validation by type

| Type | Rules |
|---|---|
| SINGLE_CHOICE | ≥2 choices, đúng 1 correct ID |
| MULTIPLE_CHOICE | ≥2 choices, ≥1 correct ID, scoring policy hợp lệ |
| TRUE_FALSE | answer boolean |
| SHORT_TEXT | normalized accepted answers hoặc manual grading policy |
| NUMERIC | target numeric, tolerance/range rõ |
| ESSAY | không có auto answer key bắt buộc, manual rubric optional |

## 6. Versioning logic

- `PATCH /question-versions/{id}` chỉ cho `DRAFT`.
- Publish version đặt immutable.
- `questions.current_version_id` có thể trỏ latest draft cho author, nhưng assessment selection phải chọn version cụ thể.
- Nếu copy question, tạo logical question mới và provenance metadata.

## 7. Search projection

Rich content JSON được trích plain text vào cột/search table để FTS. Không search trực tiếp toàn bộ JSON nested mỗi request.
