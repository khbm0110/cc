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
	"github.com/redis/go-redis/v9"

	"github.com/khbm0110/copy-trading-platform/internal/api"
	"github.com/khbm0110/copy-trading-platform/internal/api/handlers"
	"github.com/khbm0110/copy-trading-platform/internal/config"
	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

func main() {
	// ============================================================
	// STEP 1: Create TEMP logger for early errors (before config)
	// ============================================================
	tempLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn, // Conservative level for startup
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

	// DO NOT use slog.SetDefault - pass logger explicitly to all dependencies
	// This ensures testability and prevents global state issues
	// logger.SetDefault(logger) // REMOVED: Explicit logger passing only

	// ============================================================
	// STEP 4: Connect to PostgreSQL
	// ============================================================
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 5: Connect to Redis using GetRedisOptions()
	// ============================================================
	redisOpts, err := cfg.GetRedisOptions()
	if err != nil {
		logger.Error("failed to parse Redis URL", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Error("failed to connect to redis", "error", err)
			os.Exit(1)
		}
	}

	// ============================================================
	// STEP 6: Initialize Repositories
	// ============================================================
	userRepo := user.NewPostgresRepository(db)
	orderRepo := order.NewPostgresRepository(db)

	// ============================================================
	// STEP 7: Initialize KMS
	// ============================================================
	keyManager, err := kms.NewMockKMS([]byte(cfg.KMSEncryptionKey))
	if err != nil {
		logger.Error("failed to initialize kms", "error", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 8: Initialize EventBus with explicit logger
	// ============================================================
	eventBus, err := eventbus.New(eventbus.Config{
		RedisOptions:  redisOpts, // Reuse the already parsed Redis options
		ConsumerGroup: "api-group",
		ConsumerName:  "api",
	}, logger)
	if err != nil {
		logger.Error("failed to initialize eventbus", "error", err)
		os.Exit(1)
	}
	defer eventBus.Close()

	// ============================================================
	// STEP 9: Initialize Handlers with explicit logger
	// ============================================================
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, logger)
	usersHandler := handlers.NewUsersHandler(userRepo, keyManager, cfg, logger)
	ordersHandler := handlers.NewOrdersHandler(orderRepo, userRepo, logger)
	adminHandler := handlers.NewAdminHandler(userRepo, logger)
	signalsHandler := handlers.NewSignalsHandler(eventBus, userRepo, logger)

	// ============================================================
	// STEP 10: Setup Router with explicit logger
	// ============================================================
	router := api.NewRouter(
		authHandler,
		usersHandler,
		ordersHandler,
		adminHandler,
		signalsHandler,
		userRepo,
		cfg.JWTSecret,
		logger,
	)

	// ============================================================
	// STEP 11: Start HTTP Server
	// ============================================================
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.GoPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting api server", "port", cfg.GoPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("shutting down server...")

	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server forced to shutdown", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("server exited gracefully")
}
