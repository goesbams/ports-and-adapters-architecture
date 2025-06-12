package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
	}
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return val, nil
}

// Set stores a value in the cache with an optional expiration
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	err := c.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Exists checks if a key exists in the cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}

	return val > 0, nil
}

// Increment atomically increments a numeric value stored at key
func (c *RedisCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	newVal, err := c.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}

	return newVal, nil
}

// Decrement atomically decrements a numeric value stored at key
func (c *RedisCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	newVal, err := c.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}

	return newVal, nil
}

// SetObject serializes and stores an object in the cache
func (c *RedisCache) SetObject(ctx context.Context, key string, obj interface{}, expiration time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	return c.Set(ctx, key, data, expiration)
}

// GetObject retrieves and deserializes an object from the cache
func (c *RedisCache) GetObject(ctx context.Context, key string, obj interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, obj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal object: %w", err)
	}

	return nil
}

// FlushAll clears all keys from the cache
func (c *RedisCache) FlushAll(ctx context.Context) error {
	err := c.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all keys: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping tests the connection to Redis
func (c *RedisCache) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	return err
}
