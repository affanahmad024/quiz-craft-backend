package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LiveSessionDAO represents live session data from database
type LiveSessionDAO struct {
	ID                       uuid.UUID
	QuizID                   uuid.UUID
	CreatorID                uuid.UUID
	SessionType              string
	Title                    string
	Description              sql.NullString
	ScheduledAt              sql.NullTime
	StartedAt                sql.NullTime
	EndedAt                  sql.NullTime
	Status                   string
	AccessCode               sql.NullString
	IsDiscoverable           bool
	MaxParticipants          sql.NullInt64
	ProgressionMode          string
	TimerMode                string
	QuestionTimerSeconds     int
	WholeQuizTimerSeconds    sql.NullInt64
	CurrentQuestionIndex     int
	CurrentQuestionStartedAt sql.NullTime
	QuizStartedAt            sql.NullTime
	Config                   []byte // JSONB stored as bytes
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// LiveSessionAPI represents live session data for API responses
type LiveSessionAPI struct {
	ID                       uuid.UUID              `json:"id"`
	QuizID                   uuid.UUID              `json:"quiz_id"`
	CreatorID                uuid.UUID              `json:"creator_id"`
	SessionType              string                 `json:"session_type"`
	Title                    string                 `json:"title"`
	Description              string                 `json:"description,omitempty"`
	ScheduledAt              *string                `json:"scheduled_at,omitempty"`
	StartedAt                *string                `json:"started_at,omitempty"`
	EndedAt                  *string                `json:"ended_at,omitempty"`
	Status                   string                 `json:"status"`
	AccessCode               string                 `json:"access_code,omitempty"`
	IsDiscoverable           bool                   `json:"is_discoverable"`
	MaxParticipants          *int                   `json:"max_participants,omitempty"`
	ProgressionMode          string                 `json:"progression_mode"`
	TimerMode                string                 `json:"timer_mode"`
	QuestionTimerSeconds     int                    `json:"question_timer_seconds,omitempty"`
	WholeQuizTimerSeconds    *int                   `json:"whole_quiz_timer_seconds,omitempty"`
	CurrentQuestionIndex     int                    `json:"current_question_index"`
	CurrentQuestionStartedAt *string                `json:"current_question_started_at,omitempty"`
	QuizStartedAt            *string                `json:"quiz_started_at,omitempty"`
	Config                   map[string]interface{} `json:"config,omitempty"`
	CreatedAt                string                 `json:"created_at"`
	UpdatedAt                string                 `json:"updated_at"`
	ParticipantCount         int                    `json:"participant_count,omitempty"`
}

// ToAPI converts LiveSessionDAO to LiveSessionAPI
func (s *LiveSessionDAO) ToAPI() LiveSessionAPI {
	api := LiveSessionAPI{
		ID:                   s.ID,
		QuizID:               s.QuizID,
		CreatorID:            s.CreatorID,
		SessionType:          s.SessionType,
		Title:                s.Title,
		Status:               s.Status,
		IsDiscoverable:       s.IsDiscoverable,
		ProgressionMode:      s.ProgressionMode,
		TimerMode:            s.TimerMode,
		QuestionTimerSeconds: s.QuestionTimerSeconds,
		CurrentQuestionIndex: s.CurrentQuestionIndex,
		CreatedAt:            s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            s.UpdatedAt.Format(time.RFC3339),
	}

	if s.Description.Valid {
		api.Description = s.Description.String
	}

	if s.ScheduledAt.Valid {
		scheduledAt := s.ScheduledAt.Time.Format(time.RFC3339)
		api.ScheduledAt = &scheduledAt
	}

	if s.StartedAt.Valid {
		startedAt := s.StartedAt.Time.Format(time.RFC3339)
		api.StartedAt = &startedAt
	}

	if s.EndedAt.Valid {
		endedAt := s.EndedAt.Time.Format(time.RFC3339)
		api.EndedAt = &endedAt
	}

	if s.AccessCode.Valid {
		api.AccessCode = s.AccessCode.String
	}

	if s.MaxParticipants.Valid {
		maxParticipants := int(s.MaxParticipants.Int64)
		api.MaxParticipants = &maxParticipants
	}

	if s.CurrentQuestionStartedAt.Valid {
		currentQuestionStartedAt := s.CurrentQuestionStartedAt.Time.Format(time.RFC3339)
		api.CurrentQuestionStartedAt = &currentQuestionStartedAt
	}

	if s.QuizStartedAt.Valid {
		quizStartedAt := s.QuizStartedAt.Time.Format(time.RFC3339)
		api.QuizStartedAt = &quizStartedAt
	}

	if s.WholeQuizTimerSeconds.Valid {
		wholeQuizTimerSeconds := int(s.WholeQuizTimerSeconds.Int64)
		api.WholeQuizTimerSeconds = &wholeQuizTimerSeconds
	}

	// Parse JSONB config
	if len(s.Config) > 0 {
		json.Unmarshal(s.Config, &api.Config)
	}

	return api
}

// CreateLiveSessionRequest represents a request to create a live session
type CreateLiveSessionRequest struct {
	QuizID               uuid.UUID              `json:"quiz_id" binding:"required"`
	Title                string                 `json:"title" binding:"required"`
	Description          string                 `json:"description,omitempty"`
	SessionType          string                 `json:"session_type" binding:"required"`
	ScheduledAt           *string                `json:"scheduled_at,omitempty"`
	ProgressionMode       string                 `json:"progression_mode" binding:"required"`
	TimerMode             string                 `json:"timer_mode" binding:"required"`
	QuestionTimerSeconds  *int                   `json:"question_timer_seconds,omitempty"`
	WholeQuizTimerSeconds *int                   `json:"whole_quiz_timer_seconds,omitempty"`
	MaxParticipants       *int                   `json:"max_participants,omitempty"`
	IsDiscoverable        bool                   `json:"is_discoverable"`
	Config                map[string]interface{} `json:"config,omitempty"`
}

// Session constants
const (
	SessionTypePublic  = "public"
	SessionTypePrivate = "private"

	SessionStatusWaiting   = "waiting"
	SessionStatusLive      = "live"
	SessionStatusCompleted = "completed"
	SessionStatusCancelled = "cancelled"

	ProgressionModeManual    = "manual"
	ProgressionModeAutomatic = "automatic"

	TimerModePerQuestion = "per_question"
	TimerModeWholeQuiz   = "whole_quiz"
)
