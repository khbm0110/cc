package main

import (
	"context"
	"database/sql"
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
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to PostgreSQL
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

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer rdb.Close()

	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Error("failed to connect to redis", "error", err)
			os.Exit(1)
		}
	}

	// Initialize Repositories
	userRepo := user.NewPostgresRepository(db)
	orderRepo := order.NewRepository(db)

	// Initialize KMS
	keyManager, err := kms.NewKeyManager(cfg.KMSKey)
	if err != nil {
		logger.Error("failed to initialize kms", "error", err)
		os.Exit(1)
	}

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, logger)
	usersHandler := handlers.NewUsersHandler(userRepo, keyManager, cfg, logger)
	ordersHandler := handlers.NewOrdersHandler(orderRepo, logger)
	adminHandler := handlers.NewAdminHandler(userRepo, logger)

	// Setup Router
	router := api.NewRouter(authHandler, usersHandler, ordersHandler, adminHandler, userRepo, cfg.JWTSecret)

	// Start HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.GoPort,
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