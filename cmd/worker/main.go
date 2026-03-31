package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/khbm0110/copy-trading-platform/internal/config"
	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/metrics"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
	"github.com/khbm0110/copy-trading-platform/internal/validator"
	"github.com/khbm0110/copy-trading-platform/internal/worker"
)

func main() {
	// ============================================================
	// STEP 1: Create TEMP logger for early errors (before config)
	// ============================================================
	tempLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	// ============================================================
	// STEP 2: Load configuration
	// ============================================================
	cfg, err := config.Load()
	if err != nil {
		tempLogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 3: Create FINAL logger using config.LogLevel
	// ============================================================
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	if err := run(logger, cfg); err != nil {
		logger.Error("worker service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg *config.Config) error {
	// Register Prometheus metrics
	metrics.Register()

	// Database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	logger.Info("connected to database")

	// KMS (mock for local development)
	// In production, use AWS KMS for key management
	keyManager, err := kms.NewMockKMSFromPassphrase(cfg.KMSEncryptionKey)
	if err != nil {
		return fmt.Errorf("create KMS: %w", err)
	}

	// ============================================================
	// STEP 4: Redis Event Bus using GetRedisOptions()
	// ============================================================
	redisOpts, err := cfg.GetRedisOptions()
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	eb, err := eventbus.New(eventbus.Config{
		RedisOptions:  redisOpts,
		ConsumerGroup: "worker-group",
		ConsumerName:  cfg.WorkerName,
	}, logger)
	if err != nil {
		return fmt.Errorf("create event bus: %w", err)
	}
	defer eb.Close()

	// Repositories
	orderRepo := order.NewPostgresRepository(db)
	dlqRepo := order.NewPostgresDLQRepository(db)
	userRepo := user.NewPostgresRepository(db)

	// Validator
	balanceChecker := &validator.LiveBalanceChecker{}
	v := validator.New(validator.Config{
		DefaultMaxSlippage: 0.02,
	}, balanceChecker)

	// Rate Limiter
	rateLimiter := worker.NewRateLimiter(10, time.Minute)

	// Circuit Breaker Manager
	cbManager := worker.NewCircuitBreakerManager(
		worker.DefaultCircuitBreakerConfig(),
		logger,
	)

	// Worker
	w := worker.NewWorker(
		orderRepo,
		dlqRepo,
		userRepo,
		keyManager,
		eb,
		v,
		rateLimiter,
		cbManager,
		worker.Config{
			BinanceBaseURL: cfg.BinanceAPIURL,
			MaxSlippage:    0.02,
		},
		logger,
	)

	// Prometheus metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("ok"))
	})
	metricsServer := &http.Server{Addr: cfg.MetricsAddr, Handler: metricsMux}

	go func() {
		logger.Info("metrics server starting", slog.String("addr", cfg.MetricsAddr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", slog.String("error", err.Error()))
		}
	}()

	// Graceful shutdown
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- w.Start(workerCtx)
	}()

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		workerCancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = metricsServer.Shutdown(shutdownCtx)

		return nil
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("worker error: %w", err)
		}
		return nil
	}
}
