package order

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
)

// DLQRepository defines the interface for the Dead Letter Queue.
type DLQRepository interface {
	Publish(ctx context.Context, signal eventbus.TradeSignal, userID int64, errStr string) error
}

// PostgresDLQRepository implements DLQRepository using PostgreSQL.
type PostgresDLQRepository struct {
	db *sql.DB
}

// NewPostgresDLQRepository creates a new PostgresDLQRepository.
func NewPostgresDLQRepository(db *sql.DB) *PostgresDLQRepository {
	return &PostgresDLQRepository{db: db}
}

// Publish inserts a failed signal into the dead_letter_queue table.
func (r *PostgresDLQRepository) Publish(ctx context.Context, signal eventbus.TradeSignal, userID int64, errStr string) error {
	query := `
		INSERT INTO dead_letter_queue (signal_id, user_id, error)
		VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, signal.SignalID, userID, errStr)
	if err != nil {
		return fmt.Errorf("failed to publish to DLQ: %w", err)
	}
	return nil
}