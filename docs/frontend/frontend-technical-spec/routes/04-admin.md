# Routes 04 — Organization Administrator

Base: `/app/admin`

## 1. Route table

| Route | Page | Permission | Main capability |
|---|---|---|---|
| `/app/admin` | Admin Dashboard | `admin:workspace` | Counts, recent imports/security events |
| `/app/admin/users` | Users | `user:view` | List/filter/create/suspend |
| `/app/admin/users/new` | Create User | `user:create` | Create with role/temp password |
| `/app/admin/users/:userId` | User Detail | `user:view` | Profile, roles, sessions/status |
| `/app/admin/users/imports` | User Imports | `user:import` | Upload/dry-run/history |
| `/app/admin/users/imports/:jobId` | Import Result | `user:import` | Progress/errors/download |
| `/app/admin/classes` | Classes | `class:admin` | Create/archive/assign teacher |
| `/app/admin/classes/:classId` | Class Admin | `class:admin` | Enrollment/teacher management |
| `/app/admin/academic-terms` | Academic Terms | `academic:manage` | Create/activate/close |
| `/app/admin/audit-logs` | Audit Logs | `audit:view` | Filter/detail/export if allowed |
| `/app/admin/settings` | Organization Settings | `organization:update` | Branding/policies/basic config |

## 2. Admin dashboard

MVP metrics:

- Active users by role.
- Active classes.
- Pending/failed imports.
- Recent security/audit events.
- Storage usage if API provides.

No real-time chart required.

## 3. User list

Filters in URL:

```text
?role=student&status=ACTIVE&query=hs001&sort=display_name
```

Actions:

- Create.
- Suspend/activate.
- Reset temporary password.
- Change roles.
- Revoke sessions.

Destructive/security actions require confirm and reason if API requires.

## 4. User import

Flow:

```text
choose CSV
-> direct upload/file ID
-> start dry-run import job
-> poll/query job status
-> show accepted/rejected preview
-> confirm real import with new idempotency key
-> result/error file
```

Do not parse huge CSV completely in browser as source of truth. Client preview optional, backend validates final.

## 5. Academic terms

- Date/time timezone explicit.
- Activating/closing term has consequence summary.
- Current term state comes backend.

## 6. Audit logs

- Read-only.
- Server-side filter/pagination.
- Sensitive payload redacted by backend.
- Frontend does not attempt to display arbitrary JSON without safe formatting.
- Copy request/event IDs available.

## 7. Organization settings

MVP:

- Display name/logo reference.
- Default timezone.
- Locale.
- Basic exam/resource limits if API exposes.

Security-critical settings require reauth/confirmation as backend defines.
