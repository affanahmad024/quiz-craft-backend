package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// SessionAnswerDAO represents answer data from database
type SessionAnswerDAO struct {
	ID            uuid.UUID
	SessionID     uuid.UUID
	ParticipantID uuid.UUID
	QuestionID    uuid.UUID
	AnswerID      sql.NullString // UUID as string, NULL if no answer selected
	IsCorrect     bool
	PointsEarned  int
	TimeTakenMs   int
	AnsweredAt    time.Time
	CreatedAt     time.Time
}

// SessionAnswerAPI represents answer data for API responses
type SessionAnswerAPI struct {
	ID            uuid.UUID  `json:"id"`
	SessionID     uuid.UUID  `json:"session_id"`
	ParticipantID uuid.UUID  `json:"participant_id"`
	QuestionID    uuid.UUID  `json:"question_id"`
	AnswerID      *uuid.UUID `json:"answer_id,omitempty"`
	IsCorrect     bool       `json:"is_correct"`
	PointsEarned  int        `json:"points_earned"`
	TimeTakenMs   int        `json:"time_taken_ms"`
	AnsweredAt    string     `json:"answered_at"`
}

// ToAPI converts SessionAnswerDAO to SessionAnswerAPI
func (a *SessionAnswerDAO) ToAPI() SessionAnswerAPI {
	api := SessionAnswerAPI{
		ID:            a.ID,
		SessionID:     a.SessionID,
		ParticipantID: a.ParticipantID,
		QuestionID:    a.QuestionID,
		IsCorrect:     a.IsCorrect,
		PointsEarned:  a.PointsEarned,
		TimeTakenMs:   a.TimeTakenMs,
		AnsweredAt:    a.AnsweredAt.Format(time.RFC3339),
	}

	if a.AnswerID.Valid {
		answerID, _ := uuid.Parse(a.AnswerID.String)
		api.AnswerID = &answerID
	}

	return api
}

// SubmitAnswerRequest represents a request to submit an answer during live session
type SubmitAnswerRequest struct {
	QuestionID  uuid.UUID  `json:"question_id" binding:"required"`
	AnswerID    *uuid.UUID `json:"answer_id,omitempty"`
	TimeTakenMs int        `json:"time_taken_ms" binding:"required"`
}

// SubmitAnswerResponse represents the response after submitting an answer
type SubmitAnswerResponse struct {
	IsCorrect    bool `json:"is_correct"`
	PointsEarned int  `json:"points_earned"`
	CurrentScore int  `json:"current_score"`
	Rank         int  `json:"rank"`
}
