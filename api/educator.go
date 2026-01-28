package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	educatorcore "github.com/affanahmad024/quiz-craft/backend/educator/core"
	educatormodel "github.com/affanahmad024/quiz-craft/backend/educator/model"
	notificationcore "github.com/affanahmad024/quiz-craft/backend/notification/core"
	notificationws "github.com/affanahmad024/quiz-craft/backend/notification/websocket"
)

var notificationServiceInstance *notificationcore.NotificationService

// InitNotificationService initializes the notification service for educator handlers
func InitNotificationService(wsHub *notificationws.Hub) {
	notificationServiceInstance = notificationcore.NewNotificationService(wsHub)
}

// PromoteToEducator handles POST /api/v1/educator/promote
func PromoteToEducator(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	var req educatormodel.PromoteToEducatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	profile, err := educatorcore.PromoteToEducator(ctx, userID, req)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsConflict(err) {
			commons.Resp409(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.PromoteToEducator::OK [time_taken=%v [ip=%s", time.Since(st), ip)
	commons.RespOK(c, gin.H{
		"message": "Successfully promoted to educator",
		"profile": profile,
	})
}

// GetMyEducatorProfile handles GET /api/v1/educator/profile
func GetMyEducatorProfile(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	profile, err := educatorcore.GetEducatorProfile(ctx, userID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetMyEducatorProfile::OK [time_taken=%v [ip=%s", time.Since(st), ip)
	commons.RespOK(c, profile)
}

// UpdateMyEducatorProfile handles PATCH /api/v1/educator/profile
func UpdateMyEducatorProfile(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	var req educatormodel.UpdateEducatorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	err := educatorcore.UpdateEducatorProfile(ctx, userID, req)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UpdateMyEducatorProfile::OK [time_taken=%v [ip=%s", time.Since(st), ip)
	commons.RespOK(c, gin.H{
		"message": "Profile updated successfully",
	})
}

// ListEducators handles GET /api/v1/educators
func ListEducators(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	// Parse pagination parameters
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	response, err := educatorcore.ListEducators(ctx, limit, offset)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ListEducators::OK [time_taken=%v [ip=%s [count=%d", time.Since(st), ip, len(response.Educators))
	commons.RespOK(c, response)
}

// GetEducatorProfile handles GET /api/v1/educators/:id
func GetEducatorProfile(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	educatorIDStr := c.Param("id")
	educatorID, err := uuid.Parse(educatorIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-educator-id")
		return
	}

	educator, err := educatorcore.GetEducatorWithUserInfo(ctx, educatorID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetEducatorProfile::OK [time_taken=%v [ip=%s [educator_id=%s", time.Since(st), ip, educatorID)
	commons.RespOK(c, educator)
}

// SearchEducators handles GET /api/v1/educators/search
func SearchEducators(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	query := c.Query("q")
	if query == "" {
		commons.RespBadReq(c, "search-query-required")
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	response, err := educatorcore.SearchEducators(ctx, query, limit, offset)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.SearchEducators::OK [time_taken=%v [ip=%s [query=%s [count=%d", time.Since(st), ip, query, len(response.Educators))
	commons.RespOK(c, response)
}

// GetFeaturedEducators handles GET /api/v1/educators/featured
func GetFeaturedEducators(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 50 {
				limit = 50
			}
		}
	}

	educators, err := educatorcore.GetFeaturedEducators(ctx, limit)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetFeaturedEducators::OK [time_taken=%v [ip=%s [count=%d", time.Since(st), ip, len(educators))
	commons.RespOK(c, gin.H{
		"educators": educators,
		"total":     len(educators),
	})
}

// FollowEducator handles POST /api/v1/educators/:id/follow
func FollowEducator(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	educatorIDStr := c.Param("id")
	educatorID, err := uuid.Parse(educatorIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-educator-id")
		return
	}

	response, err := educatorcore.FollowEducator(ctx, userID, educatorID, notificationServiceInstance)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsConflict(err) {
			commons.Resp409(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.FollowEducator::OK [time_taken=%v [ip=%s [educator_id=%s", time.Since(st), ip, educatorID)
	commons.RespOK(c, response)
}

// UnfollowEducator handles DELETE /api/v1/educators/:id/unfollow
func UnfollowEducator(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	educatorIDStr := c.Param("id")
	educatorID, err := uuid.Parse(educatorIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-educator-id")
		return
	}

	response, err := educatorcore.UnfollowEducator(ctx, userID, educatorID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UnfollowEducator::OK [time_taken=%v [ip=%s [educator_id=%s", time.Since(st), ip, educatorID)
	commons.RespOK(c, response)
}

// GetMyFollowing handles GET /api/v1/following
func GetMyFollowing(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	// Parse pagination parameters
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	response, err := educatorcore.GetFollowing(ctx, userID, limit, offset)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetMyFollowing::OK [time_taken=%v [ip=%s [count=%d", time.Since(st), ip, len(response.Following))
	commons.RespOK(c, response)
}

// GetMyFollowers handles GET /api/v1/educator/followers
func GetMyFollowers(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	// Parse pagination parameters
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	response, err := educatorcore.GetFollowers(ctx, userID, limit, offset)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetMyFollowers::OK [time_taken=%v [ip=%s [count=%d", time.Since(st), ip, len(response.Followers))
	commons.RespOK(c, response)
}

// GetMyFollowerStats handles GET /api/v1/educator/followers/stats
func GetMyFollowerStats(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	userID, _ := commons.UsrID(ctx)

	stats, err := educatorcore.GetFollowerStats(ctx, userID)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetMyFollowerStats::OK [time_taken=%v [ip=%s [total=%d", time.Since(st), ip, stats.TotalFollowers)
	commons.RespOK(c, stats)
}

// DemoteEducator handles POST /api/v1/admin/users/:id/demote (Admin only)
func DemoteEducator(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	educatorIDStr := c.Param("id")
	educatorID, err := uuid.Parse(educatorIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-user-id")
		return
	}

	err = educatorcore.DemoteEducator(ctx, educatorID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DemoteEducator::OK [time_taken=%v [ip=%s [educator_id=%s", time.Since(st), ip, educatorID)
	commons.RespOK(c, gin.H{
		"message": "Educator demoted to user successfully",
	})
}

// GetEducatorQuizzes handles GET /api/v1/educators/:id/quizzes
func GetEducatorQuizzes(c *gin.Context) {
	ctx := commons.Ctx(c)
	st := time.Now()
	ip := c.ClientIP()

	educatorIDStr := c.Param("id")
	educatorID, err := uuid.Parse(educatorIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-educator-id")
		return
	}

	// Parse pagination parameters (unused for now)
	_ = 20     // limit
	_ = 0      // offset
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			_ = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			_ = parsed
		}
	}

	// Reuse ListUserQuizzes logic with educator ID
	c.Params = append(c.Params, gin.Param{Key: "user_id", Value: educatorIDStr})

	// Call existing quiz listing endpoint (will need to ensure it's accessible)
	commons.Log(ctx, "[api.GetEducatorQuizzes::OK [time_taken=%v [ip=%s [educator_id=%s", time.Since(st), ip, educatorID)

	// For now, return placeholder - this will be integrated with existing quiz API
	commons.RespOK(c, gin.H{
		"quizzes": []interface{}{},
		"total":   0,
		"message": "Quiz listing integration pending",
	})
}
