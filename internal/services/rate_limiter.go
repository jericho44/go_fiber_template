package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-fiber-template/internal/config"

	"github.com/redis/go-redis/v9"
)

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	// CheckLimit checks if the request is within the rate limit
	CheckLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error)
	// CheckUserLimit checks rate limit for a specific user
	CheckUserLimit(ctx context.Context, userID uint, limit int, window time.Duration) (*RateLimitResult, error)
	// CheckIPLimit checks rate limit for a specific IP address
	CheckIPLimit(ctx context.Context, ip string, limit int, window time.Duration) (*RateLimitResult, error)
	// CheckEndpointLimit checks rate limit for a specific endpoint
	CheckEndpointLimit(ctx context.Context, key string, endpoint string, limit int, window time.Duration) (*RateLimitResult, error)
	// Reset resets the rate limit for a specific key
	Reset(ctx context.Context, key string) error
	// GetStats returns current rate limit statistics
	GetStats(ctx context.Context, key string) (*RateLimitStats, error)
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool          `json:"allowed"`
	Limit         int           `json:"limit"`
	Remaining     int           `json:"remaining"`
	ResetTime     time.Time     `json:"reset_time"`
	RetryAfter    time.Duration `json:"retry_after"`
	TotalRequests int           `json:"total_requests"`
}

// RateLimitStats contains statistics about rate limiting
type RateLimitStats struct {
	Key          string        `json:"key"`
	Limit        int           `json:"limit"`
	Window       time.Duration `json:"window"`
	CurrentCount int           `json:"current_count"`
	ResetTime    time.Time     `json:"reset_time"`
	FirstRequest time.Time     `json:"first_request"`
	LastRequest  time.Time     `json:"last_request"`
}

// RedisRateLimiter implements RateLimiter using Redis
type RedisRateLimiter struct {
	client *redis.Client
	config *config.Config
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(cfg *config.Config) (*RedisRateLimiter, error) {
	if !cfg.Redis.Enabled {
		return nil, fmt.Errorf("Redis is not enabled in configuration")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddress(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisRateLimiter{
		client: client,
		config: cfg,
	}, nil
}

// CheckLimit checks if the request is within the rate limit using sliding window
func (r *RedisRateLimiter) CheckLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d-%d", now.UnixNano(), time.Now().Nanosecond()),
	})

	// Set expiration
	pipe.Expire(ctx, key, window+time.Minute)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit check: %v", err)
	}

	currentCount := int(countCmd.Val())
	allowed := currentCount < limit
	remaining := limit - currentCount - 1
	if remaining < 0 {
		remaining = 0
	}

	resetTime := now.Add(window)
	retryAfter := time.Duration(0)
	if !allowed {
		retryAfter = window
	}

	return &RateLimitResult{
		Allowed:       allowed,
		Limit:         limit,
		Remaining:     remaining,
		ResetTime:     resetTime,
		RetryAfter:    retryAfter,
		TotalRequests: currentCount + 1,
	}, nil
}

// CheckUserLimit checks rate limit for a specific user
func (r *RedisRateLimiter) CheckUserLimit(ctx context.Context, userID uint, limit int, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:user:%d", userID)
	return r.CheckLimit(ctx, key, limit, window)
}

// CheckIPLimit checks rate limit for a specific IP address
func (r *RedisRateLimiter) CheckIPLimit(ctx context.Context, ip string, limit int, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	return r.CheckLimit(ctx, key, limit, window)
}

// CheckEndpointLimit checks rate limit for a specific endpoint
func (r *RedisRateLimiter) CheckEndpointLimit(ctx context.Context, identifier string, endpoint string, limit int, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:endpoint:%s:%s", identifier, endpoint)
	return r.CheckLimit(ctx, key, limit, window)
}

// Reset resets the rate limit for a specific key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// GetStats returns current rate limit statistics
func (r *RedisRateLimiter) GetStats(ctx context.Context, key string) (*RateLimitStats, error) {
	now := time.Now()

	// Get all entries in the current window
	entries, err := r.client.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit stats: %v", err)
	}

	if len(entries) == 0 {
		return &RateLimitStats{
			Key:          key,
			CurrentCount: 0,
			ResetTime:    now,
		}, nil
	}

	firstRequestTime := time.Unix(0, int64(entries[0].Score))
	lastRequestTime := time.Unix(0, int64(entries[len(entries)-1].Score))

	return &RateLimitStats{
		Key:          key,
		CurrentCount: len(entries),
		FirstRequest: firstRequestTime,
		LastRequest:  lastRequestTime,
		ResetTime:    now,
	}, nil
}

// Close closes the Redis connection
func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}

// InMemoryRateLimiter implements RateLimiter using in-memory storage (fallback)
type InMemoryRateLimiter struct {
	requests map[string][]time.Time
	config   *config.Config
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(cfg *config.Config) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		requests: make(map[string][]time.Time),
		config:   cfg,
	}
}

// CheckLimit checks if the request is within the rate limit using in-memory storage
func (m *InMemoryRateLimiter) CheckLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	// Clean up old requests
	if requests, exists := m.requests[key]; exists {
		validRequests := make([]time.Time, 0)
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		m.requests[key] = validRequests
	}

	// Check current count
	currentCount := len(m.requests[key])
	allowed := currentCount < limit

	// Add current request if allowed
	if allowed {
		m.requests[key] = append(m.requests[key], now)
		currentCount++
	}

	remaining := limit - currentCount
	if remaining < 0 {
		remaining = 0
	}

	resetTime := now.Add(window)
	retryAfter := time.Duration(0)
	if !allowed {
		retryAfter = window
	}

	return &RateLimitResult{
		Allowed:       allowed,
		Limit:         limit,
		Remaining:     remaining,
		ResetTime:     resetTime,
		RetryAfter:    retryAfter,
		TotalRequests: currentCount,
	}, nil
}

// CheckUserLimit checks rate limit for a specific user
func (m *InMemoryRateLimiter) CheckUserLimit(ctx context.Context, userID uint, limit int, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("user:%d", userID)
	return m.CheckLimit(ctx, key, limit, window)
}

// CheckIPLimit checks rate limit for a specific IP address
func (m *InMemoryRateLimiter) CheckIPLimit(ctx context.Context, ip string, limit int, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("ip:%s", ip)
	return m.CheckLimit(ctx, key, limit, window)
}

// CheckEndpointLimit checks rate limit for a specific endpoint
func (m *InMemoryRateLimiter) CheckEndpointLimit(ctx context.Context, identifier string, endpoint string, limit int, window time.Duration) (*RateLimitResult, error) {
	key := fmt.Sprintf("endpoint:%s:%s", identifier, endpoint)
	return m.CheckLimit(ctx, key, limit, window)
}

// Reset resets the rate limit for a specific key
func (m *InMemoryRateLimiter) Reset(ctx context.Context, key string) error {
	delete(m.requests, key)
	return nil
}

// GetStats returns current rate limit statistics
func (m *InMemoryRateLimiter) GetStats(ctx context.Context, key string) (*RateLimitStats, error) {
	requests, exists := m.requests[key]
	if !exists || len(requests) == 0 {
		return &RateLimitStats{
			Key:          key,
			CurrentCount: 0,
			ResetTime:    time.Now(),
		}, nil
	}

	return &RateLimitStats{
		Key:          key,
		CurrentCount: len(requests),
		FirstRequest: requests[0],
		LastRequest:  requests[len(requests)-1],
		ResetTime:    time.Now(),
	}, nil
}

// NewRateLimiter creates a new rate limiter based on configuration
func NewRateLimiter(cfg *config.Config) (RateLimiter, error) {
	if cfg.Redis.Enabled {
		return NewRedisRateLimiter(cfg)
	}
	return NewInMemoryRateLimiter(cfg), nil
}
