package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, keys ...string) error

	// DeleteByPattern removes all keys matching a pattern
	DeleteByPattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection
	Close() error
}

// CacheKeyBuilder helps build consistent cache keys
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// Build constructs a cache key with the given parts
func (b *CacheKeyBuilder) Build(parts ...interface{}) string {
	key := b.prefix
	for _, part := range parts {
		key += ":"
		switch v := part.(type) {
		case string:
			key += v
		case int:
			key += string(rune('0' + v))
		default:
			key += ""
		}
	}
	return key
}
