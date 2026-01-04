package constants

const (
	// Date/Time formats
	RespDtFmt = "2006-01-02 15:04:05"

	// Token expiry
	AccessTokenExpiry  = 24  // hours
	RefreshTokenExpiry = 168 // hours (7 days)

	// Password constants
	MinPasswordLength     = 8
	MaxPasswordLength     = 128
	PasswordHistoryCount  = 5
	MaxFailedAttempts     = 5
	AccountLockoutMinutes = 30

	// Verification token expiry
	EmailVerificationExpiry = 24 // hours
	PasswordResetExpiry     = 1  // hours

	// Pagination defaults
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// User roles
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Token types
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)
