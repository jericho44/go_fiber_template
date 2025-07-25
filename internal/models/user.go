package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"unique;not null;size:255" validate:"required,email"`
	Password  string         `json:"-" gorm:"not null;size:255" validate:"required,min=8"`
	FirstName string         `json:"first_name" gorm:"size:100" validate:"required,min=1,max=100"`
	LastName  string         `json:"last_name" gorm:"size:100" validate:"required,min=1,max=100"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Ensure email is lowercase
	u.Email = strings.ToLower(u.Email)
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Ensure email is lowercase
	u.Email = strings.ToLower(u.Email)
	return nil
}

// Interface implementation methods for repository pattern

// GetID returns the user's ID
func (u *User) GetID() uint {
	return u.ID
}

// GetEmail returns the user's email
func (u *User) GetEmail() string {
	return u.Email
}

// SetEmail sets the user's email
func (u *User) SetEmail(email string) {
	u.Email = strings.ToLower(email)
}

// GetFirstName returns the user's first name
func (u *User) GetFirstName() string {
	return u.FirstName
}

// GetLastName returns the user's last name
func (u *User) GetLastName() string {
	return u.LastName
}

// GetIsActive returns whether the user is active
func (u *User) GetIsActive() bool {
	return u.IsActive
}

// SetIsActive sets the user's active status
func (u *User) SetIsActive(active bool) {
	u.IsActive = active
}
