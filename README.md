# E-commerce REST API

A production-ready e-commerce backend built with Go, Gin, PostgreSQL, and JWT authentication.

## Tech Stack

- **Go** — language
- **Gin** — HTTP framework
- **PostgreSQL 16** — database
- **sqlx + pgx** — database driver and query layer
- **golang-migrate** — SQL migrations (auto-applied on startup)
- **JWT (HS256)** — access tokens (15m) + refresh tokens (7d, rotation)
- **bcrypt** — password hashing
- **slog** — structured JSON logging
- **Docker + docker-compose** — containerized development

## Getting Started

### Run with Docker (recommended)

```bash
docker-compose up --build
```

App starts on `http://localhost:8080`. Migrations run automatically.

### Run locally

```bash
# 1. Start only the database
docker-compose up -d postgres

# 2. Copy env and run
cp .env.example .env
go run ./cmd/api
```

### Test

```bash
go test ./...
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `ecommerce_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | `change-me-in-production` | HMAC signing secret |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime (7 days) |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

## API

### Auth

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| POST | `/api/v1/auth/register` | Public | Register new customer |
| POST | `/api/v1/auth/login` | Public | Login, returns token pair |
| POST | `/api/v1/auth/refresh` | Public | Rotate refresh token |
| POST | `/api/v1/auth/logout` | Public | Revoke refresh token |

### Products

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/products` | Public | List products (filterable) |
| GET | `/api/v1/products/:id` | Public | Get product by ID |
| POST | `/api/v1/products` | Admin, Seller | Create product |
| PUT | `/api/v1/products/:id` | Admin, Seller | Update product |
| DELETE | `/api/v1/products/:id` | Admin, Seller | Soft-delete product |

### Categories

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/categories` | Public | List all categories |
| POST | `/api/v1/categories` | Admin | Create category |
| PUT | `/api/v1/categories/:id` | Admin | Update category |
| DELETE | `/api/v1/categories/:id` | Admin | Delete category |

### Cart

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/cart` | Auth | Get cart with totals |
| POST | `/api/v1/cart/items` | Auth | Add item to cart |
| PUT | `/api/v1/cart/items/:productId` | Auth | Update item quantity |
| DELETE | `/api/v1/cart/items/:productId` | Auth | Remove item from cart |
| DELETE | `/api/v1/cart` | Auth | Clear cart |
| POST | `/api/v1/cart/checkout` | Auth | Convert cart to order |

### Orders

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| POST | `/api/v1/orders` | Auth | Create order directly |
| GET | `/api/v1/orders` | Auth | List own orders (paginated, with items) |
| GET | `/api/v1/orders/:id` | Auth | Get order by ID |
| PATCH | `/api/v1/orders/:id/cancel` | Auth | Cancel pending order |
| PATCH | `/api/v1/orders/:id/status` | Admin, Manager | Update order status |

### Users

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/users` | Admin | List all users |
| GET | `/api/v1/users/:id` | Auth | Get user by ID |
| PUT | `/api/v1/users/:id` | Auth | Update profile |
| DELETE | `/api/v1/users/:id` | Admin | Soft-delete user |

### Seller

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/seller/products` | Seller, Admin | List own products |

## Roles

| Role | Capabilities |
|------|-------------|
| `customer` | Own profile, own orders, cart |
| `manager` | Update order status |
| `seller` | Create and manage own products |
| `admin` | Full access |

## Auth Flow

Protected routes require:
```
Authorization: Bearer <access_token>
```

On expiry, call `POST /api/v1/auth/refresh` with `{"refresh_token": "..."}` to get a new pair. The old token is revoked immediately (rotation).

## Project Structure

```
cmd/api/             # Entry point — wires all dependencies
internal/
  auth/              # JWT manager (sign / validate)
  config/            # Environment variable loading
  handler/           # Gin HTTP handlers and router
  middleware/        # JWT auth and role enforcement
  models/            # Entities and DTOs
  pkg/
    apperrors/       # Domain error types with HTTP status codes
    logger/          # slog structured logging setup
    response/        # JSON response helpers
    utils/           # Shared utilities
  repository/        # PostgreSQL data access (sqlx)
  service/           # Business logic
  testutil/          # Shared mock repositories for unit tests
migrations/          # golang-migrate SQL files
```
