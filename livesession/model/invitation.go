package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// InvitationDAO represents invitation data from database
type InvitationDAO struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	InviterID   uuid.UUID
	InviteeID   uuid.UUID
	Status      string
	InvitedAt   time.Time
	RespondedAt sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InvitationAPI represents invitation data for API responses
type InvitationAPI struct {
	ID          uuid.UUID `json:"id"`
	SessionID   uuid.UUID `json:"session_id"`
	InviterID   uuid.UUID `json:"inviter_id"`
	InviterName string    `json:"inviter_name,omitempty"`
	InviteeID   uuid.UUID `json:"invitee_id"`
	InviteeName string    `json:"invitee_name,omitempty"`
	Status      string    `json:"status"`
	InvitedAt   string    `json:"invited_at"`
	RespondedAt *string   `json:"responded_at,omitempty"`
}

// ToAPI converts InvitationDAO to InvitationAPI
func (i *InvitationDAO) ToAPI() InvitationAPI {
	api := InvitationAPI{
		ID:        i.ID,
		SessionID: i.SessionID,
		InviterID: i.InviterID,
		InviteeID: i.InviteeID,
		Status:    i.Status,
		InvitedAt: i.InvitedAt.Format(time.RFC3339),
	}

	if i.RespondedAt.Valid {
		respondedAt := i.RespondedAt.Time.Format(time.RFC3339)
		api.RespondedAt = &respondedAt
	}

	return api
}

// InviteUsersRequest represents a request to invite users to a session
type InviteUsersRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required"`
}

// Invitation status constants
const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusDeclined = "declined"
)
