package worker

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/khbm0110/copy-trading-platform/internal/metrics"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	StateClosed   CircuitState = 0
	StateHalfOpen CircuitState = 1
	StateOpen     CircuitState = 2
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	HalfOpenMaxCalls int
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		HalfOpenMaxCalls: 2,
	}
}

// CircuitBreaker implements the circuit breaker pattern per user.
type CircuitBreaker struct {
	mu             sync.Mutex
	state          CircuitState
	failureCount   int
	successCount   int
	halfOpenCalls  int
	lastFailure    time.Time
	config         CircuitBreakerConfig
	userID         string
	logger         *slog.Logger
}

// NewCircuitBreaker creates a new CircuitBreaker for a specific user.
func NewCircuitBreaker(userID string, cfg CircuitBreakerConfig, logger *slog.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		state:  StateClosed,
		config: cfg,
		userID: userID,
		logger: logger.With(slog.String("user_id", userID), slog.String("component", "circuit_breaker")),
	}
}

// Allow checks if the circuit breaker allows the request.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.config.ResetTimeout {
			cb.transition(StateHalfOpen)
			cb.halfOpenCalls = 1 // count the transition call
			return nil
		}
		return fmt.Errorf("%w: user %s, retry after %v", ErrCircuitOpen, cb.userID,
			cb.config.ResetTimeout-time.Since(cb.lastFailure))
	case StateHalfOpen:
		if cb.halfOpenCalls >= cb.config.HalfOpenMaxCalls {
			return fmt.Errorf("%w: user %s, max half-open calls reached", ErrCircuitOpen, cb.userID)
		}
		cb.halfOpenCalls++
		return nil
	default:
		return fmt.Errorf("unknown circuit breaker state: %d", cb.state)
	}
}

// RecordSuccess records a successful operation.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	if cb.state == StateHalfOpen {
		cb.transition(StateClosed)
		cb.failureCount = 0
		cb.halfOpenCalls = 0
	}
}

// RecordFailure records a failed operation.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.transition(StateOpen)
		}
	case StateHalfOpen:
		cb.transition(StateOpen)
	}
}

// State returns the current state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) transition(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	metrics.CircuitBreakerState.WithLabelValues(cb.userID).Set(float64(newState))
	cb.logger.Info("circuit breaker state transition",
		slog.String("from", oldState.String()),
		slog.String("to", newState.String()),
	)
}

// CircuitBreakerManager manages per-user circuit breakers.
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
	logger   *slog.Logger
}

// NewCircuitBreakerManager creates a new manager.
func NewCircuitBreakerManager(cfg CircuitBreakerConfig, logger *slog.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   cfg,
		logger:   logger,
	}
}

// Get returns the circuit breaker for a given user, creating one if needed.
func (m *CircuitBreakerManager) Get(userID string) *CircuitBreaker {
	m.mu.RLock()
	cb, ok := m.breakers[userID]
	m.mu.RUnlock()
	if ok {
		return cb
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, ok := m.breakers[userID]; ok {
		return cb
	}

	cb = NewCircuitBreaker(userID, m.config, m.logger)
	m.breakers[userID] = cb
	return cb
}
