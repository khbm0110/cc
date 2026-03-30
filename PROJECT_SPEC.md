# PROJECT_SPEC.md

## 1. Executive Summary

This document outlines the specifications for a copy trading platform designed for Arabic-speaking beginner traders. The platform enables users to connect their Binance accounts and automatically mirror the trades of our professional trading team, without losing custody of their funds. The business model is a subscription-based service with a profit-sharing component, facilitated through cryptocurrency payments.

## 2. Full Architecture Diagram

```mermaid
graph TD
    subgraph "User Interface (Vercel)"
        A[Next.js Frontend]
    end

    subgraph "Backend Services (Railway)"
        B[Go API Server]
        C[Worker]
        D[Reconciler]
    end

    subgraph "Data Stores"
        E[PostgreSQL]
        F[Redis Streams]
    end

    subgraph "External Services"
        G[Binance API]
        H[NOWPayments API]
        I[Prometheus]
    end

    A -- "REST API (HTTPS)" --> B
    B -- "Publish Signal" --> F
    C -- "Consume Signal" --> F
    C -- "Execute Trade" --> G
    D -- "Reconcile Orders" --> G
    B -- "User/Order Data" --> E
    C -- "Order Data" --> E
    D -- "Order Data" --> E
    B -- "Create Invoice" --> H
    H -- "Webhook" --> B
    C -- "Report Metrics" --> I
    D -- "Report Metrics" --> I
```

## 3. Complete File Structure

```
.
├── .env.example
├── .gitignore
├── docker-compose.yml
├── go.mod
├── go.sum
├── main.go
├── PROJECT_SPEC.md
├── README.md
│
├── cmd
│   ├── api
│   │   └── main.go
│   ├── reconciler
│   │   └── main.go
│   └── worker
│       └── main.go
│
├── internal
│   ├── api
│   │   ├── handlers
│   │   │   ├── auth.go
│   │   │   ├── orders.go
│   │   │   ├── payments.go
│   │   │   ├── signals.go
│   │   │   ├── subscriptions.go
│   │   │   └── users.go
│   │   ├── middleware
│   │   │   ├── auth.go
│   │   │   └── ratelimit.go
│   │   └── router.go
│   ├── auth
│   │   ├── jwt.go
│   │   └── password.go
│   ├── binance
│   │   ├── client.go
│   │   └── models.go
│   ├── config
│   │   └── config.go
│   ├── eventbus
│   │   └── eventbus.go
│   ├── kms
│   │   └── kms.go
│   ├── metrics
│   │   └── metrics.go
│   ├── order
│   │   ├── models.go
│   │   └── repository.go
│   ├── payment
│   │   ├── nowpayments.go
│   │   └── repository.go
│   ├── reconciler
│   │   └── reconciler.go
│   ├── subscription
│   │   ├── models.go
│   │   └── repository.go
│   ├── user
│   │   ├── models.go
│   │   └── repository.go
│   ├── validator
│   │   └── validator.go
│   └── worker
│       ├── circuitbreaker.go
│       ├── ratelimiter.go
│       └── worker.go
│
├── migrations
│   ├── 001_init.sql
│   ├── 002_enhanced_schema.sql
│   └── 003_add_missing_tables.sql
│
└── frontend
    ├── .env.local
    ├── .eslintrc.json
    ├── .gitignore
    ├── next.config.js
    ├── package.json
    ├── README.md
    ├── tsconfig.json
    │
    ├── public
    │   ├── favicon.ico
    │   └── vercel.svg
    │
    └── src
        ├── app
        │   ├── (auth)
        │   │   ├── login
        │   │   │   └── page.tsx
        │   │   └── register
        │   │       └── page.tsx
        │   ├── (dashboard)
        │   │   ├── dashboard
        │   │   │   └── page.tsx
        │   │   ├── orders
        │   │   │   └── page.tsx
        │   │   └── settings
        │   │       └── page.tsx
        │   ├── api
        │   │   └── auth
        │   │       └── [...nextauth]
        │   │           └── route.ts
        │   ├── globals.css
        │   └── layout.tsx
        │
        ├── components
        │   ├── ui
        │   │   ├── button.tsx
        │   │   ├── card.tsx
        │   │   └── input.tsx
        │   └── layout
        │       ├── Navbar.tsx
        │       └── Sidebar.tsx
        │
        ├── hooks
        │   └── useApi.ts
        │
        ├── lib
        │   ├── api.ts
        │   └── auth.ts
        │
        └── types
            └── index.ts
```

## 4. API Endpoints

### Auth
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login a user, returns JWT
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - Logout a user

### Users
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update user profile
- `POST /api/v1/users/me/binance` - Connect Binance API keys
- `GET /api/v1/users/me/balance` - Get Binance account balance

### Subscriptions
- `GET /api/v1/subscriptions` - List available subscription plans
- `GET /api/v1/subscriptions/me` - Get current user's subscription
- `POST /api/v1/subscriptions/subscribe` - Create a subscription invoice

### Payments
- `POST /api/v1/payments/nowpayments/webhook` - Webhook for NOWPayments

### Orders
- `GET /api/v1/orders` - Get user's trade history

### Signals (Trader only)
- `POST /api/v1/signals` - Publish a new trade signal

## 5. Database Schema

```sql
-- Existing tables
CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    max_exposure_ratio FLOAT NOT NULL,
    order_limit_per_min INT NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    plan_id INT REFERENCES plans(id),
    api_key_encrypted TEXT,
    secret_key_encrypted TEXT
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    client_order_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL,
    quantity FLOAT NOT NULL,
    price FLOAT,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Missing tables
CREATE TABLE auth (
    user_id INT PRIMARY KEY REFERENCES users(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    jwt_refresh_token TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE REFERENCES users(id),
    plan_id INT REFERENCES plans(id),
    status VARCHAR(50) NOT NULL, -- e.g., active, expired, cancelled
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    payment_id INT
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    amount_usdt DECIMAL(16, 8) NOT NULL,
    nowpayments_id VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL, -- e.g., pending, completed, failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE dead_letter_queue (
    id SERIAL PRIMARY KEY,
    signal_id VARCHAR(255),
    user_id INT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE subscriptions ADD CONSTRAINT fk_payment_id FOREIGN KEY (payment_id) REFERENCES payments(id);
```

## 6. Environment Variables

```
# .env.example

# Go Backend
GO_PORT=8080
DATABASE_URL="postgresql://user:password@host:port/dbname"
REDIS_URL="redis://host:port"
JWT_SECRET="your-jwt-secret"
KMS_ENCRYPTION_KEY="your-32-byte-aes-key"

# Binance
BINANCE_API_URL="https://api.binance.com"

# NOWPayments
NOWPAYMENTS_API_KEY="your-nowpayments-api-key"
NOWPAYMENTS_IPN_SECRET="your-nowpayments-ipn-secret"

# Frontend
NEXT_PUBLIC_API_URL="http://localhost:8080"
```

## 7. Critical Bug Fixes

1.  **MockBalanceChecker**: In `internal/validator/validator.go`, replace `MockBalanceChecker` with a real implementation that calls the `GetBalance` method on the Binance client.
2.  **Slippage Validation**: In `internal/validator/validator.go`, fetch the live price from Binance for the given symbol and compare it against `signal.Price`.
3.  **Silent Order Loss**: In `internal/worker/worker.go`, wrap order execution in a `try/catch`-like block. On error, if retries are exhausted, publish the failed signal to the `dead_letter_queue` in Redis.
4.  **MARKET Order Support**: Modify `internal/order/models.go` to include `MARKET` as an order type. Update the worker to handle signals with this type, omitting the price from the Binance API call.
5.  **GetBalance Endpoint**: Add `GetBalance(ctx context.Context, symbol string) (float64, error)` to the `BinanceClient` interface in `internal/binance/client.go` and implement it.
6.  **Binance Client Caching**: Implement a caching mechanism (e.g., a `map[userID]*binance.Client`) in the worker to store and reuse Binance clients per user, avoiding re-instantiation on every request.

## 8. Build Order

1.  Implement user authentication (registration, login, JWT).
2.  Implement the ability for users to connect their Binance API keys.
3.  Fix critical bugs, starting with the `MockBalanceChecker`.
4.  Implement real balance fetching from Binance.
5.  Implement live price fetching for slippage validation.
6.  Set up the dead letter queue for failed orders.
7.  Integrate NOWPayments for subscription management.
8.  Build the user dashboard (positions, P/L).
9.  Build the trader dashboard for signal publishing.
10. Connect the frontend to the backend using the `NEXT_PUBLIC_API_URL`.

## 9. Integration Guide

The Vercel-hosted Next.js frontend will communicate with the Railway-hosted Go backend via the `NEXT_PUBLIC_API_URL` environment variable. This variable will be set in Vercel's project settings to point to the public URL of the Railway app (e.g., `https://my-app.up.railway.app`). All API requests from the frontend will be directed to this URL.

## 10. NOWPayments Flow

1.  **User Initiates Subscription**: User selects a plan and clicks "Subscribe".
2.  **Invoice Creation**: Frontend sends a request to `POST /api/v1/subscriptions/subscribe`. The backend creates an invoice with NOWPayments, stores the `payment_id` in the `payments` table with `status: 'pending'`, and returns the payment URL to the frontend.
3.  **User Pays**: User is redirected to the NOWPayments URL to complete the payment in USDT.
4.  **Webhook Notification**: NOWPayments sends a webhook to `POST /api/v1/payments/nowpayments/webhook` with the payment status.
5.  **Webhook Verification**: The backend verifies the webhook signature using the `NOWPAYMENTS_IPN_SECRET`.
6.  **Subscription Activation**: If the payment is successful, the backend updates the `payments` table to `status: 'completed'` and the `subscriptions` table to `status: 'active'`, setting the `start_date` and `end_date`.
