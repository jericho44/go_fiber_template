package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
)

// AnomalyDetector defines the interface for detecting suspicious token usage
type AnomalyDetector interface {
	// Analyze token usage for anomalies
	AnalyzeTokenUsage(ctx context.Context, usage *models.TokenUsage) (*AnomalyResult, error)

	// Check for suspicious patterns
	CheckIPAnomalies(ctx context.Context, userID uint, ipAddress string) (*AnomalyResult, error)
	CheckUserAgentAnomalies(ctx context.Context, userID uint, userAgent string) (*AnomalyResult, error)
	CheckLocationAnomalies(ctx context.Context, userID uint, location string) (*AnomalyResult, error)
	CheckFrequencyAnomalies(ctx context.Context, sessionID string) (*AnomalyResult, error)

	// Device fingerprint analysis
	CheckDeviceFingerprintAnomalies(ctx context.Context, userID uint, fingerprint string) (*AnomalyResult, error)

	// Time-based analysis
	CheckTimeBasedAnomalies(ctx context.Context, userID uint, timestamp time.Time) (*AnomalyResult, error)
}

// AnomalyResult represents the result of anomaly detection
type AnomalyResult struct {
	IsAnomaly       bool                   `json:"is_anomaly"`
	Confidence      float64                `json:"confidence"`   // 0.0 to 1.0
	AnomalyType     string                 `json:"anomaly_type"` // ip, user_agent, location, frequency, device, time
	Description     string                 `json:"description"`
	Severity        AnomalySeverity        `json:"severity"`
	Metadata        map[string]interface{} `json:"metadata"`
	Timestamp       time.Time              `json:"timestamp"`
	Recommendations []string               `json:"recommendations"`
}

// AnomalySeverity represents the severity of an anomaly
type AnomalySeverity string

const (
	SeverityLow      AnomalySeverity = "low"
	SeverityMedium   AnomalySeverity = "medium"
	SeverityHigh     AnomalySeverity = "high"
	SeverityCritical AnomalySeverity = "critical"
)

// anomalyDetectorImpl implements AnomalyDetector
type anomalyDetectorImpl struct {
	tokenSessionRepo repositories.TokenSessionRepository
	redisService     RedisService
	config           AnomalyDetectorConfig
}

// AnomalyDetectorConfig holds configuration for anomaly detection
type AnomalyDetectorConfig struct {
	// IP-based detection
	MaxIPChangesPerHour int           `json:"max_ip_changes_per_hour"`
	IPChangeWindow      time.Duration `json:"ip_change_window"`

	// User agent detection
	MaxUserAgentChangesPerDay int           `json:"max_user_agent_changes_per_day"`
	UserAgentChangeWindow     time.Duration `json:"user_agent_change_window"`

	// Frequency detection
	MaxRequestsPerMinute int           `json:"max_requests_per_minute"`
	FrequencyWindow      time.Duration `json:"frequency_window"`

	// Device fingerprint detection
	MaxDevicesPerUser int `json:"max_devices_per_user"`

	// Time-based detection
	AllowedTimeZones         []string `json:"allowed_time_zones"`
	MaxTimeZoneChangesPerDay int      `json:"max_timezone_changes_per_day"`

	// Location detection
	MaxLocationChangesPerHour int           `json:"max_location_changes_per_hour"`
	LocationChangeWindow      time.Duration `json:"location_change_window"`
}

// DefaultAnomalyDetectorConfig returns default configuration
func DefaultAnomalyDetectorConfig() AnomalyDetectorConfig {
	return AnomalyDetectorConfig{
		MaxIPChangesPerHour:       5,
		IPChangeWindow:            time.Hour,
		MaxUserAgentChangesPerDay: 3,
		UserAgentChangeWindow:     24 * time.Hour,
		MaxRequestsPerMinute:      60,
		FrequencyWindow:           time.Minute,
		MaxDevicesPerUser:         10,
		AllowedTimeZones:          []string{"UTC", "America/New_York", "Europe/London", "Asia/Tokyo"},
		MaxTimeZoneChangesPerDay:  2,
		MaxLocationChangesPerHour: 3,
		LocationChangeWindow:      time.Hour,
	}
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(
	tokenSessionRepo repositories.TokenSessionRepository,
	redisService RedisService,
	config AnomalyDetectorConfig,
) AnomalyDetector {
	return &anomalyDetectorImpl{
		tokenSessionRepo: tokenSessionRepo,
		redisService:     redisService,
		config:           config,
	}
}

// AnalyzeTokenUsage performs comprehensive anomaly analysis on token usage
func (a *anomalyDetectorImpl) AnalyzeTokenUsage(ctx context.Context, usage *models.TokenUsage) (*AnomalyResult, error) {
	// Get session to get user ID
	session, err := a.tokenSessionRepo.GetSessionByID(ctx, usage.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Run multiple anomaly checks
	checks := []func(context.Context, uint, *models.TokenUsage) (*AnomalyResult, error){
		a.checkIPAnomaliesForUsage,
		a.checkUserAgentAnomaliesForUsage,
		a.checkFrequencyAnomaliesForUsage,
		a.checkTimeBasedAnomaliesForUsage,
	}

	var highestSeverityResult *AnomalyResult
	var maxConfidence float64

	for _, check := range checks {
		result, err := check(ctx, session.UserID, usage)
		if err != nil {
			continue // Log error in production
		}

		if result.IsAnomaly && result.Confidence > maxConfidence {
			maxConfidence = result.Confidence
			highestSeverityResult = result
		}
	}

	if highestSeverityResult != nil {
		return highestSeverityResult, nil
	}

	// No anomalies detected
	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "none",
		Description: "No anomalies detected",
		Severity:    SeverityLow,
		Timestamp:   time.Now(),
	}, nil
}

// CheckIPAnomalies checks for IP-based anomalies
func (a *anomalyDetectorImpl) CheckIPAnomalies(ctx context.Context, userID uint, ipAddress string) (*AnomalyResult, error) {
	// Count recent IP changes
	key := fmt.Sprintf("anomaly:ip_changes:user:%d", userID)

	// Check if this IP is new for this user
	ipKey := fmt.Sprintf("anomaly:user_ips:user:%d:ip:%s", userID, ipAddress)
	exists, err := a.redisService.Exists(ctx, ipKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check IP existence: %w", err)
	}

	if exists == 0 {
		// New IP, increment counter
		newCount, err := a.redisService.IncrementCounter(ctx, key, a.config.IPChangeWindow)
		if err != nil {
			return nil, fmt.Errorf("failed to increment IP change counter: %w", err)
		}

		// Remember this IP for the user
		err = a.redisService.Set(ctx, ipKey, "seen", 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to record IP: %w", err)
		}

		if newCount > int64(a.config.MaxIPChangesPerHour) {
			return &AnomalyResult{
				IsAnomaly:   true,
				Confidence:  0.8,
				AnomalyType: "ip",
				Description: fmt.Sprintf("Too many IP changes: %d in the last hour", newCount),
				Severity:    SeverityHigh,
				Metadata: map[string]interface{}{
					"ip_address":   ipAddress,
					"change_count": newCount,
					"max_allowed":  a.config.MaxIPChangesPerHour,
				},
				Timestamp: time.Now(),
				Recommendations: []string{
					"Verify user identity",
					"Consider requiring additional authentication",
					"Monitor for account compromise",
				},
			}, nil
		}
	}

	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "ip",
		Timestamp:   time.Now(),
	}, nil
}

// CheckUserAgentAnomalies checks for user agent anomalies
func (a *anomalyDetectorImpl) CheckUserAgentAnomalies(ctx context.Context, userID uint, userAgent string) (*AnomalyResult, error) {
	// Normalize user agent (remove version numbers for basic comparison)
	normalizedUA := a.normalizeUserAgent(userAgent)

	key := fmt.Sprintf("anomaly:ua_changes:user:%d", userID)
	uaKey := fmt.Sprintf("anomaly:user_agents:user:%d:ua:%s", userID, normalizedUA)

	exists, err := a.redisService.Exists(ctx, uaKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check user agent existence: %w", err)
	}

	if exists == 0 {
		// New user agent
		newCount, err := a.redisService.IncrementCounter(ctx, key, a.config.UserAgentChangeWindow)
		if err != nil {
			return nil, fmt.Errorf("failed to increment user agent counter: %w", err)
		}

		// Remember this user agent
		err = a.redisService.Set(ctx, uaKey, "seen", 7*24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to record user agent: %w", err)
		}

		if newCount > int64(a.config.MaxUserAgentChangesPerDay) {
			return &AnomalyResult{
				IsAnomaly:   true,
				Confidence:  0.6,
				AnomalyType: "user_agent",
				Description: fmt.Sprintf("Too many user agent changes: %d in the last day", newCount),
				Severity:    SeverityMedium,
				Metadata: map[string]interface{}{
					"user_agent":   userAgent,
					"change_count": newCount,
					"max_allowed":  a.config.MaxUserAgentChangesPerDay,
				},
				Timestamp: time.Now(),
				Recommendations: []string{
					"Check for automated access",
					"Verify legitimate browser usage",
				},
			}, nil
		}
	}

	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "user_agent",
		Timestamp:   time.Now(),
	}, nil
}

// CheckLocationAnomalies checks for location-based anomalies
func (a *anomalyDetectorImpl) CheckLocationAnomalies(ctx context.Context, userID uint, location string) (*AnomalyResult, error) {
	if location == "" {
		return &AnomalyResult{IsAnomaly: false, Timestamp: time.Now()}, nil
	}

	key := fmt.Sprintf("anomaly:location_changes:user:%d", userID)
	locationKey := fmt.Sprintf("anomaly:user_locations:user:%d:location:%s", userID, location)

	exists, err := a.redisService.Exists(ctx, locationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check location existence: %w", err)
	}

	if exists == 0 {
		// New location
		newCount, err := a.redisService.IncrementCounter(ctx, key, a.config.LocationChangeWindow)
		if err != nil {
			return nil, fmt.Errorf("failed to increment location counter: %w", err)
		}

		// Remember this location
		err = a.redisService.Set(ctx, locationKey, "seen", 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to record location: %w", err)
		}

		if newCount > int64(a.config.MaxLocationChangesPerHour) {
			return &AnomalyResult{
				IsAnomaly:   true,
				Confidence:  0.9,
				AnomalyType: "location",
				Description: fmt.Sprintf("Too many location changes: %d in the last hour", newCount),
				Severity:    SeverityCritical,
				Metadata: map[string]interface{}{
					"location":     location,
					"change_count": newCount,
					"max_allowed":  a.config.MaxLocationChangesPerHour,
				},
				Timestamp: time.Now(),
				Recommendations: []string{
					"Immediately verify user identity",
					"Consider account lockdown",
					"Investigate potential account compromise",
				},
			}, nil
		}
	}

	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "location",
		Timestamp:   time.Now(),
	}, nil
}

// CheckFrequencyAnomalies checks for request frequency anomalies
func (a *anomalyDetectorImpl) CheckFrequencyAnomalies(ctx context.Context, sessionID string) (*AnomalyResult, error) {
	key := fmt.Sprintf("anomaly:frequency:session:%s", sessionID)

	count, err := a.redisService.IncrementCounter(ctx, key, a.config.FrequencyWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to increment frequency counter: %w", err)
	}

	if count > int64(a.config.MaxRequestsPerMinute) {
		return &AnomalyResult{
			IsAnomaly:   true,
			Confidence:  0.7,
			AnomalyType: "frequency",
			Description: fmt.Sprintf("Too many requests: %d in the last minute", count),
			Severity:    SeverityMedium,
			Metadata: map[string]interface{}{
				"session_id":    sessionID,
				"request_count": count,
				"max_allowed":   a.config.MaxRequestsPerMinute,
			},
			Timestamp: time.Now(),
			Recommendations: []string{
				"Check for automated access",
				"Consider rate limiting",
			},
		}, nil
	}

	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "frequency",
		Timestamp:   time.Now(),
	}, nil
}

// CheckDeviceFingerprintAnomalies checks for device fingerprint anomalies
func (a *anomalyDetectorImpl) CheckDeviceFingerprintAnomalies(ctx context.Context, userID uint, fingerprint string) (*AnomalyResult, error) {
	count, err := a.tokenSessionRepo.CountSessionsByFingerprint(ctx, userID, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions by fingerprint: %w", err)
	}

	if count > int64(a.config.MaxDevicesPerUser) {
		return &AnomalyResult{
			IsAnomaly:   true,
			Confidence:  0.5,
			AnomalyType: "device",
			Description: fmt.Sprintf("Too many devices: %d active sessions", count),
			Severity:    SeverityMedium,
			Metadata: map[string]interface{}{
				"device_fingerprint": fingerprint,
				"session_count":      count,
				"max_allowed":        a.config.MaxDevicesPerUser,
			},
			Timestamp: time.Now(),
			Recommendations: []string{
				"Review active sessions",
				"Consider device management",
			},
		}, nil
	}

	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "device",
		Timestamp:   time.Now(),
	}, nil
}

// CheckTimeBasedAnomalies checks for time-based anomalies
func (a *anomalyDetectorImpl) CheckTimeBasedAnomalies(ctx context.Context, userID uint, timestamp time.Time) (*AnomalyResult, error) {
	// Check for unusual access times (e.g., 3 AM local time)
	hour := timestamp.Hour()

	// Consider 2 AM to 5 AM as unusual hours
	if hour >= 2 && hour <= 5 {
		return &AnomalyResult{
			IsAnomaly:   true,
			Confidence:  0.3,
			AnomalyType: "time",
			Description: fmt.Sprintf("Unusual access time: %02d:00", hour),
			Severity:    SeverityLow,
			Metadata: map[string]interface{}{
				"access_hour": hour,
				"timestamp":   timestamp,
			},
			Timestamp: time.Now(),
			Recommendations: []string{
				"Monitor for unusual patterns",
			},
		}, nil
	}

	return &AnomalyResult{
		IsAnomaly:   false,
		Confidence:  0.0,
		AnomalyType: "time",
		Timestamp:   time.Now(),
	}, nil
}

// Helper methods for internal use

func (a *anomalyDetectorImpl) checkIPAnomaliesForUsage(ctx context.Context, userID uint, usage *models.TokenUsage) (*AnomalyResult, error) {
	return a.CheckIPAnomalies(ctx, userID, usage.IPAddress)
}

func (a *anomalyDetectorImpl) checkUserAgentAnomaliesForUsage(ctx context.Context, userID uint, usage *models.TokenUsage) (*AnomalyResult, error) {
	return a.CheckUserAgentAnomalies(ctx, userID, usage.UserAgent)
}

func (a *anomalyDetectorImpl) checkFrequencyAnomaliesForUsage(ctx context.Context, userID uint, usage *models.TokenUsage) (*AnomalyResult, error) {
	return a.CheckFrequencyAnomalies(ctx, usage.SessionID)
}

func (a *anomalyDetectorImpl) checkTimeBasedAnomaliesForUsage(ctx context.Context, userID uint, usage *models.TokenUsage) (*AnomalyResult, error) {
	return a.CheckTimeBasedAnomalies(ctx, userID, usage.UsedAt)
}

// normalizeUserAgent removes version numbers and normalizes user agent strings
func (a *anomalyDetectorImpl) normalizeUserAgent(userAgent string) string {
	// Simple normalization - in production, you'd want more sophisticated parsing
	ua := strings.ToLower(userAgent)

	// Remove version numbers (basic approach)
	if strings.Contains(ua, "chrome") {
		return "chrome"
	}
	if strings.Contains(ua, "firefox") {
		return "firefox"
	}
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "safari"
	}
	if strings.Contains(ua, "edge") {
		return "edge"
	}

	return "unknown"
}
