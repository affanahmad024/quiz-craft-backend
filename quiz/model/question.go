package model

import (
	"errors"
	"time"
)

// Question represents a question in a quiz
type Question struct {
	ID           string    `json:"id"`
	QuizID       string    `json:"quiz_id"`
	QuestionText string    `json:"question_text"`
	QuestionType string    `json:"question_type"` // multiple_choice, true_false, short_answer
	Points       int       `json:"points"`
	OrderIndex   int       `json:"order_index"`
	Explanation  string    `json:"explanation,omitempty"`
	Meta         JSONMap   `json:"meta,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// QuestionWithAnswers represents a question with all its answers
type QuestionWithAnswers struct {
	Question
	Answers []Answer `json:"answers"`
}

// Validate validates question data
func (q *Question) Validate() error {
	if q.QuestionText == "" {
		return errors.New("question text is required")
	}
	if len(q.QuestionText) > 1000 {
		return errors.New("question text must not exceed 1000 characters")
	}
	if q.QuestionType != "multiple_choice" && q.QuestionType != "true_false" && q.QuestionType != "short_answer" {
		return errors.New("question type must be 'multiple_choice', 'true_false', or 'short_answer'")
	}
	if q.Points < 1 || q.Points > 100 {
		return errors.New("points must be between 1 and 100")
	}
	if q.OrderIndex < 0 {
		return errors.New("order index must be non-negative")
	}
	if len(q.Explanation) > 500 {
		return errors.New("explanation must not exceed 500 characters")
	}
	return nil
}

// RequiresAnswers checks if this question type requires answer options
func (q *Question) RequiresAnswers() bool {
	return q.QuestionType == "multiple_choice" || q.QuestionType == "true_false"
}

// MinimumAnswers returns the minimum number of answers required
func (q *Question) MinimumAnswers() int {
	switch q.QuestionType {
	case "multiple_choice":
		return 2 // At least 2 options
	case "true_false":
		return 2 // True and False
	case "short_answer":
		return 0 // No predefined answers
	default:
		return 0
	}
}

// ValidateWithAnswers validates the question along with its answers
func (q *Question) ValidateWithAnswers(answers []Answer) error {
	if err := q.Validate(); err != nil {
		return err
	}

	minAnswers := q.MinimumAnswers()
	if len(answers) < minAnswers {
		return errors.New("insufficient number of answers for this question type")
	}

	// For questions with predefined answers, check answer validity
	if q.RequiresAnswers() {
		for _, answer := range answers {
			if err := answer.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// QuestionReorder represents a question reordering operation
type QuestionReorder struct {
	QuestionID string `json:"question_id"`
	OrderIndex int    `json:"order_index"`
}

// QuestionInput represents input for creating a question
type QuestionInput struct {
	QuestionText string        `json:"question_text"`
	QuestionType string        `json:"question_type"` // multiple_choice, true_false, short_answer
	Points       int           `json:"points"`
	Explanation  string        `json:"explanation,omitempty"`
	Meta         JSONMap       `json:"meta,omitempty"`
	Answers      []AnswerInput `json:"answers,omitempty"` // Optional answers to create with question
}

// QuestionUpdateInput represents input for updating a question
type QuestionUpdateInput struct {
	QuestionText string  `json:"question_text,omitempty"`
	QuestionType string  `json:"question_type,omitempty"`
	Points       *int    `json:"points,omitempty"`
	Explanation  *string `json:"explanation,omitempty"`
	Meta         JSONMap `json:"meta,omitempty"`
}
