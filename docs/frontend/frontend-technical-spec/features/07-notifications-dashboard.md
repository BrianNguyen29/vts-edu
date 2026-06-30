# Feature 07 — Notifications & Dashboards

## 1. Notifications scope

- In-app list.
- Unread count.
- Mark read.
- Navigate to safe target.
- Basic pagination.

No realtime socket required in MVP. Poll 30–60 seconds or refetch on focus according policy outside exam.

## 2. Notification model

```ts
interface NotificationViewModel {
  id: string;
  type: string;
  title: string;
  body: string;
  createdAt: string;
  readAt: string | null;
  targetPath: string | null;
}
```

Target path must be validated relative same-origin route.

## 3. Mark read

Can use optimistic update because rollback is simple. On failure restore unread state and show non-blocking error.

## 4. Student dashboard

Widgets:

- Continue current class/resource.
- Upcoming assignments.
- Upcoming assessments.
- Recently published grades.
- Notifications.

Queries parallel and widget-level error boundaries.

## 5. Teacher dashboard

Widgets:

- Pending manual grading.
- Pending assignment grading.
- Upcoming assessment windows.
- Students/submissions needing attention.
- Quick create actions.
- Notifications.

## 6. Admin dashboard

Widgets:

- User/class counts.
- Failed/pending imports.
- Recent audit/security events.
- System notices if API exposes.

## 7. Dashboard performance

- Avoid a single mega endpoint only if backend already has modular endpoints; however one purpose-built dashboard summary endpoint can reduce waterfalls.
- Lazy chart.
- No heavy hero illustration initial.
- Skeleton by widget.

## 8. Error/empty states

- A failed widget does not crash dashboard.
- “Không có việc sắp đến hạn” is positive empty state.
- Permission-hidden widgets not rendered as errors.

## 9. Tests

- Unread count optimistic rollback.
- Safe notification target.
- Dashboard partial failure.
- Role-specific widget visibility.
- Mobile order priority.
