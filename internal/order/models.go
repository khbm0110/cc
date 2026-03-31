package order

import "time"

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusExecuting OrderStatus = "EXECUTING"
	StatusFilled    OrderStatus = "FILLED"
	StatusFailed    OrderStatus = "FAILED"
	StatusCanceled  OrderStatus = "CANCELED"
)

// IsTerminal returns true if the order is in a terminal state.
func (s OrderStatus) IsTerminal() bool {
	return s == StatusFilled || s == StatusFailed || s == StatusCanceled
}

// IsValid returns true if the status is a recognized value.
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusExecuting, StatusFilled, StatusFailed, StatusCanceled:
		return true
	}
	return false
}

// Order represents a copy-trade order in the system.
type Order struct {
	ID             int64       `db:"id" json:"id"`
	UserID         int64       `db:"user_id" json:"user_id"`
	ClientOrderID  string      `db:"client_order_id" json:"client_order_id"`
	Symbol         string      `db:"symbol" json:"symbol"`
	Side           string      `db:"side" json:"side"`
	Quantity       float64     `db:"quantity" json:"quantity"`
	Price          float64     `db:"price" json:"price"`
	Status         OrderStatus `db:"status" json:"status"`
	BinanceOrderID *int64      `db:"binance_order_id" json:"binance_order_id,omitempty"`
	ErrorMessage   *string     `db:"error_message" json:"error_message,omitempty"`
	RetryCount     int         `db:"retry_count" json:"retry_count"`
	SignalID       *int64      `db:"signal_id" json:"signal_id,omitempty"`
	SourcePlanID   *int64      `db:"source_plan_id" json:"source_plan_id,omitempty"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updated_at"`
}

// CreateOrderParams holds parameters for creating a new order.
type CreateOrderParams struct {
	UserID        int64   `json:"user_id"`
	ClientOrderID string  `json:"client_order_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
}

// CreateOrderFromSignalParams holds parameters for creating an order from a signal.
type CreateOrderFromSignalParams struct {
	UserID        int64   `json:"user_id"`
	SignalID      int64   `json:"signal_id"`
	SourcePlanID  int64   `json:"source_plan_id"`
	ClientOrderID string  `json:"client_order_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
}
