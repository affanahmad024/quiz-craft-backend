package api

import (
	"context"
	"strconv"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/core"
	livesessiondao "github.com/affanahmad024/quiz-craft/backend/livesession/dao"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
	livesessionws "github.com/affanahmad024/quiz-craft/backend/livesession/websocket"
	notificationcore "github.com/affanahmad024/quiz-craft/backend/notification/core"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	quizmodel "github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Global notification service for invitations
var notificationService *notificationcore.NotificationService
var liveSessionHub *livesessionws.Hub

// SetNotificationService sets the notification service for live sessions
func SetNotificationService(svc *notificationcore.NotificationService) {
	notificationService = svc
}

// SetLiveSessionHub sets the websocket hub used for manual session events.
func SetLiveSessionHub(hub *livesessionws.Hub) {
	liveSessionHub = hub
}

// CreateLiveSession creates a new live session
func CreateLiveSession(c *gin.Context) {
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

	var req model.CreateLiveSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.CreateLiveSession::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.CreateLiveSession [user_id=%s [quiz_id=%s [type=%s",
		userID, req.QuizID, req.SessionType)

	// Call CORE layer
	session, err := core.CreateSession(ctx, req, userID)
	if err != nil {
		commons.LogC(ctx, "[api.CreateLiveSession::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.CreateLiveSession::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, session.ID)
	commons.RespOKWithMsg(c, "Live session created successfully", session)
}

// GetLiveSession gets a live session by ID
func GetLiveSession(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.GetLiveSession [user_id=%s [session_id=%s", userID, sessionID)

	reconcileExpiredLiveSession(ctx, sessionID)

	// Call CORE layer
	session, err := core.GetSession(ctx, sessionID)
	if err != nil {
		commons.LogC(ctx, "[api.GetLiveSession::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetLiveSession::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, sessionID)
	commons.RespOKWithMsg(c, "Session retrieved successfully", session)
}

// GetLiveSessionQuestions gets playable questions for a live session.
func GetLiveSessionQuestions(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	reconcileExpiredLiveSession(ctx, sessionID)

	session, err := livesessiondao.FindSessionByID(ctx, sessionID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	if session.CreatorID.String() != userID.String() {
		if _, err := livesessiondao.FindParticipant(ctx, sessionID, userID); err != nil {
			commons.Resp403(c, "not-session-participant")
			return
		}
	}

	questions, err := liveQuestionsWithAnswers(ctx, session.QuizID)
	if err != nil {
		commons.LogC(ctx, "[api.GetLiveSessionQuestions::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetLiveSessionQuestions::OK [time_taken=%v [ip=%s [session_id=%s [count=%d",
		time.Since(st), ip, sessionID, len(questions))
	commons.RespOKWithMsg(c, "Session questions retrieved successfully", gin.H{
		"questions": questions,
	})
}

// JoinLiveSession joins a live session
func JoinLiveSession(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	var req model.JoinSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Access code is optional, ignore bind error
	}

	commons.Log(ctx, "[api.JoinLiveSession [user_id=%s [session_id=%s", userID, sessionID)

	// Call CORE layer
	participant, session, err := core.JoinSession(ctx, sessionID, userID, req.AccessCode)
	if err != nil {
		commons.LogC(ctx, "[api.JoinLiveSession::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsConflict(err) {
			commons.Resp409(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	response := model.JoinSessionResponse{
		Message:     "Joined session successfully",
		Participant: participant,
		Session:     session,
	}

	commons.Log(ctx, "[api.JoinLiveSession::OK [time_taken=%v [ip=%s [participant_id=%s",
		time.Since(st), ip, participant.ID)
	commons.RespOK(c, response)
}

// LeaveLiveSession leaves a live session
func LeaveLiveSession(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.LeaveLiveSession [user_id=%s [session_id=%s", userID, sessionID)

	if err := core.LeaveSession(ctx, sessionID, userID); err != nil {
		commons.LogC(ctx, "[api.LeaveLiveSession::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

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

	commons.Log(ctx, "[api.LeaveLiveSession::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, sessionID)
	commons.RespOKWithMsg(c, "Left session successfully", nil)
}

// StartLiveSession starts a live session (manual mode)
func StartLiveSession(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.StartLiveSession [user_id=%s [session_id=%s", userID, sessionID)

	// Call CORE layer
	err = core.StartSession(ctx, sessionID, userID)
	if err != nil {
		commons.LogC(ctx, "[api.StartLiveSession::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.StartLiveSession::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, sessionID)
	broadcastLiveSessionStarted(ctx, sessionID)
	commons.RespOKWithMsg(c, "Session started successfully", nil)
}

// AdvanceLiveQuestion advances to next question (manual mode)
func AdvanceLiveQuestion(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.AdvanceLiveQuestion [user_id=%s [session_id=%s", userID, sessionID)

	// Call CORE layer
	err = core.AdvanceToNextQuestion(ctx, sessionID, userID)
	if err != nil {
		commons.LogC(ctx, "[api.AdvanceLiveQuestion::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.AdvanceLiveQuestion::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, sessionID)
	broadcastLiveQuestionAdvanced(ctx, sessionID)
	commons.RespOKWithMsg(c, "Advanced to next question", nil)
}

// EndLiveSession ends a live session
func EndLiveSession(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.EndLiveSession [user_id=%s [session_id=%s", userID, sessionID)

	// Call CORE layer
	err = core.EndSession(ctx, sessionID, userID)
	if err != nil {
		commons.LogC(ctx, "[api.EndLiveSession::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.EndLiveSession::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, sessionID)
	broadcastLiveSessionEnded(ctx, sessionID)
	commons.RespOKWithMsg(c, "Session ended successfully", nil)
}

// SubmitLiveAnswer submits an answer in a live session
func SubmitLiveAnswer(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	var req model.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.SubmitLiveAnswer::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.SubmitLiveAnswer [user_id=%s [session_id=%s [question_id=%s",
		userID, sessionID, req.QuestionID)

	// Call CORE layer
	response, err := core.SubmitAnswer(ctx, sessionID, userID, req)
	if err != nil {
		commons.LogC(ctx, "[api.SubmitLiveAnswer::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

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

	commons.Log(ctx, "[api.SubmitLiveAnswer::OK [time_taken=%v [ip=%s [points=%d",
		time.Since(st), ip, response.PointsEarned)
	commons.RespOKWithMsg(c, "Answer submitted successfully", response)
}

// GetLiveLeaderboard gets the current leaderboard
func GetLiveLeaderboard(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.GetLiveLeaderboard [user_id=%s [session_id=%s", userID, sessionID)

	// Call CORE layer
	leaderboard, err := core.GetLeaderboard(ctx, sessionID)
	if err != nil {
		commons.LogC(ctx, "[api.GetLiveLeaderboard::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetLiveLeaderboard::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(leaderboard))
	commons.RespOKWithMsg(c, "Leaderboard retrieved successfully", gin.H{
		"leaderboard": leaderboard,
	})
}

// InviteToLiveSession invites users to a live session
func InviteToLiveSession(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	var req model.InviteUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.InviteToLiveSession::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.InviteToLiveSession [user_id=%s [session_id=%s [count=%d",
		userID, sessionID, len(req.UserIDs))

	// Call CORE layer
	err = core.InviteUsers(ctx, sessionID, userID, req.UserIDs, notificationService)
	if err != nil {
		commons.LogC(ctx, "[api.InviteToLiveSession::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.InviteToLiveSession::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(req.UserIDs))
	commons.RespOKWithMsg(c, "Invitations sent successfully", gin.H{
		"invited_count": len(req.UserIDs),
	})
}

// GetLiveSessionResults gets session results
func GetLiveSessionResults(c *gin.Context) {
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

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		commons.RespBadReq(c, "invalid-session-id")
		return
	}

	commons.Log(ctx, "[api.GetLiveSessionResults [user_id=%s [session_id=%s", userID, sessionID)

	// Call CORE layer
	results, err := core.GetSessionResults(ctx, sessionID, userID)
	if err != nil {
		commons.LogC(ctx, "[api.GetLiveSessionResults::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

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

	commons.Log(ctx, "[api.GetLiveSessionResults::OK [time_taken=%v [ip=%s [session_id=%s",
		time.Since(st), ip, sessionID)
	commons.RespOKWithMsg(c, "Session results retrieved successfully", results)
}

// ListMySessions lists sessions created by user
func ListMySessions(c *gin.Context) {
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	commons.Log(ctx, "[api.ListMySessions [user_id=%s", userID)

	// Call CORE layer
	response, err := core.ListMySessions(ctx, userID, limit, offset)
	if err != nil {
		commons.LogC(ctx, "[api.ListMySessions::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ListMySessions::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(response.Sessions))
	commons.RespOKWithMsg(c, "Sessions retrieved successfully", response)
}

// ListPublicSessions lists discoverable public sessions
func ListPublicSessions(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	commons.Log(ctx, "[api.ListPublicSessions")

	// Call CORE layer
	response, err := core.ListPublicSessions(ctx, limit, offset)
	if err != nil {
		commons.LogC(ctx, "[api.ListPublicSessions::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ListPublicSessions::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(response.Sessions))
	commons.RespOKWithMsg(c, "Public sessions retrieved successfully", response)
}

// ListMySessionInvitations lists sessions user is invited to
func ListMySessionInvitations(c *gin.Context) {
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

	commons.Log(ctx, "[api.ListMySessionInvitations [user_id=%s", userID)

	// Call CORE layer
	response, err := core.ListMyInvitations(ctx, userID)
	if err != nil {
		commons.LogC(ctx, "[api.ListMySessionInvitations::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ListMySessionInvitations::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(response.Sessions))
	commons.RespOKWithMsg(c, "Invitations retrieved successfully", response)
}

func broadcastLiveSessionStarted(ctx context.Context, sessionID uuid.UUID) {
	if liveSessionHub == nil {
		return
	}

	session, err := livesessiondao.FindSessionByID(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[api.broadcastLiveSessionStarted::SessionErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	questions, err := liveQuestionsWithAnswers(ctx, session.QuizID)
	if err != nil || len(questions) == 0 {
		commons.LogE(ctx, "[api.broadcastLiveSessionStarted::QuestionsErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	eventData := map[string]interface{}{
		"session_id":      sessionID.String(),
		"total_questions": len(questions),
		"first_question":  questions[0],
		"timer_mode":      session.TimerMode,
	}
	if session.QuestionTimerSeconds.Valid {
		eventData["question_timer_seconds"] = session.QuestionTimerSeconds.Int64
	}
	if session.WholeQuizTimerSeconds.Valid {
		eventData["whole_quiz_timer_seconds"] = session.WholeQuizTimerSeconds.Int64
	}

	liveSessionHub.BroadcastToSession(sessionID, livesessionws.Event{
		Type: livesessionws.EventSessionStarted,
		Data: eventData,
	})
}

func broadcastLiveQuestionAdvanced(ctx context.Context, sessionID uuid.UUID) {
	if liveSessionHub == nil {
		return
	}

	session, err := livesessiondao.FindSessionByID(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[api.broadcastLiveQuestionAdvanced::SessionErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	questions, err := liveQuestionsWithAnswers(ctx, session.QuizID)
	if err != nil || session.CurrentQuestionIndex >= len(questions) {
		commons.LogE(ctx, "[api.broadcastLiveQuestionAdvanced::QuestionsErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	liveSessionHub.BroadcastToSession(sessionID, livesessionws.Event{
		Type: livesessionws.EventQuestionAutoAdvance,
		Data: map[string]interface{}{
			"next_question_index": session.CurrentQuestionIndex,
			"next_question":       questions[session.CurrentQuestionIndex],
		},
	})
}

func broadcastLiveSessionEnded(ctx context.Context, sessionID uuid.UUID) {
	if liveSessionHub == nil {
		return
	}

	leaderboard, _ := livesessiondao.GetLeaderboard(ctx, sessionID)
	liveSessionHub.BroadcastToSession(sessionID, livesessionws.Event{
		Type: livesessionws.EventSessionEnded,
		Data: map[string]interface{}{
			"session_id":        sessionID.String(),
			"final_leaderboard": leaderboard,
		},
	})
}

func reconcileExpiredLiveSession(ctx context.Context, sessionID uuid.UUID) {
	session, err := livesessiondao.FindSessionByID(ctx, sessionID)
	if err != nil || session.Status != model.SessionStatusLive || session.ProgressionMode != model.ProgressionModeAutomatic {
		return
	}

	if session.TimerMode == model.TimerModeWholeQuiz {
		if !session.QuizStartedAt.Valid || !session.WholeQuizTimerSeconds.Valid {
			return
		}

		if liveTimerElapsed(session.QuizStartedAt.Time) >= time.Duration(session.WholeQuizTimerSeconds.Int64)*time.Second {
			if err := core.EndSessionDueToTimeout(ctx, sessionID); err != nil {
				commons.LogE(ctx, "[api.reconcileExpiredLiveSession::EndWholeQuizErr [session_id=%s [err=%v", sessionID, err)
				return
			}
			broadcastLiveSessionEnded(ctx, sessionID)
		}
		return
	}

	if session.TimerMode != model.TimerModePerQuestion || !session.CurrentQuestionStartedAt.Valid || !session.QuestionTimerSeconds.Valid {
		return
	}

	if liveTimerElapsed(session.CurrentQuestionStartedAt.Time) < time.Duration(session.QuestionTimerSeconds.Int64)*time.Second {
		return
	}

	questions, err := liveQuestionsWithAnswers(ctx, session.QuizID)
	if err != nil || len(questions) == 0 {
		commons.LogE(ctx, "[api.reconcileExpiredLiveSession::QuestionsErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	if session.CurrentQuestionIndex >= len(questions)-1 {
		if err := core.EndSessionDueToTimeout(ctx, sessionID); err != nil {
			commons.LogE(ctx, "[api.reconcileExpiredLiveSession::EndPerQuestionErr [session_id=%s [err=%v", sessionID, err)
			return
		}
		broadcastLiveSessionEnded(ctx, sessionID)
		return
	}

	if err := livesessiondao.AdvanceQuestion(ctx, sessionID); err != nil {
		commons.LogE(ctx, "[api.reconcileExpiredLiveSession::AdvanceErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	broadcastLiveQuestionAdvanced(ctx, sessionID)
}

func liveQuestionsWithAnswers(ctx context.Context, quizID uuid.UUID) ([]*quizmodel.Question, error) {
	questions, err := quizdao.ListByQuizID(ctx, quizID)
	if err != nil {
		return questions, err
	}

	for _, question := range questions {
		questionID, err := uuid.Parse(question.ID)
		if err != nil {
			continue
		}

		answers, err := quizdao.ListByQuestionIDHideCorrect(ctx, questionID)
		if err != nil {
			commons.LogE(ctx, "[api.liveQuestionsWithAnswers::AnswersErr [question_id=%s [err=%v", question.ID, err)
			continue
		}
		question.Answers = answers
	}

	return questions, nil
}

func liveTimerElapsed(startedAt time.Time) time.Duration {
	startedAt = normalizeLiveTimerStart(startedAt)
	elapsed := time.Since(startedAt)
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func normalizeLiveTimerStart(startedAt time.Time) time.Time {
	now := time.Now()
	if !startedAt.After(now.Add(2 * time.Minute)) {
		return startedAt
	}

	_, offsetSeconds := now.In(time.Local).Zone()
	if offsetSeconds != 0 {
		adjusted := startedAt.Add(-time.Duration(offsetSeconds) * time.Second)
		if !adjusted.After(now.Add(2 * time.Minute)) {
			return adjusted
		}
	}

	return now
}
