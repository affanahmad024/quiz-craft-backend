package model

import (
	"errors"
	"time"

	"github.com/lib/pq"
)

// Question represents a question in a quiz
type Question struct {
	ID                string         `json:"id"`
	QuizID            string         `json:"quiz_id"`
	SectionID         string         `json:"section_id"`
	QuestionText      string         `json:"question_text"`
	QuestionType      string         `json:"question_type"` // multiple_choice, true_false, short_answer
	Difficulty        string         `json:"difficulty,omitempty"`
	Tags              pq.StringArray `json:"tags,omitempty"`
	Points            int            `json:"points"`
	OrderIndex        int            `json:"order_index"`
	SectionOrderIndex int            `json:"section_order_index"`
	Explanation       string         `json:"explanation,omitempty"`
	Meta              JSONMap        `json:"meta,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`

	// Answers can be either []Answer (with is_correct) or []AnswerListItem (without is_correct)
	// This field is populated dynamically and not stored in DB
	Answers interface{} `json:"answers,omitempty"`
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
	if q.Difficulty != "" && q.Difficulty != "easy" && q.Difficulty != "medium" && q.Difficulty != "hard" {
		return errors.New("difficulty must be 'easy', 'medium', or 'hard'")
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
		return 2
	case "true_false":
		return 2
	case "short_answer":
		return 0
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
	QuestionType string        `json:"question_type"`
	SectionID    string        `json:"section_id,omitempty"`
	Difficulty   string        `json:"difficulty,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	Points       int           `json:"points"`
	Explanation  string        `json:"explanation,omitempty"`
	Meta         JSONMap       `json:"meta,omitempty"`
	ImageFileID  *string       `json:"image_file_id,omitempty"`
	Answers      []AnswerInput `json:"answers,omitempty"`
}

// BulkQuestionsInput represents input for creating multiple questions at once
type BulkQuestionsInput struct {
	Questions []QuestionInput `json:"questions" binding:"required,min=1"`
}

// QuestionUpdateInput represents input for updating a question
type QuestionUpdateInput struct {
	QuestionText string   `json:"question_text,omitempty"`
	QuestionType string   `json:"question_type,omitempty"`
	SectionID    *string  `json:"section_id,omitempty"`
	Difficulty   *string  `json:"difficulty,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Points       *int     `json:"points,omitempty"`
	Explanation  *string  `json:"explanation,omitempty"`
	Meta         JSONMap  `json:"meta,omitempty"`
	ImageFileID  *string  `json:"image_file_id,omitempty"`
}
