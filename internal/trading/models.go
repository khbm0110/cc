package trading

import "time"

// TradeSignal represents a trade signal from a trader.
type TradeSignal struct {
	ID          int64     `db:"id" json:"id"`
	PlanID      int64     `db:"plan_id" json:"plan_id"`
	TraderID    int64     `db:"trader_id" json:"trader_id"`
	Symbol      string    `db:"symbol" json:"symbol"`
	Side        string    `db:"side" json:"side"` // "BUY" or "SELL"
	Quantity    float64   `db:"quantity" json:"quantity"`
	Price       float64   `db:"price" json:"price"`
	TotalValue  float64   `db:"total_value" json:"total_value"`
	ExecutedAt  time.Time `db:"executed_at" json:"executed_at"`
	Status      string    `db:"status" json:"status"`
	ErrorMessage *string  `db:"error_message" json:"error_message,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// VirtualOrder represents an order in the virtual trading simulation.
type VirtualOrder struct {
	ID             int64     `db:"id" json:"id"`
	PlanID         int64     `db:"plan_id" json:"plan_id"`
	TraderID       int64     `db:"trader_id" json:"trader_id"`
	ClientOrderID  string    `db:"client_order_id" json:"client_order_id"`
	Symbol         string    `db:"symbol" json:"symbol"`
	Side           string    `db:"side" json:"side"`
	Quantity       float64   `db:"quantity" json:"quantity"`
	Price          float64   `db:"price" json:"price"`
	TotalValue     float64   `db:"total_value" json:"total_value"`
	Status         string    `db:"status" json:"status"`
	ExecutedAt     time.Time `db:"executed_at" json:"executed_at"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// VirtualPortfolio represents holdings for a trader/plan combination.
type VirtualPortfolio struct {
	ID              int64     `db:"id" json:"id"`
	PlanID          int64     `db:"plan_id" json:"plan_id"`
	TraderID        int64     `db:"trader_id" json:"trader_id"`
	Symbol          string    `db:"symbol" json:"symbol"`
	Quantity        float64   `db:"quantity" json:"quantity"`
	AvgBuyPrice     float64   `db:"avg_buy_price" json:"avg_buy_price"`
	TotalInvested   float64   `db:"total_invested" json:"total_invested"`
	CurrentValue    float64   `db:"current_value" json:"current_value"`
	ProfitLoss      float64   `db:"profit_loss" json:"profit_loss"`
	ProfitLossPct   float64   `db:"profit_loss_pct" json:"profit_loss_pct"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// VirtualPortfolioSummary represents the overall virtual portfolio for a plan.
type VirtualPortfolioSummary struct {
	TotalValue     float64            `json:"total_value"`
	TotalInvested  float64            `json:"total_invested"`
	TotalProfitLoss float64           `json:"total_profit_loss"`
	ProfitLossPct  float64             `json:"profit_loss_pct"`
	Holdings       []VirtualPortfolio `json:"holdings"`
	CashBalance    float64             `json:"cash_balance"`
}

// CreateSignalParams holds parameters for creating a new trade signal.
type CreateSignalParams struct {
	PlanID     int64   `json:"plan_id"`
	TraderID   int64   `json:"trader_id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"` // "BUY" or "SELL"
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
}

// ExecuteSignalParams holds parameters for executing a signal for users.
type ExecuteSignalParams struct {
	SignalID   int64   `json:"signal_id"`
	UserID     int64   `json:"user_id"`
	PlanID     int64   `json:"plan_id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
}
