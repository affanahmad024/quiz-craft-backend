package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ParticipantDAO represents participant data from database
type ParticipantDAO struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	UserID       uuid.UUID
	JoinedAt     time.Time
	LeftAt       sql.NullTime
	CurrentScore int
	Rank         int
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ParticipantAPI represents participant data for API responses
type ParticipantAPI struct {
	ID           uuid.UUID `json:"id"`
	SessionID    uuid.UUID `json:"session_id"`
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name,omitempty"`
	JoinedAt     string    `json:"joined_at"`
	LeftAt       *string   `json:"left_at,omitempty"`
	CurrentScore int       `json:"current_score"`
	Rank         int       `json:"rank"`
	Status       string    `json:"status"`
}

// ToAPI converts ParticipantDAO to ParticipantAPI
func (p *ParticipantDAO) ToAPI() ParticipantAPI {
	api := ParticipantAPI{
		ID:           p.ID,
		SessionID:    p.SessionID,
		UserID:       p.UserID,
		JoinedAt:     p.JoinedAt.Format(time.RFC3339),
		CurrentScore: p.CurrentScore,
		Rank:         p.Rank,
		Status:       p.Status,
	}

	if p.LeftAt.Valid {
		leftAt := p.LeftAt.Time.Format(time.RFC3339)
		api.LeftAt = &leftAt
	}

	return api
}

// LeaderboardEntry represents a participant's position in the leaderboard
type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name"`
	CurrentScore int       `json:"current_score"`
	Status       string    `json:"status"`
}

// Participant status constants
const (
	ParticipantStatusActive       = "active"
	ParticipantStatusDisconnected = "disconnected"
	ParticipantStatusCompleted    = "completed"
)
