package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/khbm0110/copy-trading-platform/internal/binance"
	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/metrics"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
	"github.com/khbm0110/copy-trading-platform/internal/validator"
)

const (
	tradeSignalsStream = "trade_signals"
	maxRetries         = 3
)

type Config struct {
	BinanceBaseURL string
	MaxSlippage    float64
}

type Worker struct {
	orderRepo   order.Repository
	dlqRepo     order.DLQRepository
	userRepo    user.Repository
	kms         kms.KeyManager
	eventBus    *eventbus.EventBus
	validator   *validator.Validator
	rateLimiter *RateLimiter
	cbManager   *CircuitBreakerManager
	config      Config
	logger      *slog.Logger
}

func NewWorker(
	orderRepo order.Repository,
	dlqRepo order.DLQRepository,
	userRepo user.Repository,
	keyManager kms.KeyManager,
	eventBus *eventbus.EventBus,
	v *validator.Validator,
	rateLimiter *RateLimiter,
	cbManager *CircuitBreakerManager,
	cfg Config,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		orderRepo:   orderRepo,
		dlqRepo:     dlqRepo,
		userRepo:    userRepo,
		kms:         keyManager,
		eventBus:    eventBus,
		validator:   v,
		rateLimiter: rateLimiter,
		cbManager:   cbManager,
		config:      cfg,
		logger:      logger.With(slog.String("component", "worker")),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	if err := w.eventBus.EnsureStream(ctx, tradeSignalsStream); err != nil {
		return fmt.Errorf("ensure stream: %w", err)
	}
	w.logger.Info("worker started, listening for trade signals")
	return w.eventBus.Subscribe(ctx, tradeSignalsStream, w.handleSignal)
}

func (w *Worker) handleSignal(ctx context.Context, msg eventbus.Message) error {
	signal := msg.Signal
	start := time.Now()

	logger := w.logger.With(
		slog.String("signal_id", signal.SignalID),
		slog.Int64("user_id", signal.UserID),
		slog.String("symbol", signal.Symbol),
		slog.String("client_order_id", signal.ClientOrderID),
	)

	logger.Info("processing trade signal")
	metrics.EventBusMessagesReceived.Inc()

	uwp, err := w.userRepo.GetUserWithPlan(ctx, signal.UserID)
	if err != nil {
		logger.Error("failed to fetch user", slog.String("error", err.Error()))
		return nil
	}

	userIDStr := strconv.FormatInt(signal.UserID, 10)
	if !w.rateLimiter.Allow(signal.UserID, uwp.Plan.OrderLimitPerMin) {
		logger.Warn("rate limit exceeded, requeuing")
		metrics.RateLimitHits.WithLabelValues(userIDStr).Inc()
		return w.eventBus.RequeueWithDelay(ctx, tradeSignalsStream, signal, 5*time.Second)
	}

	cb := w.cbManager.Get(userIDStr)
	if err := cb.Allow(); err != nil {
		logger.Warn("circuit breaker open", slog.String("error", err.Error()))
		return w.eventBus.RequeueWithDelay(ctx, tradeSignalsStream, signal, 10*time.Second)
	}

	apiKeyEnc, err := w.kms.Decrypt(ctx, uwp.User.APIKeyEncrypted)
	if err != nil {
		logger.Error("failed to decrypt API key", slog.String("error", err.Error()))
		_ = w.dlqRepo.Publish(ctx, signal, signal.UserID, "KMS decrypt API key: "+err.Error())
		return nil
	}
	secretKeyEnc, err := w.kms.Decrypt(ctx, uwp.User.SecretKeyEncrypted)
	if err != nil {
		logger.Error("failed to decrypt secret key", slog.String("error", err.Error()))
		_ = w.dlqRepo.Publish(ctx, signal, signal.UserID, "KMS decrypt secret key: "+err.Error())
		return nil
	}

	client := binance.NewRealClient(binance.ClientConfig{
		APIKey:    string(apiKeyEnc),
		SecretKey: string(secretKeyEnc),
		BaseURL:   w.config.BinanceBaseURL,
		UserID:    signal.UserID,
	}, w.logger)

	vResult := w.validator.ValidateSignal(ctx, client, signal, uwp.Plan.MaxExposureRatio)
	if !vResult.Valid {
		logger.Warn("signal validation failed", slog.Any("errors", vResult.Errors))
		metrics.OrdersTotal.WithLabelValues("validation_failed").Inc()
		return nil
	}

	ord, err := w.orderRepo.CreateOrder(ctx, order.CreateOrderParams{
		UserID:        signal.UserID,
		ClientOrderID: signal.ClientOrderID,
		Symbol:        signal.Symbol,
		Side:          signal.Side,
		Quantity:      signal.Quantity,
		Price:         signal.Price,
	})
	if err != nil {
		if errors.Is(err, order.ErrDuplicateOrder) {
			logger.Info("duplicate order, skipping", slog.Int64("existing_order_id", ord.ID))
			return nil
		}
		logger.Error("failed to create order", slog.String("error", err.Error()))
		return fmt.Errorf("create order: %w", err)
	}

	logger = logger.With(slog.Int64("order_id", ord.ID))
	if err := w.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusPending, order.StatusExecuting, nil, nil); err != nil {
		logger.Error("failed to transition to EXECUTING", slog.String("error", err.Error()))
		return fmt.Errorf("update status: %w", err)
	}

	resp, err := w.executeWithRetry(ctx, client, signal, ord.ID, cb, logger)
	if err != nil {
		logger.Error("order execution failed after retries", slog.String("error", err.Error()))
		errMsg := err.Error()
		_ = w.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFailed, nil, &errMsg)
		metrics.OrdersTotal.WithLabelValues("failed").Inc()
		cb.RecordFailure()

		if dlqErr := w.dlqRepo.Publish(ctx, signal, signal.UserID, errMsg); dlqErr != nil {
			logger.Error("failed to publish to DLQ", slog.String("error", dlqErr.Error()))
		}
		return nil
	}

	binanceID := resp.OrderID
	if err := w.orderRepo.UpdateStatus(ctx, ord.ID, order.StatusExecuting, order.StatusFilled, &binanceID, nil); err != nil {
		logger.Error("failed to update order to FILLED", slog.String("error", err.Error()))
		return fmt.Errorf("update status to filled: %w", err)
	}

	cb.RecordSuccess()
	metrics.OrdersTotal.WithLabelValues("filled").Inc()
	metrics.OrderDuration.WithLabelValues(signal.Symbol).Observe(time.Since(start).Seconds())
	logger.Info("order filled successfully", slog.Int64("binance_order_id", binanceID))
	return nil
}

func (w *Worker) executeWithRetry(ctx context.Context, client binance.Client, signal eventbus.TradeSignal, orderID int64, cb *CircuitBreaker, logger *slog.Logger) (*binance.OrderResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			_ = w.orderRepo.IncrementRetry(ctx, orderID)
		}
		resp, err := client.ExecuteOrder(ctx, binance.OrderRequest{
			Symbol:        signal.Symbol,
			Side:          signal.Side,
			Quantity:      signal.Quantity,
			Price:         signal.Price,
			ClientOrderID: signal.ClientOrderID,
		})
		if err == nil {
			return resp, nil
		}
		lastErr = err
		var apiErr *binance.APIError
		if errors.As(err, &apiErr) && !apiErr.IsRetriable() {
			return nil, fmt.Errorf("non-retriable error: %w", err)
		}
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}