package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with helper methods
type Client struct {
	client *redis.Client
	ttl    time.Duration
}

// NewClient creates a new Redis client
func NewClient(addr, password string, db int, ttl time.Duration) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: client,
		ttl:    ttl,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Get retrieves a value from Redis and unmarshals it into the target
func (c *Client) Get(ctx context.Context, key string, target any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get from Redis: %w", err)
	}

	if err := json.Unmarshal([]byte(val), target); err != nil {
		return fmt.Errorf("failed to unmarshal Redis value: %w", err)
	}

	return nil
}

// Set stores a value in Redis with the default TTL
func (c *Client) Set(ctx context.Context, key string, value any) error {
	return c.SetWithTTL(ctx, key, value, c.ttl)
}

// SetWithTTL stores a value in Redis with a custom TTL
func (c *Client) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set in Redis: %w", err)
	}

	return nil
}

// Delete removes a key from Redis
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete from Redis: %w", err)
	}
	return nil
}

// DeletePattern deletes all keys matching a pattern
func (c *Client) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan Redis keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete Redis keys: %w", err)
		}
	}

	return nil
}

// Exists checks if a key exists in Redis
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// Expire sets an expiration time on a key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}
	return nil
}

// Increment increments a key by 1
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}
	return val, nil
}

// IncrementBy increments a key by a specific amount
func (c *Client) IncrementBy(ctx context.Context, key string, amount int64) (int64, error) {
	val, err := c.client.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment by: %w", err)
	}
	return val, nil
}

// GetOrSet implements the cache-aside pattern: get from cache, or execute fn and cache the result
func (c *Client) GetOrSet(ctx context.Context, key string, target any, fn func() (any, error)) error {
	// Try to get from cache
	err := c.Get(ctx, key, target)
	if err == nil {
		return nil // Cache hit
	}

	if err != ErrCacheMiss {
		// Log the error but continue to execute fn
		// In production, you might want to use proper logging here
	}

	// Cache miss - execute the function
	result, err := fn()
	if err != nil {
		return err
	}

	// Store in cache (best effort - don't fail if caching fails)
	_ = c.Set(ctx, key, result)

	// Copy result to target
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

// ErrCacheMiss is returned when a key is not found in the cache
var ErrCacheMiss = fmt.Errorf("cache miss")
