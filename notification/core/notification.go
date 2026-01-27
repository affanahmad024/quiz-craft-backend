package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/notification/dao"
	"github.com/affanahmad024/quiz-craft/backend/notification/model"
)

// NotificationService handles notification business logic
type NotificationService struct {
	wsHub WebSocketHub // Will be set after WebSocket implementation
}

// WebSocketHub interface for real-time notification delivery
type WebSocketHub interface {
	BroadcastToUser(userID uuid.UUID, notification model.NotificationAPI)
}

// NewNotificationService creates a new notification service
func NewNotificationService(wsHub WebSocketHub) *NotificationService {
	return &NotificationService{
		wsHub: wsHub,
	}
}

// CreateNotification creates a single notification
func (s *NotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, notifType, title, message string, data map[string]interface{}) (model.NotificationAPI, error) {
	commons.Log(ctx, "[core.CreateNotification [user_id=%s [type=%s", userID, notifType)

	// Validate notification type
	if !isValidNotificationType(notifType) {
		return model.NotificationAPI{}, commons.BadInput("invalid-notification-type")
	}

	// Create notification in database
	notificationDAO, err := dao.Create(ctx, userID, notifType, title, message, data)
	if err != nil {
		commons.LogE(ctx, "[core.CreateNotification::CreateErr [user_id=%s [err=%v", userID, err)
		return model.NotificationAPI{}, err
	}

	api := notificationDAO.ToAPI()

	// Send real-time notification via WebSocket if hub is available
	if s.wsHub != nil {
		go s.wsHub.BroadcastToUser(userID, api)
	}

	commons.Log(ctx, "[core.CreateNotification::OK [notification_id=%s", api.ID)
	return api, nil
}

// CreateBulkNotifications creates notifications for multiple users
func (s *NotificationService) CreateBulkNotifications(ctx context.Context, userIDs []uuid.UUID, notifType, title, message string, data map[string]interface{}) error {
	commons.Log(ctx, "[core.CreateBulkNotifications [type=%s [user_count=%d", notifType, len(userIDs))

	if len(userIDs) == 0 {
		return nil
	}

	// Validate notification type
	if !isValidNotificationType(notifType) {
		return commons.BadInput("invalid-notification-type")
	}

	// Create notifications in database
	err := dao.CreateBulk(ctx, userIDs, notifType, title, message, data)
	if err != nil {
		commons.LogE(ctx, "[core.CreateBulkNotifications::CreateErr [type=%s [err=%v", notifType, err)
		return err
	}

	// For bulk notifications, WebSocket delivery would be handled asynchronously
	// Users will receive notifications when they fetch or via polling
	commons.Log(ctx, "[core.CreateBulkNotifications::OK [user_count=%d", len(userIDs))
	return nil
}

// GetUserNotifications gets notifications for a user with filters
func GetUserNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool, notifType *string, limit, offset int) (model.NotificationListResponse, error) {
	commons.Log(ctx, "[core.GetUserNotifications [user_id=%s [unread_only=%v", userID, unreadOnly)

	filter := model.NotificationListFilter{
		UserID:     userID,
		UnreadOnly: unreadOnly,
		Type:       notifType,
		Limit:      limit,
		Offset:     offset,
	}

	notifications, total, err := dao.ListByUser(ctx, filter)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserNotifications::ListErr [user_id=%s [err=%v", userID, err)
		return model.NotificationListResponse{}, err
	}

	// Get unread count
	unreadCount, err := dao.GetUnreadCount(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserNotifications::UnreadCountErr [user_id=%s [err=%v", userID, err)
		// Don't fail the request if unread count fails
		unreadCount = 0
	}

	// Convert to API models
	apiNotifications := make([]model.NotificationAPI, 0, len(notifications))
	for _, n := range notifications {
		apiNotifications = append(apiNotifications, n.ToAPI())
	}

	response := model.NotificationListResponse{
		Notifications: apiNotifications,
		UnreadCount:   unreadCount,
		Total:         total,
	}

	commons.Log(ctx, "[core.GetUserNotifications::OK [user_id=%s [total=%d", userID, total)
	return response, nil
}

// MarkAsRead marks a notification as read
func MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	commons.Log(ctx, "[core.MarkAsRead [notification_id=%s [user_id=%s", notificationID, userID)

	err := dao.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		commons.LogE(ctx, "[core.MarkAsRead::MarkErr [notification_id=%s [err=%v", notificationID, err)
		return err
	}

	commons.Log(ctx, "[core.MarkAsRead::OK [notification_id=%s", notificationID)
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	commons.Log(ctx, "[core.MarkAllAsRead [user_id=%s", userID)

	count, err := dao.MarkAllAsRead(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.MarkAllAsRead::MarkErr [user_id=%s [err=%v", userID, err)
		return 0, err
	}

	commons.Log(ctx, "[core.MarkAllAsRead::OK [user_id=%s [count=%d", userID, count)
	return count, nil
}

// DeleteNotification deletes a notification
func DeleteNotification(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	commons.Log(ctx, "[core.DeleteNotification [notification_id=%s [user_id=%s", notificationID, userID)

	err := dao.Delete(ctx, notificationID, userID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteNotification::DeleteErr [notification_id=%s [err=%v", notificationID, err)
		return err
	}

	commons.Log(ctx, "[core.DeleteNotification::OK [notification_id=%s", notificationID)
	return nil
}

// GetUnreadCount gets the count of unread notifications for a user
func GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	commons.Log(ctx, "[core.GetUnreadCount [user_id=%s", userID)

	count, err := dao.GetUnreadCount(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.GetUnreadCount::CountErr [user_id=%s [err=%v", userID, err)
		return 0, err
	}

	return count, nil
}

// Helper function to validate notification type
func isValidNotificationType(notifType string) bool {
	validTypes := model.ValidNotificationTypes()
	for _, vt := range validTypes {
		if vt == notifType {
			return true
		}
	}
	return false
}
