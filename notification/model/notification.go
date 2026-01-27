package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationDAO represents notification data from database (uses sql.Null* for nullable fields)
type NotificationDAO struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Message   string
	Data      []byte // JSONB stored as bytes
	IsRead    bool
	ReadAt    sql.NullTime
	CreatedAt time.Time
}

// NotificationAPI represents notification data for API responses (native Go types)
type NotificationAPI struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IsRead    bool                   `json:"is_read"`
	ReadAt    *string                `json:"read_at,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

// ToAPI converts NotificationDAO to NotificationAPI
func (n *NotificationDAO) ToAPI() NotificationAPI {
	api := NotificationAPI{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}

	if n.ReadAt.Valid {
		readAt := n.ReadAt.Time.Format(time.RFC3339)
		api.ReadAt = &readAt
	}

	// Parse JSONB data
	if len(n.Data) > 0 {
		json.Unmarshal(n.Data, &api.Data)
	}

	return api
}

// NotificationListFilter represents filters for notification listing
type NotificationListFilter struct {
	UserID     uuid.UUID
	UnreadOnly bool
	Type       *string
	Limit      int
	Offset     int
}

// NotificationListResponse represents paginated notification list
type NotificationListResponse struct {
	Notifications []NotificationAPI `json:"notifications"`
	UnreadCount   int               `json:"unread_count"`
	Total         int               `json:"total"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID  uuid.UUID              `json:"user_id" binding:"required"`
	Type    string                 `json:"type" binding:"required"`
	Title   string                 `json:"title" binding:"required"`
	Message string                 `json:"message" binding:"required"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// CreateBulkNotificationRequest represents a request to create notifications for multiple users
type CreateBulkNotificationRequest struct {
	UserIDs []uuid.UUID            `json:"user_ids" binding:"required"`
	Type    string                 `json:"type" binding:"required"`
	Title   string                 `json:"title" binding:"required"`
	Message string                 `json:"message" binding:"required"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// UnreadCountResponse represents unread notification count
type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

// Notification types constants
const (
	NotificationTypeNewQuiz                = "new_quiz"
	NotificationTypeLiveSessionScheduled   = "live_session_scheduled"
	NotificationTypeLiveSessionStarting    = "live_session_starting"
	NotificationTypeLiveSessionInvitation  = "live_session_invitation"
	NotificationTypeNewFollower            = "new_follower"
	NotificationTypeSystem                 = "system"
)

// ValidNotificationTypes returns all valid notification types
func ValidNotificationTypes() []string {
	return []string{
		NotificationTypeNewQuiz,
		NotificationTypeLiveSessionScheduled,
		NotificationTypeLiveSessionStarting,
		NotificationTypeLiveSessionInvitation,
		NotificationTypeNewFollower,
		NotificationTypeSystem,
	}
}
