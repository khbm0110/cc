package worker

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	for i := 0; i < 5; i++ {
		if !rl.Allow(1, 5) {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_DenyOverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		rl.Allow(1, 3)
	}

	if rl.Allow(1, 3) {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiter_IndependentPerUser(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	rl.Allow(1, 2)
	rl.Allow(1, 2)

	// User 1 is at limit
	if rl.Allow(1, 2) {
		t.Error("user 1 should be rate limited")
	}

	// User 2 should still be allowed
	if !rl.Allow(2, 2) {
		t.Error("user 2 should not be rate limited")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)

	rl.Allow(1, 2)
	rl.Allow(1, 2)

	if rl.Allow(1, 2) {
		t.Error("should be rate limited")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	if !rl.Allow(1, 2) {
		t.Error("should be allowed after window expiry")
	}
}

func TestRateLimiter_CustomLimit(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	// Use custom limit of 2
	rl.Allow(1, 2)
	rl.Allow(1, 2)

	if rl.Allow(1, 2) {
		t.Error("should be denied with custom limit of 2")
	}
}

func TestRateLimiter_DefaultLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	// limit=0 means use default
	for i := 0; i < 3; i++ {
		if !rl.Allow(1, 0) {
			t.Errorf("request %d should be allowed with default limit", i+1)
		}
	}
	if rl.Allow(1, 0) {
		t.Error("should be denied with default limit of 3")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	rl.Allow(1, 2)
	rl.Allow(1, 2)

	rl.Reset(1)

	if !rl.Allow(1, 2) {
		t.Error("should be allowed after reset")
	}
}

func TestRateLimiter_Count(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	if rl.Count(1) != 0 {
		t.Error("count should be 0 initially")
	}

	rl.Allow(1, 10)
	rl.Allow(1, 10)
	rl.Allow(1, 10)

	if rl.Count(1) != 3 {
		t.Errorf("count should be 3, got %d", rl.Count(1))
	}
}
