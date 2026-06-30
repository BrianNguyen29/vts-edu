# 05. Routing & Layouts

## 1. Router strategy

Dùng `createBrowserRouter` với route modules lazy-loaded.

```tsx
const router = createBrowserRouter([
  publicRoutes,
  authRoutes,
  appRoutes,
  examRoutes,
  errorRoutes,
]);
```

Route path được định nghĩa trong `app/router/paths.ts`, không rải string.

## 2. Layout hierarchy

```text
Root
├── PublicLayout
│   └── Landing / legal pages
├── AuthLayout
│   └── Login / forgot / reset
├── AppShellLayout [authenticated]
│   ├── Common workspace
│   ├── StudentWorkspaceLayout
│   ├── TeacherWorkspaceLayout
│   └── AdminWorkspaceLayout
└── ExamLayout [authenticated, minimal chrome]
    └── Attempt runner
```

## 3. Layout responsibilities

| Layout | Trách nhiệm |
|---|---|
| `PublicLayout` | Header đơn giản, public footer, no app session dependency |
| `AuthLayout` | Centered auth card, organization branding nhẹ, no sidebar |
| `AppShellLayout` | Session gate, sidebar, header, notifications, breadcrumbs |
| `RoleWorkspaceLayout` | Role/permission navigation và workspace title |
| `ExamLayout` | Timer, save state, question navigation, tối thiểu distraction |
| `ErrorLayout` | 403/404/500/maintenance |

## 4. Protected route model

Mỗi route protected có metadata:

```ts
interface RouteMeta {
  titleKey: string;
  requiredPermissions?: string[];
  anyPermission?: string[];
  allowedRolesForNavigation?: string[];
  breadcrumb?: BreadcrumbFactory;
  fullBleed?: boolean;
}
```

Guard flow:

```text
ensure session bootstrapped
  -> unauthenticated: redirect /login?returnTo=...
  -> must change password: redirect /change-password
  -> check route permission metadata
  -> forbidden: render /403
  -> render route
```

Backend 403 vẫn được xử lý độc lập nếu quyền thay đổi sau khi route đã mở.

## 5. Route catalog tóm tắt

### Public/Auth

| Route | Page | Protection | Layout |
|---|---|---|---|
| `/` | Landing/redirect | Public | Public |
| `/login` | Login | Guest-only | Auth |
| `/forgot-password` | Forgot password | Guest-only | Auth |
| `/reset-password` | Reset password | Token in query, no session required | Auth |
| `/change-password` | Forced password change | Authenticated restricted session | Auth |
| `/403` | Forbidden | Public | Error |
| `*` | Not found | Public | Error |

### Common authenticated

| Route | Page | Permission | Layout |
|---|---|---|---|
| `/app` | Workspace redirect | Authenticated | AppShell |
| `/app/profile` | Profile | Authenticated | AppShell |
| `/app/security/sessions` | Sessions | Authenticated | AppShell |
| `/app/notifications` | Notifications | Authenticated | AppShell |
| `/app/settings` | Personal settings | Authenticated | AppShell |

### Student

| Route | Page | Permission | Layout |
|---|---|---|---|
| `/app/student` | Dashboard | `student:workspace` | Student |
| `/app/student/classes` | Class list | `class:own:view` | Student |
| `/app/student/classes/:classId` | Class detail | `class:own:view` | Student |
| `/app/student/assignments` | Assignment list | `assignment:own:view` | Student |
| `/app/student/assignments/:assignmentId` | Assignment detail | `assignment:own:view` | Student |
| `/app/student/assessments` | Assessment list | `assessment:assigned:view` | Student |
| `/app/student/results/:attemptId` | Published result | `attempt:own:view` | Student |
| `/app/student/grades` | Grade summary | `grade:own:view` | Student |
| `/exam/attempts/:attemptId` | Exam runner | `attempt:own:continue` | Exam |

### Teacher

| Route | Page | Permission | Layout |
|---|---|---|---|
| `/app/teacher` | Dashboard | `teacher:workspace` | Teacher |
| `/app/teacher/classes` | Class list | `class:view` | Teacher |
| `/app/teacher/classes/:classId` | Class workspace | `class:view` | Teacher |
| `/app/teacher/question-banks` | Bank list | `question:view` | Teacher |
| `/app/teacher/questions/new` | Create question | `question:create` | Teacher |
| `/app/teacher/questions/:questionId` | Question detail/versions | `question:view` | Teacher |
| `/app/teacher/assessments` | Assessment list | `assessment:view` | Teacher |
| `/app/teacher/assessments/new` | Create assessment | `assessment:create` | Teacher |
| `/app/teacher/assessments/:assessmentId/edit` | Builder | `assessment:update` | Teacher |
| `/app/teacher/assessments/:assessmentId/results` | Results | `attempt:grade` | Teacher |
| `/app/teacher/assignments` | Assignment list | `assignment:view` | Teacher |
| `/app/teacher/gradebook/:classId` | Gradebook | `grade:view` | Teacher |

### Admin

| Route | Page | Permission | Layout |
|---|---|---|---|
| `/app/admin` | Admin dashboard | `admin:workspace` | Admin |
| `/app/admin/users` | Users | `user:view` | Admin |
| `/app/admin/users/imports` | Imports | `user:import` | Admin |
| `/app/admin/classes` | Classes | `class:admin` | Admin |
| `/app/admin/academic-terms` | Terms | `academic:manage` | Admin |
| `/app/admin/audit-logs` | Audit | `audit:view` | Admin |
| `/app/admin/settings` | Organization settings | `organization:update` | Admin |

Chi tiết nằm trong thư mục `routes/`.

## 6. Lazy loading

Route module pattern:

```ts
{
  path: 'teacher/assessments/:assessmentId/edit',
  lazy: () => import('@/pages/teacher/assessment-edit/route'),
}
```

Heavy subcomponents như TipTap editor, Recharts và PDF viewer tiếp tục lazy-load trong route.

## 7. Loader policy

Loader được dùng cho:

- Session/permission check.
- Validate required route params.
- Redirect theo state rõ ràng.
- `queryClient.ensureQueryData` khi tránh waterfall có lợi.

Không dùng loader để:

- Tự xây cache song song với TanStack Query.
- Gọi mutation.
- Chứa business logic dài.

## 8. Search params

Filter/sort/pagination có thể bookmark phải nằm trong URL:

```text
/app/teacher/questions?status=PUBLISHED&type=SINGLE_CHOICE&sort=-updated_at
```

Parse bằng schema Zod; invalid param fallback về default và optionally replace URL.

## 9. Navigation blocking

Block navigation khi:

- Form editor dirty chưa lưu.
- Assessment builder có local change chưa persist.
- Submission draft đang upload.

Không block exam navigation nội bộ; exam có autosave queue riêng. Rời toàn bộ exam route hiển thị confirm rõ.

## 10. Scroll and focus restoration

- Sau route navigation, focus vào `h1` hoặc main landmark.
- Preserve scroll cho list -> detail -> back nếu có giá trị.
- Modal close trả focus về trigger.
- Validation navigation focus field lỗi đầu tiên.

## 11. Redirect safety

`returnTo` chỉ chấp nhận relative path bắt đầu bằng `/` và thuộc origin. Không redirect tới URL tùy ý từ query param.

## 12. 404 và resource hiding

Backend có thể trả 404 cho resource cross-tenant. UI không phân biệt “không tồn tại” với “không được phép” nếu response là 404; hiển thị generic not found.
