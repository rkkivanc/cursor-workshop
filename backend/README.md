# MasterFabric Go Basic

A mobile-backend Go service built with Clean/Hexagonal Architecture and DDD principles. Exposes a **GraphQL API** for Auth, User profile, and Settings management, plus a **CLI** (`masterfabric_go`) for SDK code generation.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Quick Start](#quick-start)
- [Environment Variables](#environment-variables)
- [CLI — `masterfabric_go`](#cli--masterfabric_go)
- [GraphQL API Reference](#graphql-api-reference)
  - [Auth](#auth)
  - [User](#user--requires-bearer-token)
  - [Settings](#settings)
  - [Admin](#admin--requires-bearer-token--admin-role)
  - [Teardown](#teardown)
  - [Error format](#error-format)
- [Postman Collection](#postman-collection)
- [Makefile Reference](#makefile-reference)
- [Architecture](#architecture)
- [Mobile Client — Flutter](#mobile-client--flutter)
- [Adding a New Feature](#adding-a-new-feature)

---

## Tech Stack

| Concern        | Library / Tool                          |
|----------------|-----------------------------------------|
| Language       | Go 1.22+                                |
| API            | GraphQL — [gqlgen](https://gqlgen.com)  |
| Router         | go-chi/chi/v5                           |
| Database       | PostgreSQL — pgx/v5 + pgxpool           |
| Cache          | Redis — go-redis/v9                     |
| Message Queue  | RabbitMQ — amqp091-go                   |
| Auth           | JWT (golang-jwt/jwt/v5) + bcrypt        |
| CLI            | cobra (spf13/cobra)                     |
| Migrations     | golang-migrate/migrate/v4               |

---

## Project Structure

```
cmd/
  server/main.go                    # HTTP server entry point + DI wiring
  masterfabric_go/main.go           # Code-generation CLI entry point
internal/
  domain/
    iam/
      model/                        # User, Role entities + enums
      repository/                   # UserRepository interface
      event/                        # Domain events (UserRegistered, etc.)
      policy/                       # RBAC helpers: RequireAdmin, RequireRole, HasRole
    settings/
      model/                        # UserSettings, AppSettings entities
      repository/                   # Repository interfaces
      event/                        # Domain events
  application/
    auth/
      usecase/                      # Register, Login, RefreshTokens, Logout
      dto/                          # Auth request/response DTOs
    user/
      usecase/                      # GetProfile, UpdateProfile, DeleteAccount
      dto/                          # User request/response DTOs
    settings/
      usecase/                      # GetMySettings, UpdateMySettings, GetAppSettings
      dto/                          # Settings request/response DTOs
    admin/
      usecase/                      # ListUsers, GetUserByID, SuspendUser, ChangeRole
      dto/                          # Admin request/response DTOs
  infrastructure/
    postgres/                       # pgx repository implementations + migrations
    redis/                          # Redis client helpers
    rabbitmq/                       # RabbitMQ event bus
    auth/                           # JWT service + bcrypt helper
    graphql/
      resolver/                     # gqlgen resolvers (one file per domain)
      schema/                       # GraphQL schema files — source of truth
  codegen/
    parser/schema_parser.go         # GraphQL schema parser
    dart/                           # Dart SDK generator
  shared/
    config/                         # Viper / env config
    logger/                         # slog structured logger
    middleware/                     # Auth + RequestID middleware
    errors/                         # Domain error sentinel values
    events/                         # EventBus interface
    health/                         # Liveness + readiness check handlers
    database/                       # Postgres pool helper
    cache/                          # Redis client helper
    version/                        # Service name/version constants
deployments/
  docker-compose.yml                # Postgres, Redis, RabbitMQ, App
  Dockerfile
postman/
  masterfabric.collection.json      # Postman collection (all requests + tests)
  masterfabric.environment.json     # Postman environment (local defaults)
sdk/
  dart_go_api/                      # GENERATED — do not edit by hand
```

---

## Quick Start

### 1. Prerequisites

- Go 1.22+
- Docker + Docker Compose

### 2. Start infrastructure

```bash
make docker-infra
# starts postgres (5433), redis (6380), rabbitmq (5673 / management 15673)
```

Or start everything (infra + app container):

```bash
make docker-up
```

### 3. Configure environment

```bash
cp .env.example .env
# edit .env if needed — defaults work with docker-infra targets
```

### 4. Run the server

```bash
# build + run binary
make run

# or with hot-reload (requires: go install github.com/air-verse/air@latest)
make dev
```

Server starts at `http://localhost:8080`.

| Endpoint              | Description                    |
|-----------------------|--------------------------------|
| `GET /health`         | Liveness probe                 |
| `GET /health/ready`   | Readiness probe (DB + Redis)   |
| `POST /graphql`       | GraphQL endpoint               |
| `GET /graphql`        | GraphQL Playground (dev only)  |

---

## Environment Variables

Copy `.env.example` to `.env`. All variables and their defaults:

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=60s

# PostgreSQL
DATABASE_DSN=postgres://masterfabric:masterfabric@localhost:5433/masterfabric_basic?sslmode=disable
DATABASE_MAX_CONNS=20
DATABASE_MIN_CONNS=2
DATABASE_MAX_CONN_IDLE=5m

# Redis
REDIS_ADDR=localhost:6380
REDIS_PASSWORD=
REDIS_DB=0

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5673/
RABBITMQ_ENABLED=true
RABBITMQ_EXCHANGE=masterfabric.events

# JWT — change JWT_SECRET in production (min 32 chars)
JWT_SECRET=change-me-in-production-at-least-32-chars
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Logging: debug | info | warn | error
LOG_LEVEL=info
LOG_FORMAT=json

# GraphQL — disable introspection in production
GRAPHQL_INTROSPECTION=true
```

---

## CLI — `masterfabric_go`

The `masterfabric_go` binary is a code-generation tool. It reads the GraphQL schema and generates typed client SDKs for target platforms.

### Build the CLI

```bash
make build-cli
# produces bin/masterfabric_go
```

Or manually:

```bash
go build -o bin/masterfabric_go ./cmd/masterfabric_go
```

### Commands

#### `generate dart`

Generate a Dart SDK package from the GraphQL schema.

```bash
# default paths (schema: internal/infrastructure/graphql/schema, output: sdk/dart_go_api)
./bin/masterfabric_go generate dart

# explicit paths
./bin/masterfabric_go generate dart \
  --schema internal/infrastructure/graphql/schema \
  --output sdk/dart_go_api
```

Via Makefile:

```bash
make generate-dart   # builds CLI then runs generate dart
```

**Output** — `sdk/dart_go_api/` (never edit by hand):

```
sdk/dart_go_api/
  pubspec.yaml
  lib/
    dart_go_api.dart          # barrel export
    src/
      models/                 # enums.dart, inputs.dart, models.dart
      queries/                # documents.dart (gql DocumentNodes)
      client/                 # masterfabric_client.dart
```

> Every time a `.graphqls` schema file changes, re-run `make generate-dart` to keep the SDK in sync.

#### Adding a new SDK target (e.g. Swift)

1. Create `internal/codegen/swift/` mirroring the `dart/` package structure
2. Implement `Generate(schemaDir, outputDir string) error` as the entry point
3. Register the command in `cmd/masterfabric_go/main.go`
4. Add `make generate-swift` to the `Makefile`
5. Output goes to `sdk/swift_go_api/`

---

## GraphQL API Reference

All requests go to `POST /graphql` with `Content-Type: application/json`.  
Authenticated requests require `Authorization: Bearer <accessToken>`.

### Auth

#### Register

```graphql
mutation Register($input: RegisterInput!) {
  register(input: $input) {
    accessToken
    refreshToken
    expiresIn
    user { id email displayName avatarURL role }
  }
}
```

```json
{ "input": { "email": "user@example.com", "password": "P@ssword1234", "displayName": "Jane" } }
```

#### Login

```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    accessToken
    refreshToken
    expiresIn
    user { id email displayName avatarURL role }
  }
}
```

```json
{ "input": { "email": "user@example.com", "password": "P@ssword1234" } }
```

#### Refresh Tokens

```graphql
mutation RefreshTokens($input: RefreshInput!) {
  refreshTokens(input: $input) {
    accessToken
    refreshToken
    expiresIn
    user { id email displayName avatarURL role }
  }
}
```

```json
{ "input": { "userID": "<uuid>", "refreshToken": "<token>" } }
```

#### Logout

```graphql
mutation Logout($input: LogoutInput!) {
  logout(input: $input)
}
```

```json
{ "input": { "userID": "<uuid>", "accessToken": "<token>", "refreshToken": "<token>" } }
```

Blacklists the access token in Redis and deletes the refresh token.

---

### User  _(requires Bearer token)_

#### Get Profile

```graphql
query Me {
  me { id email displayName avatarURL bio status role createdAt updatedAt }
}
```

#### Update Profile

```graphql
mutation UpdateProfile($input: UpdateProfileInput!) {
  updateProfile(input: $input) {
    id email displayName avatarURL bio status role createdAt updatedAt
  }
}
```

```json
{ "input": { "displayName": "New Name", "avatarURL": "https://...", "bio": "Hello!" } }
```

All fields are optional — only provided fields are changed.

#### Delete Account

```graphql
mutation DeleteAccount {
  deleteAccount
}
```

---

### Settings

#### My Settings  _(requires Bearer token)_

```graphql
query MySettings {
  mySettings { id userID notificationsOn theme language timezone updatedAt }
}
```

#### Update My Settings  _(requires Bearer token)_

```graphql
mutation UpdateMySettings($input: UserSettingsInput!) {
  updateMySettings(input: $input) {
    id userID notificationsOn theme language timezone updatedAt
  }
}
```

```json
{ "input": { "notificationsOn": true, "theme": "DARK", "language": "en", "timezone": "UTC" } }
```

Available `Theme` values: `LIGHT` | `DARK` | `SYSTEM`

#### App Settings  _(public)_

```graphql
query AppSettings {
  appSettings { key value description }
}
```

---

### Admin  _(requires Bearer token + `ADMIN` role)_

All admin operations are protected by both authentication and role authorisation. A valid token with role `USER` or `MODERATOR` receives a `FORBIDDEN` error, not `UNAUTHENTICATED`.

#### List Users

```graphql
query AdminListUsers($page: Int, $pageSize: Int) {
  adminUsers(page: $page, pageSize: $pageSize) {
    users { id email displayName role status createdAt }
    totalCount
    page
    pageSize
  }
}
```

```json
{ "page": 1, "pageSize": 10 }
```

The test script saves the first returned user ID to `adminTargetUserId` in the environment for subsequent admin requests.

#### Get User By ID

```graphql
query AdminGetUser($id: UUID!) {
  adminUser(id: $id) {
    id email displayName avatarURL bio status role createdAt updatedAt
  }
}
```

```json
{ "id": "{{adminTargetUserId}}" }
```

#### Change Role

```graphql
mutation AdminChangeRole($id: UUID!, $role: UserRole!) {
  adminChangeRole(id: $id, role: $role) {
    id email role updatedAt
  }
}
```

```json
{ "id": "{{adminTargetUserId}}", "role": "MODERATOR" }
```

Available `UserRole` values: `ADMIN` | `MODERATOR` | `USER`

#### Suspend / Reactivate User

```graphql
mutation AdminSuspendUser($id: UUID!, $suspend: Boolean!) {
  adminSuspendUser(id: $id, suspend: $suspend) {
    id email status updatedAt
  }
}
```

```json
{ "id": "{{adminTargetUserId}}", "suspend": true }
```

Pass `"suspend": false` to reactivate. The operation is **idempotent** — suspending an already-suspended user returns `SUSPENDED` without error.

---

### Teardown

#### Delete Account  _(requires Bearer token)_

Permanently deletes the authenticated user's account. The Postman pre-request script includes a **production-URL safety guard** that aborts the request if `baseUrl` contains `production`, `prod.`, or `.io`.

```graphql
mutation DeleteAccount {
  deleteAccount
}
```

On success the Postman test script clears `accessToken`, `refreshToken`, `userId`, `userEmail`, and `userDisplayName` from the environment.

---

### Error format

GraphQL errors are returned in the response body regardless of HTTP status (always `200`):

```json
{
  "errors": [
    {
      "message": "invalid email or password",
      "extensions": { "code": "INVALID_CREDENTIALS" }
    }
  ],
  "data": null
}
```

| Code                  | Meaning                                         |
|-----------------------|-------------------------------------------------|
| `INVALID_CREDENTIALS` | Wrong email or password                         |
| `EMAIL_TAKEN`         | Email already registered                        |
| `USER_NOT_FOUND`      | User does not exist                             |
| `NOT_FOUND`           | Requested resource does not exist               |
| `TOKEN_EXPIRED`       | JWT or refresh token expired                    |
| `UNAUTHENTICATED`     | Missing or invalid Bearer token                 |
| `FORBIDDEN`           | Authenticated but insufficient role/permission  |

---

## Postman Collection

A complete Postman collection is included in `postman/`.

### Import

1. Open Postman
2. **Import** → select `postman/masterfabric.collection.json`
3. **Import** → select `postman/masterfabric.environment.json`
4. Select the **MasterFabric — Local** environment

### What's included

| Folder    | Requests                                                                                                                                                |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| Health    | Health Check, Readiness Check                                                                                                                           |
| Auth      | Register, Login, Login (wrong password), Register (duplicate), Refresh Tokens, Refresh (invalid token), Logout, Re-Login                               |
| User      | Me, Me (unauthenticated), Update Profile, Update Profile (partial), Me (verify after update)                                                            |
| Settings  | App Settings (public), My Settings, My Settings (unauthenticated), Update My Settings, Update My Settings (Light Theme), My Settings (verify after update) |
| Admin     | List Users, Get User By ID, Change Role to MODERATOR, Suspend User, Reactivate User, Change Role to ADMIN, Change Role back to USER, Get User (not found), Suspend Already Suspended (idempotency), List Users Page 2 (pagination), Forbidden (regular user token), List Users (no auth) |
| Teardown  | Delete Account, Me — After Delete (error path)                                                                                                          |

### Automatic token management

Collection scripts handle all token lifecycle automatically:

- **Register / Login** — saves `accessToken`, `refreshToken`, `userId` to the environment
- **Refresh Tokens** — rotates and overwrites the stored tokens
- **Logout** — clears `accessToken` and `refreshToken` from the environment
- **Re-Login** — restores tokens so subsequent folders run without manual steps
- **Admin List Users** — saves the first returned user UUID to `adminTargetUserId` for downstream admin requests

### Environment variables

| Variable             | Description                                                        |
|----------------------|--------------------------------------------------------------------|
| `baseUrl`            | Server base URL (default: `http://localhost:8080`)                 |
| `graphqlEndpoint`    | Full GraphQL URL (auto-set to `{{baseUrl}}/graphql`)               |
| `accessToken`        | JWT access token (managed by scripts)                              |
| `refreshToken`       | Refresh token (managed by scripts)                                 |
| `userId`             | Authenticated user UUID (managed by scripts)                       |
| `testEmail`          | Email used in test runs                                            |
| `testPassword`       | Password used in test runs                                         |
| `adminTargetUserId`  | UUID of the first user from Admin List Users (managed by scripts)  |
| `regularUserToken`   | Manually set — a token for a `USER`-role account (Admin FORBIDDEN test) |

### Run the full collection (Newman)

```bash
npm install -g newman

newman run postman/masterfabric.collection.json \
  --environment postman/masterfabric.environment.json \
  --delay-request 100
```

---

## Makefile Reference

```
make build          compile the server binary → bin/server
make build-cli      compile the CLI binary → bin/masterfabric_go
make run            build + run the server (requires infra)
make dev            run with air hot-reload
make generate       re-run gqlgen code generation
make generate-dart  build CLI + regenerate sdk/dart_go_api
make tidy           go mod tidy
make lint           golangci-lint run ./...
make test           go test -race -count=1 ./...
make docker-up      start all containers (infra + app)
make docker-infra   start only postgres, redis, rabbitmq
make docker-down    stop and remove all containers
make docker-logs    tail app container logs
make clean          remove bin/
```

---

## Architecture

The project follows **Clean/Hexagonal Architecture** with **DDD** principles.

```
Delivery (GraphQL resolvers)
        │
        ▼
Application (use cases + DTOs)
        │
        ▼
Domain (entities + repository interfaces)
        ▲
        │
Infrastructure (Postgres, Redis, RabbitMQ, JWT)
```

**Dependency rule (strict):**
- Domain — zero external imports
- Application — imports only Domain
- Infrastructure — imports Domain + Application
- GraphQL resolvers — import Application DTOs + Infrastructure implementations
- `codegen` — standalone; never imports Application or Domain

### Auth flow

- Access token: JWT, 15-minute TTL
- Refresh token: opaque token stored in Redis with 7-day TTL
- Logout: access token blacklisted in Redis; refresh token deleted

### Event bus

- Exchange: `masterfabric.events` (RabbitMQ topic)
- Routing keys: `iam.user.registered`, `iam.user.login`, `settings.updated`, etc.
- Consumers are idempotent

### RBAC — role-based access control

Roles are **hierarchical** and enforced at the use-case level (not at the resolver):

```
USER (1) < MODERATOR (2) < ADMIN (3)
```

`internal/domain/iam/policy/rbac.go` exposes helpers used at the top of every protected use case's `Execute()` method:

| Helper                   | Minimum role required |
|--------------------------|-----------------------|
| `policy.RequireAdmin(ctx)`     | `ADMIN`           |
| `policy.RequireModerator(ctx)` | `MODERATOR`       |
| `policy.RequireRole(ctx, r)`   | exact role `r`    |
| `policy.HasRole(ctx, r)`       | bool, no error    |

A caller with `USER` or `MODERATOR` role hitting an admin use case receives `ErrForbidden` → GraphQL error code `FORBIDDEN`. A caller with no token receives `ErrUnauthorized` → `UNAUTHENTICATED`. These are distinct sentinel errors in `internal/shared/errors/`.

### Redis key schema

```
mf:{scope}:{id}:{field}
```

Examples: `mf:auth:refresh:<userID>`, `mf:auth:blacklist:<tokenHash>`

---

## Mobile Client — Flutter

When forking this repo you likely want a Flutter front-end that talks to this API. The **MasterFabric CLI** scaffolds a complete Flutter project pre-wired with the MasterFabric architecture (MVVM + BLoC/Cubit) and ready to consume the generated `dart_go_api` SDK.

### Install

```bash
dart pub global activate masterfabric_cli
```

### Create a Flutter project

```bash
# basic
masterfabric create my_app

# with org + description
masterfabric create my_app --org com.mycompany -d "My awesome app"
```

### What gets generated

```
my_app/
├── .cursor/                    # AI-assisted development (rules, agents, skills)
├── lib/
│   ├── main.dart               # Entry point with MasterApp.runBefore
│   ├── app/
│   │   ├── app.dart            # App widget with theme + MasterApp
│   │   ├── di/injection.dart   # GetIt + Injectable DI setup
│   │   └── routes.dart         # GoRouter configuration
│   ├── theme/
│   │   ├── app_theme.dart      # Theme constants
│   │   └── theme_builder.dart  # ThemeData builder
│   └── views/
│       ├── home/               # Home view — cubit + state
│       ├── profile/            # Profile view — cubit + state
│       └── settings/           # Settings view — theme cubit
├── assets/
│   ├── app_config.json         # App configuration
│   └── i18n/en.i18n.json       # English translations (Slang)
├── android/                    # Permissions pre-configured
├── ios/                        # Info.plist permissions pre-configured
├── pubspec.yaml                # masterfabric_core dependency included
└── slang.yaml                  # i18n configuration
```

### Architecture pattern (every view)

| Layer     | Base class                    | Role                                |
|-----------|-------------------------------|-------------------------------------|
| State     | `Equatable`                   | Immutable state with `copyWith()`   |
| Cubit     | `BaseViewModelCubit<S>`       | Business logic, `@injectable`       |
| View      | `MasterViewCubit<V, S>`       | UI — `initialContent()` + `viewContent()` |

### Connect to this backend

1. Generate the Dart SDK from this repo: `make generate-dart`
2. Copy or path-reference `sdk/dart_go_api/` in the Flutter project's `pubspec.yaml`:

```yaml
dependencies:
  dart_go_api:
    path: ../masterfabric_go_basic/sdk/dart_go_api
```

3. Instantiate `MasterfabricClient` with the server URL and inject it via GetIt.

### Requirements

- Dart SDK ^3.9.2
- Flutter SDK on PATH
- `masterfabric_core` accessible (pub.dev or local path)

> Package: [pub.dev/packages/masterfabric_cli](https://pub.dev/packages/masterfabric_cli)  
> Publisher: [masterfabric.co](https://pub.dev/publishers/masterfabric.co)

---

## Adding a New Feature

1. Define model in `internal/domain/<context>/model/`
2. Define repository interface in `internal/domain/<context>/repository/`
3. Create use case in `internal/application/<context>/usecase/`
4. Create DTO in `internal/application/<context>/dto/`
5. Implement repository in `internal/infrastructure/postgres/<context>/`
6. Add GraphQL schema in `internal/infrastructure/graphql/schema/<context>.graphqls`
7. Re-run gqlgen: `make generate`
8. Wire resolver in `internal/infrastructure/graphql/resolver/`
9. Wire dependencies in `cmd/server/main.go`
10. Re-generate the SDK: `make generate-dart`
