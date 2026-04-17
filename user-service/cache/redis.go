package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client.
// All cache operations go through this — nothing else in the codebase
// touches Redis directly. One place to change if we ever swap Redis out.
type Client struct {
	rdb *redis.Client
}

// TTL constants — how long each type of data lives in cache.
// Balance changes only on payment, so 5 minutes is safe.
// Short enough that stale data is never a real problem.
// Long enough to absorb thousands of reads between payments.
const (
	BalanceTTL = 5 * time.Minute
)

// Key prefixes — consistent naming prevents collisions between
// different types of cached data.
const (
	prefixBalance = "balance:"
)

// New creates a Redis client from environment variables.
// Returns nil if REDIS_URL is not set — this makes Redis optional.
// The service works without Redis, just slower.
func New() *Client {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis:6379" // default for Docker
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: os.Getenv("REDIS_PASSWORD"), // empty = no auth
		DB:       0,                           // default DB

		// Connection pool settings — how many Redis connections to keep open.
		// 10 is enough for a service with moderate load.
		PoolSize:    10,
		DialTimeout: 3 * time.Second,
		ReadTimeout: 2 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		// Log warning but do not crash — Redis is optional
		log.Printf("⚠️  Redis not available: %v — running without cache", err)
		return nil
	}

	log.Println("✅ Redis connected")
	return &Client{rdb: rdb}
}

// ─────────────────────────────────────────────────────────────────────────────
//  Balance cache operations
// ─────────────────────────────────────────────────────────────────────────────

// GetBalance retrieves a cached wallet balance.
// Returns (balance, true) on cache hit.
// Returns (0, false) on cache miss or any error.
// Errors are logged but never returned — cache failures are silent.
func (c *Client) GetBalance(ctx context.Context, userID string) (float64, bool) {
	if c == nil {
		return 0, false // Redis not available
	}

	key := prefixBalance + userID
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cache miss — normal, not an error
		return 0, false
	}
	if err != nil {
		log.Printf("⚠️  Redis GET error for %s: %v", key, err)
		return 0, false
	}

	balance, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("⚠️  Redis parse error for %s: %v", key, err)
		return 0, false
	}

	return balance, true
}

// SetBalance stores a wallet balance in cache with TTL.
// Called after every successful DB read.
// Errors are logged but never returned.
func (c *Client) SetBalance(ctx context.Context, userID string, balance float64) {
	if c == nil {
		return
	}

	key := prefixBalance + userID
	val := strconv.FormatFloat(balance, 'f', 2, 64)

	if err := c.rdb.Set(ctx, key, val, BalanceTTL).Err(); err != nil {
		log.Printf("⚠️  Redis SET error for %s: %v", key, err)
	}
}

// InvalidateBalance removes a user's cached balance.
// Called after every payment that changes the balance.
// The next read will miss the cache and fetch fresh from PostgreSQL.
func (c *Client) InvalidateBalance(ctx context.Context, userID string) {
	if c == nil {
		return
	}

	key := prefixBalance + userID
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		log.Printf("⚠️  Redis DEL error for %s: %v", key, err)
	}

	log.Printf("🗑️  Cache invalidated for user %s", userID)
}

// InvalidateMultiple removes cached balances for multiple users at once.
// Used after a payment to invalidate both sender and receiver simultaneously.
func (c *Client) InvalidateMultiple(ctx context.Context, userIDs ...string) {
	if c == nil {
		return
	}

	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = prefixBalance + id
	}

	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		log.Printf("⚠️  Redis multi-DEL error: %v", err)
	}
}

// HealthCheck returns Redis ping latency — useful for /health endpoint.
func (c *Client) HealthCheck(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("redis not connected")
	}
	return c.rdb.Ping(ctx).Err()
}

// Close shuts down the Redis connection pool cleanly.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	return c.rdb.Close()
}
