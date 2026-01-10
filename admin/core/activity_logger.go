package core

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/admin/dao"
	"github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/gin-gonic/gin"
)

// LogActivity logs an activity to the audit trail
func LogActivity(ctx context.Context, userID string, action string, resourceType string, resourceID string, details map[string]interface{}) {
	commons.Log(ctx, "[core.LogActivity [user_id=%s [action=%s [resource_type=%s", userID, action, resourceType)

	// Extract IP and User-Agent from context
	ipAddress := getIPFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)

	log := &model.ActivityLog{
		UserID:       &userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Details:      details,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		CreatedAt:    time.Now(),
	}

	// Create log entry (fire and forget, don't fail the request if logging fails)
	go func() {
		if err := dao.CreateActivityLog(context.Background(), log); err != nil {
			commons.LogE(ctx, "[core.LogActivity::CreateErr [err=%v", err)
		}
	}()
}

// LogActivitySync logs an activity synchronously (waits for completion)
func LogActivitySync(ctx context.Context, userID string, action string, resourceType string, resourceID string, details map[string]interface{}) error {
	commons.Log(ctx, "[core.LogActivitySync [user_id=%s [action=%s", userID, action)

	ipAddress := getIPFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)

	log := &model.ActivityLog{
		UserID:       &userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Details:      details,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		CreatedAt:    time.Now(),
	}

	return dao.CreateActivityLog(ctx, log)
}

// GetActivityLogs retrieves activity logs with filters
func GetActivityLogs(ctx context.Context, filter model.ActivityLogFilter) ([]model.ActivityLog, int, error) {
	commons.Log(ctx, "[core.GetActivityLogs [page=%d", filter.Page)

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}

	logs, total, err := dao.GetActivityLogs(ctx, filter)
	if err != nil {
		commons.LogE(ctx, "[core.GetActivityLogs::DAOErr [err=%v", err)
		return nil, 0, err
	}

	return logs, total, nil
}

// Helper functions to extract context information

func getIPFromContext(ctx context.Context) string {
	// Try to get Gin context
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.ClientIP()
	}

	// Try to get from context value
	if ip, ok := ctx.Value("ip_address").(string); ok {
		return ip
	}

	return "unknown"
}

func getUserAgentFromContext(ctx context.Context) string {
	// Try to get Gin context
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.Request.UserAgent()
	}

	// Try to get from context value
	if ua, ok := ctx.Value("user_agent").(string); ok {
		return ua
	}

	return "unknown"
}
