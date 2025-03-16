package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/chats/go-user-api/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RedisCache implements the Cache interface for Redis
type RedisCache struct {
	config config.CacheConfig
	client *redis.Client
}

// NewRedis creates a new Redis cache connection
func NewRedis(config config.CacheConfig) (Cache, error) {
	return &RedisCache{
		config: config,
	}, nil
}

// Connect establishes a connection to Redis
func (c *RedisCache) Connect(ctx context.Context) error {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Password:     c.config.Password,
		DB:           c.config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   3,
	})

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	c.client = client
	log.Info().Msg("Connected to Redis successfully")
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if c.client != nil {
		log.Info().Msg("Closing Redis connection")
		return c.client.Close()
	}
	return nil
}

// Ping verifies the connection to Redis
func (c *RedisCache) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return c.client.Ping(ctx).Err()
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Key not found, return nil without error
	}
	return val, err
}

// Set stores a value in Redis
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Delete removes a key from Redis
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Clear clears all keys in Redis
func (c *RedisCache) Clear(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// GetMulti retrieves multiple values from Redis
func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	pipeline := c.client.Pipeline()

	cmds := make(map[string]*redis.StringCmd)
	for _, key := range keys {
		cmds[key] = pipeline.Get(ctx, key)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	results := make(map[string][]byte)
	for key, cmd := range cmds {
		val, err := cmd.Bytes()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		if err == nil {
			results[key] = val
		}
	}

	return results, nil
}

// GetInstance returns the Redis client instance
func (c *RedisCache) GetInstance() interface{} {
	return c.client
}
