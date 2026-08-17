# Go Estate

[中文](README.md) | English

## Project Overview

Go Estate is a Go backend service for real-estate transaction data. Built on Gin, it uses a clear Controller / Service / DAO architecture around users, authentication, and daily transaction data. Session security, role-based authorization, observability, asynchronous tasks, containerized deployment, and CI testing are designed as one engineering loop rather than as isolated features. The project is intended to demonstrate a maintainable, testable, and frontend-friendly backend service in an open-source setting.

This README is written for a technical interview review. It highlights the engineering trade-offs behind the stack, verifiable test strategy, and an end-to-end design from the client entry point to runtime observability.

### System Architecture

> The diagram below is a system architecture overview that renders directly on GitHub. It can later be replaced with a deployment topology or cloud-resource diagram.

```mermaid
flowchart TB
    Client["Client Browser\nhttp://localhost"] --> Nginx["Nginx\nUnified entry, reverse proxy, and access logs"]
    Nginx --> Home["index.html\nSwagger / Grafana landing page"]
    Nginx --> API["Gin API\n/api/v1 · /ws · Swagger · /metrics"]
    Nginx --> Grafana["Grafana"]
    Nginx --> Prometheus["Prometheus"]

    subgraph Core["Application and Data Layer"]
        API --> Services["User · Auth · Daily_Data\nController / Service"]
        Services --> MySQL[("MySQL\nSQLC · Migration")]
        Services --> Redis[("Redis\nSession Cache")]
        Services --> MinIO[("MinIO\nAvatar Object Storage")]
        API --> WS["WebSocket Manager\nPer-user multi-device connections"]
        Services --> AsynqClient["Asynq Client\nTask dispatch"]
        AsynqClient --> Redis
        Redis --> AsynqWorker["Asynq Worker\nSession cleanup tasks"]
        AsynqWorker --> Redis
    end

    Prometheus -. "scrapes /metrics" .-> API
    API --> Slog["slog\nStructured application logs"]
    Nginx --> Promtail["Promtail\nAccess-log collection"]
    Promtail --> Loki
    Loki --> Grafana
    Prometheus --> Grafana
```

## Core Technology Stack

| Category | Technology | Project usage and problem addressed |
| --- | --- | --- |
| Infrastructure | Docker, Docker Compose | Orchestrates the API, MySQL, Redis, database migration, and observability components as containers. Migration runs after MySQL becomes healthy, and the API starts only after migration completes, reducing drift between local and CI environments. |
| Infrastructure | Nginx | Provides a unified entry point, reverse proxy, and access logs for the API, Swagger, Grafana, and Prometheus. |
| Infrastructure | GitHub Actions | CI is split into unit and integration stages: it runs coverage tests and build checks first, then starts the Docker environment, waits for migration, and runs tests against a real database. |
| Storage | MySQL, golang-migrate | MySQL persists users, sessions, and daily transaction data. Versioned migrations make schema and seed-data evolution repeatable. |
| Storage | SQLC | Generates type-safe DAO code from SQL queries, avoiding handwritten scans and string-built queries. The Store also abstracts transactions so password changes and Session invalidation share one transactional boundary. |
| Storage | Redis | Caches Refresh Token Sessions. Token refresh checks the cache first, falls back to MySQL on a miss, and asynchronously repopulates the cache—balancing read performance and consistency. |
| Storage | MinIO, presigned POST | Compose starts MinIO and initializes a default avatar. The API issues upload forms constrained by file type, a 1 KiB–5 MiB size range, and expiration, so files bypass the Gin service. |
| Observability | Prometheus, Grafana | Gin exports Prometheus metrics, while Grafana uses provisioned data sources and dashboards to surface service health. |
| Observability | Loki, Promtail, slog | Compose provisions log-aggregation components, and Promtail defines Gin / Nginx log collection paths. The service uses structured `slog` fields for HTTP status, business code, latency, path, and errors, providing a consistent foundation for Loki search, diagnostics, and alerting. |
| Async processing | Asynq | Uses Redis to queue Session-cache cleanup tasks. Logout and password changes first complete the critical database state transition, then remove cache entries asynchronously with timeout and retry policies so non-critical I/O does not block the request. |
| Real-time | Gorilla WebSocket | Exposes an authenticated `/ws` demo endpoint. Connections are grouped by username and broadcast across a user’s devices; a first-time Daily_Data query asynchronously pushes its point-consumption notice. |
| Performance diagnostics | `wrk`, `pprof` | The `loadtest` environment enables pprof. The Makefile provides login / health-check load tests and CPU, heap, mutex, and block profiling commands. |

## Modules and Test Assurance

### Core Business Modules

- **User and avatars**: Supports normal-user registration, admin-created VIP accounts, and profile lookup and updates. Avatars use MinIO object storage: the service issues restricted presigned POST permissions for browser-direct uploads, and creation or updates validate that the object exists. The Service layer enforces the “owner or admin” access boundary. A password update invalidates related Sessions transactionally and clears the cache asynchronously, preventing stale tokens from remaining valid.
- **Auth**: Handles login, Access Token refresh, and device-scoped logout. Login issues Access / Refresh Tokens, writes the Refresh Token to an `HttpOnly` cookie, and records Session information with device ID, User-Agent, and client IP. Refresh validation is Redis-first with MySQL fallback.
- **Daily_Data and query points**: Offers per-day, date-range, and full-data queries. Access scope maps directly to roles: `User` can query a day, `VIP` can query a period, and `Admin` can query all data. Day and period requests use a two-part cache for query records and data. On a miss, one transaction atomically records the first query, deducts points, and fetches data as needed, preventing duplicate charges and cache penetration.
- **WebSocket demo (real-time point notices)**: `GET /ws` requires a valid Access Token. A connection manager keeps `username → multiple device connections`, using read/write pumps, a 512 B inbound limit, Ping/Pong heartbeats, write deadlines, and slow-client cleanup. A first successful query asynchronously broadcasts a `QUERY_CONSUMPTION` message such as “query succeeded; N query points deducted” to every connected device for that user.
- **Entry protection and lifecycle**: Redis-backed IP limiting protects public registration, login, and token-refresh routes (five requests per minute). On `SIGINT` / `SIGTERM`, the service stops accepting new HTTP work, drains HTTP and Asynq work, and then closes database and cache connections. pprof is enabled only in the separate `loadtest` environment.

### Layered Test Strategy and Verified Results

The project uses **Service / Controller unit tests, DAO integration tests against a real database, and Controller happy-path E2E tests**, balancing speed, business-branch coverage, and confidence in real dependencies. The following is a local verification snapshot taken on **2026-08-17**.

| Verification layer | Result from this run | Coverage focus |
| --- | --- | --- |
| Unit tests | `go test -v -short ./...` passed; 52 top-level test functions in 39 test files | Controllers, Services, DAO error paths, JWT, middleware, and the WebSocket upgrade flow |
| Aggregate statement coverage | **57.3%** with `-covermode=atomic -coverpkg=./...` | A deliberately conservative all-production-package denominator, including untested startup, configuration, and generated code |
| Core unit-test coverage | User Service **98.3%**, Daily_Data Service **97.7%**, Middleware **95.0%**; Auth / Daily_Data Controllers **91.9% / 92.3%** | High-risk authorization, point-deduction transactions, cache-hit combinations, binding, and route-security boundaries |
| Integration / E2E tests | `go test -v -p 1 -tags=integration ./...` passed | Real MySQL, Redis, and MinIO; three Controller happy-path flows for User, Auth, and Daily_Data |

- **Controller unit tests**: Gin test requests with mocked Services and Token Makers validate routes, middleware, request binding, authorization, and the common response contract.
- **Service unit tests**: `gomock` isolates Store, Redis Cache, object storage, and task-dispatch dependencies, so authorization, Session handling, avatar-upload permissions, transaction orchestration, cache fallback, and error mapping are tested directly. Daily_Data explicitly covers Double Hit, both Single Hit cases, Double Miss, first-query charging, and cache-write failure.
- **DAO integration tests**: DAO tests use the `integration` build tag and a real MySQL instance to validate SQLC-generated queries, unique constraints, Session state, query records, and data operations instead of relying only on SQL mocks.
- **Controller E2E tests**: The User flow checks registration, avatar-upload permission, admin VIP creation, login, profile read/update; Auth checks login, refresh, and device-scoped logout; Daily_Data checks User / VIP / Admin boundaries for day, period, and full-data queries.
- **CI guardrail**: `.github/workflows/test.yml` first runs `go test -v -cover -short ./...`, then starts the complete Compose dependency stack and migration, and finally runs `go test -v -p 1 -tags=integration ./...`. It catches logic regressions as well as real SQL / schema / cache / object-storage integration failures.

Reproduce locally with `go test -v -short ./...`. For real dependencies, run `docker compose up --build -d`, then run `DB_SOURCE='root:123456@tcp(127.0.0.1:3306)/go_estate?parseTime=true' go test -v -p 1 -tags=integration ./...`.

## Engineering Highlights

### Middleware: Request Context and Security Boundaries First

Business APIs compose middleware through layered route groups. Requests under `/api/v1` pass metadata validation first; protected routes then pass authentication; individual resources add role checks where needed. On failure, middleware sends the common business response and aborts the request chain, ensuring an unverified request never reaches a Controller.

- **Metadata extraction and validation**: `RequireMetadata` requires `X-Device-ID` and `User-Agent` and stores them together with the client IP in Gin Context. Auth uses this context to bind Refresh Sessions to a device, client characteristics, and IP, supporting multi-device session management, audit trails, and device-scoped logout instead of relying on a stateless token alone.
- **Token authentication**: `RequireAuth` accepts only a well-formed `Bearer` header and verifies JWT signature, expiration, and token type. Only an Access Token can reach protected resources. The verified payload is stored in Context, so Controllers and Services consume trusted identity data without reparsing tokens.
- **Role authorization**: `RequireRoles` reads the authenticated role from the payload and compares it with the role set declared by the route. VIP creation and day / period / full-data queries express authorization close to the route, creating an auditable RBAC boundary.
- **Structured request logs**: `SlogMiddleware` records status, business code, method, path, source IP, latency, and internal errors after each request. It derives Info / Warn / Error severity from HTTP status and business code, making application logs useful for diagnostics and alerting.

### Route Bridging and Centralized Response Delivery

The routing layer organizes public metadata routes and authenticated routes, then wires each module with its Controller and authorization policy. Controllers follow a single “return data or return an error” model. `response.Wrapper` bridges that model to Gin’s native Handler: successful paths receive the common success structure, while failures flow through one error handler.

- Controllers focus on request binding, DTO conversion, and Service calls rather than repeating JSON serialization, status-code decisions, and logging at every branch.
- Successful responses always use `code / msg / data`; error responses use `code / msg`. The frontend can therefore build interceptors and user feedback around one stable contract.
- `NoRoute` also returns the common result structure instead of a framework-default 404 page. It intentionally retains HTTP 404 because a missing route is a protocol-level error; normal business failures use the business-code policy below.

### `BizError` and Unified Responses: A Frontend-Friendly Error Contract

`BizError` is the cross-layer error language. It contains a stable public `Code`, a user-presentable `Msg`, and an underlying `Err` retained only for server-side diagnosis. Services use `WithErr` to preserve root causes from databases, caches, or task systems without exposing implementation details to clients.

| Error category | Business-code range | HTTP policy | Frontend value |
| --- | --- | --- | --- |
| Validation | `400xx` | `200` | Displays field or format feedback; validation errors are translated into readable messages. |
| Authentication / authorization | `401xx` | `200` | Triggers login, token refresh, or a permission prompt instead of treating a business rejection as a network failure. |
| Resource / conflict | `404xx`, `409xx` | `200` | Reliably distinguishes an absent resource from expected conflicts such as duplicate username or email. |
| Server failure | `500xx` or unknown errors | `500` | Lets the client show a fallback message while the server retains the original error for investigation. |

- `response.Fail` uses `errors.As` to identify request-binding errors and `BizError`. The former becomes a common validation response; the latter is emitted by business code. Unknown errors are collapsed into the generic `50000` response to prevent internal-detail leakage.
- Expected business failures return HTTP `200` and convey the result through `Code + Msg`. This gives browser, mobile, and gateway clients one stable rule for login expiry, insufficient permission, and resource conflicts.
- Only `500xx` business failures and unclassified errors return HTTP `500`. The `NoRoute` behavior above is the intentional HTTP `404` exception for a protocol-level error, while still using the same JSON result structure.
- `response.Fail` records errors in Gin Context, and `SlogMiddleware` emits them with the associated `BizCode`. This creates a two-track error model: stable for the frontend and traceable for backend operators.

## Quick Start

### Prerequisites

- Docker Desktop with Docker Compose installed.
- The Compose configuration contains local-development default credentials. Use secure environment variables or a secret manager before any production deployment.

### Start the Full Stack

```bash
git clone https://github.com/raozhaizhu/go-estate.git
cd go-estate
docker compose up -d --build
```

For legacy Docker Compose installations:

```bash
docker-compose up -d --build
```

This starts the API, MySQL, Redis, Migration, Nginx, Prometheus, Grafana, Loki, Promtail, MinIO, and the MinIO initialization task. The API starts after a successful migration. When the services are ready, visit:

- API health check: `http://localhost:8080/ping`
- **System landing page**: `http://localhost/`. The root path serves `index.html` with basic service information and links to Swagger and Grafana.
- Swagger: `http://localhost:8080/swagger/index.html`
- Grafana: `http://localhost/grafana/`
- Prometheus: `http://localhost/prometheus/`
- MinIO Console: `http://localhost:9001/`

The WebSocket demo endpoint is `ws://localhost:8080/ws`. Connect with a valid Access Token in the `Authorization: Bearer <token>` header; after a first Daily_Data query that charges points, the client receives a `QUERY_CONSUMPTION` notification.

If the existing image is sufficient, omit `--build`: `docker compose up -d`.
