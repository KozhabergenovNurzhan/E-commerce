# E-commerce REST API

A production-ready e-commerce backend built with Go, Gin, PostgreSQL, Redis, AWS S3, and JWT authentication. Deployed on AWS EC2 with managed RDS database.

## Tech Stack

### Backend
- **Go 1.24** — language
- **Gin** — HTTP framework
- **PostgreSQL 16** — primary database (sqlx + pgx)
- **Redis 7** — caching (products, categories) and idempotency
- **golang-migrate** — SQL migrations (auto-applied on startup)
- **JWT (HS256)** — access tokens (15m) + refresh tokens (7d, rotation)
- **bcrypt** — password hashing
- **slog** — structured logging (text for terminal, JSON for Grafana/Prometheus)

### Cloud Infrastructure (AWS)
- **EC2** — application hosting (Amazon Linux 2023, t3.micro)
- **RDS PostgreSQL** — managed database with automated backups and SSL
- **S3** — object storage for product images
- **IAM** — scoped credentials for S3 access

### DevOps
- **Docker + docker-compose** — containerized development and deployment
- **Multi-stage Dockerfile** — minimal production image

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────────┐         ┌──────────────┐
│  EC2 (Gin API)  │────────▶│   AWS RDS    │
│   - JWT Auth    │         │ (PostgreSQL) │
│   - Validation  │         └──────────────┘
│   - Business    │
│     Logic       │         ┌──────────────┐
│                 │────────▶│    Redis     │
│                 │         │  (cache +    │
│                 │         │ idempotency) │
│                 │         └──────────────┘
│                 │
│                 │         ┌──────────────┐
│                 │────────▶│    AWS S3    │
│                 │         │ (images)     │
└─────────────────┘         └──────────────┘
```

## Getting Started

### Run with Docker (recommended)

```bash
docker-compose up --build
```

App starts on `http://localhost:8080`. PostgreSQL, Redis and migrations run automatically.

### Run locally

```bash
docker-compose up -d postgres redis
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
| `DB_SSLMODE` | `disable` | SSL mode (`require` for RDS) |
| `JWT_SECRET` | `change-me-in-production` | HMAC signing secret |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime (7 days) |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | _(empty)_ | Redis password |
| `AWS_REGION` | `us-east-1` | AWS region for S3 |
| `AWS_ACCESS_KEY_ID` | _(required)_ | IAM user access key |
| `AWS_SECRET_ACCESS_KEY` | _(required)_ | IAM user secret key |
| `S3_BUCKET` | _(required)_ | S3 bucket name for images |

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

### File Upload (S3)

| Method | Path | Access | Description |
|--------|------|--------|-------------|
| POST | `/api/v1/upload/product-image` | Admin, Seller | Upload image to S3, returns public URL |

Constraints: max 5MB, JPEG/PNG/WebP only. Files stored under `products/<uuid>.<ext>` in the configured S3 bucket.

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
| `seller` | Create and manage own products, upload images |
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

Responses are cached in Redis for 24 hours, scoped per user. The middleware uses **fail-closed** strategy — if Redis is unavailable, requests return `503 Service Unavailable` to prevent duplicate operations.

## Order Status Flow

```
pending → confirmed → shipping → delivered
pending → cancelled
```

Only forward transitions are allowed. Cancellation is only possible from `pending`.

## Deployment

### AWS Production Setup

The application is designed to run on AWS with the following services:

1. **EC2 instance** (t3.micro) running Docker Compose
2. **RDS PostgreSQL** (db.t4g.micro) for managed database with SSL
3. **S3 bucket** for product images with public-read access
4. **IAM user** with scoped S3 permissions

Application connects to RDS via SSL (`DB_SSLMODE=require`) and uploads files to S3 using AWS SDK v2.

### Quick Deploy to EC2

```bash
git clone https://github.com/KozhabergenovNurzhan/E-commerce.git
cd E-commerce

nano .env

docker-compose up -d --build

curl http://localhost:8080/health
```

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
  storage/           # AWS S3 client for file uploads
  testutil/          # Shared mock repositories for unit tests
migrations/          # golang-migrate SQL files
```