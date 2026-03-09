# go-clean-auth-api

Production-style REST API in Go that demonstrates clean architecture, authentication strategies, PostgreSQL design, middleware, validation, structured logging, testing, and Docker-based deployment readiness.

## Project Overview

`go-clean-auth-api` is a backend portfolio project designed to reflect real-world engineering practices:

- Clean architecture separation (`controllers -> usecases -> repositories -> database`)
- Echo-based REST API design
- PostgreSQL access via `pgx`
- Dual authentication:
  - JWT for user-protected endpoints
  - API key/secret for selected machine-to-machine endpoints
- Request validation and centralized error handling
- Structured JSON logging
- Pagination/filtering/sorting patterns
- Migration tooling and Docker deployment

## Architecture

### Layered responsibilities

- `controllers`:
  - HTTP transport concerns only (bind, validate, parse params, return JSON)
- `usecase`:
  - Business logic and orchestration
  - Authorization/ownership checks
  - Validation beyond transport layer
- `repository`:
  - Persistence contracts and PostgreSQL implementations
  - Parameterized, safe SQL operations
- `domain`:
  - Core entities (`users`, `api_clients`, `projects`, `tasks`)

### Request lifecycle

1. Request enters Echo with middleware chain (request ID, logging, CORS, gzip, recover, security headers, auth).
2. Controller binds and validates request DTOs.
3. Controller invokes usecase with normalized input.
4. Usecase applies business rules and calls repositories.
5. Repository performs DB operations through `pgx`.
6. Response wrapper returns consistent envelope.

## Folder Structure

```text
cmd/api
cmd/migrate
internal/config
internal/database
internal/domain
internal/repository
internal/repository/postgres
internal/usecase
internal/http/controllers
internal/http/routes
internal/http/middleware
internal/http/requests
internal/http/responses
internal/pkg/logger
internal/pkg/validator
internal/pkg/apperror
migrations
docs
```

## Features

- `users`
  - Register user
  - Login and JWT issuance
- `api_clients`
  - Register API client
  - Generate API key + secret (secret shown once)
- `projects`
  - Create/read/update/delete (soft delete)
- `tasks`
  - Create/read/update/delete (soft delete)
  - Assign tasks to users
  - Update completion status
  - List with pagination, filtering, sorting
- `auth`
  - JWT middleware (Bearer validation)
  - API key middleware (`X-API-Key`, `X-API-Secret`)
- `health`
  - Liveness endpoint

## Tech Stack

- Go
- Echo
- PostgreSQL
- pgx (`pgxpool`)
- JWT (`github.com/golang-jwt/jwt/v5`)
- bcrypt password hashing
- Docker + docker-compose
- Makefile
- OpenAPI (`docs/openapi.yaml`)

## Quick Start

### 1) Configure environment

```bash
cp .env.example .env
```

Update values in `.env` as needed.

### 2) Start PostgreSQL and API with Docker

```bash
make docker-up
```

### 3) Run migrations

```bash
make migrate
```

### 4) Run API locally (optional)

```bash
make run
```

API base URL: `http://localhost:8080`

Swagger UI (via compose): `http://localhost:8081`
OpenAPI file: `http://localhost:8080/docs/openapi.yaml`

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_ENV` | No | `development` | Runtime environment |
| `APP_PORT` | No | `8080` | API port |
| `LOG_LEVEL` | No | `INFO` | Structured logger level |
| `DATABASE_URL` | Yes | - | PostgreSQL DSN |
| `JWT_SECRET` | Yes | - | JWT signing secret |
| `JWT_TTL_MINUTES` | No | `60` | JWT lifetime |
| `DEFAULT_PAGE_SIZE` | No | `20` | Default pagination size |
| `MAX_PAGE_SIZE` | No | `100` | Max pagination size |

## Authentication

### JWT flow

1. `POST /api/v1/auth/register`
2. `POST /api/v1/auth/login`
3. Use token in header:

```http
Authorization: Bearer <jwt-token>
```

### API key flow

1. Authenticated user creates API client via `POST /api/v1/api-clients`
2. Store returned `api_key` + `api_secret` securely
3. Call API-key protected endpoint using:

```http
X-API-Key: <key>
X-API-Secret: <secret>
```

## Example API Requests

### Register user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com","password":"password123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}'
```

### Create project (JWT)

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Platform API","description":"Core backend services"}'
```

### List tasks with pagination/filter/sort (JWT)

```bash
curl "http://localhost:8080/api/v1/tasks?page=1&limit=20&status=pending&sort_by=created_at&sort_order=desc" \
  -H "Authorization: Bearer $TOKEN"
```

### List tasks with API key auth

```bash
curl "http://localhost:8080/api/v1/client/projects/1/tasks?page=1&limit=10" \
  -H "X-API-Key: $API_KEY" \
  -H "X-API-Secret: $API_SECRET"
```

## Response Format

### Success response

```json
{
  "success": true,
  "data": {
    "id": 1
  }
}
```

### Message response

```json
{
  "success": true,
  "message": "project deleted"
}
```

### Validation error response

```json
{
  "success": false,
  "message": "validation failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      { "field": "Email", "tag": "required" }
    ]
  }
}
```

### Internal error response

```json
{
  "success": false,
  "message": "internal server error",
  "error": {
    "code": "INTERNAL_ERROR"
  }
}
```

### Paginated response

```json
{
  "success": true,
  "data": [],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 101,
    "total_pages": 6
  }
}
```

## Database and Migrations

- Migration files:
  - `migrations/001_init.up.sql`
  - `migrations/001_init.down.sql`
- Includes:
  - foreign keys
  - indexes
  - timestamp fields (`created_at`, `updated_at`)
  - soft deletes (`deleted_at`)

## Middleware Stack

- Request ID
- Structured request logging
- Recover panic handling
- CORS
- Gzip
- Security headers
- JWT auth middleware
- API key auth middleware

## Testing

Run tests:

```bash
make test
```

Current test coverage includes:

- Usecase unit tests (`auth`, `tasks`)
- HTTP handler tests (`auth`, `tasks`)

## Make Commands

- `make run` - run API locally
- `make test` - run all tests
- `make migrate` - apply SQL migrations
- `make migrate-down` - rollback SQL migrations
- `make swagger` - generate Swagger docs if `swag` is installed
- `make docker-up` - build and start containers
- `make docker-down` - stop containers

## Deployment Notes (Linux)

For production deployment on Linux servers:

- run behind Nginx or Traefik reverse proxy
- terminate TLS at proxy or load balancer
- store secrets in environment or secret manager
- use managed PostgreSQL with backups
- enable log shipping and metrics collection
- run migrations as part of release workflow

## Swagger/OpenAPI

- OpenAPI spec: `docs/openapi.yaml`
- Swagger UI via docker-compose: `http://localhost:8081`

(If you want screenshots, capture Swagger UI locally and place assets under `docs/`.)
