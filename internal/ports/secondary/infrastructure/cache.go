package infrastructure

import (
	"context"
	"time"
)

// Cache defines the port for caching operations
type Cache interface {

	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with an optional expiration
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
}
