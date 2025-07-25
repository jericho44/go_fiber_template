package models

import (
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist represents a blacklisted JWT token
type TokenBlacklist struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	TokenHash     string         `json:"-" gorm:"unique;not null;size:255;index"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	ExpiresAt     time.Time      `json:"expires_at" gorm:"not null;index"`
	BlacklistedAt time.Time      `json:"blacklisted_at" gorm:"default:CURRENT_TIMESTAMP;index"`
	Reason        string         `json:"reason" gorm:"size:100;default:'logout'"`
	User          User           `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the TokenBlacklist model
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// IsExpired checks if the token has expired
func (tb *TokenBlacklist) IsExpired() bool {
	return time.Now().After(tb.ExpiresAt)
}

// GetID returns the token blacklist entry's ID
func (tb *TokenBlacklist) GetID() uint {
	return tb.ID
}

// GetTokenHash returns the token hash
func (tb *TokenBlacklist) GetTokenHash() string {
	return tb.TokenHash
}

// GetUserID returns the user ID
func (tb *TokenBlacklist) GetUserID() uint {
	return tb.UserID
}

// GetExpiresAt returns the expiration time
func (tb *TokenBlacklist) GetExpiresAt() time.Time {
	return tb.ExpiresAt
}

// GetReason returns the blacklist reason
func (tb *TokenBlacklist) GetReason() string {
	return tb.Reason
}
