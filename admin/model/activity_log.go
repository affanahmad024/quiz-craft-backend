package model

import "time"

// ActivityLog represents a system activity for audit trail
type ActivityLog struct {
	ID           string                 `json:"id" db:"id"`
	UserID       *string                `json:"user_id,omitempty" db:"user_id"`
	Action       string                 `json:"action" db:"action"`
	ResourceType *string                `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty" db:"resource_id"`
	Details      map[string]interface{} `json:"details" db:"details"`
	IPAddress    *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// Activity action constants
const (
	// User actions
	ActionUserCreated     = "user.created"
	ActionUserUpdated     = "user.updated"
	ActionUserDeleted     = "user.deleted"
	ActionUserDisabled    = "user.disabled"
	ActionUserEnabled     = "user.enabled"
	ActionUserRoleChanged = "user.role_changed"

	// Quiz actions
	ActionQuizCreated     = "quiz.created"
	ActionQuizPublished   = "quiz.published"
	ActionQuizUnpublished = "quiz.unpublished"
	ActionQuizDeleted     = "quiz.deleted"

	// Attempt actions
	ActionAttemptStarted   = "attempt.started"
	ActionAttemptCompleted = "attempt.completed"

	// Admin actions
	ActionAdminLogin  = "admin.login"
	ActionAdminAccess = "admin.access"
)

// ActivityLogFilter represents filters for querying activity logs
type ActivityLogFilter struct {
	UserID       *string
	Action       *string
	ResourceType *string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	Limit        int
}
