# VTS EDU API

Minimal Go API scaffold for the MVP demo.

## Local development

```bash
cd apps/api
cp ../../config/render.env.example .env
# Edit .env with local values, then:
go run ./cmd/server
```

## Endpoints

- `GET /healthz` — liveness
- `GET /readyz` — readiness
- `GET /api/v1/auth/csrf-token` — CSRF token endpoint
- `POST /api/v1/auth/login` — login with org code + username/password; sets refresh cookie
- `POST /api/v1/auth/refresh` — rotate refresh session; returns new access token (requires CSRF)
- `POST /api/v1/auth/logout` — revoke current refresh session and clear cookie (requires CSRF)
- `GET /api/v1/me` — current actor from Bearer token
- `GET /api/v1/attempts/{attempt_id}` — authenticated tenant-scoped attempt snapshot
- `PUT /api/v1/attempts/{attempt_id}/answers/{attempt_item_id}` — save/update answer for an owned in-progress attempt (requires CSRF)
- `POST /api/v1/attempts/{attempt_id}/submit` — submit an owned in-progress attempt (requires CSRF)

Unsafe cookie-backed endpoints require the `X-CSRF-Token` header to match the `vts_csrf` cookie.

## Demo credentials

Seeded by `supabase/migrations/000004_auth_passwords_and_demo_seed.sql`:

- Organization code: `school-a`
- Username: `hs001`
- Password: `Password123!`
- Demo attempt id: `00000000-0000-4000-8000-000000000001`

## Build

```bash
docker build -t vts-edu-api -f apps/api/Dockerfile apps/api
```

## Notes

- Run `go mod tidy` after enabling network to fetch `chi/v5`.
- River workers and complex async grading are intentionally placeholders for the scaffold.
