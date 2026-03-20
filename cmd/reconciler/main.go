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

	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/metrics"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/reconciler"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("reconciler service failed", slog.String("error", err.Error()))
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

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
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

	// Repositories
	orderRepo := order.NewPostgresRepository(db)
	userRepo := user.NewPostgresRepository(db)

	// Reconciler configuration
	cfg := reconciler.DefaultConfig()
	cfg.BinanceBaseURL = getEnv("BINANCE_BASE_URL", "https://api.binance.com")

	// Create reconciler
	r := reconciler.New(orderRepo, userRepo, keyManager, cfg, logger)

	// Prometheus metrics server
	metricsAddr := getEnv("METRICS_ADDR", ":9091")
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
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
	reconcilerCtx, reconcilerCancel := context.WithCancel(context.Background())
	defer reconcilerCancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.Start(reconcilerCtx)
	}()

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		reconcilerCancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = metricsServer.Shutdown(shutdownCtx)

		return nil
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("reconciler error: %w", err)
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
