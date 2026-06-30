# 11. Testing Strategy

## 1. Test pyramid

```text
Few:    End-to-end / load / security
Some:   Integration tests with real PostgreSQL
Many:   Unit tests for domain and application policies
```

Không mock PostgreSQL behavior quá mức cho transaction/concurrency.

## 2. Unit tests

Ưu tiên:

- State transition.
- Score calculation.
- Time window.
- Permission policy.
- Question validation.
- Attempt deadline/accommodation.
- Grade final score selection.
- Refresh token family logic.

Dùng table-driven tests.

Ví dụ:

```go
func TestAttemptSubmitTransitions(t *testing.T) {
    tests := []struct {
        name    string
        status  Status
        wantErr error
    }{
        {"in progress", StatusInProgress, nil},
        {"already submitted", StatusSubmitted, nil}, // idempotent domain behavior
        {"terminated", StatusTerminated, ErrInvalidState},
    }
    // ...
}
```

## 3. Integration tests

Chạy PostgreSQL thật bằng test container hoặc ephemeral DB.

Test:

- Migrations.
- sqlc queries.
- Unique/foreign/check constraints.
- Transaction rollback.
- `FOR UPDATE` concurrency.
- Revision conflict.
- Tenant scoping.
- River transactional enqueue.
- Refresh rotation race.

Mỗi test cô lập bằng transaction rollback hoặc database/schema riêng; phải tương thích với River/multi-connection behavior.

## 4. HTTP/API tests

- `httptest.Server` hoặc handler trực tiếp.
- Validate status, headers, response schema.
- Auth missing/invalid.
- Permission matrix.
- Error Problem Details.
- Idempotency replay.
- Pagination/filter.

## 5. End-to-end scenarios

### Teacher → assessment → student → grade

1. Admin tạo teacher/student.
2. Teacher tạo class và enroll student.
3. Teacher tạo/publish question.
4. Teacher tạo/publish assessment.
5. Student start, save, refresh page, resume, submit.
6. Worker grade.
7. Teacher review essay nếu có.
8. Teacher publish grade.
9. Student xem result.

### Assignment flow

1. Teacher publish assignment.
2. Student upload file, submit.
3. Teacher grade và feedback.
4. Student xem grade đã publish.

## 6. Concurrency tests

Bắt buộc:

- Hai request start attempt cùng lúc.
- Hai answer saves cùng expected revision.
- Hai submit requests cùng lúc.
- Refresh token dùng đồng thời hai lần.
- Hai teacher override cùng grade version.
- Publish assessment duplicate.

Kết quả phải deterministic và không tạo dữ liệu trùng.

## 7. Time tests

Inject fake clock để kiểm tra:

- Trước opens_at.
- Đúng opens_at.
- Gần expires_at.
- Quá hạn.
- Accommodation.
- DST/timezone chỉ ở conversion layer; backend core dùng UTC.

## 8. Load tests

Công cụ có thể dùng k6.

Scenarios:

| Scenario | Pattern |
|---|---|
| Login burst | 300 users trong 1–2 phút |
| Start burst | 300 start gần đồng thời |
| Autosave | Mỗi user 1 request/10–15 giây |
| Submit spike | 200 submit trong 30 giây |
| Gradebook read | 10 giáo viên mở class 50 học sinh |
| Export | 5 export lớn async |

Đo:

- p50/p95/p99.
- Error rate.
- DB CPU/connections/locks.
- Queue latency.
- Slow queries.

## 9. Security tests

Xem checklist [09-security.md](09-security.md). Thêm automated negative tests cho từng endpoint có resource ID.

## 10. Migration tests

CI:

1. Start empty PostgreSQL.
2. `goose up`.
3. `sqlc vet/generate`.
4. Run integration tests.
5. Optional down/up cho migration reversible.
6. Kiểm tra schema diff không ngoài ý muốn.

## 11. Test data

- Fixtures không dùng dữ liệu học sinh thật.
- Random/factory data deterministic bằng seed.
- Có organization A/B để test isolation.
- Có teacher assigned/unassigned.
- Có attempts ở mọi state.

## 12. Coverage policy

Không đặt mục tiêu coverage tổng cao vô nghĩa. Bắt buộc coverage thực chất cho:

- Domain transitions.
- Grading.
- Auth/token.
- Authorization.
- Attempt save/submit.
- Grade override.

PR thay invariant phải thêm test failure-before/fix-after.
