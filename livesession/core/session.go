package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/dao"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
	notificationcore "github.com/affanahmad024/quiz-craft/backend/notification/core"
	notificationmodel "github.com/affanahmad024/quiz-craft/backend/notification/model"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
)

const (
	BasePoints           = 100
	MaxSpeedBonusPercent = 50 // 50% bonus for fastest answers
)

// CreateSession creates a new live session
func CreateSession(ctx context.Context, req model.CreateLiveSessionRequest, creatorID uuid.UUID) (model.LiveSessionAPI, error) {
	commons.Log(ctx, "[core.CreateSession [creator_id=%s [quiz_id=%s", creatorID, req.QuizID)

	// Validate session type
	if req.SessionType != model.SessionTypePublic && req.SessionType != model.SessionTypePrivate {
		return model.LiveSessionAPI{}, commons.BadInput("invalid-session-type")
	}

	// Validate progression mode
	if req.ProgressionMode != model.ProgressionModeManual && req.ProgressionMode != model.ProgressionModeAutomatic {
		return model.LiveSessionAPI{}, commons.BadInput("invalid-progression-mode")
	}

	// Public sessions must be scheduled
	if req.SessionType == model.SessionTypePublic && req.ScheduledAt == nil {
		return model.LiveSessionAPI{}, commons.BadInput("public-sessions-must-be-scheduled")
	}

	// Automatic mode requires timer
	if req.ProgressionMode == model.ProgressionModeAutomatic && req.QuestionTimerSeconds <= 0 {
		req.QuestionTimerSeconds = 30 // Default 30 seconds
	}

	// Verify quiz exists and user has access
	quiz, err := quizdao.FindByID(ctx, req.QuizID)
	if err != nil {
		commons.LogE(ctx, "[core.CreateSession::QuizNotFound [quiz_id=%s [err=%v", req.QuizID, err)
		return model.LiveSessionAPI{}, commons.NotFound("quiz-not-found")
	}

	// Only creator can create sessions for their quiz
	if quiz.CreatorID != creatorID.String() {
		return model.LiveSessionAPI{}, commons.Forbidden("not-quiz-owner")
	}

	// Create session in database
	sessionDAO, err := dao.CreateSession(ctx, req, creatorID)
	if err != nil {
		commons.LogE(ctx, "[core.CreateSession::CreateErr [err=%v", err)
		return model.LiveSessionAPI{}, err
	}

	api := sessionDAO.ToAPI()

	commons.Log(ctx, "[core.CreateSession::OK [session_id=%s", api.ID)
	return api, nil
}

// GetSession gets a session by ID
func GetSession(ctx context.Context, sessionID uuid.UUID) (model.LiveSessionAPI, error) {
	commons.Log(ctx, "[core.GetSession [session_id=%s", sessionID)

	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[core.GetSession::FindErr [session_id=%s [err=%v", sessionID, err)
		return model.LiveSessionAPI{}, err
	}

	api := sessionDAO.ToAPI()

	// Get participant count
	count, _ := dao.GetParticipantCount(ctx, sessionID)
	api.ParticipantCount = count

	return api, nil
}

// JoinSession allows a user to join a session
func JoinSession(ctx context.Context, sessionID, userID uuid.UUID, accessCode string) (model.ParticipantAPI, model.LiveSessionAPI, error) {
	commons.Log(ctx, "[core.JoinSession [session_id=%s [user_id=%s", sessionID, userID)

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return model.ParticipantAPI{}, model.LiveSessionAPI{}, err
	}

	// Check session status
	if sessionDAO.Status != model.SessionStatusWaiting && sessionDAO.Status != model.SessionStatusLive {
		return model.ParticipantAPI{}, model.LiveSessionAPI{}, commons.BadInput("session-not-joinable")
	}

	// Validate access
	if sessionDAO.SessionType == model.SessionTypePrivate {
		// Check access code OR invitation
		if accessCode != "" {
			if !sessionDAO.AccessCode.Valid || sessionDAO.AccessCode.String != accessCode {
				return model.ParticipantAPI{}, model.LiveSessionAPI{}, commons.Forbidden("invalid-access-code")
			}
		} else {
			// Check if user has invitation
			hasInvite, _ := dao.HasInvitation(ctx, sessionID, userID)
			if !hasInvite {
				return model.ParticipantAPI{}, model.LiveSessionAPI{}, commons.Forbidden("not-invited")
			}
		}
	}

	// Check max participants
	if sessionDAO.MaxParticipants.Valid {
		count, _ := dao.GetParticipantCount(ctx, sessionID)
		if count >= int(sessionDAO.MaxParticipants.Int64) {
			return model.ParticipantAPI{}, model.LiveSessionAPI{}, commons.Conflict("session-full")
		}
	}

	// Add participant
	participantDAO, err := dao.AddParticipant(ctx, sessionID, userID)
	if err != nil {
		if commons.IsConflict(err) {
			// Already joined, return existing participant
			participantDAO, _ = dao.FindParticipant(ctx, sessionID, userID)
		} else {
			commons.LogE(ctx, "[core.JoinSession::AddErr [err=%v", err)
			return model.ParticipantAPI{}, model.LiveSessionAPI{}, err
		}
	}

	sessionAPI := sessionDAO.ToAPI()
	participantAPI := participantDAO.ToAPI()

	commons.Log(ctx, "[core.JoinSession::OK [participant_id=%s", participantAPI.ID)
	return participantAPI, sessionAPI, nil
}

// StartSession starts a session (manual mode)
func StartSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	commons.Log(ctx, "[core.StartSession [session_id=%s [user_id=%s", sessionID, userID)

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Only creator can start
	if sessionDAO.CreatorID.String() != userID.String() {
		return commons.Forbidden("not-session-creator")
	}

	// Only manual mode allows manual start
	if sessionDAO.ProgressionMode != model.ProgressionModeManual {
		return commons.BadInput("session-uses-automatic-progression")
	}

	// Check status
	if sessionDAO.Status != model.SessionStatusWaiting {
		return commons.BadInput("session-already-started-or-ended")
	}

	// Start session
	err = dao.StartSession(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[core.StartSession::StartErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[core.StartSession::OK [session_id=%s", sessionID)
	return nil
}

// AdvanceToNextQuestion advances to next question (manual mode)
func AdvanceToNextQuestion(ctx context.Context, sessionID, userID uuid.UUID) error {
	commons.Log(ctx, "[core.AdvanceToNextQuestion [session_id=%s [user_id=%s", sessionID, userID)

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Only creator can advance
	if sessionDAO.CreatorID.String() != userID.String() {
		return commons.Forbidden("not-session-creator")
	}

	// Only manual mode allows manual advance
	if sessionDAO.ProgressionMode != model.ProgressionModeManual {
		return commons.BadInput("session-uses-automatic-progression")
	}

	// Check status
	if sessionDAO.Status != model.SessionStatusLive {
		return commons.BadInput("session-not-live")
	}

	// Advance question
	err = dao.AdvanceQuestion(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[core.AdvanceToNextQuestion::AdvanceErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[core.AdvanceToNextQuestion::OK [session_id=%s", sessionID)
	return nil
}

// EndSession ends a session
func EndSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	commons.Log(ctx, "[core.EndSession [session_id=%s [user_id=%s", sessionID, userID)

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Only creator can end
	if sessionDAO.CreatorID.String() != userID.String() {
		return commons.Forbidden("not-session-creator")
	}

	// Check status
	if sessionDAO.Status != model.SessionStatusLive {
		return commons.BadInput("session-not-live")
	}

	// End session
	err = dao.EndSession(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[core.EndSession::EndErr [err=%v", err)
		return err
	}

	// Mark all participants as completed
	participants, _ := dao.ListParticipants(ctx, sessionID)
	for _, p := range participants {
		if p.Status == model.ParticipantStatusActive {
			dao.UpdateParticipantStatus(ctx, p.ID, model.ParticipantStatusCompleted)
		}
	}

	commons.Log(ctx, "[core.EndSession::OK [session_id=%s", sessionID)
	return nil
}

// SubmitAnswer submits an answer for a live session question
func SubmitAnswer(ctx context.Context, sessionID, userID uuid.UUID, req model.SubmitAnswerRequest) (model.SubmitAnswerResponse, error) {
	commons.Log(ctx, "[core.SubmitAnswer [session_id=%s [user_id=%s [question_id=%s", sessionID, userID, req.QuestionID)

	response := model.SubmitAnswerResponse{}

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return response, err
	}

	// Check session is live
	if sessionDAO.Status != model.SessionStatusLive {
		return response, commons.BadInput("session-not-live")
	}

	// Get participant
	participantDAO, err := dao.FindParticipant(ctx, sessionID, userID)
	if err != nil {
		return response, commons.NotFound("not-a-participant")
	}

	// Check if already answered
	hasAnswered, _ := dao.HasAnsweredQuestion(ctx, sessionID, participantDAO.ID, req.QuestionID)
	if hasAnswered {
		return response, commons.Conflict("already-answered")
	}

	// Get question and validate
	_, err = quizdao.FindQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return response, commons.NotFound("question-not-found")
	}

	// Check answer is correct
	isCorrect := false
	if req.AnswerID != nil {
		answer, err := quizdao.FindAnswerByID(ctx, *req.AnswerID)
		if err == nil && answer.IsCorrect {
			isCorrect = true
		}
	}

	// Calculate points with speed bonus
	pointsEarned := 0
	if isCorrect {
		pointsEarned = CalculatePoints(req.TimeTakenMs, sessionDAO.QuestionTimerSeconds*1000)
	}

	// Save answer
	_, err = dao.SaveAnswer(ctx, sessionID, participantDAO.ID, req.QuestionID, req.AnswerID, isCorrect, pointsEarned, req.TimeTakenMs)
	if err != nil {
		commons.LogE(ctx, "[core.SubmitAnswer::SaveErr [err=%v", err)
		return response, err
	}

	// Update participant score
	err = dao.UpdateParticipantScore(ctx, participantDAO.ID, pointsEarned)
	if err != nil {
		commons.LogE(ctx, "[core.SubmitAnswer::UpdateScoreErr [err=%v", err)
	}

	// Get updated participant info
	updatedParticipant, _ := dao.FindParticipantByID(ctx, participantDAO.ID)

	response.IsCorrect = isCorrect
	response.PointsEarned = pointsEarned
	response.CurrentScore = updatedParticipant.CurrentScore
	response.Rank = updatedParticipant.Rank

	commons.Log(ctx, "[core.SubmitAnswer::OK [participant_id=%s [points=%d", participantDAO.ID, pointsEarned)
	return response, nil
}

// GetLeaderboard gets the current leaderboard
func GetLeaderboard(ctx context.Context, sessionID uuid.UUID) ([]model.LeaderboardEntry, error) {
	commons.Log(ctx, "[core.GetLeaderboard [session_id=%s", sessionID)

	leaderboard, err := dao.GetLeaderboard(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[core.GetLeaderboard::GetErr [err=%v", err)
		return []model.LeaderboardEntry{}, err
	}

	return leaderboard, nil
}

// InviteUsers invites users to a private session
func InviteUsers(ctx context.Context, sessionID, inviterID uuid.UUID, userIDs []uuid.UUID, notificationService *notificationcore.NotificationService) error {
	commons.Log(ctx, "[core.InviteUsers [session_id=%s [inviter_id=%s [count=%d", sessionID, inviterID, len(userIDs))

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Only creator can invite
	if sessionDAO.CreatorID.String() != inviterID.String() {
		return commons.Forbidden("not-session-creator")
	}

	// Only private sessions can have invitations
	if sessionDAO.SessionType != model.SessionTypePrivate {
		return commons.BadInput("only-private-sessions-can-invite")
	}

	// Create invitations
	err = dao.CreateBulkInvitations(ctx, sessionID, inviterID, userIDs)
	if err != nil {
		commons.LogE(ctx, "[core.InviteUsers::CreateErr [err=%v", err)
		return err
	}

	// Send notifications if service is available
	if notificationService != nil {
		go func() {
			inviterName := "A user" // TODO: Get actual inviter name
			title := fmt.Sprintf("%s invited you to a live quiz", inviterName)
			message := fmt.Sprintf("Join '%s' and compete with your friends!", sessionDAO.Title)

			data := map[string]interface{}{
				"session_id":  sessionDAO.ID.String(),
				"inviter_id":  inviterID.String(),
				"access_code": sessionDAO.AccessCode.String,
			}

			if sessionDAO.ScheduledAt.Valid {
				data["scheduled_at"] = sessionDAO.ScheduledAt.Time.Format("2006-01-02T15:04:05Z07:00")
			}

			notificationService.CreateBulkNotifications(
				context.Background(),
				userIDs,
				notificationmodel.NotificationTypeLiveSessionInvitation,
				title,
				message,
				data,
			)
		}()
	}

	commons.Log(ctx, "[core.InviteUsers::OK [count=%d", len(userIDs))
	return nil
}

// CalculatePoints calculates points based on answer speed
func CalculatePoints(timeTakenMs, maxTimeMs int) int {
	if timeTakenMs > maxTimeMs {
		timeTakenMs = maxTimeMs
	}

	// Calculate speed bonus (0-50%)
	timeRatio := float64(timeTakenMs) / float64(maxTimeMs)
	speedBonus := (1.0 - timeRatio) * (float64(MaxSpeedBonusPercent) / 100.0)

	points := int(float64(BasePoints) * (1.0 + speedBonus))
	return points
}
