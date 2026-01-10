package model

import (
	"errors"
	"time"
)

// Answer represents an answer option for a question
type Answer struct {
	ID         string    `json:"id"`
	QuestionID string    `json:"question_id"`
	AnswerText string    `json:"answer_text"`
	IsCorrect  bool      `json:"is_correct"`
	OrderIndex int       `json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate validates answer data
func (a *Answer) Validate() error {
	if a.AnswerText == "" {
		return errors.New("answer text is required")
	}
	if len(a.AnswerText) > 500 {
		return errors.New("answer text must not exceed 500 characters")
	}
	if a.OrderIndex < 0 {
		return errors.New("order index must be non-negative")
	}
	return nil
}

// AnswerListItem represents a simplified answer for list views
// Used when we don't want to expose the correct answer
type AnswerListItem struct {
	ID         string `json:"id"`
	AnswerText string `json:"answer_text"`
	OrderIndex int    `json:"order_index"`
}

// ToListItem converts an Answer to AnswerListItem (hides is_correct)
func (a *Answer) ToListItem() AnswerListItem {
	return AnswerListItem{
		ID:         a.ID,
		AnswerText: a.AnswerText,
		OrderIndex: a.OrderIndex,
	}
}

// AnswerInput represents the input for creating/updating an answer
type AnswerInput struct {
	AnswerText string `json:"answer_text"`
	IsCorrect  bool   `json:"is_correct"`
	OrderIndex int    `json:"order_index"`
}

// Validate validates answer input
func (ai *AnswerInput) Validate() error {
	if ai.AnswerText == "" {
		return errors.New("answer text is required")
	}
	if len(ai.AnswerText) > 500 {
		return errors.New("answer text must not exceed 500 characters")
	}
	if ai.OrderIndex < 0 {
		return errors.New("order index must be non-negative")
	}
	return nil
}

// AnswerReorder represents an answer reordering operation
type AnswerReorder struct {
	AnswerID   string `json:"answer_id"`
	OrderIndex int    `json:"order_index"`
}
