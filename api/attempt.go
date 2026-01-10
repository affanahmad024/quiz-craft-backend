package api

import (
	"strconv"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/core"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/gin-gonic/gin"
)

// StartAttempt starts a new quiz attempt
func StartAttempt(c *gin.Context) {
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
	input := model.StartAttemptInput{
		QuizID: quizID,
	}

	commons.Log(ctx, "[api.StartAttempt [user_id=%s [quiz_id=%s", userID, quizID)

	attempt, err := core.StartAttempt(ctx, userID.String(), input)
	if err != nil {
		commons.LogC(ctx, "[api.StartAttempt::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.StartAttempt::OK [time_taken=%v [ip=%s [attempt_id=%s",
		time.Since(st), ip, attempt.ID)
	commons.RespOKWithMsg(c, "Quiz attempt started successfully", attempt)
}

// GetAttempt retrieves an attempt by ID
func GetAttempt(c *gin.Context) {
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

	attemptID := c.Param("id")
	commons.Log(ctx, "[api.GetAttempt [user_id=%s [attempt_id=%s", userID, attemptID)

	attempt, err := core.GetAttempt(ctx, attemptID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetAttempt::KO [time_taken=%v [ip=%s [attempt_id=%s [err=%v",
			time.Since(st), ip, attemptID, err)

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

	commons.Log(ctx, "[api.GetAttempt::OK [time_taken=%v [ip=%s [attempt_id=%s",
		time.Since(st), ip, attemptID)
	commons.RespOKWithMsg(c, "Attempt retrieved successfully", attempt)
}

// SubmitAnswer submits an answer for a question
func SubmitAnswer(c *gin.Context) {
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

	attemptID := c.Param("id")

	var req model.SubmitAnswerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.SubmitAnswer::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.SubmitAnswer [user_id=%s [attempt_id=%s [question_id=%s",
		userID, attemptID, req.QuestionID)

	userAnswer, err := core.SubmitAnswer(ctx, attemptID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.SubmitAnswer::KO [time_taken=%v [ip=%s [attempt_id=%s [err=%v",
			time.Since(st), ip, attemptID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.SubmitAnswer::OK [time_taken=%v [ip=%s [user_answer_id=%s",
		time.Since(st), ip, userAnswer.ID)
	commons.RespOKWithMsg(c, "Answer submitted successfully", userAnswer)
}

// FinishAttempt finishes an attempt and calculates score
func FinishAttempt(c *gin.Context) {
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

	attemptID := c.Param("id")
	commons.Log(ctx, "[api.FinishAttempt [user_id=%s [attempt_id=%s", userID, attemptID)

	attempt, err := core.FinishAttempt(ctx, attemptID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.FinishAttempt::KO [time_taken=%v [ip=%s [attempt_id=%s [err=%v",
			time.Since(st), ip, attemptID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.FinishAttempt::OK [time_taken=%v [ip=%s [attempt_id=%s [score=%v",
		time.Since(st), ip, attemptID, attempt.Score)
	commons.RespOKWithMsg(c, "Attempt finished successfully", attempt)
}

// GetAttemptResult gets the full result of an attempt
func GetAttemptResult(c *gin.Context) {
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

	attemptID := c.Param("id")
	commons.Log(ctx, "[api.GetAttemptResult [user_id=%s [attempt_id=%s", userID, attemptID)

	result, err := core.GetAttemptResult(ctx, attemptID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetAttemptResult::KO [time_taken=%v [ip=%s [attempt_id=%s [err=%v",
			time.Since(st), ip, attemptID, err)

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

	commons.Log(ctx, "[api.GetAttemptResult::OK [time_taken=%v [ip=%s [attempt_id=%s",
		time.Since(st), ip, attemptID)
	commons.RespOKWithMsg(c, "Attempt result retrieved successfully", result)
}

// ListMyAttempts lists all attempts for the current user
func ListMyAttempts(c *gin.Context) {
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
	filter := model.AttemptListFilter{
		Page:  1,
		Limit: 20,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	if quizID := c.Query("quiz_id"); quizID != "" {
		filter.QuizID = &quizID
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	commons.Log(ctx, "[api.ListMyAttempts [user_id=%s [page=%d [limit=%d",
		userID, filter.Page, filter.Limit)

	attempts, total, err := core.ListUserAttempts(ctx, userID.String(), filter)
	if err != nil {
		commons.LogC(ctx, "[api.ListMyAttempts::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	response := gin.H{
		"attempts": attempts,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
	}

	commons.Log(ctx, "[api.ListMyAttempts::OK [time_taken=%v [ip=%s [count=%d [total=%d",
		time.Since(st), ip, len(attempts), total)
	commons.RespOKWithMsg(c, "Attempts retrieved successfully", response)
}

// GetQuizStats gets statistics for a quiz (creator only)
func GetQuizStats(c *gin.Context) {
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
	commons.Log(ctx, "[api.GetQuizStats [user_id=%s [quiz_id=%s", userID, quizID)

	stats, err := core.GetQuizStats(ctx, quizID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetQuizStats::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
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

	commons.Log(ctx, "[api.GetQuizStats::OK [time_taken=%v [ip=%s [quiz_id=%s [total_attempts=%d",
		time.Since(st), ip, quizID, stats.TotalAttempts)
	commons.RespOKWithMsg(c, "Quiz stats retrieved successfully", stats)
}
