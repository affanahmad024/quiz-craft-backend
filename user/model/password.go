package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserPasswordDAO represents password data from database
type UserPasswordDAO struct {
	ID                      int64
	UserID                  uuid.UUID
	PasswordHash            string
	PasswordHistory         []byte // JSONB
	VerificationToken       sql.NullString
	VerificationTokenExpiry sql.NullTime
	ResetToken              sql.NullString
	ResetTokenExpiry        sql.NullTime
	FailedLoginAttempts     int
	LastFailedLogin         sql.NullTime
	AccountLockedUntil      sql.NullTime
	LastPasswordChange      time.Time
	MustChangePassword      bool
	TwoFactorEnabled        bool
	TwoFactorSecret         sql.NullString
	Meta                    []byte // JSONB
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// PasswordHistory represents password history entry
type PasswordHistory struct {
	Hash      string    `json:"hash"`
	ChangedAt time.Time `json:"changed_at"`
}

// GetPasswordHistory parses password history from JSONB
func (p *UserPasswordDAO) GetPasswordHistory() []PasswordHistory {
	var history []PasswordHistory
	if len(p.PasswordHistory) > 0 {
		json.Unmarshal(p.PasswordHistory, &history)
	}
	return history
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Token  string    `json:"token" binding:"required"`
}

// ResendVerificationRequest represents resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	Token       string    `json:"token" binding:"required"`
	NewPassword string    `json:"new_password" binding:"required,min=8"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
