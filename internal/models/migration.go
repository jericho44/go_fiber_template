package models

import (
	"time"
)

// Migration represents a database migration record
type Migration struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Version   string    `json:"version" gorm:"unique;not null;size:255" validate:"required"`
	AppliedAt time.Time `json:"applied_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for the Migration model
func (Migration) TableName() string {
	return "migrations"
}
