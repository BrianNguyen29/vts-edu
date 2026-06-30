# 02. Frontend Architecture

## 1. Lựa chọn kiến trúc

Sử dụng **feature-based modular architecture lấy cảm hứng từ Feature-Sliced Design**, giản lược cho dự án solo.

Các layer:

```text
app       -> composition root, providers, router
pages     -> route-level composition
widgets   -> large reusable page sections
features  -> user actions/use cases
entities  -> domain presentation models/components
shared    -> generic UI, API, utilities, configuration
```

### Import direction

```text
app
 ↓
pages
 ↓
widgets
 ↓
features
 ↓
entities
 ↓
shared
```

Layer thấp không được import layer cao. Feature A không import nội bộ Feature B; nếu cần phối hợp, page/widget composition thực hiện orchestration hoặc trích abstraction xuống entity/shared.

## 2. Vì sao không dùng MVC/Atomic Design thuần

### Không dùng MVC

MVC không mô tả tốt boundary của React app hiện đại, nơi component, query, form và route thường gắn theo user capability.

### Không dùng Atomic Design làm top-level architecture

Atomic Design hữu ích cho design system nhưng dễ tạo thư mục `atoms/molecules/organisms` không phản ánh nghiệp vụ. Trong dự án này:

- UI primitives tương đương atom nằm trong `shared/ui`.
- Domain components nằm trong `entities`.
- User actions nằm trong `features`.
- Page sections lớn nằm trong `widgets`.

## 3. Trách nhiệm từng layer

| Layer | Trách nhiệm | Ví dụ | Không được làm |
|---|---|---|---|
| `app` | Bootstrap, provider, router, global error | `AppProviders`, `router.tsx` | Domain logic |
| `pages` | Compose route screen | `StudentDashboardPage` | Gọi raw fetch, chứa reusable business logic lớn |
| `widgets` | Khối UI lớn tái sử dụng ở page | `UpcomingWorkPanel` | Tự sở hữu API contract riêng |
| `features` | Hành động người dùng | `login`, `save-answer`, `publish-assessment` | Import feature khác tùy tiện |
| `entities` | Mô hình trình bày domain | `ClassCard`, `QuestionSummary` | Orchestrate workflow nhiều bước |
| `shared` | Generic foundation | button, dialog, api, date | Biết domain cụ thể |

## 4. Slice structure

Mỗi feature/entity nên có public API nhỏ:

```text
features/save-answer/
├── api/
│   ├── mutation.ts
│   └── types.ts
├── model/
│   ├── queue.ts
│   └── use-save-answer.ts
├── ui/
│   └── save-indicator.tsx
├── lib/
│   └── revision.ts
├── __tests__/
└── index.ts
```

Bên ngoài chỉ import từ `index.ts`:

```ts
import { useSaveAnswer, SaveIndicator } from '@/features/save-answer';
```

Không import sâu:

```ts
// Cấm
import { useSaveAnswer } from '@/features/save-answer/model/use-save-answer';
```

Ngoại lệ: test nội bộ slice có thể import module private.

## 5. Composition root

`app/` chịu trách nhiệm khởi tạo theo thứ tự:

```text
runtime config
  -> telemetry/logger
  -> auth session store
  -> API client
  -> QueryClient
  -> i18n
  -> theme
  -> router
  -> render
```

Nếu bootstrap config thất bại, render `BootstrapErrorScreen`, không render app với config rỗng.

## 6. Dependency injection thực dụng

Không dùng DI container. Dependency được tạo bằng factory và truyền qua module/provider.

Ví dụ:

```ts
export interface AppServices {
  api: ApiClient;
  auth: AuthSessionStore;
  examDrafts: ExamDraftRepository;
  logger: FrontendLogger;
  clock: Clock;
}
```

Production và test tạo implementation khác nhau.

## 7. Page composition

Page chỉ nên:

1. Đọc route params/search params.
2. Gọi feature/entity hooks cấp cao.
3. Compose widget/component.
4. Chọn loading/error/empty state.
5. Cập nhật title/breadcrumb nếu cần.

Ví dụ:

```tsx
export function TeacherClassPage() {
  const { classId } = useRequiredParams(['classId']);
  const classQuery = useClassDetail(classId);

  return (
    <PageContainer>
      <ClassHeader query={classQuery} />
      <Tabs>
        <ClassOverviewTab classId={classId} />
        <ClassStudentsTab classId={classId} />
        <ClassResourcesTab classId={classId} />
      </Tabs>
    </PageContainer>
  );
}
```

Không đặt chuỗi fetch/update trực tiếp trong JSX.

## 8. Domain boundaries phía frontend

| Domain | Frontend responsibility |
|---|---|
| Auth | Session bootstrap, route gate, login UI, token refresh coordination |
| Classes | Navigation, list/detail, enrollment management UI |
| Resources | File metadata, direct upload, preview, access state |
| Questions | Editor, version list, preview, filters |
| Assessments | Builder, validation summary, publish flow |
| Attempts | Runtime, timer, durable pending answers, submit |
| Assignments | Submission editor/upload, status |
| Gradebook | Grid view/edit, publish UX, decimal formatting |
| Notifications | Inbox, unread badge, navigation to target |

Backend vẫn sở hữu business invariant. Frontend chỉ biểu diễn và hỗ trợ workflow.

## 9. Cross-cutting concerns

Các concern dùng chung đặt ở `shared` hoặc `app`:

- API transport.
- Error normalization.
- Authentication session.
- Permission evaluation.
- Runtime config.
- i18n.
- Date/time.
- Decimal display.
- Telemetry.
- Feature flags.
- File upload/download.
- IndexedDB adapter.

## 10. Error boundaries

Phân tầng:

```text
RootErrorBoundary
  -> LayoutErrorBoundary
      -> RouteErrorBoundary
          -> Widget error state / QueryErrorResetBoundary
```

- Root: lỗi bootstrap hoặc render không phục hồi.
- Route: lỗi route module/loader.
- Widget: lỗi query cục bộ, cho retry không làm sập page.
- Exam: boundary riêng, không tự reset dữ liệu pending.

## 11. Anti-patterns

Cấm:

- `components/` toàn cục chứa hàng trăm file không ownership.
- `utils.ts` khổng lồ.
- API call trong `useEffect` thủ công.
- Mirror query data vào context/store.
- Component 1.000 dòng vừa fetch vừa validate vừa render.
- Global event emitter.
- Feature import vòng tròn.
- `index.ts` barrel export toàn repository gây circular dependency.
- Boolean props khó hiểu như `small`, `blue`, `teacherMode`; dùng variant rõ.
- Permission check bằng string role rải rác trong JSX.
