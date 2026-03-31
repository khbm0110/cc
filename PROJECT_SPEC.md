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
        E[PostgreSQL Database]
        F[Redis Streams]
    end

    subgraph "External Services"
        G[Binance API]
        H[NOWPayments API]
        I[Prometheus]
    end

    A -- "User requests via HTTPS" --> B;
    B -- "Publishes trade signal to stream" --> F;
    C -- "Consumes trade signal from stream" --> F;
    C -- "Executes trades via API" --> G;
    D -- "Reconciles order status via API" --> G;
    B -- "Reads/writes user and order data" --> E;
    C -- "Writes order data" --> E;
    D -- "Updates order status" --> E;
    B -- "Creates payment invoice via API" --> H;
    H -- "Sends payment status via Webhook" --> B;
    C -- "Reports execution metrics" --> I;
    D -- "Reports reconciliation metrics" --> I;
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

CREATE TABLE auth (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    jwt_refresh_token TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    amount_usdt DECIMAL(16, 8) NOT NULL,
    nowpayments_id VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    plan_id INT REFERENCES plans(id),
    payment_id INT REFERENCES payments(id),
    status VARCHAR(50) NOT NULL, -- e.g., active, expired, cancelled
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE
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
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE TABLE dead_letter_queue (
    id SERIAL PRIMARY KEY,
    signal_id VARCHAR(255),
    user_id INT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
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

### 1. MockBalanceChecker
In `internal/validator/validator.go`, the `BalanceChecker` is a mock. It needs to be replaced with a real implementation that calls the Binance client.

**Current Code:**
```go
// internal/validator/validator.go

type MockBalanceChecker struct{}

func (m *MockBalanceChecker) CheckBalance(ctx context.Context, userID int, required float64, symbol string) (bool, error) {
    // MOCK: Always returns true
    return true, nil
}
```

**Fixed Code:**
```go
// internal/validator/validator.go

type LiveBalanceChecker struct {
    BinanceClient binance.Client
}

func (l *LiveBalanceChecker) CheckBalance(ctx context.Context, apiKey, apiSecret string, required float64, symbol string) (bool, error) {
    balance, err := l.BinanceClient.GetBalance(ctx, apiKey, apiSecret, symbol)
    if err != nil {
        return false, fmt.Errorf("failed to get balance: %w", err)
    }
    return balance >= required, nil
}
```

### 2. Slippage Validation
In `internal/validator/validator.go`, the validator does not check for slippage. It must fetch the live price from Binance and compare it against the signal's price.

**Current Code:**
```go
// internal/validator/validator.go

func (v *Validator) ValidateSignal(ctx context.Context, signal *models.Signal) error {
    // ... other validations
    if signal.Price <= 0 {
        return errors.New("price must be positive")
    }
    // No slippage check
    return nil
}
```

**Fixed Code:**
```go
// internal/validator/validator.go

const maxSlippage = 0.01 // 1%

func (v *Validator) ValidateSignal(ctx context.Context, signal *models.Signal, binanceClient binance.Client) error {
    // ... other validations
    livePrice, err := binanceClient.GetTickerPrice(ctx, signal.Symbol)
    if err != nil {
        return fmt.Errorf("could not fetch live price: %w", err)
    }

    priceDiff := math.Abs(signal.Price - livePrice)
    slippage := priceDiff / livePrice

    if slippage > maxSlippage {
        return fmt.Errorf("slippage of %.2f%% exceeds max of %.2f%%", slippage*100, maxSlippage*100)
    }

    return nil
}
```

### 3. Silent Order Loss
In `internal/worker/worker.go`, an order execution failure is not handled, leading to silent loss. Failed orders must be retried and eventually sent to a dead-letter queue.

**Current Code:**
```go
// internal/worker/worker.go

func (w *Worker) processSignal(ctx context.Context, signal *models.Signal, user *models.User) {
    // ...
    order, err := w.binanceClient.CreateOrder(ctx, user.ApiKey, user.ApiSecret, signal)
    if err != nil {
        // ERROR IS IGNORED
        log.Printf("Error creating order for user %d: %v", user.ID, err)
    }
    // ...
}
```

**Fixed Code:**
```go
// internal/worker/worker.go

func (w *Worker) processSignal(ctx context.Context, signal *models.Signal, user *models.User) {
    // ...
    var order *models.Order
    var err error

    // Retry logic
    for i := 0; i < maxRetries; i++ {
        order, err = w.binanceClient.CreateOrder(ctx, user.ApiKey, user.ApiSecret, signal)
        if err == nil {
            break // Success
        }
        log.Printf("Attempt %d: Error creating order for user %d: %v", i+1, user.ID, err)
        time.Sleep(retryDelay)
    }

    if err != nil {
        log.Printf("Final attempt failed for user %d. Publishing to dead letter queue.", user.ID)
        if dlqErr := w.dlq.Publish(ctx, signal, user.ID, err); dlqErr != nil {
            log.Printf("CRITICAL: Failed to publish to dead letter queue: %v", dlqErr)
        }
        return
    }
    // ... process successful order
}
```

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

This section details the complete subscription payment flow using NOWPayments, including error handling.

1.  **User Initiates Subscription**: The user selects a subscription plan on the frontend and clicks "Subscribe". The frontend sends a request to the backend: `POST /api/v1/subscriptions/subscribe` with the plan ID.

2.  **Invoice Creation**:
    *   The backend receives the request, validates the user's JWT, and finds the corresponding plan.
    *   It creates a record in the `payments` table with `status: 'pending'`, `user_id`, and the `amount_usdt` from the plan.
    *   It then calls the NOWPayments API to create an invoice, passing the payment amount and a unique `order_id` (e.g., the internal payment ID).
    *   NOWPayments returns a payment URL. The backend updates the `payments` record with the `nowpayments_id` from the response.
    *   The backend returns the `payment_url` to the frontend.

3.  **User Payment**:
    *   The frontend redirects the user to the `payment_url`.
    *   The user completes the payment using a supported cryptocurrency (e.g., USDT).

4.  **Webhook Notification**:
    *   After the payment state changes (e.g., `waiting`, `confirming`, `finished`, `failed`, `expired`), NOWPayments sends an HTTP POST request (webhook) to our backend endpoint: `POST /api/v1/payments/nowpayments/webhook`.
    *   The request body contains the payment details, and the headers contain a signature (`x-nowpayments-sig`).

5.  **Webhook Verification & Processing**:
    *   Our backend receives the webhook. It first verifies the signature by computing an HMAC-SHA512 hash of the request body using the `NOWPAYMENTS_IPN_SECRET` and comparing it to the `x-nowpayments-sig` header.
    *   **If signature is invalid**: The webhook is discarded, and an error is logged. An HTTP 400 Bad Request is returned.
    *   **If signature is valid**: The backend parses the payment status.
        *   `finished`: The payment was successful. The backend finds the corresponding `payments` record, updates its status to `'completed'`, and activates the user's subscription in the `subscriptions` table (sets `status: 'active'`, `start_date`, and `end_date`).
        *   `failed` or `expired`: The payment was not completed. The backend updates the `payments` record status to `'failed'` or `'expired'`. The subscription is not activated.
        *   Other statuses (`waiting`, `confirming`): These are intermediate states. The backend can log them but typically takes no action, waiting for a final status.

### Error Handling Scenarios:

*   **Webhook Endpoint Fails**: If our endpoint returns a non-200 status code or is down, NOWPayments will automatically retry sending the webhook periodically for up to 24 hours. This ensures we eventually receive the payment status.

*   **Payment Expires**: If the user doesn't pay within the allotted time (e.g., 20 minutes), NOWPayments sends a webhook with the `expired` status. Our system will update the corresponding `payments` record to `'expired'`, and the user will have to start the subscription process over.

*   **Duplicate Payments / Double Webhook**:
    *   Our system uses the unique `nowpayments_id` to identify payments. The `payments` table has a `UNIQUE` constraint on this column.
    *   If a webhook for a `finished` payment is received twice, the first one will update the subscription to `'active'`. The second webhook will be processed, but since the subscription is already active, no further state change will occur. The system should be idempotent: processing the same successful webhook multiple times should not result in multiple new subscriptions or extensions.
    *   If a user somehow initiates and pays for a second subscription while the first is still active, this will create a new payment record. When the webhook for the second payment arrives, our business logic must decide how to handle it. A safe approach is to flag it for manual review. A more automated approach could be to extend the user's current subscription `end_date`.
