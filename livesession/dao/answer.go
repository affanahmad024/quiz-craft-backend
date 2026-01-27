package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
)

// SaveAnswer saves a participant's answer
func SaveAnswer(ctx context.Context, sessionID, participantID, questionID uuid.UUID, answerID *uuid.UUID, isCorrect bool, pointsEarned, timeTakenMs int) (model.SessionAnswerDAO, error) {
	commons.Log(ctx, "[dao.SaveAnswer [session_id=%s [participant_id=%s [question_id=%s", sessionID, participantID, questionID)

	answer := model.SessionAnswerDAO{}

	var answerIDSQL sql.NullString
	if answerID != nil {
		answerIDSQL = sql.NullString{String: answerID.String(), Valid: true}
	}

	qry := `INSERT INTO live_session_answers (session_id, participant_id, question_id, answer_id, is_correct, points_earned, time_taken_ms, answered_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			RETURNING id, session_id, participant_id, question_id, answer_id, is_correct, points_earned, time_taken_ms, answered_at, created_at`

	err := db.QueryRowContext(ctx, qry, sessionID, participantID, questionID, answerIDSQL, isCorrect, pointsEarned, timeTakenMs).Scan(
		&answer.ID,
		&answer.SessionID,
		&answer.ParticipantID,
		&answer.QuestionID,
		&answer.AnswerID,
		&answer.IsCorrect,
		&answer.PointsEarned,
		&answer.TimeTakenMs,
		&answer.AnsweredAt,
		&answer.CreatedAt,
	)

	if err != nil {
		// Check for duplicate
		if err.Error() == "pq: duplicate key value violates unique constraint \"live_session_answers_session_id_participant_id_question_id_key\"" {
			return answer, commons.Conflict("already-answered")
		}
		commons.LogE(ctx, "[dao.SaveAnswer::SQLErr [err=%v", err)
		return answer, err
	}

	commons.Log(ctx, "[dao.SaveAnswer::OK [answer_id=%s", answer.ID)
	return answer, nil
}

// GetParticipantAnswers gets all answers for a participant in a session
func GetParticipantAnswers(ctx context.Context, sessionID, participantID uuid.UUID) ([]model.SessionAnswerDAO, error) {
	commons.Log(ctx, "[dao.GetParticipantAnswers [session_id=%s [participant_id=%s", sessionID, participantID)

	answers := []model.SessionAnswerDAO{}
	qry := `SELECT id, session_id, participant_id, question_id, answer_id, is_correct, points_earned, time_taken_ms, answered_at, created_at
			FROM live_session_answers
			WHERE session_id = $1 AND participant_id = $2
			ORDER BY answered_at ASC`

	rows, err := db.QueryContext(ctx, qry, sessionID, participantID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetParticipantAnswers::SQLErr [err=%v", err)
		return answers, err
	}
	defer rows.Close()

	for rows.Next() {
		var answer model.SessionAnswerDAO
		err := rows.Scan(
			&answer.ID,
			&answer.SessionID,
			&answer.ParticipantID,
			&answer.QuestionID,
			&answer.AnswerID,
			&answer.IsCorrect,
			&answer.PointsEarned,
			&answer.TimeTakenMs,
			&answer.AnsweredAt,
			&answer.CreatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetParticipantAnswers::ScanErr [err=%v", err)
			continue
		}
		answers = append(answers, answer)
	}

	commons.Log(ctx, "[dao.GetParticipantAnswers::OK [count=%d", len(answers))
	return answers, nil
}

// GetSessionAnswers gets all answers for a session
func GetSessionAnswers(ctx context.Context, sessionID uuid.UUID) ([]model.SessionAnswerDAO, error) {
	commons.Log(ctx, "[dao.GetSessionAnswers [session_id=%s", sessionID)

	answers := []model.SessionAnswerDAO{}
	qry := `SELECT id, session_id, participant_id, question_id, answer_id, is_correct, points_earned, time_taken_ms, answered_at, created_at
			FROM live_session_answers
			WHERE session_id = $1
			ORDER BY answered_at ASC`

	rows, err := db.QueryContext(ctx, qry, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetSessionAnswers::SQLErr [err=%v", err)
		return answers, err
	}
	defer rows.Close()

	for rows.Next() {
		var answer model.SessionAnswerDAO
		err := rows.Scan(
			&answer.ID,
			&answer.SessionID,
			&answer.ParticipantID,
			&answer.QuestionID,
			&answer.AnswerID,
			&answer.IsCorrect,
			&answer.PointsEarned,
			&answer.TimeTakenMs,
			&answer.AnsweredAt,
			&answer.CreatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetSessionAnswers::ScanErr [err=%v", err)
			continue
		}
		answers = append(answers, answer)
	}

	commons.Log(ctx, "[dao.GetSessionAnswers::OK [count=%d", len(answers))
	return answers, nil
}

// HasAnsweredQuestion checks if participant has already answered a question
func HasAnsweredQuestion(ctx context.Context, sessionID, participantID, questionID uuid.UUID) (bool, error) {
	var exists bool
	qry := `SELECT EXISTS(SELECT 1 FROM live_session_answers WHERE session_id = $1 AND participant_id = $2 AND question_id = $3)`

	err := db.QueryRowContext(ctx, qry, sessionID, participantID, questionID).Scan(&exists)
	if err != nil {
		commons.LogE(ctx, "[dao.HasAnsweredQuestion::SQLErr [err=%v", err)
		return false, err
	}

	return exists, nil
}

// GetQuestionStats gets statistics for a question in a session
func GetQuestionStats(ctx context.Context, sessionID, questionID uuid.UUID) (int, int, int, error) {
	var totalAnswers, correctAnswers, avgTimeMs int

	qry := `SELECT
				COUNT(*) as total_answers,
				SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) as correct_answers,
				COALESCE(AVG(time_taken_ms), 0) as avg_time_ms
			FROM live_session_answers
			WHERE session_id = $1 AND question_id = $2`

	err := db.QueryRowContext(ctx, qry, sessionID, questionID).Scan(&totalAnswers, &correctAnswers, &avgTimeMs)
	if err != nil {
		commons.LogE(ctx, "[dao.GetQuestionStats::SQLErr [err=%v", err)
		return 0, 0, 0, err
	}

	return totalAnswers, correctAnswers, avgTimeMs, nil
}
