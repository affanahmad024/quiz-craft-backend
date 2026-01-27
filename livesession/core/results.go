package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/dao"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
)

// GetSessionResults gets detailed results after session ends
func GetSessionResults(ctx context.Context, sessionID, userID uuid.UUID) (model.SessionResultsResponse, error) {
	commons.Log(ctx, "[core.GetSessionResults [session_id=%s [user_id=%s", sessionID, userID)

	response := model.SessionResultsResponse{}

	// Get session
	sessionDAO, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return response, err
	}

	// Session should be completed
	if sessionDAO.Status != model.SessionStatusCompleted {
		return response, commons.BadInput("session-not-completed")
	}

	response.Session = sessionDAO.ToAPI()

	// Get final leaderboard
	leaderboard, err := dao.GetLeaderboard(ctx, sessionID)
	if err == nil {
		response.FinalLeaderboard = leaderboard
		response.TotalParticipants = len(leaderboard)
	}

	// Get my performance if I'm a participant
	myParticipant, err := dao.FindParticipant(ctx, sessionID, userID)
	if err == nil {
		// Get my answers
		myAnswers, _ := dao.GetParticipantAnswers(ctx, sessionID, myParticipant.ID)
		participantAnswers := []model.ParticipantAnswer{}

		for _, answer := range myAnswers {
			// Get question details
			question, qErr := quizdao.FindQuestionByID(ctx, answer.QuestionID)
			if qErr != nil {
				continue
			}

			participantAnswer := model.ParticipantAnswer{
				QuestionID:   answer.QuestionID,
				QuestionText: question.QuestionText,
				IsCorrect:    answer.IsCorrect,
				PointsEarned: answer.PointsEarned,
				TimeTakenMs:  answer.TimeTakenMs,
			}

			// Get answer text if answered
			if answer.AnswerID.Valid {
				answerUUID, _ := uuid.Parse(answer.AnswerID.String)
				participantAnswer.AnswerID = &answerUUID

				answerObj, aErr := quizdao.FindAnswerByID(ctx, answerUUID)
				if aErr == nil {
					participantAnswer.AnswerText = answerObj.AnswerText
				}
			}

			// Get correct answer
			correctAnswers, cErr := quizdao.ListByQuestionID(ctx, answer.QuestionID)
			if cErr == nil {
				for _, ans := range correctAnswers {
					if ans.IsCorrect {
						participantAnswer.CorrectAnswer = ans.AnswerText
						break
					}
				}
			}

			participantAnswers = append(participantAnswers, participantAnswer)
		}

		myPerformance := &model.ParticipantDetails{
			Participant: myParticipant.ToAPI(),
			Answers:     participantAnswers,
		}
		response.MyPerformance = myPerformance
	}

	// Get question stats for creator or all participants in private sessions
	canSeeStats := false
	if sessionDAO.CreatorID.String() == userID.String() {
		canSeeStats = true
	} else if sessionDAO.SessionType == model.SessionTypePrivate {
		// Private sessions: all participants can see stats
		_, err := dao.FindParticipant(ctx, sessionID, userID)
		if err == nil {
			canSeeStats = true
		}
	}

	if canSeeStats {
		// Get quiz questions
		questions, _ := quizdao.ListByQuizID(ctx, sessionDAO.QuizID)
		questionStats := []model.QuestionStat{}

		for _, question := range questions {
			questionID, err := uuid.Parse(question.ID)
			if err != nil {
				continue
			}

			totalAnswers, correctAnswers, avgTimeMs, err := dao.GetQuestionStats(ctx, sessionID, questionID)
			if err != nil {
				continue
			}

			correctRate := 0.0
			if totalAnswers > 0 {
				correctRate = float64(correctAnswers) / float64(totalAnswers) * 100.0
			}

			stat := model.QuestionStat{
				QuestionID:     questionID,
				QuestionText:   question.QuestionText,
				TotalAnswers:   totalAnswers,
				CorrectAnswers: correctAnswers,
				CorrectRate:    correctRate,
				AverageTimeMs:  avgTimeMs,
			}
			questionStats = append(questionStats, stat)
		}

		response.QuestionStats = questionStats
	}

	commons.Log(ctx, "[core.GetSessionResults::OK [session_id=%s", sessionID)
	return response, nil
}

// ListMySessions lists sessions created by user
func ListMySessions(ctx context.Context, userID uuid.UUID, limit, offset int) (model.SessionListResponse, error) {
	commons.Log(ctx, "[core.ListMySessions [user_id=%s", userID)

	filter := model.SessionListFilter{
		CreatorID: &userID,
		Limit:     limit,
		Offset:    offset,
	}

	sessions, total, err := dao.ListSessions(ctx, filter)
	if err != nil {
		return model.SessionListResponse{}, err
	}

	apiSessions := make([]model.LiveSessionAPI, 0, len(sessions))
	for _, s := range sessions {
		apiSessions = append(apiSessions, s.ToAPI())
	}

	return model.SessionListResponse{
		Sessions: apiSessions,
		Total:    total,
	}, nil
}

// ListPublicSessions lists discoverable public sessions
func ListPublicSessions(ctx context.Context, limit, offset int) (model.SessionListResponse, error) {
	commons.Log(ctx, "[core.ListPublicSessions")

	sessions, total, err := dao.ListDiscoverableSessions(ctx, limit, offset)
	if err != nil {
		return model.SessionListResponse{}, err
	}

	apiSessions := make([]model.LiveSessionAPI, 0, len(sessions))
	for _, s := range sessions {
		api := s.ToAPI()
		// Add participant count
		count, _ := dao.GetParticipantCount(ctx, s.ID)
		api.ParticipantCount = count
		apiSessions = append(apiSessions, api)
	}

	return model.SessionListResponse{
		Sessions: apiSessions,
		Total:    total,
	}, nil
}

// ListMyInvitations lists sessions user is invited to
func ListMyInvitations(ctx context.Context, userID uuid.UUID) (model.SessionListResponse, error) {
	commons.Log(ctx, "[core.ListMyInvitations [user_id=%s", userID)

	// Get user invitations
	invitations, err := dao.ListUserInvitations(ctx, userID)
	if err != nil {
		return model.SessionListResponse{}, err
	}

	apiSessions := make([]model.LiveSessionAPI, 0, len(invitations))
	for _, inv := range invitations {
		// Get session details
		session, err := dao.FindSessionByID(ctx, inv.SessionID)
		if err != nil {
			continue
		}
		apiSessions = append(apiSessions, session.ToAPI())
	}

	return model.SessionListResponse{
		Sessions: apiSessions,
		Total:    len(apiSessions),
	}, nil
}
