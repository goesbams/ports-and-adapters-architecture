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

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Increment atomically increments a numeric value stored at key
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement atomically decrements a numeric value stored at key
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// SetObject serializes and stores an object in the cache
	SetObject(ctx context.Context, key string, obj interface{}, expiration time.Duration) error

	// GetObject retrieves and deserializes an object from the cache
	GetObject(ctx context.Context, key string, obj interface{}) error

	// FlushAll clears all keys from the cache
	FlushAll(ctx context.Context) error
}
