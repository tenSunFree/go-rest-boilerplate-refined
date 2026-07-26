# luma-lang-go

![CI](https://github.com/tenSunFree/luma-lang-go/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/tenSunFree/luma-lang-go/graph/badge.svg)](https://codecov.io/gh/tenSunFree/luma-lang-go)
[![Go](https://img.shields.io/badge/Go-1.25.0-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Gin](https://img.shields.io/badge/Framework-Gin-00ACD7)](https://gin-gonic.com)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-4CAF50)](#architecture)
[![PostgreSQL](https://img.shields.io/badge/DB-PostgreSQL-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Cache-Redis%20%2B%20Ristretto-DC382D?logo=redis&logoColor=white)](https://redis.io)
[![Auth](https://img.shields.io/badge/Auth-JWT-000000?logo=jsonwebtokens&logoColor=white)](#authentication--account-security)
[![Docker](https://img.shields.io/badge/Container-Docker-2496ED?logo=docker&logoColor=white)](https://www.docker.com)
[![Docs](https://img.shields.io/badge/API%20Docs-Swagger-85EA2D?logo=swagger&logoColor=black)](#api-documentation)
[![Testing](https://img.shields.io/badge/Testing-Unit%20%2B%20Integration%20%2B%20E2E-FF9800)](#testing)
[![CodeRabbit Reviews](https://img.shields.io/badge/Code%20Review-CodeRabbit-FF6B35)](https://coderabbit.ai)

---

## Introduction

luma-lang-go is a Go backend built with Gin, PostgreSQL, sqlx, Redis, Ristretto, JWT, Clean Architecture, Swagger, and Docker.

It demonstrates backend architecture, authentication security, live-course APIs, content delivery, caching and temporary state management, observability, automated testing, and CI quality gates.

This project is intended for independent learning, technical practice, and portfolio demonstration.

It is based on and extended from:
[go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)

---

## Related App Client

This backend can be used together with a Kotlin Multiplatform client:

- [luma-lang-kmp](https://github.com/tenSunFree/luma-lang-kmp)

The app project provides a cross-platform client foundation built with Kotlin Multiplatform, Compose Multiplatform, Clean Architecture, MVI ViewModel, Navigation3, and Koin.

Together, the two repositories demonstrate a full-stack mobile architecture covering:

- Cross-platform Android and iOS development
- RESTful API integration
- JWT authentication flows
- User profile management
- Content delivery
- Live-course functionality with Agora RTC integration

---

## Preview

<p align="left">
  <img src="https://i.postimg.cc/597mrH0y/2026-06-23-071232.png" width="500"/>
  <img src="https://i.postimg.cc/pVGBS9Tx/2026-06-23-071318.png" width="500"/>
  <img src="https://i.postimg.cc/HWZtNJxk/2026-06-23-071331.png" width="500"/>
  <img src="https://i.postimg.cc/Kcq5WKzm/2026-06-23-071343.png" width="500"/>
</p>

---

## Features

### Authentication & Account Security
- JWT-based authentication with access and refresh tokens
- Refresh-token rotation and server-side revocation through Redis
- User registration with email OTP verification before account activation
- Login brute-force protection with attempt tracking and temporary lockout
- Forgot-password and reset-password flows using time-limited reset tokens
- Asynchronous email delivery for OTP and password-reset messages
- Change-password endpoint with current-password verification
- Password-change timestamps used as token-revocation cutoffs
- Constant-time comparison paths to reduce timing-based account/OTP enumeration
- Audit logging for authentication-sensitive operations

### Live Course Sessions
- Live-course creation and management for teachers (`POST /teacher/lives/start`, `/end`)
- Agora RTC token generation for teachers and students
- Separate teacher camera and screen-sharing stream UIDs
- Live-course listing filtered by status: `scheduled`, `live`, `ended`, `cancelled`
- Student join, leave, and token-renewal endpoints (`/lives/:liveId/join`, `/leave`, `/renew-token`)
- Live participant tracking and per-user Agora UID allocation
- Live-session reminder creation and deletion, reflected in the live-course list response
- WebSocket chat URL returned as part of the join flow
- Database constraint preventing multiple active live sessions for the same teacher

### Content Delivery
- Language-learning content listing, detail retrieval, and full-text search
- Pagination support across content and user listing endpoints
- PostgreSQL-backed content storage
- JSON-based content import and seed-data tooling for local development

### Caching & Temporary State
- In-process caching with Ristretto
- Shared caching and temporary state storage with Redis
- Cache-first user lookup with PostgreSQL fallback on miss
- Cache invalidation after account-state changes
- `singleflight` request coalescing to prevent duplicate database round-trips under cache-miss contention
- Redis-backed refresh-token, OTP, and security-attempt state with expiration policies

### Observability & Reliability
- Structured application logging with Zap, request-scoped fields, and request IDs
- HTTP access logs and authentication audit logs
- Prometheus-compatible metrics for cache, mailer, and connection-pool behavior
- OpenTelemetry tracing and trace-context propagation across HTTP and use-case layers
- Health (`/health`) and readiness (`/ready`) endpoints, with readiness checks for PostgreSQL and Redis connectivity

### API Security
- JWT authentication middleware
- Per-IP request rate limiting and request body size limits
- CORS configuration
- Security headers: `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`
- Content-Security-Policy (context-aware: relaxed for `/swagger/*`, locked down elsewhere)
- HSTS enabled in production
- Password hashing with bcrypt

### Engineering Practices
- Versioned SQL migrations with matching up/down files and a dedicated migration CLI
- Generated Swagger/OpenAPI specifications with CI drift validation
- Static analysis (`go vet`), linting (`golangci-lint`), and security scanning (`gosec`)
- Formatting validation with `gofmt` / `goimports`
- Local scripts that mirror CI checks (`scripts/check.sh`, `scripts/check-integration.sh`)
- Local coverage tooling that mirrors CI's test invocation and renders a browsable HTML report (`scripts/coverage.sh`)
- Conventional Pull Request workflow with automated AI-assisted review via CodeRabbit

---

## Architecture

The project follows Clean Architecture principles, separating responsibilities into distinct layers:

```
HTTP Request
    │
    ▼
 Routes → Middleware → Handlers
    │
    ▼
 Use Cases (business layer)
    │
    ▼
 Repository Interfaces
    │
    ▼
 PostgreSQL / Redis / Ristretto / External Services (Agora, Mailer)
```

**HTTP layer** — routing, request parsing, validation, auth middleware, response serialization
`internal/http/routes`, `internal/http/handlers`, `internal/http/middlewares`, `internal/http/datatransfers`

**Business layer** — use cases, domain entities, business validation
`internal/business/domain`, `internal/business/usecases`

**Data layer** — PostgreSQL/sqlx access, Redis, Ristretto, records, migrations
`internal/datasources/drivers`, `internal/datasources/repositories`, `internal/datasources/caches`, `internal/datasources/records`, `internal/datasources/migration`

**Infrastructure layer** — JWT, logging, mailer, observability, validators, password hashing, audit
`pkg/jwt`, `pkg/logger`, `pkg/mailer`, `pkg/observability`, `pkg/validators`, `pkg/helpers`, `pkg/audit`

Dependencies point inward through interfaces, keeping business logic independent of HTTP framework and infrastructure implementations.

---

## Tech Stack

- **Go 1.25.0** — primary language for the API, CLI tools, and migrations
- **Gin** — HTTP framework for routing, middleware, and request handling
- **PostgreSQL** — primary relational database
- **sqlx** — typed query mapping on top of `database/sql`, with explicit SQL control
- **Redis** — shared cache and temporary state store for refresh tokens, OTP state, and attempt counters
- **Ristretto** — in-process cache to reduce repeated lookups within an application instance
- **JWT (`golang-jwt/jwt`)** — access/refresh token authentication
- **bcrypt** — password hashing
- **Agora RTC** — token generation and channel authorization for live sessions
- **Swagger / OpenAPI (`swaggo`)** — generated REST API documentation
- **Zap** — structured application logging
- **Prometheus client** — runtime and application metrics
- **OpenTelemetry** — HTTP tracing and trace-context propagation across application layers
- **testify** — test assertions and mocked behavior verification
- **Testcontainers for Go** — disposable PostgreSQL/Redis containers for integration tests
- **golangci-lint / gosec** — aggregated linting and security static analysis
- **Docker / Docker Compose** — local infrastructure and container builds
- **GitHub Actions** — CI for formatting, linting, tests, integration tests, and Swagger drift checks
- **CodeRabbit** — automated AI-assisted Pull Request review

---

## API Documentation

Swagger documentation is generated from annotations in the Go source.

After starting the API, the Swagger UI is available at:

```
http://localhost:8080/swagger/index.html
```

Base path: `/api/v1`

Main route groups: `/auth`, `/users`, `/contents`, `/live-courses`, `/lives`, `/teacher/lives`

Generated spec files: `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`

---

## Testing

**Unit tests** — domain logic, use cases, middleware, JWT handling, logging, and mailer, with mocked dependencies

```bash
go test ./...
```

**Integration tests** — run against disposable PostgreSQL/Redis containers via Testcontainers (Docker required)

```bash
make test-integration
```

**End-to-end tests** — cover registration, OTP verification, login, refresh-token rotation/revocation, logout, password change/reset, and brute-force lockout behavior

**Local quality checks** — runs formatting, static analysis, linting, unit tests, dependency consistency, migration validation, Swagger drift detection, and build checks.

```bash
bash scripts/check.sh
```

**Local integration checks** — runs tests tagged `integration` against real PostgreSQL and Redis via Testcontainers. Docker must be running locally.

```bash
bash scripts/check-integration.sh
```

**Coverage reporting** — the CI unit-test job generates a Go coverage report and uploads it to [Codecov](https://codecov.io) for pushes to `main`/`develop` and for Pull Requests targeting those branches. Uploads use the repository's `CODECOV_TOKEN` GitHub Actions secret. The current report includes unit-test coverage only; coverage from the Testcontainers integration-test job is not yet merged into it. The `codecov/project` and `codecov/patch` checks are currently informational and are not yet configured as required branch-protection checks.

**Local coverage report** — runs the same test invocation as CI (`-race -covermode=atomic`, falling back to a non-race run if a working C compiler/cgo isn't available locally), converts the resulting Go coverage profile to LCOV, and renders a navigable HTML report (per-directory percentages, expandable file tree) before opening it in your default browser.

```bash
bash scripts/coverage.sh
```

Requires [`gcov2lcov`](https://github.com/jandelgado/gcov2lcov) plus either `genhtml` (from the `lcov` package) or [`@lcov-viewer/cli`](https://www.npmjs.com/package/@lcov-viewer/cli) to render the tree-view HTML report; falls back to Go's built-in `go tool cover -html` output if neither is available. Set `NO_OPEN=1` to skip auto-opening the browser. Coverage exclusions are maintained solely in `codecov.yml`, so the local report intentionally mirrors CI's raw output rather than filtering separately.

---

## Environment

**Required:** Go 1.25.0, Docker, Docker Compose, Git
**Optional:** `golangci-lint`, `swag` CLI, `gcov2lcov` + `genhtml`/`lcov-viewer` (for local HTML coverage reports)

**Local runtime services:** PostgreSQL 16, Redis 7, API on port `8080`

Create the local environment file from the example, then review and update the values:

```bash
cp internal/config/.env.example internal/config/.env
```

Do not commit the real `.env` file or production credentials to version control.

---

## Local Development

```bash
git clone https://github.com/tenSunFree/luma-lang-go.git
cd luma-lang-go
```

**Start with Docker Compose** (API + PostgreSQL + Redis):

```bash
docker compose -f deploy/docker-compose.yml up --build
```

**Or run the API directly** once PostgreSQL, Redis, and `.env` are ready:

```bash
go run ./cmd/api
```

**Run migrations:**

```bash
go run ./cmd/migration -up     # roll back with -down
```

**Seed development data:**

```bash
go run ./cmd/seed
```

**Run local checks:**

```bash
bash scripts/check.sh
```

---

## Continuous Integration

GitHub Actions validates every Pull Request through automated quality gates:

- Dependency validation and formatting checks
- Static analysis (`go vet`) and `golangci-lint`
- Unit tests and Testcontainers-based integration tests
- Swagger drift checks
- Binary compilation

Local check scripts mirror CI behavior so failures can be caught before pushing.

---

## Project Structure

```
luma-lang-go
├── .github
│   ├── pull_request_template.md
│   └── workflows
│       └── ci.yml
├── cmd
│   ├── api
│   │   ├── main.go
│   │   └── server
│   ├── generate_contents_json
│   ├── import_contents
│   ├── migration
│   │   └── migrations
│   └── seed
│       └── seeders
├── data
│   ├── contents.json
│   └── lessons.json
├── deploy
│   ├── Dockerfile
│   └── docker-compose.yml
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal
│   ├── apperror
│   ├── business
│   │   ├── domain
│   │   └── usecases
│   │       ├── auth
│   │       ├── contents
│   │       ├── lives
│   │       └── users
│   ├── config
│   ├── constants
│   ├── datasources
│   │   ├── caches
│   │   ├── drivers
│   │   ├── migration
│   │   ├── records
│   │   └── repositories
│   ├── http
│   │   ├── auth
│   │   ├── datatransfers
│   │   ├── handlers
│   │   ├── middlewares
│   │   └── routes
│   └── test
│       ├── mocks
│       └── testenv
├── pkg
│   ├── audit
│   ├── clock
│   ├── helpers
│   ├── jwt
│   ├── logger
│   ├── mailer
│   ├── observability
│   └── validators
├── scripts
│   ├── check-integration.sh
│   ├── check.sh
│   ├── coverage.sh
│   └── pre-push
├── go.mod
├── go.sum
├── makefile
├── README.md
└── rest.http
```

---

## Credits

This project is created for independent learning and demonstration purposes.

Special thanks to the original author for the open-source foundation:
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)

---

## Notes

Image resources are included for learning and demonstration purposes only. Please do not use them commercially without confirming the relevant usage rights.

If any resource infringes upon third-party rights, please contact me so it can be reviewed and removed.

---

## License

This repository is currently intended for learning and portfolio demonstration.

Before distributing or using it commercially, add an explicit open-source license and confirm the licensing and usage rights of all inherited code, dependencies, and third-party assets.