package model

import (
	"time"

	quizmodel "github.com/affanahmad024/quiz-craft/backend/quiz/model"
)

// PlatformStats represents overall platform statistics
type PlatformStats struct {
	// User statistics
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"` // Last 30 days
	NewUsersToday     int `json:"new_users_today"`
	NewUsersThisWeek  int `json:"new_users_this_week"`
	NewUsersThisMonth int `json:"new_users_this_month"`
	DisabledUsers     int `json:"disabled_users"`

	// Quiz statistics
	TotalQuizzes        int `json:"total_quizzes"`
	PublishedQuizzes    int `json:"published_quizzes"`
	PrivateQuizzes      int `json:"private_quizzes"`
	QuizzesCreatedToday int `json:"quizzes_created_today"`
	TotalQuestions      int `json:"total_questions"`

	// Attempt statistics
	TotalAttempts     int     `json:"total_attempts"`
	CompletedAttempts int     `json:"completed_attempts"`
	AttemptsToday     int     `json:"attempts_today"`
	AverageScore      float64 `json:"average_score"`

	// Storage statistics (placeholders for future implementation)
	TotalUploads  int     `json:"total_uploads"`
	StorageUsedMB float64 `json:"storage_used_mb"`

	// System health
	DatabaseStatus    string     `json:"database_status"`
	APIResponseTimeMs float64    `json:"api_response_time_ms"`
	LastBackupAt      *time.Time `json:"last_backup_at,omitempty"`

	GeneratedAt time.Time `json:"generated_at"`
}

// UserActivity represents user activity over time
type UserActivity struct {
	Date           string `json:"date"`
	NewUsers       int    `json:"new_users"`
	ActiveUsers    int    `json:"active_users"`
	QuizzesCreated int    `json:"quizzes_created"`
	AttemptsCount  int    `json:"attempts_count"`
}

// AdminUserDetail represents detailed user info for admin
type AdminUserDetail struct {
	// Basic info
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	PhoneNumber   *string    `json:"phone_number,omitempty"`
	Role          string     `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	Enabled       bool       `json:"enabled"`

	// Statistics
	QuizzesCreated int     `json:"quizzes_created"`
	QuizzesTaken   int     `json:"quizzes_taken"`
	AverageScore   float64 `json:"average_score"`

	// Metadata
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// AdminQuizDetail represents detailed quiz info for admin
type AdminQuizDetail struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Description *string                  `json:"description,omitempty"`
	Creator     quizmodel.UserSummary    `json:"creator"`
	Category    *string                  `json:"category,omitempty"`
	Difficulty  *string                  `json:"difficulty,omitempty"`
	IsPublished bool                     `json:"is_published"`
	IsPublic    bool                     `json:"is_public"`
	Questions   []*quizmodel.Question    `json:"questions,omitempty"`

	// Statistics
	QuestionCount int     `json:"question_count"`
	AttemptCount  int     `json:"attempt_count"`
	ViewCount     int     `json:"view_count"`
	AverageScore  float64 `json:"average_score"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserListResponse represents paginated user list for admin
type UserListResponse struct {
	Users      []AdminUserDetail `json:"users"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// QuizListResponse represents paginated quiz list for admin
type QuizListResponse struct {
	Quizzes []AdminQuizDetail `json:"quizzes"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
}

// ActivityLogResponse represents paginated activity log list
type ActivityLogResponse struct {
	Logs  []ActivityLog `json:"logs"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}
