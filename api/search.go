package api

import (
	"strconv"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/core"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/gin-gonic/gin"
)

// SearchQuizzes searches quizzes with filters and pagination
func SearchQuizzes(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	// User ID is optional for search
	var userIDPtr *string
	if userID, ok := commons.UsrID(ctx); ok {
		userIDStr := userID.String()
		userIDPtr = &userIDStr
	}

	// Parse query parameters
	var input model.SearchQuizzesInput
	if err := c.ShouldBindQuery(&input); err != nil {
		commons.LogW(ctx, "[api.SearchQuizzes::BindErr [err=%v", err)
		commons.RespBadReq(c, "invalid-query-parameters")
		return
	}

	commons.Log(ctx, "[api.SearchQuizzes [query=%s [page=%d", input.Query, input.Page)

	result, err := core.SearchQuizzes(ctx, input, userIDPtr)
	if err != nil {
		commons.LogC(ctx, "[api.SearchQuizzes::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.SearchQuizzes::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(result.Quizzes))
	commons.RespOKWithMsg(c, "Search completed successfully", result)
}

// GetCategories gets all categories with statistics
func GetCategories(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	commons.Log(ctx, "[api.GetCategories")

	categories, err := core.GetCategories(ctx)
	if err != nil {
		commons.LogC(ctx, "[api.GetCategories::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	response := gin.H{
		"categories": categories,
	}

	commons.Log(ctx, "[api.GetCategories::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(categories))
	commons.RespOKWithMsg(c, "Categories retrieved successfully", response)
}

// GetPopularTags gets most popular tags
func GetPopularTags(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	// Parse limit parameter
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	commons.Log(ctx, "[api.GetPopularTags [limit=%d", limit)

	tags, err := core.GetPopularTags(ctx, limit)
	if err != nil {
		commons.LogC(ctx, "[api.GetPopularTags::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	response := gin.H{
		"tags": tags,
	}

	commons.Log(ctx, "[api.GetPopularTags::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(tags))
	commons.RespOKWithMsg(c, "Tags retrieved successfully", response)
}

// GetQuizByShareCode retrieves a quiz by share code
func GetQuizByShareCode(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	shareCode := c.Param("code")
	commons.Log(ctx, "[api.GetQuizByShareCode [code=%s", shareCode)

	quiz, err := core.GetQuizByShareCode(ctx, shareCode)
	if err != nil {
		commons.LogC(ctx, "[api.GetQuizByShareCode::KO [time_taken=%v [ip=%s [code=%s [err=%v",
			time.Since(st), ip, shareCode, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetQuizByShareCode::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quiz.ID)
	commons.RespOKWithMsg(c, "Quiz retrieved successfully", quiz)
}
