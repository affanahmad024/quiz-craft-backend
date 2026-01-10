package core

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/admin/dao"
	"github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/google/uuid"
)

// GetPlatformStats retrieves comprehensive platform statistics
func GetPlatformStats(ctx context.Context, adminID string) (*model.PlatformStats, error) {
	commons.Log(ctx, "[core.GetPlatformStats [admin_id=%s", adminID)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return nil, commons.Forbidden("admin-access-required")
	}

	stats, err := dao.GetPlatformStats(ctx)
	if err != nil {
		commons.LogE(ctx, "[core.GetPlatformStats::DAOErr [err=%v", err)
		return nil, err
	}

	// Log admin access
	go LogActivity(ctx, adminID, model.ActionAdminAccess, "platform", "stats", map[string]interface{}{
		"admin_name": admin.Name,
	})

	commons.Log(ctx, "[core.GetPlatformStats::OK")
	return stats, nil
}

// GetUserActivityTimeline retrieves daily activity timeline
func GetUserActivityTimeline(ctx context.Context, adminID string, days int) ([]model.UserActivity, error) {
	commons.Log(ctx, "[core.GetUserActivityTimeline [admin_id=%s [days=%d", adminID, days)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return nil, commons.Forbidden("admin-access-required")
	}

	// Validate days parameter
	if days < 1 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	// Calculate start date
	startDate := time.Now().AddDate(0, 0, -days)

	activities, err := dao.GetDailyUserActivity(ctx, startDate, days)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserActivityTimeline::DAOErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.GetUserActivityTimeline::OK [count=%d", len(activities))
	return activities, nil
}
