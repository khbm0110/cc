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

	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/metrics"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
	"github.com/khbm0110/copy-trading-platform/internal/validator"
	"github.com/khbm0110/copy-trading-platform/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("worker service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// Register Prometheus metrics
	metrics.Register()

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://copytrading:copytrading@localhost:5432/copytrading?sslmode=disable")
	db, err := sql.Open("postgres", dbURL)
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
	passphrase := getEnv("KMS_PASSPHRASE", "dev-passphrase-change-me!!")
	keyManager, err := kms.NewMockKMSFromPassphrase(passphrase)
	if err != nil {
		return fmt.Errorf("create KMS: %w", err)
	}

	// Redis Event Bus
	eb, err := eventbus.New(eventbus.Config{
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,
		ConsumerGroup: "worker-group",
		ConsumerName:  getEnv("WORKER_NAME", "worker-1"),
	}, logger)
	if err != nil {
		return fmt.Errorf("create event bus: %w", err)
	}
	defer eb.Close()

	// Repositories
	orderRepo := order.NewPostgresRepository(db)
	userRepo := user.NewPostgresRepository(db)

	// Validator
	balanceChecker := validator.NewMockBalanceChecker()
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
		userRepo,
		keyManager,
		eb,
		v,
		rateLimiter,
		cbManager,
		worker.Config{
			BinanceBaseURL: getEnv("BINANCE_BASE_URL", "https://api.binance.com"),
			MaxSlippage:    0.02,
		},
		logger,
	)

	// Prometheus metrics server
	metricsAddr := getEnv("METRICS_ADDR", ":9090")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("ok"))
	})
	metricsServer := &http.Server{Addr: metricsAddr, Handler: metricsMux}

	go func() {
		logger.Info("metrics server starting", slog.String("addr", metricsAddr))
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

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
