# 10. Exam Runtime Frontend Specification

## 1. Mục tiêu

Exam runner là subsystem có yêu cầu độ tin cậy cao nhất. Mục tiêu:

- Không mất answer đã được server xác nhận.
- Giữ answer chưa xác nhận qua reload/browser crash.
- Server time kiểm soát deadline.
- Submit idempotent.
- UI vẫn rõ ràng khi offline/chậm mạng.
- Không dựa vào service worker như nguồn duy nhất cho sync.

## 2. Exam route

```text
/exam/attempts/:attemptId
```

Dùng `ExamLayout` riêng:

- Không sidebar dashboard.
- Header tối thiểu: tên bài, timer, save status, submit.
- Question navigation.
- Main question area.
- Accessibility controls.

## 3. Pre-start flow

```text
student opens assigned assessment
-> fetch assessment availability summary
-> show rules, duration, attempts remaining
-> capability check (IndexedDB, storage)
-> user confirms start
-> create stable idempotency key
-> POST start attempt
-> receive attempt_id, status, server_time, expires_at (tối thiểu)
-> GET /attempts/{attempt_id} để lấy snapshot metadata đầy đủ
-> initialize IndexedDB attempt record
-> navigate exam route
```

Double click start dùng cùng in-flight mutation/key.

## 4. Attempt bootstrap

```text
load local attempt record
+ GET attempt runtime snapshot
+ reconcile server acknowledged answers/revisions
+ restore pending operations
+ calculate server clock offset
+ start sync worker
+ render questions
```

Nếu server attempt terminal:

- Stop sync.
- Reconcile status.
- Redirect result/status page.
- Không cho tiếp tục sửa.

## 5. State model

```ts
type SaveState =
  | 'idle'
  | 'local-pending'
  | 'syncing'
  | 'saved'
  | 'offline'
  | 'conflict'
  | 'fatal';

interface LocalAnswerState {
  attemptItemId: string;
  payload: AnswerPayload;
  localRevision: number;
  acknowledgedServerRevision: number;
  clientMutationId: string;
  updatedAt: number;
  saveState: SaveState;
}
```

## 6. IndexedDB schema

Database: `lms-exam-runtime`, versioned migrations.

Object stores:

```text
attempts
  key: attemptId
  fields: userId, organizationId, status, expiresAt, updatedAt

answers
  key: [attemptId, attemptItemId]
  fields: payload, localRevision, acknowledgedRevision, updatedAt

operations
  key: operationId
  indexes: attemptId, status, createdAt
  fields: attemptId, attemptItemId, expectedRevision, payload,
          status, retryCount, nextRetryAt, createdAt

metadata
  key: string
```

Không lưu access token.

## 7. Answer write flow

```text
user changes answer
-> normalize/canonicalize payload
-> update in-memory UI immediately
-> transaction IndexedDB:
     save current answer
     enqueue/replace pending operation for attemptItem
-> mark local-pending
-> schedule network sync after debounce 500–1000ms
```

IndexedDB write phải hoàn tất trước khi UI hiển thị “đã lưu cục bộ”.

## 8. Operation coalescing

Với cùng attemptItem chưa gửi:

- Giữ payload mới nhất.
- Tăng local revision.
- Không cần gửi từng keystroke.

Nếu operation đang in-flight, tạo operation kế tiếp hoặc đánh dấu dirty-after-flight.

Essay text có debounce dài hơn lựa chọn trắc nghiệm nhưng phải persist local nhanh.

## 9. Network save request

Request gồm:

```json
{
  "expected_revision": 7,
  "client_mutation_id": "uuid",
  "answer": {
    "type": "SINGLE_CHOICE",
    "choice_ids": ["..."]
  }
}
```

Response:

```json
{
  "data": {
    "attempt_item_id": "...",
    "acknowledged_client_mutation_id": "uuid",
    "revision": 8,
    "saved_at": "2026-06-29T10:30:00Z"
  }
}
```

Ack flow:

```text
2xx with revision=N+1 (server acknowledged new revision)
-> IndexedDB transaction:
   update acknowledged revision
   delete completed operation
-> update memory state saved
```

Chỉ sau transaction local này mới có thể hiển thị “Đã lưu”.

## 10. Conflict handling

Nếu 409 revision conflict:

1. Fetch server answer/revision cho attemptItem hoặc attempt.
2. So sánh local operation timestamp/revision theo protocol backend.
3. Không tự overwrite server nếu ambiguity.
4. Với single active tab policy, conflict thường là duplicate tab; show blocking warning.
5. Cho lựa chọn reload server state hoặc contact support theo exam policy.

Không merge essay text tự động nếu có khả năng mất dữ liệu; giữ local copy để người dùng/support phục hồi.

## 11. Retry and backoff

Retry save queue:

```text
network error/offline
-> operation remains pending
-> exponential backoff with jitter
-> listen online event
-> retry when online and nextRetryAt reached
```

Giới hạn backoff, ví dụ 1s, 2s, 5s, 10s, 20s, tối đa 30s. Khi user bấm submit, flush ưu tiên ngay.

Không retry 401 trực tiếp; API auth layer refresh trước.

## 12. Save status UX

| State | UI |
|---|---|
| `local-pending` | “Đang lưu cục bộ…” |
| `syncing` | “Đang đồng bộ…” |
| `saved` | “Đã lưu lúc HH:mm:ss” |
| `offline` | Banner “Mất kết nối — câu trả lời được giữ trên thiết bị” |
| `conflict` | Blocking alert, không tiếp tục submit âm thầm |
| `fatal` | Hướng dẫn không đóng tab và liên hệ giám sát |

Không dùng màu duy nhất; có icon và text.

## 13. Timer

### Offset calculation

Ghi timestamp trước request và sau response, dùng midpoint để ước lượng offset.

```text
serverNow = Date.now() + offsetMs
remaining = expiresAt - serverNow
```

### Rendering

- Tick UI mỗi 1 giây.
- Logic expiry không phụ thuộc interval chính xác.
- Khi tab background, khi visible lại tính từ clock, không trừ counter cũ.
- Warning ở 10 phút, 5 phút, 1 phút theo cấu hình và không spam screen reader.

### Resync

Heartbeat server có thể trả server time. Nếu offset thay đổi nhỏ, cập nhật. Nếu thay đổi lớn, log và hiển thị stable timer theo server expiry; không kéo dài thời gian client.

## 14. Expiry

Khi clock đạt expiry:

```text
stop accepting new edits
-> persist current local input
-> flush pending best effort
-> POST /attempts/{attempt_id}/submit với idempotency key
-> show finalizing state
```

Backend quyết định accept/expire. Client không tự tuyên bố bài đã nộp thành công nếu chưa có response.

> **Offline/deadline policy (P0-13):** Offline durability không đồng nghĩa được chấp nhận sau deadline. Answer nhập offline gần/qua hạn có thể bị server từ chối theo server time. UI phải hiển thị rõ ràng: "Câu trả lời được giữ trên thiết bị nhưng có thể không được chấp nhận nếu đã quá hạn."

## 15. Submit flow

```text
user clicks Submit
-> confirmation summary: unanswered, flagged
-> lock navigation/edit locally
-> persist current input to IndexedDB
-> flush pending operations with deadline
-> POST /attempts/{attempt_id}/submit using stable idempotency key
   -> 2xx: mark local attempt terminal, cleanup completed ops, navigate status
   -> network error: keep submit intent, show retry/finalizing
   -> 409 already submitted: fetch attempt, reconcile terminal
   -> validation pending answers: sync/retry per response
```

Submit idempotency key được lưu trong attempt record để survive reload.

## 16. Submit intent persistence

Nếu user đã xác nhận submit nhưng request chưa rõ kết quả:

```ts
interface SubmitIntent {
  attemptId: string;
  idempotencyKey: string;
  createdAt: number;
  status: 'pending' | 'confirmed';
}
```

Reload sẽ reconcile trước khi cho tiếp tục edit.

## 17. Multiple tabs

- Broadcast `ATTEMPT_ACTIVE` theo attempt ID.
- Tab thứ hai hiển thị warning và ưu tiên read-only cho đến khi xác minh.
- Không dùng BroadcastChannel làm security.
- Server vẫn enforce active attempt/revision.

## 18. Page lifecycle

### `beforeunload`

Chỉ cảnh báo nếu:

- Có pending operation.
- Submit đang uncertain.
- Attempt đang active.

Không dựa vào `beforeunload` để gửi answer cuối cùng.

### Visibility

Ghi event local/optional backend khi tab hidden/visible theo policy, nhưng không xem đó là bằng chứng gian lận.

## 19. Service worker policy

- Static shell precache có thể thêm sau.
- Không giao answer sync duy nhất cho Background Sync API.
- IndexedDB operation queue và page foreground sync là cơ chế chính.
- Service worker update không được reload exam page tự động.
- Nếu có bản mới trong lúc thi, trì hoãn activation/reload đến khi attempt terminal.

## 20. Question rendering

Question renderer registry:

```ts
const questionRenderers: Record<QuestionType, QuestionRenderer> = {
  SINGLE_CHOICE: SingleChoiceQuestion,
  MULTIPLE_CHOICE: MultipleChoiceQuestion,
  TRUE_FALSE: TrueFalseQuestion,
  SHORT_TEXT: ShortTextQuestion,
  NUMERIC: NumericQuestion,
  ESSAY: EssayQuestion,
};
```

Unknown type: render unsupported state, không crash toàn bài.

Question snapshot content là read-only; không dùng current question bank data.

## 21. Accessibility

- Question heading có số/thứ tự rõ.
- Choice group dùng fieldset/legend hoặc ARIA tương đương.
- Keyboard navigation không cướp phím nhập text.
- Timer warning dùng polite announcement có kiểm soát.
- Save failure dùng persistent status region.
- Question navigator có current/answered/flagged semantics ngoài màu.
- Focus chuyển hợp lý khi next/previous.

## 22. Exam test matrix

| Scenario | Expected |
|---|---|
| Chọn answer rồi reload trước network ack | Restore từ IndexedDB và sync |
| Network mất 5 phút | Continue local, status offline, sync khi online; answer gần/qua hạn có thể bị server từ chối |
| Hai save cùng question | Latest payload, monotonic revision |
| Access token hết hạn | Refresh single-flight, save tiếp tục |
| Browser background 10 phút | Timer đúng khi quay lại |
| Device clock sai | Server offset giữ deadline đúng |
| Double click submit | Một idempotent submit |
| Submit response mất | Reconcile bằng same key/GET attempt |
| Server đã expire | UI terminal, không cho edit |
| IndexedDB quota/error | Blocking safety message trước/đang thi |
| Duplicate tab | Warning/conflict handling |
| App deploy version mới | Không force reload active exam |

## 23. Cleanup

Sau terminal confirmed:

- Xóa operations completed.
- Giữ minimal attempt receipt trong thời gian ngắn nếu cần support.
- Không giữ answer content lâu hơn policy.
- Cleanup theo user ID để tránh account sau trên shared device đọc draft trước.
