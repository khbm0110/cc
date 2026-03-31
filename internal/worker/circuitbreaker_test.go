package worker

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker("user-1", DefaultCircuitBreakerConfig(), testLogger())

	if cb.State() != StateClosed {
		t.Errorf("expected CLOSED, got %s", cb.State())
	}
	if err := cb.Allow(); err != nil {
		t.Errorf("expected allow, got error: %v", err)
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     time.Second,
		HalfOpenMaxCalls: 1,
	}
	cb := NewCircuitBreaker("user-2", cfg, testLogger())

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != StateOpen {
		t.Errorf("expected OPEN after %d failures, got %s", 3, cb.State())
	}

	if err := cb.Allow(); err == nil {
		t.Error("expected error when circuit is open")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenMaxCalls: 1,
	}
	cb := NewCircuitBreaker("user-3", cfg, testLogger())

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("expected OPEN, got %s", cb.State())
	}

	time.Sleep(150 * time.Millisecond)

	if err := cb.Allow(); err != nil {
		t.Errorf("expected allow after reset timeout, got: %v", err)
	}

	if cb.State() != StateHalfOpen {
		t.Errorf("expected HALF_OPEN, got %s", cb.State())
	}
}

func TestCircuitBreaker_ClosesOnSuccessInHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenMaxCalls: 1,
	}
	cb := NewCircuitBreaker("user-4", cfg, testLogger())

	cb.RecordFailure()
	cb.RecordFailure()

	time.Sleep(150 * time.Millisecond)
	_ = cb.Allow() // transitions to half-open

	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Errorf("expected CLOSED after success in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_ReOpensOnFailureInHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenMaxCalls: 1,
	}
	cb := NewCircuitBreaker("user-5", cfg, testLogger())

	cb.RecordFailure()
	cb.RecordFailure()

	time.Sleep(150 * time.Millisecond)
	_ = cb.Allow() // transitions to half-open

	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("expected OPEN after failure in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenMaxCalls(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenMaxCalls: 2,
	}
	cb := NewCircuitBreaker("user-6", cfg, testLogger())

	cb.RecordFailure()
	cb.RecordFailure()

	time.Sleep(150 * time.Millisecond)

	// First call: allowed
	if err := cb.Allow(); err != nil {
		t.Errorf("first half-open call should be allowed: %v", err)
	}
	// Second call: allowed
	if err := cb.Allow(); err != nil {
		t.Errorf("second half-open call should be allowed: %v", err)
	}
	// Third call: denied
	if err := cb.Allow(); err == nil {
		t.Error("third half-open call should be denied")
	}
}

func TestCircuitBreakerManager_GetOrCreate(t *testing.T) {
	mgr := NewCircuitBreakerManager(DefaultCircuitBreakerConfig(), testLogger())

	cb1 := mgr.Get("user-a")
	cb2 := mgr.Get("user-a")
	cb3 := mgr.Get("user-b")

	if cb1 != cb2 {
		t.Error("expected same circuit breaker for same user")
	}
	if cb1 == cb3 {
		t.Error("expected different circuit breakers for different users")
	}
}
