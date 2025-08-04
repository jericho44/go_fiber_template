package models

import (
	"time"

	"gorm.io/gorm"
)

// TokenSession represents an active token session with fingerprinting and tracking
type TokenSession struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	UserID            uint           `json:"user_id" gorm:"not null;index"`
	SessionID         string         `json:"session_id" gorm:"unique;not null;size:255;index"`
	RefreshTokenHash  string         `json:"-" gorm:"unique;not null;size:255;index"`
	DeviceFingerprint string         `json:"device_fingerprint" gorm:"not null;size:255;index"`
	IPAddress         string         `json:"ip_address" gorm:"not null;size:45"`
	UserAgent         string         `json:"user_agent" gorm:"type:text"`
	Location          string         `json:"location" gorm:"size:255"`
	IsActive          bool           `json:"is_active" gorm:"default:true;index"`
	LastUsedAt        time.Time      `json:"last_used_at" gorm:"default:CURRENT_TIMESTAMP;index"`
	ExpiresAt         time.Time      `json:"expires_at" gorm:"not null;index"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User           User            `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	TokenUsages    []TokenUsage    `json:"token_usages" gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
	TokenRotations []TokenRotation `json:"token_rotations" gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}

// TokenUsage tracks individual token usage for anomaly detection
type TokenUsage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"not null;size:255;index"`
	TokenType string    `json:"token_type" gorm:"not null;size:20"` // access, refresh
	IPAddress string    `json:"ip_address" gorm:"not null;size:45"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	Endpoint  string    `json:"endpoint" gorm:"size:255"`
	UsedAt    time.Time `json:"used_at" gorm:"default:CURRENT_TIMESTAMP;index"`
	IsAnomaly bool      `json:"is_anomaly" gorm:"default:false;index"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenRotation tracks token rotation history
type TokenRotation struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SessionID      string    `json:"session_id" gorm:"not null;size:255;index"`
	OldTokenHash   string    `json:"old_token_hash" gorm:"not null;size:255"`
	NewTokenHash   string    `json:"new_token_hash" gorm:"not null;size:255"`
	RotationReason string    `json:"rotation_reason" gorm:"size:100"` // refresh, security, anomaly
	IPAddress      string    `json:"ip_address" gorm:"not null;size:45"`
	UserAgent      string    `json:"user_agent" gorm:"type:text"`
	RotatedAt      time.Time `json:"rotated_at" gorm:"default:CURRENT_TIMESTAMP;index"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName returns the table name for the TokenSession model
func (TokenSession) TableName() string {
	return "token_sessions"
}

// TableName returns the table name for the TokenUsage model
func (TokenUsage) TableName() string {
	return "token_usages"
}

// TableName returns the table name for the TokenRotation model
func (TokenRotation) TableName() string {
	return "token_rotations"
}

// IsExpired checks if the token session has expired
func (ts *TokenSession) IsExpired() bool {
	return time.Now().After(ts.ExpiresAt)
}

// UpdateLastUsed updates the last used timestamp
func (ts *TokenSession) UpdateLastUsed() {
	ts.LastUsedAt = time.Now()
}

// Deactivate marks the session as inactive
func (ts *TokenSession) Deactivate() {
	ts.IsActive = false
}

// GetID returns the session ID
func (ts *TokenSession) GetID() uint {
	return ts.ID
}

// GetSessionID returns the session identifier
func (ts *TokenSession) GetSessionID() string {
	return ts.SessionID
}

// GetUserID returns the user ID
func (ts *TokenSession) GetUserID() uint {
	return ts.UserID
}

// GetDeviceFingerprint returns the device fingerprint
func (ts *TokenSession) GetDeviceFingerprint() string {
	return ts.DeviceFingerprint
}

// IsActive returns whether the session is active
func (ts *TokenSession) GetIsActive() bool {
	return ts.IsActive
}
