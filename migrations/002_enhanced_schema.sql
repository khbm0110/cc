-- Migration: 002_enhanced_schema.sql
-- Description: Enhanced schema for centralized copy trading platform
-- Features: Traders, Plans with Virtual Balance, Subscriptions, Signals

BEGIN;

-- ============================================================================
-- Traders table (Professional traders assigned by admin)
-- ============================================================================
CREATE TABLE IF NOT EXISTS traders (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER REFERENCES users(id), -- Link to system user (optional)
    name            VARCHAR(100) NOT NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    avatar_url      VARCHAR(500),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    max_trades_per_day INTEGER NOT NULL DEFAULT 10,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_traders_email ON traders(email);
CREATE INDEX IF NOT EXISTS idx_traders_is_active ON traders(is_active);

-- ============================================================================
-- Enhanced Plans table (Now includes trader assignment and virtual balance)
-- ============================================================================
ALTER TABLE plans 
ADD COLUMN IF NOT EXISTS trader_id INTEGER REFERENCES traders(id),
ADD COLUMN IF NOT EXISTS virtual_balance NUMERIC(20,2) NOT NULL DEFAULT 3000.00,
ADD COLUMN IF NOT EXISTS subscription_price NUMERIC(20,2) NOT NULL DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS min_investment NUMERIC(20,2) NOT NULL DEFAULT 100.00,
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

-- ============================================================================
-- User Subscriptions table (Which plan each user subscribed to)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id                  SERIAL PRIMARY KEY,
    user_id             INTEGER NOT NULL REFERENCES users(id),
    plan_id             INTEGER NOT NULL REFERENCES plans(id),
    subscribed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    monthly_fee_paid    NUMERIC(20,2) NOT NULL DEFAULT 0.00,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_subscription UNIQUE (user_id, plan_id, is_active)
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_plan_id ON user_subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_is_active ON user_subscriptions(is_active);

-- ============================================================================
-- Trade Signals table (Trades executed by traders in virtual mode)
-- ============================================================================
CREATE TABLE IF NOT EXISTS trade_signals (
    id              SERIAL PRIMARY KEY,
    plan_id         INTEGER NOT NULL REFERENCES plans(id),
    trader_id       INTEGER NOT NULL REFERENCES traders(id),
    symbol          VARCHAR(20) NOT NULL,
    side            VARCHAR(10) NOT NULL, -- 'BUY' or 'SELL'
    quantity        NUMERIC(20,8) NOT NULL,
    price           NUMERIC(20,8) NOT NULL,
    total_value     NUMERIC(20,8) NOT NULL,
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_signal_side CHECK (side IN ('BUY', 'SELL')),
    CONSTRAINT chk_signal_status CHECK (status IN ('PENDING', 'EXECUTING', 'COMPLETED', 'FAILED'))
);

CREATE INDEX IF NOT EXISTS idx_trade_signals_plan_id ON trade_signals(plan_id);
CREATE INDEX IF NOT EXISTS idx_trade_signals_trader_id ON trade_signals(trader_id);
CREATE INDEX IF NOT EXISTS idx_trade_signals_executed_at ON trade_signals(executed_at);

-- ============================================================================
-- Virtual Portfolio (Holdings for each trader/plan combination)
-- ============================================================================
CREATE TABLE IF NOT EXISTS virtual_portfolios (
    id              SERIAL PRIMARY KEY,
    plan_id         INTEGER NOT NULL REFERENCES plans(id),
    trader_id       INTEGER NOT NULL REFERENCES traders(id),
    symbol          VARCHAR(20) NOT NULL,
    quantity        NUMERIC(20,8) NOT NULL DEFAULT 0,
    avg_buy_price   NUMERIC(20,8) NOT NULL DEFAULT 0,
    total_invested  NUMERIC(20,2) NOT NULL DEFAULT 0,
    current_value   NUMERIC(20,2) NOT NULL DEFAULT 0,
    profit_loss     NUMERIC(20,2) NOT NULL DEFAULT 0,
    profit_loss_pct NUMERIC(10,4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_plan_trader_symbol UNIQUE (plan_id, trader_id, symbol)
);

CREATE INDEX IF NOT EXISTS idx_virtual_portfolios_plan_id ON virtual_portfolios(plan_id);
CREATE INDEX IF NOT EXISTS idx_virtual_portfolios_trader_id ON virtual_portfolios(trader_id);

-- ============================================================================
-- Virtual Orders (Orders for virtual trading simulation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS virtual_orders (
    id              SERIAL PRIMARY KEY,
    plan_id         INTEGER NOT NULL REFERENCES plans(id),
    trader_id       INTEGER NOT NULL REFERENCES traders(id),
    client_order_id VARCHAR(255) NOT NULL,
    symbol          VARCHAR(20) NOT NULL,
    side            VARCHAR(10) NOT NULL,
    quantity        NUMERIC(20,8) NOT NULL,
    price           NUMERIC(20,8) NOT NULL,
    total_value     NUMERIC(20,2) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_virtual_order UNIQUE (plan_id, trader_id, client_order_id),
    CONSTRAINT chk_vorder_side CHECK (side IN ('BUY', 'SELL')),
    CONSTRAINT chk_vorder_status CHECK (status IN ('PENDING', 'EXECUTING', 'FILLED', 'FAILED', 'CANCELED'))
);

CREATE INDEX IF NOT EXISTS idx_virtual_orders_plan_id ON virtual_orders(plan_id);
CREATE INDEX IF NOT EXISTS idx_virtual_orders_trader_id ON virtual_orders(trader_id);

-- ============================================================================
-- User Orders History (Real orders executed for users)
-- ============================================================================
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS signal_id INTEGER REFERENCES trade_signals(id),
ADD COLUMN IF NOT EXISTS source_plan_id INTEGER REFERENCES plans(id);

CREATE INDEX IF NOT EXISTS idx_orders_signal_id ON orders(signal_id);
CREATE INDEX IF NOT EXISTS idx_orders_source_plan_id ON orders(source_plan_id);

-- ============================================================================
-- Seed sample traders (for testing)
-- ============================================================================
INSERT INTO traders (name, email, is_active, max_trades_per_day)
VALUES 
    ('Ahmed Trader', 'ahmed@copytrader.com', true, 15),
    ('Sara Trading', 'sara@copytrader.com', true, 10),
    ('Crypto Master', 'master@copytrader.com', true, 20)
ON CONFLICT (email) DO NOTHING;

-- ============================================================================
-- Update existing plans with new structure
-- ============================================================================
UPDATE plans SET 
    virtual_balance = 3000.00,
    subscription_price = 0.00,
    min_investment = 1000.00,
    description = 'Plan with $3000 virtual balance for simulation'
WHERE name = 'basic';

UPDATE plans SET 
    virtual_balance = 5000.00,
    subscription_price = 49.00,
    min_investment = 3000.00,
    description = 'Pro plan with $5000 virtual balance'
WHERE name = 'pro';

UPDATE plans SET 
    virtual_balance = 10000.00,
    subscription_price = 199.00,
    min_investment = 10000.00,
    description = 'Enterprise plan with $10000 virtual balance'
WHERE name = 'enterprise';

COMMIT;
