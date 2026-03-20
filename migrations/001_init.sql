-- Migration: 001_init.sql
-- Description: Initial schema for Copy Trading Platform

BEGIN;

-- Plans table
CREATE TABLE IF NOT EXISTS plans (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    max_exposure_ratio  NUMERIC(5,4) NOT NULL DEFAULT 0.1000,
    order_limit_per_min INTEGER NOT NULL DEFAULT 10,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id                  SERIAL PRIMARY KEY,
    name                VARCHAR(255) NOT NULL,
    plan_id             INTEGER NOT NULL REFERENCES plans(id),
    api_key_encrypted   BYTEA NOT NULL,
    secret_key_encrypted BYTEA NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_plan_id ON users(plan_id);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES users(id),
    client_order_id VARCHAR(255) NOT NULL,
    symbol          VARCHAR(20) NOT NULL,
    side            VARCHAR(10) NOT NULL,
    quantity        NUMERIC(20,8) NOT NULL,
    price           NUMERIC(20,8) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    binance_order_id BIGINT,
    error_message   TEXT,
    retry_count     INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_client_order UNIQUE (user_id, client_order_id),
    CONSTRAINT chk_order_status CHECK (status IN ('PENDING', 'EXECUTING', 'FILLED', 'FAILED', 'CANCELED'))
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status_updated ON orders(status, updated_at);
CREATE INDEX IF NOT EXISTS idx_orders_client_order_id ON orders(client_order_id);

-- Seed default plans
INSERT INTO plans (name, max_exposure_ratio, order_limit_per_min)
VALUES
    ('basic', 0.0500, 5),
    ('pro', 0.1000, 20),
    ('enterprise', 0.2000, 60)
ON CONFLICT (name) DO NOTHING;

COMMIT;
