package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Quiz represents a quiz in the system
type Quiz struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Description      string         `json:"description,omitempty"`
	CreatorID        string         `json:"creator_id"`
	Category         string         `json:"category,omitempty"`
	Difficulty       string         `json:"difficulty,omitempty"` // easy, medium, hard
	TimeLimitMinutes *int           `json:"time_limit_minutes,omitempty"`
	PassingScore     int            `json:"passing_score"`
	IsPublished      bool           `json:"is_published"`
	IsPublic         bool           `json:"is_public"`
	ThumbnailURL     string         `json:"thumbnail_url,omitempty"`
	Tags             pq.StringArray `json:"tags,omitempty"`

	// Phase 6: Advanced Features
	ShareCode          *string    `json:"share_code,omitempty"`
	ShareCount         int        `json:"share_count"`
	ViewCount          int        `json:"view_count"`
	IsTemplate         bool       `json:"is_template"`
	TemplateUsageCount int        `json:"template_usage_count"`
	BasedOnTemplateID  *string    `json:"based_on_template_id,omitempty"`

	Meta             JSONMap              `json:"meta,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	DeletedAt        *time.Time           `json:"deleted_at,omitempty"`

	// Latest attempt information (populated only when listing quizzes for a specific user)
	LatestAttempt    *LatestAttemptInfo   `json:"latest_attempt,omitempty"`
}

// LatestAttemptInfo contains summary of the latest quiz attempt
type LatestAttemptInfo struct {
	AttemptID   string     `json:"attempt_id"`
	Status      string     `json:"status"` // in_progress, submitted, expired, abandoned
	StartedAt   time.Time  `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	Score       *float64   `json:"score,omitempty"`
	MaxScore    *int       `json:"max_score,omitempty"`
	Percentage  *float64   `json:"percentage,omitempty"`
	Passed      *bool      `json:"passed,omitempty"`
}

// JSONMap is a custom type for JSONB fields
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONMap: value is not []byte")
	}

	result := make(JSONMap)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// QuizListItem represents a simplified quiz for list views
type QuizListItem struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	CreatorID    string    `json:"creator_id"`
	Category     string    `json:"category,omitempty"`
	Difficulty   string    `json:"difficulty,omitempty"`
	IsPublished  bool      `json:"is_published"`
	IsPublic     bool      `json:"is_public"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	QuestionCount *int     `json:"question_count,omitempty"` // Added via JOIN
}

// QuizWithQuestions represents a quiz with all its questions
type QuizWithQuestions struct {
	Quiz
	Questions []Question `json:"questions"`
}

// Validate validates quiz data
func (q *Quiz) Validate() error {
	if len(q.Title) < 3 {
		return errors.New("title must be at least 3 characters long")
	}
	if len(q.Title) > 255 {
		return errors.New("title must not exceed 255 characters")
	}
	if len(q.Description) > 2000 {
		return errors.New("description must not exceed 2000 characters")
	}
	if q.Difficulty != "" && q.Difficulty != "easy" && q.Difficulty != "medium" && q.Difficulty != "hard" {
		return errors.New("difficulty must be 'easy', 'medium', or 'hard'")
	}
	if q.TimeLimitMinutes != nil && (*q.TimeLimitMinutes < 1 || *q.TimeLimitMinutes > 300) {
		return errors.New("time limit must be between 1 and 300 minutes")
	}
	if q.PassingScore < 0 || q.PassingScore > 100 {
		return errors.New("passing score must be between 0 and 100")
	}
	if q.Category != "" && len(q.Category) > 100 {
		return errors.New("category must not exceed 100 characters")
	}
	return nil
}

// CanPublish checks if a quiz can be published
func (q *Quiz) CanPublish() error {
	if q.IsPublished {
		return errors.New("quiz is already published")
	}
	// Note: We'll check for questions in the business logic layer
	// since this model doesn't have access to questions
	return nil
}

// IsOwnedBy checks if the quiz is owned by the given user ID
func (q *Quiz) IsOwnedBy(userID string) bool {
	return q.CreatorID == userID
}

// IsAccessibleBy checks if a user can access this quiz
func (q *Quiz) IsAccessibleBy(userID string) bool {
	// Owner can always access
	if q.IsOwnedBy(userID) {
		return true
	}
	// Others can access if it's published and public
	return q.IsPublished && q.IsPublic
}

// QuizListFilter represents filters for listing quizzes
type QuizListFilter struct {
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	Category    string `json:"category,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
	IsPublished *bool  `json:"is_published,omitempty"`
	Search      string `json:"search,omitempty"` // Search in title/description
}

// QuizInput represents input for creating a quiz
type QuizInput struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Category         string   `json:"category,omitempty"`
	Difficulty       string   `json:"difficulty,omitempty"`
	TimeLimitMinutes *int     `json:"time_limit_minutes,omitempty"`
	PassingScore     int      `json:"passing_score"`
	IsPublic         bool     `json:"is_public"`
	Tags             []string `json:"tags,omitempty"`
	Meta             JSONMap  `json:"meta,omitempty"`
}

// QuizUpdateInput represents input for updating a quiz
type QuizUpdateInput struct {
	Title            string   `json:"title,omitempty"`
	Description      *string  `json:"description,omitempty"`
	Category         *string  `json:"category,omitempty"`
	Difficulty       *string  `json:"difficulty,omitempty"`
	TimeLimitMinutes *int     `json:"time_limit_minutes,omitempty"`
	PassingScore     *int     `json:"passing_score,omitempty"`
	IsPublic         *bool    `json:"is_public,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Meta             JSONMap  `json:"meta,omitempty"`
}
