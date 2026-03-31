# Copy Trading Platform

A production-ready Copy Trading Platform for Binance built in Go. Allows beginner users to automatically copy trades from professional traders.

## Architecture

- **Clean Architecture** with clear module separation
- **Event-Driven Design** using Redis Streams
- **Defensive Programming** with circuit breakers, rate limiting, and retry with exponential backoff
- **Multi-tenant**: Each user has their own API keys, limits, and subscriptions

### Components

| Component | Description |
|-----------|-------------|
| **Worker** | Listens for trade signals, validates, rate-limits, and executes orders on Binance |
| **Reconciler** | Periodically scans stale orders and reconciles with Binance |
| **Event Bus** | Redis Streams-based pub/sub with consumer groups |
| **Validator** | Validates slippage, stop-loss/take-profit, balance, and exposure |
| **Binance Client** | Per-user API client with HMAC-SHA256 signatures |
| **KMS** | Key management for encrypting/decrypting API keys at rest |
| **Metrics** | Prometheus metrics for orders, rate limits, circuit breakers |

## Project Structure

```
.
├── cmd/
│   ├── worker/main.go          # Worker service entry point
│   └── reconciler/main.go      # Reconciler service entry point
├── internal/
│   ├── order/                   # Order models, DB repository, idempotency
│   ├── user/                    # User/Plan models and repository
│   ├── binance/                 # Binance API client (per-user)
│   ├── validator/               # Trade signal validation
│   ├── eventbus/                # Redis Streams event bus
│   ├── worker/                  # Worker logic, rate limiter, circuit breaker
│   ├── reconciler/              # Stale order reconciliation
│   ├── kms/                     # Key management (mock + interface)
│   └── metrics/                 # Prometheus metrics
├── migrations/
│   └── 001_init.sql             # Database schema
├── docker-compose.yml           # Local development stack
└── go.mod
```

## Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Redis 7+
- Docker & Docker Compose (optional, for local development)

## Quick Start

### 1. Start Infrastructure (Docker Compose)

```bash
docker-compose up -d
```

This starts PostgreSQL and Redis locally.

### 2. Run Database Migrations

```bash
psql -h localhost -U copytrading -d copytrading -f migrations/001_init.sql
```

Or with Docker:

```bash
docker exec -i copy-trading-postgres psql -U copytrading -d copytrading < migrations/001_init.sql
```

### 3. Build Services

```bash
go build -o bin/worker ./cmd/worker
go build -o bin/reconciler ./cmd/reconciler
```

### 4. Run Worker

```bash
./bin/worker
```

### 5. Run Reconciler

```bash
./bin/reconciler
```

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://copytrading:copytrading@localhost:5432/copytrading?sslmode=disable` | PostgreSQL connection string |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | `` | Redis password |
| `KMS_PASSPHRASE` | `dev-passphrase-change-me!!` | Encryption passphrase (use real KMS in production) |
| `BINANCE_BASE_URL` | `https://api.binance.com` | Binance API base URL |
| `WORKER_NAME` | `worker-1` | Unique worker consumer name |
| `METRICS_ADDR` | `:9090` (worker) / `:9091` (reconciler) | Prometheus metrics endpoint |

## Database Schema

### Plans
- `id`, `name`, `max_exposure_ratio`, `order_limit_per_min`
- Seeded with `basic`, `pro`, `enterprise` plans

### Users
- `id`, `name`, `plan_id` (FK), `api_key_encrypted`, `secret_key_encrypted`
- API keys encrypted at rest using AES-256-GCM

### Orders
- `id`, `user_id` (FK), `client_order_id` (unique per user), `symbol`, `side`, `quantity`, `price`, `status`
- Status lifecycle: `PENDING → EXECUTING → FILLED/FAILED/CANCELED`
- Indexed on `user_id`, `status+updated_at`, `client_order_id`
- Idempotent creation via `(user_id, client_order_id)` unique constraint

## Key Features

### Order Processing Pipeline
1. Signal received from Redis Stream
2. User and plan fetched from DB
3. Per-user rate limiting checked
4. Per-user circuit breaker checked
5. Signal validated (slippage, SL/TP, balance, exposure)
6. Order created in DB (idempotent)
7. Status: PENDING → EXECUTING
8. API keys decrypted via KMS
9. Order executed on Binance with retry + exponential backoff
10. Status: EXECUTING → FILLED/FAILED

### Reliability Mechanisms
- **Idempotency**: Unique constraint on `(user_id, client_order_id)` prevents duplicate orders
- **Atomic Transitions**: Status updates use optimistic locking (`WHERE status = $from`)
- **Rate Limiting**: Per-user sliding window rate limiter
- **Circuit Breaker**: Per-user circuit breaker (CLOSED → OPEN → HALF_OPEN)
- **Retry with Backoff**: Exponential backoff for retriable Binance API errors
- **Reconciler**: Catches orphaned/stuck orders and syncs with Binance

### Security
- API keys encrypted at rest using AES-256-GCM
- Mock KMS for development, swap with AWS/GCP KMS in production
- No global API key singleton — per-user Binance clients

### Monitoring
- Prometheus metrics at `/metrics` endpoint
- Structured JSON logging with `slog`
- Full traceability: `user_id`, `order_id`, `client_order_id`, `symbol` in every log

## Running Tests

```bash
go test ./... -v
```

## Production Considerations

1. **Replace MockKMS** with AWS KMS, GCP KMS, or HashiCorp Vault
2. **Replace MockBalanceChecker** with real Binance balance queries
3. **Add TLS** for all external connections
4. **Scale workers** horizontally (Redis consumer groups handle partitioning)
5. **Add alerting** on circuit breaker state changes and error rates
6. **Set up Grafana dashboards** from Prometheus metrics
