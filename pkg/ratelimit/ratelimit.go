package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/issue-tracker/pkg/cache"
)

// Limiter provides rate limiting functionality
type Limiter struct {
	cache  cache.Cache
	limit  int           // Maximum number of requests
	window time.Duration // Time window
}

// NewLimiter creates a new rate limiter
func NewLimiter(cache cache.Cache, limit int, window time.Duration) *Limiter {
	return &Limiter{
		cache:  cache,
		limit:  limit,
		window: window,
	}
}

// Allow checks if a request is allowed for the given key
// Returns: (allowed bool, remaining int, resetTime time.Time, error)
func (l *Limiter) Allow(ctx context.Context, key string) (bool, int, time.Time, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(l.window.Seconds()))

	// Get current count
	countStr, err := l.cache.Get(ctx, windowKey)
	count := 0
	if err == nil && countStr != "" {
		// Parse existing count
		fmt.Sscanf(countStr, "%d", &count)
	}

	// Check if limit exceeded
	if count >= l.limit {
		resetTime := now.Truncate(l.window).Add(l.window)
		return false, 0, resetTime, nil
	}

	// Increment counter
	count++
	err = l.cache.Set(ctx, windowKey, fmt.Sprintf("%d", count), l.window)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	remaining := l.limit - count
	resetTime := now.Truncate(l.window).Add(l.window)

	return true, remaining, resetTime, nil
}

// AllowN checks if N requests are allowed for the given key
func (l *Limiter) AllowN(ctx context.Context, key string, n int) (bool, int, time.Time, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(l.window.Seconds()))

	// Get current count
	countStr, err := l.cache.Get(ctx, windowKey)
	count := 0
	if err == nil && countStr != "" {
		fmt.Sscanf(countStr, "%d", &count)
	}

	// Check if limit exceeded
	if count+n > l.limit {
		resetTime := now.Truncate(l.window).Add(l.window)
		return false, 0, resetTime, nil
	}

	// Increment counter by n
	count += n
	err = l.cache.Set(ctx, windowKey, fmt.Sprintf("%d", count), l.window)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	remaining := l.limit - count
	resetTime := now.Truncate(l.window).Add(l.window)

	return true, remaining, resetTime, nil
}

// Reset clears the rate limit for a given key
func (l *Limiter) Reset(ctx context.Context, key string) error {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(l.window.Seconds()))
	return l.cache.Delete(ctx, windowKey)
}

// Limit returns the maximum number of requests allowed
func (l *Limiter) Limit() int {
	return l.limit
}

// Window returns the time window for rate limiting
func (l *Limiter) Window() time.Duration {
	return l.window
}
