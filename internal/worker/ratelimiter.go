package worker

import (
	"sync"
	"time"
)

// RateLimiter implements a per-user sliding window rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	windows  map[int64]*userWindow
	limit    int
	interval time.Duration
}

type userWindow struct {
	timestamps []time.Time
}

// NewRateLimiter creates a new per-user rate limiter.
func NewRateLimiter(defaultLimit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		windows:  make(map[int64]*userWindow),
		limit:    defaultLimit,
		interval: interval,
	}
}

// Allow checks if a user is allowed to make a request. Returns true if allowed.
func (rl *RateLimiter) Allow(userID int64, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limit == 0 {
		limit = rl.limit
	}

	w, ok := rl.windows[userID]
	if !ok {
		w = &userWindow{}
		rl.windows[userID] = w
	}

	now := time.Now()
	cutoff := now.Add(-rl.interval)

	// Remove expired timestamps
	valid := w.timestamps[:0]
	for _, ts := range w.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	w.timestamps = valid

	if len(w.timestamps) >= limit {
		return false
	}

	w.timestamps = append(w.timestamps, now)
	return true
}

// Reset clears the rate limit state for a user.
func (rl *RateLimiter) Reset(userID int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.windows, userID)
}

// Count returns the current request count for a user in the current window.
func (rl *RateLimiter) Count(userID int64) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	w, ok := rl.windows[userID]
	if !ok {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-rl.interval)

	count := 0
	for _, ts := range w.timestamps {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}
