# 10. Background Jobs & Scheduling

## 1. Queue choice

River chạy trên PostgreSQL. Mục tiêu:

- Transactional enqueue.
- Retry/backoff.
- Scheduled jobs.
- Không thêm Redis/message broker ở MVP.

## 2. Job catalog

| Job kind | Trigger | Idempotency key | Retry |
|---|---|---|---|
| `grading.auto_grade_attempt` | Attempt submitted/expired | `attempt_id + grading_version` | Có |
| `attempts.expire_sweep` | Schedule định kỳ | time bucket | Có |
| `files.inspect` | Upload complete | `file_id + object_version` | Có |
| `files.generate_preview` | File ready phù hợp | `file_id + preview_version` | Có |
| `notifications.fanout` | Domain event | `event_id` | Có |
| `email.send` | Notification channel | `notification_recipient_id` | Có |
| `users.import` | Admin import | `import_job_id` | Có, giới hạn |
| `gradebook.export` | Export request | `export_job_id` | Có |
| `analytics.rebuild_class` | Event/schedule | `class_id + projection_version` | Có |

## 3. Job payload rules

Payload chỉ chứa identifiers và version:

```json
{
  "attempt_id": "...",
  "organization_id": "...",
  "grading_version": 1
}
```

Không nhét full answer/PII vào payload.

## 4. Transactional enqueue example

```go
err := txManager.WithinTx(ctx, func(ctx context.Context, q *db.Queries) error {
    if err := attempts.MarkSubmitted(ctx, q, attemptID, now); err != nil {
        return err
    }

    _, err := riverClient.InsertTx(ctx, tx, AutoGradeArgs{
        AttemptID: attemptID,
    }, nil)
    return err
})
```

Tên API thực tế phụ thuộc version River; code phải theo docs/version đã pin.

## 5. Worker requirements

Mọi worker phải:

- Idempotent.
- Respect context cancellation.
- Có timeout.
- Log job ID/kind/attempt count.
- Phân loại retryable vs permanent error.
- Không log payload nhạy cảm.
- Có metric duration/failure.

## 6. Auto-grade idempotency

1. Load attempt terminal status.
2. Nếu grading run cùng version đã completed: return success.
3. Create/upsert grading run.
4. Grade từng item từ immutable snapshot.
5. Upsert item results theo unique key.
6. Mark manual review required nếu có essay.
7. Update attempt grading status.
8. Update linked grade entry chỉ nếu source version phù hợp và không phá override.

## 7. Retry policy

| Error | Retry? |
|---|:---:|
| DB transient/network | Có |
| Object storage 5xx | Có |
| Email 429/5xx | Có với backoff |
| Invalid payload/schema | Không |
| Resource not found do bug/stale | Thường không; log/mark discarded |
| Permission error trong job nội bộ | Không; security alert |

Exponential backoff có jitter. Số lần retry tùy job; email và file có thể nhiều hơn grading logic lỗi permanent.

## 8. Scheduled jobs

- **Attempt expiry là request-time reconciliation:** mọi read/write runtime và `POST /attempts/{id}/submit` kiểm tra `expires_at` theo server time; nếu quá hạn chuyển `EXPIRED` trong cùng request.
- Attempt expiry sweep: chỉ là safety-net, ví dụ mỗi 1 phút hoặc phù hợp tải; không dùng cron làm nguồn duy nhất.
- Cleanup expired idempotency keys: hàng ngày.
- Cleanup abandoned upload intents: hàng giờ/ngày.
- Notification digest: tùy product.
- Backup không được triển khai chỉ bằng app scheduler; dùng platform/provider job (Supabase backup/self-script).

## 9. Scheduler safety

Nếu nhiều process:

- River periodic jobs hoặc unique job mechanism.
- Không dựa vào in-memory cron trên mọi replica nếu có thể chạy trùng.
- Job vẫn phải idempotent dù scheduler đảm bảo uniqueness.

## 10. Dead jobs

Cần admin/operator query hoặc River UI restricted:

- Job kind.
- Error summary.
- Attempts.
- Next retry.
- Related resource ID.

Không expose River UI public.
