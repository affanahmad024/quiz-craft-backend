package model

import "time"

// QuizStatistics represents detailed statistics for a quiz
type QuizStatistics struct {
	QuizID     string `json:"quiz_id"`
	Title      string `json:"title"`
	CreatorID  string `json:"creator_id"`

	// Attempt statistics
	TotalAttempts      int `json:"total_attempts"`
	UniqueUsers        int `json:"unique_users"`
	CompletedAttempts  int `json:"completed_attempts"`
	InProgressAttempts int `json:"in_progress_attempts"`
	AbandonedAttempts  int `json:"abandoned_attempts"`
	ExpiredAttempts    int `json:"expired_attempts"`

	// Score statistics
	AverageScore       float64  `json:"average_score"`
	HighestScore       float64  `json:"highest_score"`
	LowestScore        float64  `json:"lowest_score"`
	MedianScore        *float64 `json:"median_score,omitempty"`
	PassRate           float64  `json:"pass_rate"` // Percentage

	// Time statistics
	AverageTimeSeconds *float64 `json:"average_time_seconds,omitempty"`
	FastestTimeSeconds *float64 `json:"fastest_time_seconds,omitempty"`

	// Engagement statistics
	ViewCount  int `json:"view_count"`
	ShareCount int `json:"share_count"`

	// Question-level stats (optional)
	QuestionStats []QuestionStat `json:"question_stats,omitempty"`
}

// QuestionStat represents statistics for a single question
type QuestionStat struct {
	QuestionID       string  `json:"question_id"`
	QuestionText     string  `json:"question_text"`
	TimesAnswered    int     `json:"times_answered"`
	CorrectAnswers   int     `json:"correct_answers"`
	IncorrectAnswers int     `json:"incorrect_answers"`
	SuccessRate      float64 `json:"success_rate"` // Percentage
}

// UserStatistics represents aggregate statistics for a user
type UserStatistics struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`

	// Quiz creation stats
	QuizzesCreated    int `json:"quizzes_created"`
	PublishedQuizzes  int `json:"published_quizzes"`
	TotalQuizViews    int `json:"total_quiz_views"`
	TotalQuizAttempts int `json:"total_quiz_attempts"`

	// Quiz attempt stats
	QuizzesTaken      int     `json:"quizzes_taken"`
	AttemptsCompleted int     `json:"attempts_completed"`
	AverageScore      float64 `json:"average_score"`
	TotalPointsEarned float64 `json:"total_points_earned"`
	QuizzesPassed     int     `json:"quizzes_passed"`
	PassRate          float64 `json:"pass_rate"` // Percentage

	// Streak and engagement
	CurrentStreak  int       `json:"current_streak"`  // Days
	LongestStreak  int       `json:"longest_streak"`  // Days
	LastActivityAt time.Time `json:"last_activity_at"`
	JoinedAt       time.Time `json:"joined_at"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Category      string  `json:"category"`
	QuizCount     int     `json:"quiz_count"`
	TotalAttempts int     `json:"total_attempts"`
	AverageScore  float64 `json:"average_score"`
}

// TagCount represents a tag with its usage count
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// SearchQuizzesInput defines search/filter parameters
type SearchQuizzesInput struct {
	// Search
	Query string `form:"q"`

	// Filters
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Tags       string `form:"tags"` // Comma-separated
	CreatorID  string `form:"creator_id"`
	IsTemplate *bool  `form:"is_template"`

	// Sorting
	SortBy    string `form:"sort_by"`    // newest, popular, most_attempted
	SortOrder string `form:"sort_order"` // asc, desc

	// Pagination
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// SearchQuizzesResponse represents search results
type SearchQuizzesResponse struct {
	Quizzes    []QuizSummary `json:"quizzes"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
	Facets     SearchFacets  `json:"facets"`
}

// SearchFacets provides counts for filter options
type SearchFacets struct {
	Categories   []FacetCount `json:"categories"`
	Difficulties []FacetCount `json:"difficulties"`
	Tags         []FacetCount `json:"tags"`
}

// FacetCount represents a facet value with count
type FacetCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// QuizSummary is a lightweight quiz representation for search results
type QuizSummary struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Description      *string     `json:"description,omitempty"`
	Category         *string     `json:"category,omitempty"`
	Difficulty       *string     `json:"difficulty,omitempty"`
	QuestionCount    int         `json:"question_count"`
	AttemptCount     int         `json:"attempt_count"`
	AverageScore     float64     `json:"average_score"`
	Creator          UserSummary `json:"creator"`
	Tags             []string    `json:"tags"`
	ThumbnailURL     *string     `json:"thumbnail_url,omitempty"`
	ViewCount        int         `json:"view_count"`
	IsTemplate       bool        `json:"is_template"`
	TemplateUsageCount int       `json:"template_usage_count,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

// UserSummary is a lightweight user representation
type UserSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SaveAsTemplateInput represents input for saving quiz as template
type SaveAsTemplateInput struct {
	TemplateName        string `json:"template_name" binding:"required"`
	TemplateDescription string `json:"template_description"`
	MakePublic          bool   `json:"make_public"`
}

// UseTemplateInput represents input for creating quiz from template
type UseTemplateInput struct {
	Title string `json:"title" binding:"required"`
}

// ShareLinkResponse represents share link information
type ShareLinkResponse struct {
	ShareCode string `json:"share_code"`
	ShareURL  string `json:"share_url"`
}
