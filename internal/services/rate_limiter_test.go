package services

import (
	"context"
	"testing"
	"time"

	"go-fiber-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRateLimiter_CheckLimit(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    3,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		result, err := limiter.CheckLimit(ctx, "test_key", 3, time.Second)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
		assert.Equal(t, 3, result.Limit)
		assert.Equal(t, 3-i-1, result.Remaining)
	}

	// 4th request should be denied
	result, err := limiter.CheckLimit(ctx, "test_key", 3, time.Second)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, 0, result.Remaining)
	assert.Greater(t, result.RetryAfter, time.Duration(0))
}

func TestInMemoryRateLimiter_CheckUserLimit(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    2,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	// Test user rate limiting
	userID := uint(123)

	// First 2 requests should be allowed
	for i := 0; i < 2; i++ {
		result, err := limiter.CheckUserLimit(ctx, userID, 2, time.Second)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
	}

	// 3rd request should be denied
	result, err := limiter.CheckUserLimit(ctx, userID, 2, time.Second)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestInMemoryRateLimiter_CheckIPLimit(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    2,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	ip := "192.168.1.1"

	// First 2 requests should be allowed
	for i := 0; i < 2; i++ {
		result, err := limiter.CheckIPLimit(ctx, ip, 2, time.Second)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
	}

	// 3rd request should be denied
	result, err := limiter.CheckIPLimit(ctx, ip, 2, time.Second)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestInMemoryRateLimiter_CheckEndpointLimit(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    2,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	identifier := "user:123"
	endpoint := "/api/v1/auth/login"

	// First 2 requests should be allowed
	for i := 0; i < 2; i++ {
		result, err := limiter.CheckEndpointLimit(ctx, identifier, endpoint, 2, time.Second)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
	}

	// 3rd request should be denied
	result, err := limiter.CheckEndpointLimit(ctx, identifier, endpoint, 2, time.Second)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestInMemoryRateLimiter_Reset(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    1,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	// Make a request to reach the limit
	result, err := limiter.CheckLimit(ctx, "test_key", 1, time.Second)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Next request should be denied
	result, err = limiter.CheckLimit(ctx, "test_key", 1, time.Second)
	require.NoError(t, err)
	assert.False(t, result.Allowed)

	// Reset the limit
	err = limiter.Reset(ctx, "test_key")
	require.NoError(t, err)

	// Request should now be allowed again
	result, err = limiter.CheckLimit(ctx, "test_key", 1, time.Second)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestInMemoryRateLimiter_GetStats(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    3,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	// Initially no stats
	stats, err := limiter.GetStats(ctx, "test_key")
	require.NoError(t, err)
	assert.Equal(t, 0, stats.CurrentCount)

	// Make some requests
	for i := 0; i < 2; i++ {
		_, err := limiter.CheckLimit(ctx, "test_key", 3, time.Second)
		require.NoError(t, err)
	}

	// Check stats
	stats, err = limiter.GetStats(ctx, "test_key")
	require.NoError(t, err)
	assert.Equal(t, 2, stats.CurrentCount)
	assert.Equal(t, "test_key", stats.Key)
}

func TestInMemoryRateLimiter_WindowExpiry(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    1,
			Window: 100 * time.Millisecond,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	// Make a request to reach the limit
	result, err := limiter.CheckLimit(ctx, "test_key", 1, 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Next request should be denied
	result, err = limiter.CheckLimit(ctx, "test_key", 1, 100*time.Millisecond)
	require.NoError(t, err)
	assert.False(t, result.Allowed)

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Request should now be allowed again
	result, err = limiter.CheckLimit(ctx, "test_key", 1, 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestInMemoryRateLimiter_DifferentKeys(t *testing.T) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    1,
			Window: time.Second,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	// Make requests with different keys
	result1, err := limiter.CheckLimit(ctx, "key1", 1, time.Second)
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	result2, err := limiter.CheckLimit(ctx, "key2", 1, time.Second)
	require.NoError(t, err)
	assert.True(t, result2.Allowed)

	// Both keys should now be at their limit
	result1, err = limiter.CheckLimit(ctx, "key1", 1, time.Second)
	require.NoError(t, err)
	assert.False(t, result1.Allowed)

	result2, err = limiter.CheckLimit(ctx, "key2", 1, time.Second)
	require.NoError(t, err)
	assert.False(t, result2.Allowed)
}

func TestNewRateLimiter_InMemoryFallback(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
		Rate: config.RateConfig{
			Max:    10,
			Window: time.Minute,
		},
	}

	limiter, err := NewRateLimiter(cfg)
	require.NoError(t, err)
	assert.IsType(t, &InMemoryRateLimiter{}, limiter)
}

func TestRateLimitResult_Headers(t *testing.T) {
	result := &RateLimitResult{
		Allowed:       false,
		Limit:         100,
		Remaining:     0,
		ResetTime:     time.Now().Add(time.Minute),
		RetryAfter:    time.Minute,
		TotalRequests: 101,
	}

	assert.False(t, result.Allowed)
	assert.Equal(t, 100, result.Limit)
	assert.Equal(t, 0, result.Remaining)
	assert.Equal(t, time.Minute, result.RetryAfter)
	assert.Equal(t, 101, result.TotalRequests)
}

func TestRateLimitStats_Structure(t *testing.T) {
	now := time.Now()
	stats := &RateLimitStats{
		Key:          "test_key",
		Limit:        100,
		Window:       time.Minute,
		CurrentCount: 50,
		ResetTime:    now,
		FirstRequest: now.Add(-30 * time.Second),
		LastRequest:  now.Add(-5 * time.Second),
	}

	assert.Equal(t, "test_key", stats.Key)
	assert.Equal(t, 100, stats.Limit)
	assert.Equal(t, time.Minute, stats.Window)
	assert.Equal(t, 50, stats.CurrentCount)
	assert.Equal(t, now, stats.ResetTime)
}

// Benchmark tests
func BenchmarkInMemoryRateLimiter_CheckLimit(b *testing.B) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    1000,
			Window: time.Minute,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := limiter.CheckLimit(ctx, "benchmark_key", 1000, time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInMemoryRateLimiter_CheckUserLimit(b *testing.B) {
	cfg := &config.Config{
		Rate: config.RateConfig{
			Max:    1000,
			Window: time.Minute,
		},
	}

	limiter := NewInMemoryRateLimiter(cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := limiter.CheckUserLimit(ctx, uint(i%100), 1000, time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}
}
