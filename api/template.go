package api

import (
	"strconv"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/core"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/gin-gonic/gin"
)

// SaveQuizAsTemplate saves a quiz as a reusable template
func SaveQuizAsTemplate(c *gin.Context) {
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

	var req model.SaveAsTemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.SaveQuizAsTemplate::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.SaveQuizAsTemplate [user_id=%s [quiz_id=%s", userID, quizID)

	templateID, err := core.SaveQuizAsTemplate(ctx, quizID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.SaveQuizAsTemplate::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

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

	result := gin.H{
		"template_id": templateID,
	}

	commons.Log(ctx, "[api.SaveQuizAsTemplate::OK [time_taken=%v [ip=%s [template_id=%s",
		time.Since(st), ip, templateID)
	commons.RespOKWithMsg(c, "Quiz saved as template successfully", result)
}

// ListTemplates lists available quiz templates
func ListTemplates(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	// User ID is optional for templates
	var userIDPtr *string
	if userID, ok := commons.UsrID(ctx); ok {
		userIDStr := userID.String()
		userIDPtr = &userIDStr
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

	category := c.Query("category")
	myTemplates := c.Query("my_templates") == "true"

	commons.Log(ctx, "[api.ListTemplates [page=%d [my_templates=%v", page, myTemplates)

	templates, total, err := core.ListTemplates(ctx, userIDPtr, page, limit, category, myTemplates)
	if err != nil {
		commons.LogC(ctx, "[api.ListTemplates::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	response := gin.H{
		"templates": templates,
		"total":     total,
		"page":      page,
		"limit":     limit,
	}

	commons.Log(ctx, "[api.ListTemplates::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(templates))
	commons.RespOKWithMsg(c, "Templates retrieved successfully", response)
}

// CreateQuizFromTemplate creates a new quiz from a template
func CreateQuizFromTemplate(c *gin.Context) {
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

	templateID := c.Param("id")

	var req model.UseTemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.CreateQuizFromTemplate::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.CreateQuizFromTemplate [user_id=%s [template_id=%s", userID, templateID)

	quizID, err := core.CreateQuizFromTemplate(ctx, templateID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.CreateQuizFromTemplate::KO [time_taken=%v [ip=%s [template_id=%s [err=%v",
			time.Since(st), ip, templateID, err)

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

	result := gin.H{
		"quiz_id": quizID,
	}

	commons.Log(ctx, "[api.CreateQuizFromTemplate::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz created from template successfully", result)
}
