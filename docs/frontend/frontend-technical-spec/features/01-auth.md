# Feature 01 — Authentication & Session

## 1. Scope

- Login bằng organization code + username/password.
- Refresh session bootstrap.
- Logout current/all sessions.
- Change/reset password.
- Current actor and permissions.
- Cross-tab logout.
- Route gating.

## 2. Slice map

```text
features/auth/
├── login/
├── refresh-session/
├── logout/
├── change-password/
└── revoke-session/

entities/user/
├── model/actor.ts
└── ui/user-avatar.tsx

shared/auth/
├── auth-session-store.ts
├── auth-broadcast.ts
└── permissions.ts
```

## 3. Session state

```ts
type AuthStatus =
  | 'bootstrapping'
  | 'authenticated'
  | 'anonymous'
  | 'restricted'
  | 'degraded';
```

- `restricted`: phải đổi password.
- `degraded`: refresh chưa xác định do network/5xx; cho retry, không coi là anonymous ngay.

## 4. Login form

Fields:

- `organizationCode`.
- `username`.
- `password`.

Client validation chỉ format/required. Invalid credentials response hiển thị generic.

## 5. Refresh coordinator

Responsibilities:

- Single-flight promise.
- Prevent recursive refresh.
- Update access token atomically.
- Signal auth store subscribers.
- Return retryable vs terminal failure.

## 6. Actor model

```ts
interface CurrentActor {
  id: string;
  organizationId: string;
  displayName: string;
  roles: string[];
  permissions: string[];
  mustChangePassword: boolean;
}
```

Do not derive permission list locally from role.

## 7. Cache interactions

On login:

- Clear prior QueryClient.
- Set actor query if response sufficient.
- Preload role home essentials optionally.

On logout:

- Cancel and clear queries.
- Clear access token.
- Cleanup user-scoped persisted data.
- Broadcast event.

## 8. Errors

| Code/category | UI |
|---|---|
| Invalid credentials | Inline general error |
| Suspended account | General error + admin contact guidance |
| Rate limit | Retry time |
| Network | Retry without clearing inputs |
| Refresh reuse/revoked | Clear session, login |
| Must change password | Redirect restricted flow |

## 9. Tests

- Login success/invalid/rate limit.
- Refresh boot success/401/network.
- 20 concurrent 401 -> one refresh.
- Logout clears token/cache and other tab.
- Safe returnTo.
- Permission utilities.
- No token in local/session storage.
