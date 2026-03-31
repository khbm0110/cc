package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

// Config holds all application configuration.
type Config struct {
	GoPort               string
	DatabaseURL          string
	RedisURL             string
	JWTSecret            string
	KMSEncryptionKey     string
	BinanceAPIURL        string
	NowPaymentsAPIKey    string
	NowPaymentsIPNSecret string
	WorkerName           string
	MetricsAddr          string
	LogLevel             string // debug, info, warn, error
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{
		GoPort:               getEnv("GO_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		RedisURL:             getEnv("REDIS_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		KMSEncryptionKey:     getEnv("KMS_ENCRYPTION_KEY", ""),
		BinanceAPIURL:        getEnv("BINANCE_API_URL", "https://api.binance.com"),
		NowPaymentsAPIKey:    getEnv("NOWPAYMENTS_API_KEY", ""),
		NowPaymentsIPNSecret: getEnv("NOWPAYMENTS_IPN_SECRET", ""),
		WorkerName:           getEnv("WORKER_NAME", "worker-1"),
		MetricsAddr:          getEnv("METRICS_ADDR", ":9090"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.KMSEncryptionKey) != 32 {
		return fmt.Errorf("KMS_ENCRYPTION_KEY must be exactly 32 bytes")
	}
	if c.NowPaymentsAPIKey == "" {
		return fmt.Errorf("NOWPAYMENTS_API_KEY is required")
	}
	if c.NowPaymentsIPNSecret == "" {
		return fmt.Errorf("NOWPAYMENTS_IPN_SECRET is required")
	}
	// Validate LogLevel
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid
	default:
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}
	return nil
}

// GetRedisOptions parses the RedisURL and returns redis.Options.
// Uses redis.ParseURL() for proper URL parsing.
// Returns an error if parsing fails.
func (c *Config) GetRedisOptions() (*redis.Options, error) {
	opts, err := redis.ParseURL(c.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	return opts, nil
}

// GetLogLevel returns the slog.Level corresponding to the LogLevel string.
func (c *Config) GetLogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
