package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserPasswordDAO represents password data from database
type UserPasswordDAO struct {
	ID                      uuid.UUID
	UserID                  uuid.UUID
	PasswordHash            string
	PasswordHistory         []byte // JSONB
	VerificationToken       sql.NullString
	VerificationTokenExpiry sql.NullTime
	ResetToken              sql.NullString
	ResetTokenExpiry        sql.NullTime
	TwoFactorEnabled        bool
	TwoFactorSecret         sql.NullString
	Meta                    []byte // JSONB - contains security metadata
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// PasswordMeta represents the security metadata stored in JSONB
type PasswordMeta struct {
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LastFailedLogin     *time.Time `json:"last_failed_login,omitempty"`
	AccountLockedUntil  *time.Time `json:"account_locked_until,omitempty"`
	LastPasswordChange  *time.Time `json:"last_password_change,omitempty"`
	MustChangePassword  bool       `json:"must_change_password"`
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

// GetMeta parses security metadata from JSONB
func (p *UserPasswordDAO) GetMeta() PasswordMeta {
	var meta PasswordMeta
	if len(p.Meta) > 0 {
		json.Unmarshal(p.Meta, &meta)
	}
	return meta
}

// IsAccountLocked checks if the account is currently locked due to failed login attempts
func (p *UserPasswordDAO) IsAccountLocked() bool {
	meta := p.GetMeta()
	if meta.AccountLockedUntil == nil {
		return false
	}
	return time.Now().Before(*meta.AccountLockedUntil)
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
