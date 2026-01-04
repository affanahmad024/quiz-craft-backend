package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/user/core"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
)

// GetUserProfile gets current user's profile
func GetUserProfile(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	// Extract user ID from context (set by auth middleware)
	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.LogW(ctx, "[api.GetUserProfile::NoUserID [time_taken=%v [ip=%s", time.Since(st), ip)
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.GetUserProfile [user_id=%s", userID)

	// Call CORE layer
	user, err := core.GetUserProfile(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.GetUserProfile::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetUserProfile::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Profile retrieved successfully", user)
}

// UpdateProfile updates current user's profile
func UpdateProfile(c *gin.Context) {
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

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.UpdateProfile::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.UpdateProfile [user_id=%s [req=%+v", userID, req)

	err := core.UpdateUserProfile(ctx, userID, req)
	if err != nil {
		commons.LogC(ctx, "[api.UpdateProfile::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UpdateProfile::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Profile updated successfully", nil)
}

// DeleteAccount soft deletes current user's account
func DeleteAccount(c *gin.Context) {
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

	commons.Log(ctx, "[api.DeleteAccount [user_id=%s", userID)

	err := core.DeleteUser(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.DeleteAccount::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DeleteAccount::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Account deleted successfully. You can contact support within 30 days to recover your account.", nil)
}

// ListUsers lists all users with pagination
func ListUsers(c *gin.Context) {
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := model.UserListFilter{
		Page:  page,
		Limit: limit,
	}

	commons.Log(ctx, "[api.ListUsers [user_id=%s [page=%d [limit=%d", userID, page, limit)

	resp, err := core.ListUsers(ctx, filter)
	if err != nil {
		commons.LogC(ctx, "[api.ListUsers::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ListUsers::OK [time_taken=%v [ip=%s [user_id=%s [total=%d",
		time.Since(st), ip, userID, resp.Total)
	commons.RespOKWithMsg(c, "Users retrieved successfully", resp)
}

// GetUserStats gets user statistics
func GetUserStats(c *gin.Context) {
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

	commons.Log(ctx, "[api.GetUserStats [user_id=%s", userID)

	stats, err := core.GetUserStats(ctx)
	if err != nil {
		commons.LogC(ctx, "[api.GetUserStats::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetUserStats::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Statistics retrieved successfully", stats)
}

// GetUserByID gets user by ID (admin)
func GetUserByID(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	currentUserID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse user ID from URL
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		commons.LogW(ctx, "[api.GetUserByID::InvalidUUID [id=%s [err=%v", userIDStr, err)
		commons.RespBadReq(c, "invalid-user-id")
		return
	}

	commons.Log(ctx, "[api.GetUserByID [current_user=%s [target_user=%s", currentUserID, userID)

	user, err := core.GetUserByID(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.GetUserByID::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetUserByID::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "User retrieved successfully", user)
}

// RestoreUser restores a soft deleted user (admin)
func RestoreUser(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	currentUserID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse user ID from URL
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		commons.LogW(ctx, "[api.RestoreUser::InvalidUUID [id=%s [err=%v", userIDStr, err)
		commons.RespBadReq(c, "invalid-user-id")
		return
	}

	commons.Log(ctx, "[api.RestoreUser [admin_user=%s [target_user=%s", currentUserID, userID)

	err = core.RestoreUser(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.RestoreUser::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.RestoreUser::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Account restored successfully", nil)
}

// DisableUser disables a user account (admin)
func DisableUser(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	currentUserID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse user ID from URL
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		commons.LogW(ctx, "[api.DisableUser::InvalidUUID [id=%s [err=%v", userIDStr, err)
		commons.RespBadReq(c, "invalid-user-id")
		return
	}

	commons.Log(ctx, "[api.DisableUser [admin_user=%s [target_user=%s", currentUserID, userID)

	err = core.DisableUser(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.DisableUser::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DisableUser::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "User disabled successfully", nil)
}

// EnableUser enables a user account (admin)
func EnableUser(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	currentUserID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse user ID from URL
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		commons.LogW(ctx, "[api.EnableUser::InvalidUUID [id=%s [err=%v", userIDStr, err)
		commons.RespBadReq(c, "invalid-user-id")
		return
	}

	commons.Log(ctx, "[api.EnableUser [admin_user=%s [target_user=%s", currentUserID, userID)

	err = core.EnableUser(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.EnableUser::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.EnableUser::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "User enabled successfully", nil)
}
