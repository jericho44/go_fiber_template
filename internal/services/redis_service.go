package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService provides Redis operations for caching and token management
type RedisService interface {
	// Basic operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Token blacklist operations
	BlacklistToken(ctx context.Context, tokenHash string, expiration time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	BlacklistTokenWithMetadata(ctx context.Context, tokenHash string, metadata map[string]interface{}, expiration time.Duration) error

	// Session operations
	SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error
	GetSession(ctx context.Context, sessionID string, dest interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
	ExtendSessionExpiry(ctx context.Context, sessionID string, expiration time.Duration) error

	// Enhanced token tracking
	TrackTokenUsage(ctx context.Context, tokenHash string, metadata map[string]interface{}) error
	GetTokenUsageStats(ctx context.Context, tokenHash string) (map[string]interface{}, error)

	// Anomaly detection operations
	IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error)
	GetCounter(ctx context.Context, key string) (int64, error)
	SetCounterWithExpiry(ctx context.Context, key string, value int64, expiration time.Duration) error

	// Advanced security operations
	RecordSecurityEvent(ctx context.Context, eventType string, metadata map[string]interface{}, expiration time.Duration) error
	GetSecurityEvents(ctx context.Context, eventType string, limit int) ([]map[string]interface{}, error)

	// Rate limiting enhancements
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error)
	ResetRateLimit(ctx context.Context, key string) error

	// Health check
	Ping(ctx context.Context) error
	Close() error
}

// redisServiceImpl implements RedisService
type redisServiceImpl struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service
func NewRedisService(addr, password string, db int) RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisServiceImpl{
		client: rdb,
	}
}

// Set stores a key-value pair with expiration
func (r *redisServiceImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var data string

	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = string(jsonData)
	}

	err := r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get retrieves a value by key
func (r *redisServiceImpl) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return val, nil
}

// Del deletes one or more keys
func (r *redisServiceImpl) Del(ctx context.Context, keys ...string) error {
	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Exists checks if keys exist
func (r *redisServiceImpl) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count, nil
}

// BlacklistToken adds a token to the blacklist
func (r *redisServiceImpl) BlacklistToken(ctx context.Context, tokenHash string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", tokenHash)
	return r.Set(ctx, key, "blacklisted", expiration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *redisServiceImpl) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", tokenHash)
	count, err := r.Exists(ctx, key)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// SetSession stores session data
func (r *redisServiceImpl) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Set(ctx, key, data, expiration)
}

// GetSession retrieves session data
func (r *redisServiceImpl) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return nil
}

// DeleteSession removes session data
func (r *redisServiceImpl) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Del(ctx, key)
}

// IncrementCounter increments a counter with expiration
func (r *redisServiceImpl) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.Pipeline()

	// Increment the counter
	incrCmd := pipe.Incr(ctx, key)

	// Set expiration if this is the first increment
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter %s: %w", key, err)
	}

	return incrCmd.Val(), nil
}

// GetCounter gets the current counter value
func (r *redisServiceImpl) GetCounter(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Counter doesn't exist, return 0
		}
		return 0, fmt.Errorf("failed to get counter %s: %w", key, err)
	}

	return val, nil
}

// Ping checks Redis connectivity
func (r *redisServiceImpl) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// BlacklistTokenWithMetadata adds a token to the blacklist with metadata
func (r *redisServiceImpl) BlacklistTokenWithMetadata(ctx context.Context, tokenHash string, metadata map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", tokenHash)

	// Store metadata along with blacklist status
	data := map[string]interface{}{
		"status":         "blacklisted",
		"blacklisted_at": time.Now().Unix(),
		"metadata":       metadata,
	}

	return r.Set(ctx, key, data, expiration)
}

// ExtendSessionExpiry extends the expiry time of a session
func (r *redisServiceImpl) ExtendSessionExpiry(ctx context.Context, sessionID string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Expire(ctx, key, expiration).Err()
}

// TrackTokenUsage tracks token usage with metadata
func (r *redisServiceImpl) TrackTokenUsage(ctx context.Context, tokenHash string, metadata map[string]interface{}) error {
	key := fmt.Sprintf("token:usage:%s", tokenHash)

	// Get current usage data
	var usageData map[string]interface{}
	existingData, err := r.Get(ctx, key)
	if err != nil {
		// Initialize new usage data
		usageData = map[string]interface{}{
			"first_used":  time.Now().Unix(),
			"usage_count": 1,
			"last_used":   time.Now().Unix(),
			"metadata":    []map[string]interface{}{metadata},
		}
	} else {
		// Parse existing data
		err = json.Unmarshal([]byte(existingData), &usageData)
		if err != nil {
			return fmt.Errorf("failed to parse existing usage data: %w", err)
		}

		// Update usage data
		usageData["usage_count"] = usageData["usage_count"].(float64) + 1
		usageData["last_used"] = time.Now().Unix()

		// Add new metadata (keep last 10 entries)
		metadataList, ok := usageData["metadata"].([]interface{})
		if !ok {
			metadataList = []interface{}{}
		}

		metadataList = append(metadataList, metadata)
		if len(metadataList) > 10 {
			metadataList = metadataList[len(metadataList)-10:]
		}
		usageData["metadata"] = metadataList
	}

	// Store updated usage data with 24-hour expiry
	return r.Set(ctx, key, usageData, 24*time.Hour)
}

// GetTokenUsageStats retrieves token usage statistics
func (r *redisServiceImpl) GetTokenUsageStats(ctx context.Context, tokenHash string) (map[string]interface{}, error) {
	key := fmt.Sprintf("token:usage:%s", tokenHash)

	data, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var usageStats map[string]interface{}
	err = json.Unmarshal([]byte(data), &usageStats)
	if err != nil {
		return nil, fmt.Errorf("failed to parse usage stats: %w", err)
	}

	return usageStats, nil
}

// SetCounterWithExpiry sets a counter value with expiration
func (r *redisServiceImpl) SetCounterWithExpiry(ctx context.Context, key string, value int64, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set counter %s: %w", key, err)
	}
	return nil
}

// RecordSecurityEvent records a security event with metadata
func (r *redisServiceImpl) RecordSecurityEvent(ctx context.Context, eventType string, metadata map[string]interface{}, expiration time.Duration) error {
	// Create event data
	eventData := map[string]interface{}{
		"type":      eventType,
		"timestamp": time.Now().Unix(),
		"metadata":  metadata,
	}

	// Use a list to store multiple events of the same type
	key := fmt.Sprintf("security:events:%s", eventType)

	// Convert to JSON
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Add to list (left push to keep newest first)
	err = r.client.LPush(ctx, key, string(jsonData)).Err()
	if err != nil {
		return fmt.Errorf("failed to record security event: %w", err)
	}

	// Trim list to keep only last 100 events
	err = r.client.LTrim(ctx, key, 0, 99).Err()
	if err != nil {
		return fmt.Errorf("failed to trim security events list: %w", err)
	}

	// Set expiration on the list
	err = r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration on security events: %w", err)
	}

	return nil
}

// GetSecurityEvents retrieves security events of a specific type
func (r *redisServiceImpl) GetSecurityEvents(ctx context.Context, eventType string, limit int) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("security:events:%s", eventType)

	// Get events from list (0 to limit-1)
	events, err := r.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	var result []map[string]interface{}
	for _, eventStr := range events {
		var eventData map[string]interface{}
		err = json.Unmarshal([]byte(eventStr), &eventData)
		if err != nil {
			continue // Skip malformed events
		}
		result = append(result, eventData)
	}

	return result, nil
}

// CheckRateLimit checks if a rate limit is exceeded
func (r *redisServiceImpl) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	// Use sliding window rate limiting with Redis
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe := r.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})

	// Set expiration
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("failed to check rate limit: %w", err)
	}

	currentCount := countCmd.Val()

	// Check if limit is exceeded
	isExceeded := currentCount >= int64(limit)

	return isExceeded, currentCount, nil
}

// ResetRateLimit resets a rate limit counter
func (r *redisServiceImpl) ResetRateLimit(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *redisServiceImpl) Close() error {
	return r.client.Close()
}
