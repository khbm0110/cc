package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// OrdersTotal tracks total orders processed, labeled by status outcome.
	OrdersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "copytrading",
			Subsystem: "orders",
			Name:      "total",
			Help:      "Total number of orders processed by status.",
		},
		[]string{"status"},
	)

	// OrderDuration tracks order processing duration in seconds.
	OrderDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "copytrading",
			Subsystem: "orders",
			Name:      "duration_seconds",
			Help:      "Order processing duration from signal to completion.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"symbol"},
	)

	// RateLimitHits tracks how many times users hit their rate limits.
	RateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "copytrading",
			Subsystem: "ratelimit",
			Name:      "hits_total",
			Help:      "Total rate limit hits by user.",
		},
		[]string{"user_id"},
	)

	// BinanceAPIErrors tracks Binance API call failures.
	BinanceAPIErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "copytrading",
			Subsystem: "binance",
			Name:      "api_errors_total",
			Help:      "Total Binance API errors by type.",
		},
		[]string{"error_type"},
	)

	// BinanceAPIDuration tracks Binance API call latency.
	BinanceAPIDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "copytrading",
			Subsystem: "binance",
			Name:      "api_duration_seconds",
			Help:      "Binance API call duration.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method"},
	)

	// CircuitBreakerState tracks circuit breaker state changes per user.
	CircuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "copytrading",
			Subsystem: "circuit_breaker",
			Name:      "state",
			Help:      "Circuit breaker state (0=closed, 1=half-open, 2=open).",
		},
		[]string{"user_id"},
	)

	// ReconcilerOrdersProcessed tracks orders processed by the reconciler.
	ReconcilerOrdersProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "copytrading",
			Subsystem: "reconciler",
			Name:      "orders_processed_total",
			Help:      "Total orders processed by the reconciler.",
		},
	)

	// EventBusMessagesReceived tracks messages received from the event bus.
	EventBusMessagesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "copytrading",
			Subsystem: "eventbus",
			Name:      "messages_received_total",
			Help:      "Total messages received from the event bus.",
		},
	)
)

// Register registers all metrics with the default Prometheus registry.
func Register() {
	prometheus.MustRegister(
		OrdersTotal,
		OrderDuration,
		RateLimitHits,
		BinanceAPIErrors,
		BinanceAPIDuration,
		CircuitBreakerState,
		ReconcilerOrdersProcessed,
		EventBusMessagesReceived,
	)
}
