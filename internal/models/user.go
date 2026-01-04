package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex:idx_users_email,where:deleted_at IS NULL;not null" json:"email"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber   string         `gorm:"type:varchar(20)" json:"phone_number,omitempty"`
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // Soft delete
	Meta          Meta           `gorm:"type:jsonb;default:'{}'" json:"meta,omitempty"`

	// Associations
	Password *UserPassword `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// Meta represents flexible JSONB metadata for users
type Meta map[string]interface{}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to generate UUID if not set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserResponse is the sanitized response without sensitive data
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	PhoneNumber   string    `json:"phone_number,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Meta          Meta      `json:"meta,omitempty"`
}

// ToResponse converts User to UserResponse (removes sensitive data)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		PhoneNumber:   u.PhoneNumber,
		EmailVerified: u.EmailVerified,
		Enabled:       u.Enabled,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Meta:          u.Meta,
	}
}
