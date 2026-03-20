package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrDuplicateOrder    = errors.New("duplicate client_order_id for user")
	ErrInvalidTransition = errors.New("invalid order status transition")
)

// Repository defines the interface for order persistence.
type Repository interface {
	// CreateOrder inserts a new order with PENDING status (idempotent).
	CreateOrder(ctx context.Context, params CreateOrderParams) (*Order, error)
	// GetByID retrieves an order by its primary key.
	GetByID(ctx context.Context, id int64) (*Order, error)
	// GetByClientOrderID retrieves an order by user_id + client_order_id (unique).
	GetByClientOrderID(ctx context.Context, userID int64, clientOrderID string) (*Order, error)
	// UpdateStatus atomically transitions an order to a new status.
	UpdateStatus(ctx context.Context, id int64, from, to OrderStatus, binanceOrderID *int64, errMsg *string) error
	// IncrementRetry bumps the retry counter for an order.
	IncrementRetry(ctx context.Context, id int64) error
	// FindStaleOrders returns orders stuck in non-terminal states past the given threshold.
	FindStaleOrders(ctx context.Context, olderThan time.Duration, limit int) ([]Order, error)
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateOrder(ctx context.Context, params CreateOrderParams) (*Order, error) {
	query := `
		INSERT INTO orders (user_id, client_order_id, symbol, side, quantity, price, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, client_order_id) DO NOTHING
		RETURNING id, user_id, client_order_id, symbol, side, quantity, price, status,
		          binance_order_id, error_message, retry_count, created_at, updated_at`

	var o Order
	err := r.db.QueryRowContext(ctx, query,
		params.UserID, params.ClientOrderID, params.Symbol, params.Side,
		params.Quantity, params.Price, StatusPending,
	).Scan(
		&o.ID, &o.UserID, &o.ClientOrderID, &o.Symbol, &o.Side,
		&o.Quantity, &o.Price, &o.Status,
		&o.BinanceOrderID, &o.ErrorMessage, &o.RetryCount,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ON CONFLICT DO NOTHING returns no rows — fetch existing
			existing, fetchErr := r.GetByClientOrderID(ctx, params.UserID, params.ClientOrderID)
			if fetchErr != nil {
				return nil, fmt.Errorf("duplicate order, failed to fetch existing: %w", fetchErr)
			}
			return existing, ErrDuplicateOrder
		}
		return nil, fmt.Errorf("insert order: %w", err)
	}
	return &o, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*Order, error) {
	query := `
		SELECT id, user_id, client_order_id, symbol, side, quantity, price, status,
		       binance_order_id, error_message, retry_count, created_at, updated_at
		FROM orders WHERE id = $1`

	var o Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.ClientOrderID, &o.Symbol, &o.Side,
		&o.Quantity, &o.Price, &o.Status,
		&o.BinanceOrderID, &o.ErrorMessage, &o.RetryCount,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return &o, nil
}

func (r *PostgresRepository) GetByClientOrderID(ctx context.Context, userID int64, clientOrderID string) (*Order, error) {
	query := `
		SELECT id, user_id, client_order_id, symbol, side, quantity, price, status,
		       binance_order_id, error_message, retry_count, created_at, updated_at
		FROM orders WHERE user_id = $1 AND client_order_id = $2`

	var o Order
	err := r.db.QueryRowContext(ctx, query, userID, clientOrderID).Scan(
		&o.ID, &o.UserID, &o.ClientOrderID, &o.Symbol, &o.Side,
		&o.Quantity, &o.Price, &o.Status,
		&o.BinanceOrderID, &o.ErrorMessage, &o.RetryCount,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by client_order_id: %w", err)
	}
	return &o, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id int64, from, to OrderStatus, binanceOrderID *int64, errMsg *string) error {
	if !isValidTransition(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}

	query := `
		UPDATE orders
		SET status = $1, binance_order_id = COALESCE($2, binance_order_id),
		    error_message = COALESCE($3, error_message), updated_at = NOW()
		WHERE id = $4 AND status = $5`

	result, err := r.db.ExecContext(ctx, query, to, binanceOrderID, errMsg, id, from)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: order %d not in expected state %s", ErrInvalidTransition, id, from)
	}
	return nil
}

func (r *PostgresRepository) IncrementRetry(ctx context.Context, id int64) error {
	query := `UPDATE orders SET retry_count = retry_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment retry: %w", err)
	}
	return nil
}

func (r *PostgresRepository) FindStaleOrders(ctx context.Context, olderThan time.Duration, limit int) ([]Order, error) {
	query := `
		SELECT id, user_id, client_order_id, symbol, side, quantity, price, status,
		       binance_order_id, error_message, retry_count, created_at, updated_at
		FROM orders
		WHERE status IN ($1, $2) AND updated_at < $3
		ORDER BY updated_at ASC
		LIMIT $4`

	threshold := time.Now().UTC().Add(-olderThan)
	rows, err := r.db.QueryContext(ctx, query, StatusPending, StatusExecuting, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("find stale orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.ClientOrderID, &o.Symbol, &o.Side,
			&o.Quantity, &o.Price, &o.Status,
			&o.BinanceOrderID, &o.ErrorMessage, &o.RetryCount,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stale order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// isValidTransition checks allowed order status transitions.
func isValidTransition(from, to OrderStatus) bool {
	switch from {
	case StatusPending:
		return to == StatusExecuting || to == StatusCanceled || to == StatusFailed
	case StatusExecuting:
		return to == StatusFilled || to == StatusFailed || to == StatusCanceled
	default:
		return false // terminal states cannot transition
	}
}
