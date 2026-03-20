package reconciler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/khbm0110/copy-trading-platform/internal/binance"
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/metrics"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

// Config holds reconciler configuration.
type Config struct {
	ScanInterval   time.Duration
	StaleThreshold time.Duration
	BatchSize      int
	BinanceBaseURL string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		ScanInterval:   30 * time.Second,
		StaleThreshold: 5 * time.Minute,
		BatchSize:      50,
	}
}

// Reconciler periodically scans for stale orders and reconciles with Binance.
type Reconciler struct {
	orderRepo order.Repository
	userRepo  user.Repository
	kms       kms.KeyManager
	config    Config
	logger    *slog.Logger
}

// New creates a new Reconciler.
func New(
	orderRepo order.Repository,
	userRepo user.Repository,
	keyManager kms.KeyManager,
	cfg Config,
	logger *slog.Logger,
) *Reconciler {
	return &Reconciler{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		kms:       keyManager,
		config:    cfg,
		logger:    logger.With(slog.String("component", "reconciler")),
	}
}

// Start begins the reconciliation loop. Blocks until context is canceled.
func (r *Reconciler) Start(ctx context.Context) error {
	r.logger.Info("reconciler started",
		slog.Duration("scan_interval", r.config.ScanInterval),
		slog.Duration("stale_threshold", r.config.StaleThreshold),
	)

	ticker := time.NewTicker(r.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("reconciler shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := r.reconcile(ctx); err != nil {
				r.logger.Error("reconciliation cycle failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context) error {
	staleOrders, err := r.orderRepo.FindStaleOrders(ctx, r.config.StaleThreshold, r.config.BatchSize)
	if err != nil {
		return fmt.Errorf("find stale orders: %w", err)
	}

	if len(staleOrders) == 0 {
		return nil
	}

	r.logger.Info("found stale orders", slog.Int("count", len(staleOrders)))

	for _, ord := range staleOrders {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := r.reconcileOrder(ctx, ord); err != nil {
			r.logger.Error("failed to reconcile order",
				slog.Int64("order_id", ord.ID),
				slog.String("error", err.Error()),
			)
			continue
		}
		metrics.ReconcilerOrdersProcessed.Inc()
	}

	return nil
}

func (r *Reconciler) reconcileOrder(ctx context.Context, ord order.Order) error {
	logger := r.logger.With(
		slog.Int64("order_id", ord.ID),
		slog.Int64("user_id", ord.UserID),
		slog.String("symbol", ord.Symbol),
		slog.String("client_order_id", ord.ClientOrderID),
		slog.String("current_status", string(ord.Status)),
	)

	logger.Info("reconciling stale order")

	// For PENDING orders that never reached Binance, mark as FAILED
	if ord.Status == order.StatusPending {
		logger.Info("stale PENDING order, marking as FAILED")
		errMsg := "order stuck in PENDING state, marked as FAILED by reconciler"
		return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusPending, order.StatusFailed, nil, &errMsg)
	}

	// For EXECUTING orders, query Binance for actual status
	if ord.Status == order.StatusExecuting {
		return r.reconcileExecutingOrder(ctx, ord, logger)
	}

	return nil
}

func (r *Reconciler) reconcileExecutingOrder(ctx context.Context, ord order.Order, logger *slog.Logger) error {
	// Get user credentials
	u, err := r.userRepo.GetUserByID(ctx, ord.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found, marking order as FAILED")
			errMsg := "user not found during reconciliation"
			return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFailed, nil, &errMsg)
		}
		return fmt.Errorf("get user: %w", err)
	}

	// Decrypt API keys
	apiKey, err := r.kms.Decrypt(ctx, u.APIKeyEncrypted)
	if err != nil {
		logger.Error("failed to decrypt API key during reconciliation")
		errMsg := "failed to decrypt API key during reconciliation"
		return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFailed, nil, &errMsg)
	}

	secretKey, err := r.kms.Decrypt(ctx, u.SecretKeyEncrypted)
	if err != nil {
		logger.Error("failed to decrypt secret key during reconciliation")
		errMsg := "failed to decrypt secret key during reconciliation"
		return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFailed, nil, &errMsg)
	}

	// Create per-user Binance client
	client := binance.NewRealClient(binance.ClientConfig{
		APIKey:    string(apiKey),
		SecretKey: string(secretKey),
		BaseURL:   r.config.BinanceBaseURL,
		UserID:    ord.UserID,
	}, r.logger)

	// Query order status from Binance
	resp, err := client.QueryOrderStatus(ctx, ord.Symbol, ord.ClientOrderID)
	if err != nil {
		var apiErr *binance.APIError
		if errors.As(err, &apiErr) {
			// If order doesn't exist on Binance, mark as FAILED
			if apiErr.Code == -2013 { // Order does not exist
				logger.Info("order not found on Binance, marking as FAILED")
				errMsg := "order not found on Binance during reconciliation"
				return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFailed, nil, &errMsg)
			}
		}
		return fmt.Errorf("query order status: %w", err)
	}

	// Map Binance status to our status
	logger.Info("binance order status",
		slog.String("binance_status", resp.Status),
		slog.Int64("binance_order_id", resp.OrderID),
	)

	binanceOrderID := resp.OrderID

	switch resp.Status {
	case "FILLED":
		return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFilled, &binanceOrderID, nil)
	case "CANCELED", "EXPIRED", "REJECTED":
		errMsg := fmt.Sprintf("binance status: %s", resp.Status)
		return r.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFailed, &binanceOrderID, &errMsg)
	case "NEW", "PARTIALLY_FILLED":
		// Still active on Binance, update timestamp to prevent re-processing
		logger.Info("order still active on Binance, updating timestamp",
			slog.String("binance_status", resp.Status),
		)
		return nil
	default:
		logger.Warn("unknown Binance status", slog.String("status", resp.Status))
		return nil
	}
}
