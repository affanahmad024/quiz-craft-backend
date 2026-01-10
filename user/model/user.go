package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserDAO represents user data from database (uses sql.Null* for nullable fields)
type UserDAO struct {
	ID            uuid.UUID
	Email         string
	Name          string
	PhoneNumber   sql.NullString
	Role          string // Phase 7: user role (user, admin)
	EmailVerified bool
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
	LastLoginAt   sql.NullTime // Phase 7: track last login
	Meta          []byte       // JSONB stored as bytes
}

// UserAPI represents user data for API responses (native Go types)
type UserAPI struct {
	ID            uuid.UUID              `json:"id"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	PhoneNumber   string                 `json:"phone_number,omitempty"`
	Role          string                 `json:"role"`               // Phase 7: user role
	EmailVerified bool                   `json:"email_verified"`
	Enabled       bool                   `json:"enabled"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	LastLoginAt   *string                `json:"last_login_at,omitempty"` // Phase 7: last login
	Meta          map[string]interface{} `json:"meta,omitempty"`
}

// ToAPI converts UserDAO to UserAPI
func (u *UserDAO) ToAPI() UserAPI {
	api := UserAPI{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
		Enabled:       u.Enabled,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     u.UpdatedAt.Format(time.RFC3339),
	}

	if u.PhoneNumber.Valid {
		api.PhoneNumber = u.PhoneNumber.String
	}

	if u.LastLoginAt.Valid {
		lastLogin := u.LastLoginAt.Time.Format(time.RFC3339)
		api.LastLoginAt = &lastLogin
	}

	// Parse JSONB meta
	if len(u.Meta) > 0 {
		json.Unmarshal(u.Meta, &api.Meta)
	}

	return api
}

// SignupRequest represents user signup request
type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Name        string `json:"name" binding:"required,min=2,max=255"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// SignupResponse represents signup response
type SignupResponse struct {
	UserID                 uuid.UUID `json:"user_id"`
	Email                  string    `json:"email"`
	EmailVerificationSent  bool      `json:"email_verification_sent"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"`
	TokenType    string  `json:"token_type"`
	User         UserAPI `json:"user"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name        *string                 `json:"name,omitempty"`
	PhoneNumber *string                 `json:"phone_number,omitempty"`
	Meta        *map[string]interface{} `json:"meta,omitempty"`
}

// UserListFilter represents filters for user listing
type UserListFilter struct {
	Page           int
	Limit          int
	EmailVerified  *bool
	Enabled        *bool
	IncludeDeleted bool
}

// UserListResponse represents paginated user list
type UserListResponse struct {
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
	Users []UserAPI `json:"users"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	TotalUsers   int `json:"total_users"`
	ActiveUsers  int `json:"active_users"`
	DeletedUsers int `json:"deleted_users"`
}
