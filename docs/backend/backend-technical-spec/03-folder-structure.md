# 03. Repository & Folder Structure

## 1. Monorepo tổng thể

```text
lms-platform/
├── AGENTS.md
├── README.md
├── package.json
├── pnpm-lock.yaml
├── pnpm-workspace.yaml
├── Makefile
├── compose.yaml
├── .env.example
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── security.yml
│       └── release.yml
│
├── apps/
│   ├── web/                         # React + Vite + TypeScript
│   │   ├── src/
│   │   ├── public/
│   │   └── package.json
│   │
│   └── api/                         # Go backend
│       ├── cmd/
│       │   ├── server/              # MVP: API + worker + scheduler trong một process
│       │   │   └── main.go
│       │   ├── api/                 # Scale phase: API-only process
│       │   │   └── main.go
│       │   ├── worker/              # Scale phase: River workers
│       │   │   └── main.go
│       │   ├── scheduler/           # Scale phase: leader-safe scheduler (optional separate process)
│       │   │   └── main.go
│       │   └── migrate/             # Migration CLI wrapper
│       │       └── main.go
│       │
│       ├── internal/
│       │   ├── app/                 # Composition root, lifecycle
│       │   ├── platform/            # Shared technical adapters
│       │   └── modules/             # Feature-first modules
│       │
│       ├── db/
│       │   ├── migrations/          # Goose SQL migrations
│       │   ├── queries/             # sqlc queries theo module
│       │   ├── schema/              # Optional schema snapshots
│       │   └── seed/                # Dev/test seed
│       │
│       ├── gen/
│       │   └── db/                   # sqlc generated code; không sửa tay
│       │
│       ├── openapi/
│       │   └── openapi.json          # Generated artifact
│       │
│       ├── static/
│       │   └── web/                  # Vite dist được copy khi single binary deploy
│       │
│       ├── tests/
│       │   ├── integration/
│       │   ├── fixtures/
│       │   └── testutil/
│       │
│       ├── go.mod
│       ├── go.sum
│       ├── sqlc.yaml
│       └── Dockerfile
│
├── packages/
│   ├── api-client/                  # Generated TS client/types from OpenAPI
│   ├── ui/
│   └── config/
│
└── docs/
    └── backend-technical-spec/
```

## 2. Go backend chi tiết

```text
apps/api/internal/
├── app/
│   ├── app.go                       # Build dependencies, Start/Stop
│   ├── config.go                    # Typed configuration
│   ├── routes.go                    # Mount module routes
│   └── health.go                    # Readiness/liveness
│
├── platform/
│   ├── authn/
│   │   ├── jwt.go                   # Sign/verify access JWT
│   │   ├── refresh_token.go         # Random token/hash/rotation helpers
│   │   └── password.go              # Argon2id
│   ├── authz/
│   │   ├── permission.go
│   │   └── authorizer.go
│   ├── clock/
│   │   ├── clock.go                 # Interface
│   │   └── system.go
│   ├── config/
│   ├── database/
│   │   ├── pool.go
│   │   ├── transaction.go
│   │   └── errors.go
│   ├── errors/
│   │   ├── problem.go               # RFC 9457 response model
│   │   └── mapping.go
│   ├── httpx/
│   │   ├── middleware/
│   │   ├── pagination.go
│   │   ├── response.go
│   │   └── idempotency.go
│   ├── jobs/
│   │   ├── client.go
│   │   ├── workers.go
│   │   └── kinds.go
│   ├── logging/
│   ├── mail/
│   ├── storage/
│   └── validation/
│
└── modules/
    ├── auth/
    │   ├── domain/
    │   ├── application/
    │   ├── repository/
    │   ├── transport/http/
    │   └── module.go
    ├── users/
    ├── organizations/
    ├── academics/
    ├── resources/
    ├── questions/
    ├── assessments/
    ├── attempts/
    ├── grading/
    ├── assignments/
    ├── gradebook/
    ├── notifications/
    ├── files/
    └── audit/
```

## 3. Cấu trúc module chuẩn

Ví dụ `attempts`:

```text
modules/attempts/
├── domain/
│   ├── attempt.go                   # Entity và state transitions
│   ├── answer.go                    # Answer value/payload rules
│   ├── status.go
│   ├── errors.go
│   └── policy.go
│
├── application/
│   ├── start_attempt.go
│   ├── save_answer.go
│   ├── submit_attempt.go
│   ├── get_attempt.go
│   ├── expire_attempt.go
│   ├── dto.go
│   └── ports.go                     # Interfaces do application layer cần
│
├── repository/
│   ├── postgres.go                  # Adapter dùng sqlc
│   └── mapper.go
│
├── transport/http/
│   ├── handlers.go
│   ├── inputs.go
│   ├── outputs.go
│   └── routes.go
│
├── jobs/
│   └── expire_worker.go
│
├── module.go                        # Constructor và public facade
└── *_test.go
```

## 4. Trách nhiệm từng lớp

| Thư mục | Trách nhiệm | Cấm |
|---|---|---|
| `domain` | State, invariant, pure policy | SQL, HTTP, JSON response |
| `application` | Use case, transaction, orchestration, authorization call | Chi tiết router, raw SQL |
| `repository` | Map domain ↔ DB, gọi sqlc | Quyết định business workflow |
| `transport/http` | Parse input, gọi use case, map output/error | Truy cập DB trực tiếp |
| `jobs` | Worker adapter gọi application use case | Copy lại domain logic |
| `module.go` | Dependency wiring/public service | Global mutable state |

## 5. Interface policy

Không tạo interface cho mọi struct. Chỉ tạo khi:

- Application service cần thay adapter.
- Cần fake trong unit test.
- Có nhiều implementation hợp lệ.
- Boundary cần ngăn dependency ngược.

Ví dụ:

```go
// application/ports.go
package application

type AttemptRepository interface {
    GetForUpdate(ctx context.Context, tx pgx.Tx, orgID, attemptID uuid.UUID) (domain.Attempt, error)
    SaveAnswer(ctx context.Context, tx pgx.Tx, params SaveAnswerParams) (domain.Answer, error)
}
```

Không đặt một thư mục global `interfaces/`.

## 6. Generated code

Không sửa tay:

- `apps/api/gen/db/**`
- `apps/api/openapi/openapi.json`
- `packages/api-client/src/generated/**`

Nguồn sinh:

```text
migrations/schema + db/queries -> sqlc -> gen/db
Go API types/routes             -> Huma -> openapi.json
openapi.json                    -> openapi-typescript/client -> packages/api-client
```

## 7. Naming conventions

| Loại | Quy ước | Ví dụ |
|---|---|---|
| Package | lowercase, singular khi tự nhiên | `gradebook`, `attempts` |
| Table | snake_case plural | `attempt_answers` |
| Column | snake_case | `organization_id` |
| Route | kebab-case/plural nouns | `/grade-items` |
| Go exported type | PascalCase | `SubmitAttemptInput` |
| Permission | `resource:action` | `assessment:publish` |
| Job kind | dotted namespace | `grading.auto_grade_attempt` |
| Audit action | dotted namespace | `grade.override` |

## 8. Import rules

- `platform` không import `modules`.
- Module A không import repository của module B.
- Cross-module call qua public application facade hoặc port.
- Không có cyclic dependency.
- `cmd/*` chỉ làm composition và lifecycle.
