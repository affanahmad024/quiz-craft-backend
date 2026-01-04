package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserPassword represents user authentication credentials
type UserPassword struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID                  uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	PasswordHash            string     `gorm:"type:varchar(255);not null" json:"-"`
	Salt                    string     `gorm:"type:varchar(255);not null" json:"-"`
	EmailVerificationToken  *string    `gorm:"type:varchar(255);index:idx_user_passwords_email_token,where:email_verification_token IS NOT NULL" json:"-"`
	EmailTokenExpiresAt     *time.Time `json:"-"`
	PasswordResetToken      *string    `gorm:"type:varchar(255);index:idx_user_passwords_reset_token,where:password_reset_token IS NOT NULL" json:"-"`
	ResetTokenExpiresAt     *time.Time `json:"-"`
	TOTPSecret              *string    `gorm:"type:varchar(255)" json:"-"`
	TOTPEnabled             bool       `gorm:"default:false" json:"totp_enabled"`
	BackupCodes             []string   `gorm:"type:text[]" json:"-"`
	LoggedSessions          []string   `gorm:"type:text[];default:'{}'" json:"-"`
	PasswordHistory         PasswordHistoryJSON `gorm:"type:jsonb;default:'[]'" json:"-"`
	CreatedAt               time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Meta                    PasswordMeta `gorm:"type:jsonb;default:'{}'" json:"-"`

	// Associations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// PasswordMeta represents metadata for password management
type PasswordMeta struct {
	LastPasswordChange  *time.Time `json:"last_password_change,omitempty"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LastFailedLogin     *time.Time `json:"last_failed_login,omitempty"`
	AccountLockedUntil  *time.Time `json:"account_locked_until,omitempty"`
	PasswordExpiresAt   *time.Time `json:"password_expires_at,omitempty"`
}

// PasswordHistoryJSON represents password history
type PasswordHistoryJSON []PasswordHistoryEntry

// PasswordHistoryEntry represents a single password history entry
type PasswordHistoryEntry struct {
	PasswordHash string    `json:"password_hash"`
	ChangedAt    time.Time `json:"changed_at"`
}

// TableName specifies the table name for UserPassword model
func (UserPassword) TableName() string {
	return "user_passwords"
}

// BeforeCreate hook to generate UUID if not set
func (up *UserPassword) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}

// IsAccountLocked checks if the account is currently locked
func (up *UserPassword) IsAccountLocked() bool {
	if up.Meta.AccountLockedUntil == nil {
		return false
	}
	return time.Now().Before(*up.Meta.AccountLockedUntil)
}

// IsEmailTokenValid checks if email verification token is valid
func (up *UserPassword) IsEmailTokenValid() bool {
	if up.EmailVerificationToken == nil || up.EmailTokenExpiresAt == nil {
		return false
	}
	return time.Now().Before(*up.EmailTokenExpiresAt)
}

// IsResetTokenValid checks if password reset token is valid
func (up *UserPassword) IsResetTokenValid() bool {
	if up.PasswordResetToken == nil || up.ResetTokenExpiresAt == nil {
		return false
	}
	return time.Now().Before(*up.ResetTokenExpiresAt)
}

// IncrementFailedLoginAttempts increments failed login counter
func (up *UserPassword) IncrementFailedLoginAttempts() {
	up.Meta.FailedLoginAttempts++
	now := time.Now()
	up.Meta.LastFailedLogin = &now

	// Lock account after 5 failed attempts
	if up.Meta.FailedLoginAttempts >= 5 {
		lockUntil := now.Add(30 * time.Minute)
		up.Meta.AccountLockedUntil = &lockUntil
	}
}

// ResetFailedLoginAttempts resets the failed login counter
func (up *UserPassword) ResetFailedLoginAttempts() {
	up.Meta.FailedLoginAttempts = 0
	up.Meta.LastFailedLogin = nil
	up.Meta.AccountLockedUntil = nil
}

// AddPasswordToHistory adds current password to history
func (up *UserPassword) AddPasswordToHistory() {
	entry := PasswordHistoryEntry{
		PasswordHash: up.PasswordHash,
		ChangedAt:    time.Now(),
	}

	// Keep only last 5 passwords in history
	if len(up.PasswordHistory) >= 5 {
		up.PasswordHistory = up.PasswordHistory[1:]
	}

	up.PasswordHistory = append(up.PasswordHistory, entry)
	now := time.Now()
	up.Meta.LastPasswordChange = &now
}

// IsPasswordInHistory checks if password was used before
func (up *UserPassword) IsPasswordInHistory(passwordHash string) bool {
	for _, entry := range up.PasswordHistory {
		if entry.PasswordHash == passwordHash {
			return true
		}
	}
	return up.PasswordHash == passwordHash
}
