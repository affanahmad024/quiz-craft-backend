package model

import (
	"github.com/google/uuid"
)

// SessionResultsResponse represents the complete results after a session ends
type SessionResultsResponse struct {
	Session          LiveSessionAPI      `json:"session"`
	FinalLeaderboard []LeaderboardEntry  `json:"final_leaderboard"`
	MyPerformance    *ParticipantDetails `json:"my_performance,omitempty"`
	QuestionStats    []QuestionStat      `json:"question_stats,omitempty"`
	TotalParticipants int                `json:"total_participants"`
}

// ParticipantDetails represents detailed performance of a participant
type ParticipantDetails struct {
	Participant ParticipantAPI       `json:"participant"`
	Answers     []ParticipantAnswer  `json:"answers"`
}

// ParticipantAnswer represents a participant's answer with question details
type ParticipantAnswer struct {
	QuestionID    uuid.UUID  `json:"question_id"`
	QuestionText  string     `json:"question_text"`
	AnswerID      *uuid.UUID `json:"answer_id,omitempty"`
	AnswerText    string     `json:"answer_text,omitempty"`
	CorrectAnswer string     `json:"correct_answer"`
	IsCorrect     bool       `json:"is_correct"`
	PointsEarned  int        `json:"points_earned"`
	TimeTakenMs   int        `json:"time_taken_ms"`
}

// QuestionStat represents statistics for a question in the session
type QuestionStat struct {
	QuestionID       uuid.UUID `json:"question_id"`
	QuestionText     string    `json:"question_text"`
	TotalAnswers     int       `json:"total_answers"`
	CorrectAnswers   int       `json:"correct_answers"`
	CorrectRate      float64   `json:"correct_rate"`
	AverageTimeMs    int       `json:"average_time_ms"`
}

// JoinSessionRequest represents a request to join a session
type JoinSessionRequest struct {
	AccessCode string `json:"access_code,omitempty"`
}

// JoinSessionResponse represents the response after joining a session
type JoinSessionResponse struct {
	Message     string         `json:"message"`
	Participant ParticipantAPI `json:"participant"`
	Session     LiveSessionAPI `json:"session"`
}

// SessionListFilter represents filters for listing sessions
type SessionListFilter struct {
	CreatorID   *uuid.UUID
	SessionType *string
	Status      *string
	Limit       int
	Offset      int
}

// SessionListResponse represents paginated session list
type SessionListResponse struct {
	Sessions []LiveSessionAPI `json:"sessions"`
	Total    int              `json:"total"`
}
