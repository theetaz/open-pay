package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements sliding window rate limiting backed by Redis.
type RedisRateLimiter struct {
	client *redis.Client
	config *Config
}

// Config holds rate limit configuration.
type Config struct {
	// DefaultLimit is the fallback requests-per-window.
	DefaultLimit int
	// Window is the sliding window duration.
	Window time.Duration
	// EndpointLimits maps "METHOD:path_pattern" to a specific limit.
	// Example: "POST:/v1/payment" → 100, "GET:/v1/payment/list" → 200
	EndpointLimits map[string]int
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultLimit: 100,
		Window:       time.Minute,
		EndpointLimits: map[string]int{
			"POST:/v1/payments":       100,
			"GET:/v1/payments":        200,
			"GET:/v1/payments/*":      300,
			"POST:/v1/payment-links":  50,
			"GET:/v1/payment-links":   100,
			"GET:/v1/payment-links/*": 100,
			"PUT:/v1/payment-links/*": 50,
			"DELETE:/v1/payment-links/*": 50,
		},
	}
}

// NewRedisRateLimiter creates a Redis-backed rate limiter.
func NewRedisRateLimiter(client *redis.Client, cfg *Config) *RedisRateLimiter {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &RedisRateLimiter{client: client, config: cfg}
}

// AllowWithEndpoint checks rate limit using a composite key of clientID + endpoint.
func (r *RedisRateLimiter) AllowWithEndpoint(clientID, method, path string) (allowed bool, limit int, remaining int, resetAt time.Time) {
	limit = r.resolveLimit(method, path)
	key := fmt.Sprintf("rl:%s:%s:%s", clientID, method, normalizePath(path))
	return r.check(key, limit)
}

// Allow implements the RateLimiter interface for backwards compatibility (per-key only).
func (r *RedisRateLimiter) Allow(key string) (allowed bool, limit int, remaining int, resetAt time.Time) {
	return r.check("rl:"+key, r.config.DefaultLimit)
}

func (r *RedisRateLimiter) check(key string, limit int) (bool, int, int, time.Time) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-r.config.Window)
	resetAt := now.Add(r.config.Window)

	pipe := r.client.Pipeline()

	// Remove entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixMicro(), 10))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMicro()), Member: now.UnixMicro()})

	// Set TTL on the key
	pipe.Expire(ctx, key, r.config.Window+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		// On Redis error, allow the request (fail-open)
		return true, limit, limit, resetAt
	}

	count := int(countCmd.Val())
	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	if count >= limit {
		return false, limit, 0, resetAt
	}

	return true, limit, remaining, resetAt
}

func (r *RedisRateLimiter) resolveLimit(method, path string) int {
	// Try exact match first
	key := method + ":" + path
	if limit, ok := r.config.EndpointLimits[key]; ok {
		return limit
	}

	// Try wildcard match (replace last segment with *)
	normalized := normalizePath(path)
	wildcard := method + ":" + normalized
	if limit, ok := r.config.EndpointLimits[wildcard]; ok {
		return limit
	}

	return r.config.DefaultLimit
}

// normalizePath replaces UUID path segments with * for wildcard matching.
func normalizePath(path string) string {
	// Simple normalization: replace UUID-like segments
	result := make([]byte, 0, len(path))
	i := 0
	for i < len(path) {
		if path[i] == '/' {
			result = append(result, '/')
			i++
			// Check if next segment looks like a UUID (36 chars with dashes)
			j := i
			for j < len(path) && path[j] != '/' {
				j++
			}
			seg := path[i:j]
			if len(seg) == 36 && seg[8] == '-' && seg[13] == '-' {
				result = append(result, '*')
			} else {
				result = append(result, seg...)
			}
			i = j
		} else {
			result = append(result, path[i])
			i++
		}
	}
	return string(result)
}
