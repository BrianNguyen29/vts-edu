# Routes 01 — Public & Authentication

## 1. Route table

| Method-like action | URL | Page | Protection | Layout | Main API |
|---|---|---|---|---|---|
| Navigate | `/` | Landing or role redirect | Public | Public | Optional runtime config |
| Navigate | `/login` | Login | Guest-only | Auth | `POST /auth/login` |
| Navigate | `/forgot-password` | Forgot Password | Guest-only | Auth | `POST /auth/forgot-password` |
| Navigate | `/reset-password?token=...` | Reset Password | Public token flow | Auth | `POST /auth/reset-password` |
| Navigate | `/change-password` | Forced/normal change | Auth required | Auth | `POST /auth/change-password` |
| Navigate | `/403` | Forbidden | Public | Error | — |
| Navigate | `/maintenance` | Maintenance | Public | Error | health/config optional |
| Navigate | `*` | Not Found | Public | Error | — |

## 2. `/`

Behavior:

```text
not bootstrapped -> app bootstrap
anonymous -> /login
student -> /app/student
teacher -> /app/teacher
admin -> /app/admin
multiple workspaces -> last safe workspace or chooser
```

No permanent role redirect based only on stale local preference.

## 3. Login page

### Fields

- Organization code.
- Username.
- Password.
- Remember username optional; never remember password/token.

### Request

```json
{
  "organization_code": "school-a",
  "username": "hs001",
  "password": "***"
}
```

### States

- Idle.
- Submitting.
- Invalid credentials generic.
- Account suspended.
- Rate limited with retry guidance.
- Network error.
- Must change password.

### Security

- Do not reveal whether organization/user exists beyond backend message policy.
- Password autocomplete `current-password`.
- Organization/username autocomplete policy explicit.
- `returnTo` validated same-origin relative path.

## 4. Forgot password

MVP may be enabled mainly for teacher/admin with email. Student account may require admin reset.

UI always displays generic accepted message after 202 to avoid account enumeration.

## 5. Reset password

- Token parsed from query but never stored.
- After successful reset, remove token from URL by navigation replace.
- Password strength rules reflect backend minimum but backend validates final.
- Session/token reuse error has clear restart flow.

## 6. Change password

Two modes:

1. Normal authenticated change: current + new password.
2. Forced first-login: backend restricted access token/flag; only new password flow according to API contract.

Success:

- Backend may revoke sessions.
- Clear auth state and re-login, or refresh actor according to response.

## 7. Guest-only route

If authenticated user opens `/login`, redirect to role workspace unless explicit `reauth=1` flow exists.

## 8. Error pages

### 403

- Explain insufficient permission.
- Link back to dashboard.
- Show request ID if from API error.
- Do not reveal hidden resource details.

### 404

- Generic not found.
- Safe navigation options.

### Maintenance

- Retry button.
- No infinite auto-refresh.
- Support/release info safe.
