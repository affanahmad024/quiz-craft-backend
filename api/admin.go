package api

import (
	"context"
	"strconv"
	"time"

	admincore "github.com/affanahmad024/quiz-craft/backend/admin/core"
	adminmodel "github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ========== User Management Endpoints ==========

// ListUsers lists all users with admin view
func AdminListUsers(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse query parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var role *string
	if roleStr := c.Query("role"); roleStr != "" {
		role = &roleStr
	}

	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if e, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = &e
		}
	}

	var emailVerified *bool
	if evStr := c.Query("email_verified"); evStr != "" {
		if ev, err := strconv.ParseBool(evStr); err == nil {
			emailVerified = &ev
		}
	}

	var search *string
	if searchStr := c.Query("search"); searchStr != "" {
		search = &searchStr
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	commons.Log(ctx, "[api.ListUsers [admin_id=%s [page=%d", userID, page)

	users, total, err := admincore.ListUsers(ctx, userID.String(), page, limit, role, enabled, emailVerified, search, sortBy, sortOrder)
	if err != nil {
		commons.LogC(ctx, "[api.ListUsers::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	totalPages := (total + limit - 1) / limit
	response := adminmodel.UserListResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	commons.Log(ctx, "[api.ListUsers::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(users))
	commons.RespOKWithMsg(c, "Users retrieved successfully", response)
}

// GetUserDetails gets detailed user info for admin
func GetUserDetails(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	targetUserID := c.Param("id")
	commons.Log(ctx, "[api.GetUserDetails [admin_id=%s [target_user_id=%s", userID, targetUserID)

	user, err := admincore.GetUserDetail(ctx, userID.String(), targetUserID)
	if err != nil {
		commons.LogC(ctx, "[api.GetUserDetails::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetUserDetails::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "User details retrieved successfully", user)
}

// ChangeUserRole changes a user's role
func ChangeUserRole(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	targetUserID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.ChangeUserRole::BindErr [admin_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.ChangeUserRole [admin_id=%s [target_user_id=%s [new_role=%s", userID, targetUserID, req.Role)

	err := admincore.ChangeUserRole(ctx, userID.String(), targetUserID, req.Role)
	if err != nil {
		commons.LogC(ctx, "[api.ChangeUserRole::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ChangeUserRole::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "User role updated successfully", nil)
}

// DisableUser disables a user account
func AdminDisableUser(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	targetUserID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.DisableUser::BindErr [admin_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.DisableUser [admin_id=%s [target_user_id=%s", userID, targetUserID)

	err := admincore.DisableUser(ctx, userID.String(), targetUserID, req.Reason)
	if err != nil {
		commons.LogC(ctx, "[api.DisableUser::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DisableUser::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "User account disabled successfully", nil)
}

// EnableUser enables a user account
func AdminEnableUser(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	targetUserID := c.Param("id")
	commons.Log(ctx, "[api.EnableUser [admin_id=%s [target_user_id=%s", userID, targetUserID)

	err := admincore.EnableUser(ctx, userID.String(), targetUserID)
	if err != nil {
		commons.LogC(ctx, "[api.EnableUser::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.EnableUser::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "User account enabled successfully", nil)
}

// DeleteUser permanently deletes a user
func DeleteUser(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	targetUserID := c.Param("id")

	// Require permanent=true query parameter
	permanent := c.Query("permanent") == "true"
	if !permanent {
		commons.RespBadReq(c, "permanent-flag-required")
		return
	}

	commons.Log(ctx, "[api.DeleteUser [admin_id=%s [target_user_id=%s [permanent=%v", userID, targetUserID, permanent)

	err := admincore.DeleteUser(ctx, userID.String(), targetUserID, permanent)
	if err != nil {
		commons.LogC(ctx, "[api.DeleteUser::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DeleteUser::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "User permanently deleted successfully", nil)
}

// ========== Quiz Moderation Endpoints ==========

// ListQuizzesAdmin lists all quizzes with admin view
func ListQuizzesAdmin(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse query parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var isPublished *bool
	if pubStr := c.Query("is_published"); pubStr != "" {
		if pub, err := strconv.ParseBool(pubStr); err == nil {
			isPublished = &pub
		}
	}

	var isPublic *bool
	if pubStr := c.Query("is_public"); pubStr != "" {
		if pub, err := strconv.ParseBool(pubStr); err == nil {
			isPublic = &pub
		}
	}

	var creatorID *string
	if cid := c.Query("creator_id"); cid != "" {
		creatorID = &cid
	}

	var category *string
	if cat := c.Query("category"); cat != "" {
		category = &cat
	}

	var search *string
	if searchStr := c.Query("search"); searchStr != "" {
		search = &searchStr
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	includeDeleted := c.Query("include_deleted") == "true"

	commons.Log(ctx, "[api.ListQuizzesAdmin [admin_id=%s [page=%d", userID, page)

	quizzes, total, err := admincore.ListQuizzesAdmin(ctx, userID.String(), page, limit, isPublished, isPublic, creatorID, category, search, sortBy, sortOrder, includeDeleted)
	if err != nil {
		commons.LogC(ctx, "[api.ListQuizzesAdmin::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	response := adminmodel.QuizListResponse{
		Quizzes: quizzes,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}

	commons.Log(ctx, "[api.ListQuizzesAdmin::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(quizzes))
	commons.RespOKWithMsg(c, "Quizzes retrieved successfully", response)
}

// GetQuizDetailsAdmin gets detailed quiz info for admin
func GetQuizDetailsAdmin(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	commons.Log(ctx, "[api.GetQuizDetailsAdmin [admin_id=%s [quiz_id=%s", userID, quizID)

	quiz, err := admincore.GetQuizDetailAdmin(ctx, userID.String(), quizID)
	if err != nil {
		commons.LogC(ctx, "[api.GetQuizDetailsAdmin::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetQuizDetailsAdmin::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "Quiz details retrieved successfully", quiz)
}

// DeleteQuizAdmin deletes a quiz
func DeleteQuizAdmin(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	permanent := c.Query("permanent") == "true"

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.DeleteQuizAdmin::BindErr [admin_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.DeleteQuizAdmin [admin_id=%s [quiz_id=%s [permanent=%v", userID, quizID, permanent)

	err := admincore.DeleteQuizAdmin(ctx, userID.String(), quizID, req.Reason, permanent)
	if err != nil {
		commons.LogC(ctx, "[api.DeleteQuizAdmin::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	deleteType := "soft deleted"
	if permanent {
		deleteType = "permanently deleted"
	}

	commons.Log(ctx, "[api.DeleteQuizAdmin::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "Quiz "+deleteType+" successfully", nil)
}

// ========== Statistics Endpoints ==========

// GetPlatformStats gets platform-wide statistics
func GetPlatformStats(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.GetPlatformStats [admin_id=%s", userID)

	stats, err := admincore.GetPlatformStats(ctx, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetPlatformStats::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetPlatformStats::OK [time_taken=%v [ip=%s",
		time.Since(st), ip)
	commons.RespOKWithMsg(c, "Platform statistics retrieved successfully", stats)
}

// GetUserActivity gets daily user activity timeline
func GetUserActivity(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	commons.Log(ctx, "[api.GetUserActivity [admin_id=%s [days=%d", userID, days)

	activities, err := admincore.GetUserActivityTimeline(ctx, userID.String(), days)
	if err != nil {
		commons.LogC(ctx, "[api.GetUserActivity::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	response := gin.H{
		"activity": activities,
	}

	commons.Log(ctx, "[api.GetUserActivity::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(activities))
	commons.RespOKWithMsg(c, "Activity timeline retrieved successfully", response)
}

// GetActivityLogs gets activity logs with filters
func GetActivityLogs(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Check admin role
	if !isAdminCheck(ctx, userID.String()) {
		commons.Resp403(c, "admin-access-required")
		return
	}

	// Parse query parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	filter := adminmodel.ActivityLogFilter{
		Page:  page,
		Limit: limit,
	}

	if uid := c.Query("user_id"); uid != "" {
		filter.UserID = &uid
	}

	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filter.ResourceType = &resourceType
	}

	// Parse date filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	commons.Log(ctx, "[api.GetActivityLogs [admin_id=%s [page=%d", userID, page)

	logs, total, err := admincore.GetActivityLogs(ctx, filter)
	if err != nil {
		commons.LogC(ctx, "[api.GetActivityLogs::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	response := adminmodel.ActivityLogResponse{
		Logs:  logs,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	commons.Log(ctx, "[api.GetActivityLogs::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(logs))
	commons.RespOKWithMsg(c, "Activity logs retrieved successfully", response)
}

// Helper function to check if user is admin
func isAdminCheck(ctx context.Context, userID string) bool {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	user, err := userdao.FindByID(ctx, userUUID)
	if err != nil {
		return false
	}
	return user.Role == "admin"
}
