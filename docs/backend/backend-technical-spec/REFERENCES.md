# References — Official/Primary Sources

Snapshot reviewed: **2026-06-29**.

## Go

- [Go Release History](https://go.dev/doc/devel/release)
- [Go `argon2` package](https://pkg.go.dev/golang.org/x/crypto/argon2)
- [Go `slog` package](https://pkg.go.dev/log/slog)

## HTTP/API

- [go-chi/chi](https://github.com/go-chi/chi)
- [Huma Documentation](https://huma.rocks/)
- [Huma OpenAPI Generation](https://huma.rocks/features/openapi-generation/)
- [OpenAPI Specification 3.1](https://swagger.io/specification/)
- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457)

## PostgreSQL & data access

- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/)
- [PostgreSQL Versioning Policy](https://www.postgresql.org/support/versioning/)
- [PostgreSQL Row Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [pgx/v5 package](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [pgxpool package](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)
- [sqlc: Using Go and pgx](https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html)
- [goose migrations](https://pressly.github.io/goose/)

## Background jobs

- [River repository and transactional enqueue overview](https://github.com/riverqueue/river)

## JWT & application security

- [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [OWASP OAuth 2.0 Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/OAuth2_Cheat_Sheet.html)
- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [OWASP API1:2023 Broken Object Level Authorization](https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/)
- [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/)

## Notes

- Luôn kiểm tra release/security advisory mới trước khi pin version production.
- Tài liệu này ưu tiên primary sources; blog/thư viện wrapper chỉ nên dùng sau review.
