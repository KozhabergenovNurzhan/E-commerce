# E-commerce REST API

A production-ready e-commerce backend built with Go, Gin, PostgreSQL, Redis, and JWT authentication.

## Tech Stack

- **Go** — language
- **Gin** — HTTP framework
- **PostgreSQL 16** — primary database (sqlx + pgx)
- **Redis 7** — caching (products, categories) and idempotency
- **golang-migrate** — SQL migrations (auto-applied on startup)
- **JWT (HS256)** — access tokens (15m) + refresh tokens (7d, rotation)
- **bcrypt** — password hashing
- **slog** — structured logging (text for terminal, JSON for Grafana/Prometheus)
- **Docker + docker-compose** — containerized development

## Getting Started

### Run with Docker (recommended)

```bash
docker-compose up --build
```

App starts on `http://localhost:8080`. PostgreSQL, Redis and migrations run automatically.

### Run locally

```bash
# 1. Start database and cache
docker-compose up -d postgres redis

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
| `GIN_MODE` | `debug` | Gin mode (`debug` / `release`) |
| `LOG_LEVEL` | `info` | Log level (`debug` / `info` / `warn` / `error`) |
| `LOG_FORMAT` | `text` | Log format (`text` for terminal, `json` for Grafana) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `ecommerce_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | `change-me-in-production` | HMAC signing secret |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime (7 days) |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | _(empty)_ | Redis password |

## API

Full Postman guide: [`docs/postman-guide.md`](docs/postman-guide.md)

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
| GET | `/api/v1/products` | Public | List products (search, filter by category, price) |
| GET | `/api/v1/products/:id` | Public | Get product by ID |
| POST | `/api/v1/products` | Admin, Seller | Create product |
| PUT | `/api/v1/products/:id` | Admin, Seller | Update product |
| DELETE | `/api/v1/products/:id` | Admin, Seller | Soft-delete product |

### Categories

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/categories` | Public | List all categories |
| POST | `/api/v1/categories` | Admin | Create category |
| PUT | `/api/v1/categories/:id` | Admin, Manager | Update category |
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
| GET | `/api/v1/orders` | Auth | List own orders (paginated) |
| GET | `/api/v1/orders/:id` | Auth | Get order by ID |
| PATCH | `/api/v1/orders/:id/cancel` | Auth | Cancel pending order |
| PATCH | `/api/v1/orders/:id/status` | Admin, Manager | Update order status |

### Users

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/users` | Admin | List all users |
| GET | `/api/v1/users/:id` | Auth | Get user by ID |
| PUT | `/api/v1/users/:id` | Auth | Update own profile |
| DELETE | `/api/v1/users/:id` | Admin | Soft-delete user |

### Seller

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/seller/products` | Seller, Admin | List own products |

## Roles

| Role | Capabilities |
|------|-------------|
| `customer` | Own profile, cart, own orders |
| `manager` | Update order status, update categories |
| `seller` | Create and manage own products |
| `admin` | Full access |

## Auth Flow

Protected routes require:
```
Authorization: Bearer <access_token>
```

On expiry, call `POST /api/v1/auth/refresh` with `{"refresh_token": "..."}` to get a new pair. The old token is revoked immediately (rotation).

## Idempotency

`POST /api/v1/orders` and `POST /api/v1/cart/checkout` support idempotent requests via the `Idempotency-Key` header. Duplicate requests with the same key return the cached response without re-processing.

```
Idempotency-Key: <any-unique-string>
```

Responses are cached in Redis for 24 hours, scoped per user.

## Order Status Flow

```
pending → confirmed → shipping → delivered
pending → cancelled
```

Only forward transitions are allowed. Cancellation is only possible from `pending`.

## Project Structure

```
cmd/api/             # Entry point — wires all dependencies
docs/                # Postman guide and API documentation
internal/
  auth/              # JWT manager (sign / validate)
  cache/             # Redis: product cache + idempotency store
  config/            # Environment variable loading
  handler/           # Gin HTTP handlers and router
  middleware/        # JWT auth, role enforcement, idempotency
  models/            # Entities and DTOs
  pkg/
    apperrors/       # Domain error types with HTTP status codes
    logger/          # slog setup (text / json)
    response/        # JSON response helpers
    utils/           # Shared utilities
  repository/        # PostgreSQL data access (sqlx)
  service/           # Business logic
  testutil/          # Shared mock repositories for unit tests
migrations/          # golang-migrate SQL files
```
