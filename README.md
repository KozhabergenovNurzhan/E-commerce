# E-commerce REST API

A production-ready e-commerce backend built with Go, Gin, PostgreSQL, and JWT authentication.

## Tech Stack

- **Go 1.25.4** — language
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

# Force re-run (bypass cache)
go test -count=1 ./...
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
| GET | `/api/v1/categories` | Public | List categories |
| POST | `/api/v1/products` | Admin | Create product |
| PUT | `/api/v1/products/:id` | Admin | Update product |
| DELETE | `/api/v1/products/:id` | Admin | Soft-delete product |

### Users
| Method | Path | Access | Description |
|--------|------|--------|-------------|
| GET | `/api/v1/users` | Admin | List all users |
| GET | `/api/v1/users/:id` | Auth | Get user by ID |
| PUT | `/api/v1/users/:id` | Owner / Admin | Update profile |
| DELETE | `/api/v1/users/:id` | Admin | Soft-delete user |

### Orders
| Method | Path | Access | Description |
|--------|------|--------|-------------|
| POST | `/api/v1/orders` | Auth | Create order |
| GET | `/api/v1/orders` | Auth | List own orders |
| GET | `/api/v1/orders/:id` | Owner / Admin | Get order by ID |
| PATCH | `/api/v1/orders/:id/cancel` | Owner | Cancel pending order |
| PATCH | `/api/v1/orders/:id/status` | Admin / Manager | Update order status |

### Cart
| Method | Path | Access | Description |
|--------|------|--------|-------------|
| POST | `/api/v1/cart` | Auth | Add item to cart |
| GET | `/api/v1/cart` | Auth | Get cart with totals |
| PUT | `/api/v1/cart/:product_id` | Auth | Update item quantity |
| DELETE | `/api/v1/cart/:product_id` | Auth | Remove item from cart |
| DELETE | `/api/v1/cart` | Auth | Clear cart |

## Roles

| Role | Capabilities |
|------|-------------|
| `customer` | Own profile, own orders, cart |
| `manager` | Update order status (confirmed → shipping → delivered) |
| `seller` | Create and manage own products |
| `admin` | Full access |

## Auth Flow

Requests to protected routes require:
```
Authorization: Bearer <access_token>
```

On token expiry, call `POST /api/v1/auth/refresh` with `{"refresh_token": "..."}` to get a new pair. The old refresh token is revoked immediately (rotation).

## Project Structure

```
cmd/api/          # Entry point
config/           # Environment config
internal/
  auth/           # JWT manager (stateless signing/validation)
  domain/         # Entities and DTOs
  handler/        # Gin HTTP handlers and router
  middleware/      # Auth and role middleware
  repository/     # PostgreSQL data access layer
  server/         # HTTP server with graceful shutdown
  service/        # Business logic
  testutil/       # Shared mock repositories for tests
migrations/       # SQL migration files
pkg/
  apperrors/      # Domain error types
  logger/         # slog setup
  response/       # JSON response helpers
  utils/          # Shared utilities (timezone-aware Now())
```
