package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/notification/core"
	"github.com/affanahmad024/quiz-craft/backend/notification/model"
)

// GetNotifications gets notifications for current user
func GetNotifications(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.LogW(ctx, "[api.GetNotifications::NoUserID [time_taken=%v [ip=%s", time.Since(st), ip)
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse query parameters
	unreadOnlyStr := c.DefaultQuery("unread_only", "false")
	unreadOnly := unreadOnlyStr == "true"
	notifType := c.Query("type")
	var notifTypePtr *string
	if notifType != "" {
		notifTypePtr = &notifType
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	commons.Log(ctx, "[api.GetNotifications [user_id=%s [unread_only=%v [limit=%d [offset=%d",
		userID, unreadOnly, limit, offset)

	// Call CORE layer
	response, err := core.GetUserNotifications(ctx, userID, unreadOnly, notifTypePtr, limit, offset)
	if err != nil {
		commons.LogC(ctx, "[api.GetNotifications::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetNotifications::OK [time_taken=%v [ip=%s [user_id=%s [count=%d",
		time.Since(st), ip, userID, len(response.Notifications))
	commons.RespOKWithMsg(c, "Notifications retrieved successfully", response)
}

// GetUnreadCount gets unread notification count for current user
func GetUnreadCount(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.LogW(ctx, "[api.GetUnreadCount::NoUserID [time_taken=%v [ip=%s", time.Since(st), ip)
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.GetUnreadCount [user_id=%s", userID)

	// Call CORE layer
	count, err := core.GetUnreadCount(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.GetUnreadCount::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)
		commons.RespISE(c, err)
		return
	}

	response := model.UnreadCountResponse{
		UnreadCount: count,
	}

	commons.Log(ctx, "[api.GetUnreadCount::OK [time_taken=%v [ip=%s [user_id=%s [count=%d",
		time.Since(st), ip, userID, count)
	commons.RespOKWithMsg(c, "Unread count retrieved successfully", response)
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.LogW(ctx, "[api.MarkNotificationAsRead::NoUserID [time_taken=%v [ip=%s", time.Since(st), ip)
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse notification ID from path
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		commons.LogW(ctx, "[api.MarkNotificationAsRead::InvalidID [notification_id=%s [err=%v",
			notificationIDStr, err)
		commons.RespBadReq(c, "invalid-notification-id")
		return
	}

	commons.Log(ctx, "[api.MarkNotificationAsRead [user_id=%s [notification_id=%s",
		userID, notificationID)

	// Call CORE layer
	err = core.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		commons.LogC(ctx, "[api.MarkNotificationAsRead::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.MarkNotificationAsRead::OK [time_taken=%v [ip=%s [user_id=%s [notification_id=%s",
		time.Since(st), ip, userID, notificationID)
	commons.RespOKWithMsg(c, "Notification marked as read", nil)
}

// MarkAllNotificationsAsRead marks all notifications as read for current user
func MarkAllNotificationsAsRead(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.LogW(ctx, "[api.MarkAllNotificationsAsRead::NoUserID [time_taken=%v [ip=%s", time.Since(st), ip)
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.MarkAllNotificationsAsRead [user_id=%s", userID)

	// Call CORE layer
	count, err := core.MarkAllAsRead(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.MarkAllNotificationsAsRead::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)
		commons.RespISE(c, err)
		return
	}

	response := map[string]interface{}{
		"count": count,
	}

	commons.Log(ctx, "[api.MarkAllNotificationsAsRead::OK [time_taken=%v [ip=%s [user_id=%s [count=%d",
		time.Since(st), ip, userID, count)
	commons.RespOKWithMsg(c, "All notifications marked as read", response)
}

// DeleteNotification deletes a notification
func DeleteNotification(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.LogW(ctx, "[api.DeleteNotification::NoUserID [time_taken=%v [ip=%s", time.Since(st), ip)
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse notification ID from path
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		commons.LogW(ctx, "[api.DeleteNotification::InvalidID [notification_id=%s [err=%v",
			notificationIDStr, err)
		commons.RespBadReq(c, "invalid-notification-id")
		return
	}

	commons.Log(ctx, "[api.DeleteNotification [user_id=%s [notification_id=%s",
		userID, notificationID)

	// Call CORE layer
	err = core.DeleteNotification(ctx, notificationID, userID)
	if err != nil {
		commons.LogC(ctx, "[api.DeleteNotification::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DeleteNotification::OK [time_taken=%v [ip=%s [user_id=%s [notification_id=%s",
		time.Since(st), ip, userID, notificationID)
	commons.RespOKWithMsg(c, "Notification deleted successfully", nil)
}
