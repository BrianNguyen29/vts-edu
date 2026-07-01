# Huma v2 feasibility spike — academics slice

**Branch:** `spike/huma-academics`
**ADR:** [ADR-0010 — Huma + sqlc staged groundwork](../adr/0010-huma-sqlc-staged-groundwork.md)
**Goal:** Collect bounded go/no-go evidence for adopting Huma v2 on at least
one non-auth, non-CSRF, non-streaming slice (academics) before deciding
whether to migrate the rest of the API.

## Scope

- Add Huma v2.38.0 (single new dep) and `adapters/humachi`.
- Mount a Huma v2 sub-router at `/_spike/huma` on the existing `*chi.Mux`
  for **two academics operations only**:
  - `GET  /_spike/huma/academic-terms`  — `ListTerms`
  - `POST /_spike/huma/academic-terms`  — `CreateTerm`
- Existing routes are untouched. Auth, refresh cookie, CSRF, and the
  handler-shared `auth.ActorFromRequest` / `csrf.Validate` helpers are
  reused **as-is** (no rewrites).
- The `openapi-skeleton.yaml` is unchanged. The Huma-generated OpenAPI
  doc is **evidence only**.

## Out of scope (deferred)

- Migrating any production handler to Huma.
- Auth/refresh/CSRF/streaming handlers (intentionally avoided; see ADR-0010).
- Replacing the `openapi-skeleton.yaml` with a Huma-generated spec.
- The in-process scheduler, River, or any DB/sqlc changes.

## Implementation

### Files

- `apps/api/go.mod` — adds `github.com/danielgtaylor/huma/v2 v2.38.0`.
- `apps/api/internal/features/academics/spike_huma.go` — spike wiring,
  two operations, `{data,error}` envelope types.
- `apps/api/internal/features/academics/spike_huma_test.go` — 4 unit tests.
- `apps/api/cmd/server/main.go` — hoists `academicsSvc` and `authIssuer`
  to outer scope and calls `academics.MountHumaSpike(r, ...)` before the
  existing routes are registered. The mount is a no-op in `DB_SKIP` mode.

### Wiring

- A child `*chi.Mux` carries `middleware.RequestID` + `spikeMiddleware`
  (injects the `*http.Request` into the standard context under
  `spikeRequestKey{}`). `humachi` mounts Huma on that child; the parent
  router mounts the child at `/_spike/huma`. The child is a sub-router
  so it can be unmounted in one line for rollback.
- Handlers read the request via `spikeRequestFromContext`, then call
  the existing `auth.ActorFromRequest` and `csrf.Validate` helpers.
- Huma config disables the default `$schema` field embed
  (`cfg.CreateHooks = nil`, `cfg.Transformers = nil`) so responses match
  the production envelope shape exactly.
- Response types are concrete `Status int` + `Body struct` types with
  `Data *T` and `Error *spikeErrBody` as pointer/omitempty fields, so a
  single response type covers both success and error paths.
- Huma's handler signature in v2.38 is `func(context.Context, *I) (*O, error)`,
  so the `*http.Request` is delivered via the chi-level middleware into
  the standard context, not via `huma.Context`.

## Tests

```bash
cd apps/api
go test ./internal/features/academics/ -run TestHumaSpike -v
```

All 4 tests pass:

| Test                                              | Verifies                                                          |
| ------------------------------------------------- | ----------------------------------------------------------------- |
| `TestHumaSpike_ListTerms_PreservesEnvelopeAndRequestID` | 200 with `{data:[Term]}`; X-Request-Id context flows through. |
| `TestHumaSpike_ListTerms_ForbiddenForStudent`     | 403 with `{error:{code,message,request_id}}`; student role gated. |
| `TestHumaSpike_CreateTerm_ValidationCatchesBadDate` | 400 on `start_date > end_date`; CSRF passes; role check passes.   |
| `TestHumaSpike_CreateTerm_HappyPathEnvelope`      | 201 with `{data:Term}` for admin + valid CSRF.                    |

`go test ./...` passes (no regressions in auth, attempts, resources,
gradebook, etc.). `go vet ./...` and `gofmt -l .` are clean.

## Evidence

### What worked

- Huma v2.38 mounted cleanly on a `*chi.Mux` sub-router; existing routes
  were not perturbed.
- The `{data}` success envelope and the `{error:{code,message,request_id}}`
  error envelope are both preserved by Huma when the response type is a
  struct with a `Body` field (success) and a Status field that drives
  the HTTP code (error).
- Body validation via Huma's `minLength`/`required` tags works on the
  request body.
- Auth/CSRF plumbing reuses the existing helpers with no rewrites.
- Huma generates a valid OpenAPI 3.1 document for the spike operations
  (see "Caveats" below for the error-content-type divergence).

### What required non-obvious knobs

- **Disable `$schema` embed.** `huma.DefaultConfig` registers a
  `SchemaLinkTransformer` that adds `{"$schema":".../X.json"}` to every
  response body. To match our envelope exactly, the spike sets
  `cfg.CreateHooks = nil` and `cfg.Transformers = nil`. A production
  migration would need a per-feature Huma config helper that strips
  this transformer but keeps the OpenAPI doc generation.
- **Use concrete response types.** Generic `spikeEnvelope[T any]` does
  not work with Huma v2.38's `processOutputType` — it has to see
  a concrete `Status int` + `Body struct` pattern.
- **Pointer + omitempty for union-style bodies.** Because Huma exposes
  exactly one response type per operation, the spike uses a struct with
  `Data *T` and `Error *spikeErrBody` (both pointers, both `omitempty`)
  so a single type covers both success and error paths. This is a
  spike-specific workaround; Huma also supports `OneOf`/`AnyOf` bodies,
  which the spike report leaves for follow-up.
- **Request injection via middleware.** Huma v2.38's handler signature
  is `func(context.Context, *I) (*O, error)`. To read headers/cookies
  in the handler, the chi-level `spikeMiddleware` adds the `*http.Request`
  to the standard context. Huma's `huma.Context` is still available
  inside middlewares via `humachi.Unwrap` if needed.
- **Huma v2.38 removed `huma.DefaultConfig` as a value** (now a function)
  and removed `Config.OpenAPI.Path`/`DocsPath` fields. This was a
  surprise only the first time; current spike code uses the function
  form and does not touch the removed fields.

### Caveats (would need to be resolved for a production migration)

- **Error response content type.** Huma's generated OpenAPI doc
  advertises `application/problem+json` for error responses (RFC 9457).
  Our envelope is `application/json`. The spike does not change the
  client contract — the actual response is `application/json` — but a
  full migration would need to either: (a) update the client to expect
  `problem+json` for errors, or (b) override Huma's default
  error-content-type registration.
- **Huma `NewError` override is not needed** in the spike because the
  spike handlers return a response struct rather than a `StatusError`.
  This means a future migration of all handlers would not need the
  `huma.NewError` global override that the spike initially tried.
- **X-Request-Id response header is not set** by chi's
  `middleware.RequestID`. The spike's `X-Request-Id` is on the request
  header and in the standard context, but not echoed to the response
  by chi. Production handlers set it from context in
  `response.go::respondError`. A future spike on the response-middleware
  layer would address this systematically.
- **OpenAPI doc divergence.** Huma generates its own OpenAPI doc; the
  production `openapi-skeleton.yaml` is hand-curated. For a full
  migration, either the skeleton must absorb the Huma-generated spec
  (likely the better long-term answer) or the Huma spec must be
  hand-tuned to match.

## DX observations

- Huma's per-operation type registration is verbose compared to a
  hand-written handler: each operation needs a response struct, a body
  type, and a registration call. The verbose ceremony pays off for
  complex APIs (validation, generated clients) but is friction for
  small operations.
- Generic response types do not work with v2.38; concrete types are
  required. This pushes repeated type definitions for similar
  operations.
- The `Status int` + `Body struct` pattern is opinionated; deviating
  (e.g. union bodies via `OneOf`) requires more boilerplate.

## Go / no-go recommendation

**GO** for the bounded spike scope (one slice, one feature, with the
caveats above noted in a follow-up spike plan). The spike confirms:

1. Huma v2 mounts cleanly on the existing `*chi.Mux` without disrupting
   existing routes.
2. The `{data,error}` envelope is preservable, but only with
   non-default config (`CreateHooks = nil`, `Transformers = nil`) and
   concrete response types.
3. Body validation via Huma tags is a real win for hand-rolled
   request DTOs.
4. Auth/CSRF/role checks integrate cleanly via the existing helpers.

**No-go triggers for a full migration** (i.e. reasons to pause before
scaling this spike to gradebook or attempts):

- The `application/problem+json` error-content-type divergence must be
  resolved first, ideally by extending the spike to test a different
  approach (e.g. explicit per-operation error schemas or a Huma
  middleware that rewrites error content type).
- The X-Request-Id response-header gap must be closed (or the spike
  report must accept the inconsistency for the bounded duration of
  the migration).
- The OpenAPI doc divergence must be resolved (either absorb the
  Huma-generated spec or commit to maintaining two specs).

**Suggested next spike**: extend the bounded spike to one streaming
candidate (resources download) to test how Huma handles
`http.Flusher`/SSE, which is the other risk axis the roadmap calls out
before recommending a full Huma adoption.
