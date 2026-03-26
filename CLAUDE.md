# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Layout

```
├── cmd/api/main.go          # Entry point — wires all dependencies
├── config/config.go         # Environment variable loading (DB + JWT + logging)
├── internal/
│   ├── domain/              # Entities and DTOs (user, product, order, token)
│   ├── repository/          # Data access layer — interfaces + sqlx implementations
│   ├── service/             # Business logic (auth, stock validation, token rotation)
│   └── handler/             # Gin HTTP handlers, JWT middleware, router
├── pkg/
│   ├── apperrors/           # Custom AppError type with HTTP status codes
│   ├── response/            # JSON response helpers (OK, Created, Paginated, Error…)
│   └── logger/              # slog JSON structured logging
└── migrations/              # golang-migrate SQL files (auto-run on startup)
```

## Commands

All commands run from the project root.

```bash
# Install / tidy dependencies (required after first clone or go.mod changes)
go mod tidy

# Run locally (requires PostgreSQL on localhost:5432)
go run ./cmd/api

# Build binary
go build -o bin/ecommerce ./cmd/api

# Test
go test ./...
go test -v -run TestName ./internal/service/...

# Docker — starts app + PostgreSQL, runs migrations automatically
docker-compose up --build
docker-compose down
```

## Configuration

Copy `.env.example` to `.env` for local development. All vars have working defaults except `JWT_SECRET` (change before going to production).

| Variable         | Default                    |
|------------------|----------------------------|
| PORT             | 8080                       |
| DB_HOST          | localhost                  |
| DB_PORT          | 5432                       |
| DB_USER          | postgres                   |
| DB_PASSWORD      | postgres                   |
| DB_NAME          | ecommerce_db               |
| DB_SSLMODE       | disable                    |
| JWT_SECRET       | change-me-in-production    |
| JWT_ACCESS_TTL   | 15m                        |
| JWT_REFRESH_TTL  | 168h (7 days)              |
| LOG_LEVEL        | info                       |

## Architecture

Request flow: `handler → service → repository → PostgreSQL`

Dependency injection is done manually in `cmd/api/main.go` — repos first, then services (tokenSvc takes both tokenRepo + userRepo), then handlers, then router.

**Key patterns:**
- Soft deletes: `is_active` boolean on users and products
- UUID primary keys on all main entities
- Order creation uses a DB transaction to atomically insert order + items
- Stock validated in `order_service.go` before creating an order
- Passwords hashed with bcrypt; never stored plaintext
- `AppError` type maps domain errors to HTTP status codes in `pkg/response`

## Auth Flow

- `POST /api/v1/auth/register` → returns `{user, tokens}`
- `POST /api/v1/auth/login` → returns `{user, tokens}`
- `POST /api/v1/auth/refresh` → `{refresh_token}` → returns new `{access_token, refresh_token, expires_in}`
- `POST /api/v1/auth/logout` → revokes refresh token, returns 204

Access tokens are short-lived JWTs (HS256). Refresh tokens are random 32-byte values stored as SHA-256 hashes in the `refresh_tokens` table. On refresh, the old token is revoked and a new pair is issued (rotation).

Protected routes require `Authorization: Bearer <access_token>`. The `JWTMiddleware` in `internal/handler/middleware.go` sets `userID` and `userRole` in Gin context. Use `mustUserID(c)` inside protected handlers to retrieve the UUID.

## API Routes

**Public:**
- `GET /health`
- `POST /api/v1/auth/register|login|refresh|logout`
- `GET /api/v1/products`, `GET /api/v1/products/:id`
- `GET /api/v1/categories`

**Protected (Bearer token required):**
- `GET/PUT/DELETE /api/v1/users/:id`, `GET /api/v1/users`
- `POST/PUT/DELETE /api/v1/products/:id`
- `POST/GET /api/v1/orders`, `GET /api/v1/orders/:id`, `PATCH /api/v1/orders/:id/cancel`

## Database

- Driver: `pgx` (via `jackc/pgx/v5/stdlib`) for the app connection; `golang-migrate` uses the `postgres://` URL (`lib/pq` as indirect dep)
- Migrations in `migrations/` are applied automatically at startup via `file://migrations` path
- `config.DBConfig.DSN()` → key-value string for sqlx; `MigrateURL()` → `postgres://` URL for golang-migrate
