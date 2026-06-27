package model

import (
	"errors"
	"time"
)

// QuizAttempt represents a user's attempt to take a quiz
type QuizAttempt struct {
	ID               string     `json:"id"`
	QuizID           string     `json:"quiz_id"`
	UserID           string     `json:"user_id"`
	Status           string     `json:"status"` // in_progress, submitted, expired, abandoned
	StartedAt        time.Time  `json:"started_at"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	TimeLimitSeconds *int       `json:"time_limit_seconds,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	Score            *float64   `json:"score,omitempty"`
	MaxScore         *int       `json:"max_score,omitempty"`
	Percentage       *float64   `json:"percentage,omitempty"`
	Passed           *bool      `json:"passed,omitempty"`
	Meta             JSONMap    `json:"meta,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UserAnswer represents an answer submitted by a user during an attempt
type UserAnswer struct {
	ID           string    `json:"id"`
	AttemptID    string    `json:"attempt_id"`
	QuestionID   string    `json:"question_id"`
	AnswerID     *string   `json:"answer_id,omitempty"`
	AnswerText   *string   `json:"answer_text,omitempty"`
	IsCorrect    *bool     `json:"is_correct,omitempty"`
	PointsEarned float64   `json:"points_earned"`
	AnsweredAt   time.Time `json:"answered_at"`
	Meta         JSONMap   `json:"meta,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AttemptWithAnswers represents an attempt with all user answers
type AttemptWithAnswers struct {
	QuizAttempt
	Answers []UserAnswer `json:"answers"`
}

// AttemptResult represents the result of a quiz attempt with details
type AttemptResult struct {
	QuizAttempt
	QuizTitle           string                  `json:"quiz_title"`
	QuizDescription     string                  `json:"quiz_description"`
	Answers             []UserAnswerWithDetails `json:"answers"`
	DifficultyBreakdown []DifficultyAttemptStat `json:"difficulty_breakdown,omitempty"`
	TotalQuestions      int                     `json:"total_questions"`
	AnsweredCount       int                     `json:"answered_count"`
}

// DifficultyAttemptStat summarizes attempt outcomes grouped by question difficulty.
type DifficultyAttemptStat struct {
	Difficulty    string `json:"difficulty"`
	Total         int    `json:"total"`
	Correct       int    `json:"correct"`
	Wrong         int    `json:"wrong"`
	Unanswered    int    `json:"unanswered"`
	PendingReview int    `json:"pending_review"`
}

// UserAnswerWithDetails represents a user answer with question and correct answer details
type UserAnswerWithDetails struct {
	UserAnswer
	QuestionText       string  `json:"question_text"`
	QuestionType       string  `json:"question_type"`
	QuestionDifficulty string  `json:"question_difficulty,omitempty"`
	QuestionPoints     int     `json:"question_points"`
	QuestionMeta       JSONMap `json:"question_meta,omitempty"`
	Explanation        *string `json:"explanation,omitempty"`
	SelectedAnswerText *string `json:"selected_answer_text,omitempty"`
	CorrectAnswerText  *string `json:"correct_answer_text,omitempty"`
}

// AttemptListItem represents a simplified attempt for list views
type AttemptListItem struct {
	ID          string     `json:"id"`
	QuizID      string     `json:"quiz_id"`
	QuizTitle   string     `json:"quiz_title"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	Score       *float64   `json:"score,omitempty"`
	MaxScore    *int       `json:"max_score,omitempty"`
	Percentage  *float64   `json:"percentage,omitempty"`
	Passed      *bool      `json:"passed,omitempty"`
}

// QuizStats represents statistics for a quiz (creator view)
type QuizStats struct {
	QuizID        string   `json:"quiz_id"`
	TotalAttempts int      `json:"total_attempts"`
	AverageScore  *float64 `json:"average_score,omitempty"`
}

// StartAttemptInput represents input for starting a quiz attempt
type StartAttemptInput struct {
	QuizID string `json:"quiz_id" binding:"required"`
}

// SubmitAnswerInput represents input for submitting an answer
type SubmitAnswerInput struct {
	QuestionID string  `json:"question_id" binding:"required"`
	AnswerID   *string `json:"answer_id,omitempty"`
	AnswerText *string `json:"answer_text,omitempty"`
}

// AttemptListFilter represents filters for listing attempts
type AttemptListFilter struct {
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	QuizID *string `json:"quiz_id,omitempty"`
	Status *string `json:"status,omitempty"`
}

// Validate validates the StartAttemptInput
func (i *StartAttemptInput) Validate() error {
	if i.QuizID == "" {
		return errors.New("quiz_id is required")
	}
	return nil
}

// Validate validates the SubmitAnswerInput
func (i *SubmitAnswerInput) Validate() error {
	if i.QuestionID == "" {
		return errors.New("question_id is required")
	}
	if i.AnswerID != nil && i.AnswerText != nil {
		return errors.New("provide either answer_id or answer_text, not both")
	}
	if i.AnswerID == nil && i.AnswerText == nil {
		return errors.New("either answer_id or answer_text must be provided")
	}
	if i.AnswerText != nil && len(*i.AnswerText) > 2000 {
		return errors.New("answer_text must not exceed 2000 characters")
	}
	return nil
}

// IsInProgress checks if attempt is still in progress
func (a *QuizAttempt) IsInProgress() bool {
	return a.Status == "in_progress"
}

// IsCompleted checks if attempt is completed (submitted or expired)
func (a *QuizAttempt) IsCompleted() bool {
	return a.Status == "submitted" || a.Status == "expired"
}

// IsExpired checks if attempt has expired based on time limit
func (a *QuizAttempt) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// CanSubmitAnswer checks if user can submit answers
func (a *QuizAttempt) CanSubmitAnswer() bool {
	return a.IsInProgress() && !a.IsExpired()
}

// CanFinish checks if attempt can be finished
func (a *QuizAttempt) CanFinish() bool {
	return a.IsInProgress()
}

// TimeRemaining returns the time remaining in seconds (if time limit exists)
func (a *QuizAttempt) TimeRemaining() *int {
	if a.ExpiresAt == nil {
		return nil
	}
	remaining := int(time.Until(*a.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}

// CalculatePercentage calculates the percentage score
func (a *QuizAttempt) CalculatePercentage() float64 {
	if a.MaxScore == nil || *a.MaxScore == 0 {
		return 0
	}
	if a.Score == nil {
		return 0
	}
	return (*a.Score / float64(*a.MaxScore)) * 100
}

// DeterminePassed determines if user passed based on passing score
func (a *QuizAttempt) DeterminePassed(passingScore int) bool {
	percentage := a.CalculatePercentage()
	return percentage >= float64(passingScore)
}
