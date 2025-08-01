package models

import (
	"time"
)

// PasswordHistory represents a user's password history entry
type PasswordHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	PasswordHash string    `json:"-" gorm:"not null;size:255"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName returns the table name for the PasswordHistory model
func (PasswordHistory) TableName() string {
	return "password_history"
}
