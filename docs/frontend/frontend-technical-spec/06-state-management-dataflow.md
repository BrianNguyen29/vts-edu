# 06. State Management & Dataflow

## 1. State taxonomy

Không có một “global store” chứa mọi thứ. Mỗi loại state dùng công cụ phù hợp.

| Loại state | Công cụ | Ví dụ | Persistent |
|---|---|---|---|
| Server state | TanStack Query | classes, questions, grades | Query cache memory |
| Auth session | Memory external store | access token, actor, bootstrap status | Không |
| Form state | React Hook Form | question editor, assignment form | Không hoặc draft explicit |
| Route state | URL params/search params | filter, sort, cursor | URL |
| Local UI state | `useState`/`useReducer` | open dialog, active tab | Không |
| UI preferences | Context + localStorage | theme, sidebar collapsed, locale | Có, không nhạy cảm |
| Exam pending answers | IndexedDB repository | answer chưa ack, revision | Có |
| Ephemeral upload state | Feature reducer | progress, cancellation | Không |
| Runtime configuration | Read-only app service | API base URL, flags | Memory |

## 2. Nguyên tắc

1. Không mirror query response vào Context/Zustand.
2. Không lưu access token trong Query cache.
3. Không lưu form state vào Query cache trước khi backend nhận.
4. State có thể bookmark/share phải nằm trong URL.
5. State cần survive crash nhưng chưa được server xác nhận dùng IndexedDB.
6. Query cache phải clear khi logout hoặc đổi organization context.

## 3. Standard dataflow

```text
User action
  -> feature UI handler
  -> mutation/query hook
  -> openapi-react-query
  -> openapi-fetch middleware
  -> HTTP request
  -> Go API
  -> HTTP response / Problem Details
  -> normalized result/error
  -> TanStack Query cache update/invalidate
  -> React re-render
  -> toast/focus/navigation side effect
```

## 4. Query conventions

### Query option factory

```ts
export function classDetailQueryOptions(classId: string) {
  return apiQuery.queryOptions('get', '/classes/{class_id}', {
    params: { path: { class_id: classId } },
  });
}
```

Feature/page dùng:

```ts
const query = useQuery(classDetailQueryOptions(classId));
```

### Stale time policy

| Data | `staleTime` đề xuất |
|---|---:|
| Current user/permissions | 1–5 phút; refresh explicit sau auth action |
| Reference data | 10–30 phút |
| Class list | 30–60 giây |
| Question list | 15–30 giây |
| Assessment builder detail | 0–10 giây; refetch theo action |
| Exam attempt runtime | Không background refetch tùy tiện |
| Notifications | 30–60 giây hoặc manual refresh |
| Gradebook | 0–15 giây, invalidate sau mutation |

Không áp dụng một default stale time cho mọi query.

## 5. Query key rules

Nếu dùng openapi-react-query, key được tạo từ method/path/params. Custom derived query cần key factory:

```ts
export const gradebookKeys = {
  root: ['gradebook'] as const,
  class: (classId: string) => [...gradebookKeys.root, 'class', classId] as const,
};
```

Key phải bao gồm:

- Organization context nếu client có thể switch organization trong một session.
- Resource ID.
- Filter/sort/search params canonical.

## 6. Mutation conventions

Mutation phải khai báo:

- Request type từ OpenAPI.
- Pending UI.
- Error mapping.
- Cache invalidation/update.
- Success feedback.
- Idempotency nếu endpoint yêu cầu.

Ví dụ:

```ts
export function usePublishAssessment(assessmentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => publishAssessment(assessmentId),
    onSuccess: (response) => {
      queryClient.setQueryData(
        assessmentDetailKey(assessmentId),
        response,
      );
      queryClient.invalidateQueries({ queryKey: assessmentListKey() });
    },
  });
}
```

## 7. Optimistic update policy

Chỉ optimistic update khi:

- Operation dễ rollback.
- Conflict hiếm.
- Không tác động invariant nhạy cảm.

Có thể optimistic:

- Mark notification read.
- Toggle local preference đã có backend.
- Rename non-published folder nếu API có version check.

Không optimistic:

- Publish assessment.
- Submit attempt.
- Grade override.
- Change role.
- Finalize manual grade.

## 8. Retry policy

### Query

- Retry tối đa 2 lần cho network/5xx có khả năng transient.
- Không retry 400/401/403/404/409/422.
- Không retry exam answer bằng Query default; dùng sync queue policy riêng.

### Mutation

- Không retry write mặc định nếu không idempotent.
- Endpoint có idempotency key có thể retry có kiểm soát.
- Upload part/direct upload có retry riêng.

## 9. Auth store

Auth store nhỏ dùng `useSyncExternalStore`:

```ts
type AuthStatus = 'bootstrapping' | 'authenticated' | 'anonymous';

interface AuthSnapshot {
  status: AuthStatus;
  accessToken: string | null;
  actor: CurrentActor | null;
}
```

Store không chứa class list, notification list hoặc profile edit form.

## 10. URL state

Ví dụ question list:

```text
?query=ham-so&type=SINGLE_CHOICE&status=PUBLISHED&sort=-updated_at
```

Flow:

```text
URL search params
  -> Zod parse/default
  -> query options
  -> API request
  -> table render
```

Thay filter nên reset cursor.

## 11. Form state

React Hook Form là owner của field values/dirty/touched/errors. Query data chỉ dùng làm `defaultValues` một lần hoặc `reset()` có kiểm soát sau fetch.

Không làm:

```ts
useEffect(() => setValue('title', query.data.title), [query.data]);
```

cho từng field vì dễ ghi đè người dùng đang nhập.

## 12. Exam state

Exam có ba lớp state:

```text
Server snapshot/query data
  + In-memory current UI answer
  + IndexedDB pending operations
```

- Server snapshot: question, deadline, last acknowledged revision.
- In-memory: input đang thao tác.
- IndexedDB: operation cần gửi/retry.

Chi tiết ở `10-exam-runtime-frontend.md`.

## 13. Multi-tab coordination

Dùng `BroadcastChannel` cho:

- Logout.
- Session revoked.
- Token refresh completed/failed signal nếu cần.
- Attempt opened in another tab warning.

Không broadcast access token raw. Mỗi tab nhận token qua refresh flow riêng hoặc update in-process only.

## 14. Query cache lifecycle

### Login

```text
clear anonymous cache
-> set authenticated actor
-> preload role dashboard essentials
```

### Logout

```text
cancel queries
-> clear QueryClient
-> clear access token
-> close sensitive IndexedDB handles
-> remove user-scoped drafts according to policy
-> broadcast logout
-> navigate /login
```

Exam drafts của một attempt đã submit có thể cleanup; pending draft chưa rõ trạng thái phải được reconcile trước khi xóa.

## 15. Derived state

Không lưu derived state nếu tính rẻ:

```ts
const overdue = assignment.status !== 'submitted' && isPast(dueAt);
```

Memoize chỉ khi profiling hoặc referential stability thực sự cần.
