package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/issue-tracker/pkg/ratelimit"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Limiter *ratelimit.Limiter
	// KeyFunc extracts the key from the request (e.g., IP, user ID)
	// If nil, defaults to IP-based rate limiting
	KeyFunc func(r *http.Request) string
}

// RateLimit creates a rate limiting middleware
func RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	if config.KeyFunc == nil {
		// Default to IP-based rate limiting
		config.KeyFunc = func(r *http.Request) string {
			return getClientIP(r)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := config.KeyFunc(r)
			ctx := r.Context()

			allowed, remaining, resetTime, err := config.Limiter.Allow(ctx, key)
			if err != nil {
				// Log error but don't block request if rate limiter fails
				// This ensures availability over strict rate limiting
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limiter.Limit()))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				retryAfter := int(resetTime.Sub(time.Now()).Seconds())
				if retryAfter < 0 {
					retryAfter = 0
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(fmt.Sprintf(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded. Try again after %d seconds"}}`, retryAfter)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (set by some proxies)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// RemoteAddr includes port, strip it
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
