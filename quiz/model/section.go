package model

import "time"

// QuizSection represents a creator-defined section within a quiz.
type QuizSection struct {
	ID               string     `json:"id"`
	QuizID           string     `json:"quiz_id"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	Instructions     string     `json:"instructions,omitempty"`
	OrderIndex       int        `json:"order_index"`
	TimeLimitMinutes *int       `json:"time_limit_minutes,omitempty"`
	IsDefault        bool       `json:"is_default"`
	Meta             JSONMap    `json:"meta,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Questions        []Question `json:"questions,omitempty"`
}

// QuizSectionInput is used to create a section.
type QuizSectionInput struct {
	Title            string  `json:"title" binding:"required"`
	Description      string  `json:"description,omitempty"`
	Instructions     string  `json:"instructions,omitempty"`
	TimeLimitMinutes *int    `json:"time_limit_minutes,omitempty"`
	Meta             JSONMap `json:"meta,omitempty"`
}

// QuizSectionUpdateInput is used to update section details.
type QuizSectionUpdateInput struct {
	Title            *string `json:"title,omitempty"`
	Description      *string `json:"description,omitempty"`
	Instructions     *string `json:"instructions,omitempty"`
	TimeLimitMinutes *int    `json:"time_limit_minutes,omitempty"`
	Meta             JSONMap `json:"meta,omitempty"`
}

// SectionReorderInput represents section ordering updates.
type SectionReorderInput struct {
	Sections []SectionReorderItem `json:"sections" binding:"required"`
}

type SectionReorderItem struct {
	SectionID  string `json:"section_id" binding:"required"`
	OrderIndex int    `json:"order_index"`
}

// QuestionSectionReorderInput represents question movement/order across sections.
type QuestionSectionReorderInput struct {
	Questions []QuestionSectionReorderItem `json:"questions" binding:"required"`
}

type QuestionSectionReorderItem struct {
	QuestionID        string `json:"question_id" binding:"required"`
	SectionID         string `json:"section_id" binding:"required"`
	SectionOrderIndex int    `json:"section_order_index"`
	OrderIndex        int    `json:"order_index"`
}

// QuestionSectionUpdateInput moves one question to a section.
type QuestionSectionUpdateInput struct {
	SectionID string `json:"section_id" binding:"required"`
}
