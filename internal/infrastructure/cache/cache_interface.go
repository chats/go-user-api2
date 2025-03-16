package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Connect establishes a connection to the cache
	Connect(ctx context.Context) error

	// Close closes the cache connection
	Close() error

	// Ping verifies the connection to the cache
	Ping(ctx context.Context) error

	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with an optional expiration time
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error

	// Clear clears all keys in the cache
	Clear(ctx context.Context) error

	// GetMulti retrieves multiple values from the cache
	GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)

	// GetInstance returns the cache client instance
	GetInstance() interface{}
}
