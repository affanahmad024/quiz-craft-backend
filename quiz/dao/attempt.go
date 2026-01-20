package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// CreateAttempt creates a new quiz attempt
func CreateAttempt(ctx context.Context, attempt *model.QuizAttempt) error {
	commons.Log(ctx, "[dao.CreateAttempt [user_id=%s [quiz_id=%s", attempt.UserID, attempt.QuizID)

	metaJSON, _ := json.Marshal(attempt.Meta)
	qry := `INSERT INTO quiz_attempts (
				id, quiz_id, user_id, status, started_at,
				time_limit_seconds, expires_at, max_score, meta,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
			) RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		attempt.ID,
		attempt.QuizID,
		attempt.UserID,
		attempt.Status,
		attempt.StartedAt,
		attempt.TimeLimitSeconds,
		attempt.ExpiresAt,
		attempt.MaxScore,
		metaJSON,
	).Scan(&attempt.CreatedAt, &attempt.UpdatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateAttempt::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateAttempt::OK [attempt_id=%s", attempt.ID)
	return nil
}

// FindAttemptByID finds an attempt by ID
func FindAttemptByID(ctx context.Context, id uuid.UUID) (*model.QuizAttempt, error) {
	commons.Log(ctx, "[dao.FindAttemptByID [id=%s", id)

	attempt := &model.QuizAttempt{}
	qry := `SELECT id, quiz_id, user_id, status, started_at, submitted_at,
			time_limit_seconds, expires_at, score, max_score, percentage, passed,
			meta, created_at, updated_at
			FROM quiz_attempts
			WHERE id = $1`

	var submittedAtSQL sql.NullTime
	var timeLimitSQL sql.NullInt32
	var expiresAtSQL sql.NullTime
	var scoreSQL sql.NullFloat64
	var maxScoreSQL sql.NullInt32
	var percentageSQL sql.NullFloat64
	var passedSQL sql.NullBool
	var metaJSON []byte

	err := db.QueryRowContext(ctx, qry, id).Scan(
		&attempt.ID,
		&attempt.QuizID,
		&attempt.UserID,
		&attempt.Status,
		&attempt.StartedAt,
		&submittedAtSQL,
		&timeLimitSQL,
		&expiresAtSQL,
		&scoreSQL,
		&maxScoreSQL,
		&percentageSQL,
		&passedSQL,
		&metaJSON,
		&attempt.CreatedAt,
		&attempt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("attempt-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindAttemptByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	// Handle nullable fields
	if submittedAtSQL.Valid {
		attempt.SubmittedAt = &submittedAtSQL.Time
	}
	if timeLimitSQL.Valid {
		timeLimit := int(timeLimitSQL.Int32)
		attempt.TimeLimitSeconds = &timeLimit
	}
	if expiresAtSQL.Valid {
		attempt.ExpiresAt = &expiresAtSQL.Time
	}
	if scoreSQL.Valid {
		attempt.Score = &scoreSQL.Float64
	}
	if maxScoreSQL.Valid {
		maxScore := int(maxScoreSQL.Int32)
		attempt.MaxScore = &maxScore
	}
	if percentageSQL.Valid {
		attempt.Percentage = &percentageSQL.Float64
	}
	if passedSQL.Valid {
		attempt.Passed = &passedSQL.Bool
	}

	// Parse meta JSON
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &attempt.Meta)
	}

	return attempt, nil
}

// UpdateAttemptStatus updates the status of an attempt
func UpdateAttemptStatus(ctx context.Context, attemptID uuid.UUID, status string) error {
	commons.Log(ctx, "[dao.UpdateAttemptStatus [attempt_id=%s [status=%s", attemptID, status)

	qry := `UPDATE quiz_attempts
			SET status = $1, updated_at = NOW()
			WHERE id = $2`

	result, err := db.ExecContext(ctx, qry, status, attemptID)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateAttemptStatus::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("attempt-not-found")
	}

	return nil
}

// FinishAttempt updates attempt with final score and status
func FinishAttempt(ctx context.Context, attempt *model.QuizAttempt) error {
	commons.Log(ctx, "[dao.FinishAttempt [attempt_id=%s [status=%s", attempt.ID, attempt.Status)

	qry := `UPDATE quiz_attempts
			SET status = $1, submitted_at = $2, score = $3,
				max_score = $4, percentage = $5, passed = $6, updated_at = NOW()
			WHERE id = $7`

	result, err := db.ExecContext(ctx, qry,
		attempt.Status,
		attempt.SubmittedAt,
		attempt.Score,
		attempt.MaxScore,
		attempt.Percentage,
		attempt.Passed,
		attempt.ID,
	)
	if err != nil {
		commons.LogE(ctx, "[dao.FinishAttempt::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("attempt-not-found")
	}

	commons.Log(ctx, "[dao.FinishAttempt::OK [attempt_id=%s", attempt.ID)
	return nil
}

// ListUserAttempts lists all attempts for a user with optional filters
func ListUserAttempts(ctx context.Context, userID uuid.UUID, filter model.AttemptListFilter) ([]*model.AttemptListItem, int, error) {
	commons.Log(ctx, "[dao.ListUserAttempts [user_id=%s [page=%d [limit=%d", userID, filter.Page, filter.Limit)

	// Build WHERE clause
	whereClause := "WHERE a.user_id = $1"
	args := []interface{}{userID}
	argCount := 1

	if filter.QuizID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND a.quiz_id = $%d", argCount)
		quizUUID, _ := uuid.Parse(*filter.QuizID)
		args = append(args, quizUUID)
	}

	if filter.Status != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND a.status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	// Count total
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM quiz_attempts a %s", whereClause)
	var total int
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUserAttempts::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Query attempts with quiz info
	offset := (filter.Page - 1) * filter.Limit
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	qry := fmt.Sprintf(`
		SELECT a.id, a.quiz_id, q.title, a.status, a.started_at, a.submitted_at,
			   a.score, a.max_score, a.percentage, a.passed
		FROM quiz_attempts a
		JOIN quizzes q ON a.quiz_id = q.id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitArg, offsetArg)

	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUserAttempts::QueryErr [err=%v", err)
		return nil, 0, err
	}
	defer rows.Close()

	attempts := make([]*model.AttemptListItem, 0)
	for rows.Next() {
		item := &model.AttemptListItem{}
		var submittedAtSQL sql.NullTime
		var scoreSQL sql.NullFloat64
		var maxScoreSQL sql.NullInt32
		var percentageSQL sql.NullFloat64
		var passedSQL sql.NullBool

		err := rows.Scan(
			&item.ID,
			&item.QuizID,
			&item.QuizTitle,
			&item.Status,
			&item.StartedAt,
			&submittedAtSQL,
			&scoreSQL,
			&maxScoreSQL,
			&percentageSQL,
			&passedSQL,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListUserAttempts::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if submittedAtSQL.Valid {
			item.SubmittedAt = &submittedAtSQL.Time
		}
		if scoreSQL.Valid {
			item.Score = &scoreSQL.Float64
		}
		if maxScoreSQL.Valid {
			maxScore := int(maxScoreSQL.Int32)
			item.MaxScore = &maxScore
		}
		if percentageSQL.Valid {
			item.Percentage = &percentageSQL.Float64
		}
		if passedSQL.Valid {
			item.Passed = &passedSQL.Bool
		}

		attempts = append(attempts, item)
	}

	commons.Log(ctx, "[dao.ListUserAttempts::OK [count=%d [total=%d", len(attempts), total)
	return attempts, total, nil
}

// GetQuizStats gets statistics for a quiz
func GetQuizStats(ctx context.Context, quizID uuid.UUID) (*model.QuizStats, error) {
	commons.Log(ctx, "[dao.GetQuizStats [quiz_id=%s", quizID)

	stats := &model.QuizStats{
		QuizID: quizID.String(),
	}

	qry := `SELECT COUNT(*) as total_attempts,
				   COALESCE(AVG(percentage), 0) as average_score
			FROM quiz_attempts
			WHERE quiz_id = $1 AND status = 'submitted'`

	var avgScore sql.NullFloat64
	err := db.QueryRowContext(ctx, qry, quizID).Scan(&stats.TotalAttempts, &avgScore)
	if err != nil {
		commons.LogE(ctx, "[dao.GetQuizStats::SQLErr [err=%v", err)
		return nil, err
	}

	if avgScore.Valid {
		stats.AverageScore = &avgScore.Float64
	}

	commons.Log(ctx, "[dao.GetQuizStats::OK [total=%d", stats.TotalAttempts)
	return stats, nil
}

// CreateOrUpdateUserAnswer creates or updates a user's answer (upsert)
func CreateOrUpdateUserAnswer(ctx context.Context, answer *model.UserAnswer) error {
	commons.Log(ctx, "[dao.CreateOrUpdateUserAnswer [attempt_id=%s [question_id=%s", answer.AttemptID, answer.QuestionID)

	metaJSON, _ := json.Marshal(answer.Meta)
	qry := `INSERT INTO user_answers (
				id, attempt_id, question_id, answer_id, answer_text,
				is_correct, points_earned, answered_at, meta,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
			)
			ON CONFLICT (attempt_id, question_id)
			DO UPDATE SET
				answer_id = EXCLUDED.answer_id,
				answer_text = EXCLUDED.answer_text,
				is_correct = EXCLUDED.is_correct,
				points_earned = EXCLUDED.points_earned,
				answered_at = EXCLUDED.answered_at,
				meta = EXCLUDED.meta,
				updated_at = NOW()
			RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		answer.ID,
		answer.AttemptID,
		answer.QuestionID,
		answer.AnswerID,
		answer.AnswerText,
		answer.IsCorrect,
		answer.PointsEarned,
		answer.AnsweredAt,
		metaJSON,
	).Scan(&answer.CreatedAt, &answer.UpdatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateOrUpdateUserAnswer::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateOrUpdateUserAnswer::OK [user_answer_id=%s", answer.ID)
	return nil
}

// GetUserAnswersByAttempt gets all user answers for an attempt
func GetUserAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*model.UserAnswer, error) {
	commons.Log(ctx, "[dao.GetUserAnswersByAttempt [attempt_id=%s", attemptID)

	qry := `SELECT id, attempt_id, question_id, answer_id, answer_text,
			is_correct, points_earned, answered_at, meta, created_at, updated_at
			FROM user_answers
			WHERE attempt_id = $1
			ORDER BY answered_at ASC`

	rows, err := db.QueryContext(ctx, qry, attemptID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetUserAnswersByAttempt::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	answers := make([]*model.UserAnswer, 0)
	for rows.Next() {
		answer := &model.UserAnswer{}
		var answerIDSQL sql.NullString
		var answerTextSQL sql.NullString
		var isCorrectSQL sql.NullBool
		var metaJSON []byte

		err := rows.Scan(
			&answer.ID,
			&answer.AttemptID,
			&answer.QuestionID,
			&answerIDSQL,
			&answerTextSQL,
			&isCorrectSQL,
			&answer.PointsEarned,
			&answer.AnsweredAt,
			&metaJSON,
			&answer.CreatedAt,
			&answer.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetUserAnswersByAttempt::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if answerIDSQL.Valid {
			answer.AnswerID = &answerIDSQL.String
		}
		if answerTextSQL.Valid {
			answer.AnswerText = &answerTextSQL.String
		}
		if isCorrectSQL.Valid {
			answer.IsCorrect = &isCorrectSQL.Bool
		}

		// Parse meta JSON
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &answer.Meta)
		}

		answers = append(answers, answer)
	}

	commons.Log(ctx, "[dao.GetUserAnswersByAttempt::OK [count=%d", len(answers))
	return answers, nil
}

// GetAttemptResult gets full attempt result with answers and details
func GetAttemptResult(ctx context.Context, attemptID uuid.UUID) (*model.AttemptResult, error) {
	commons.Log(ctx, "[dao.GetAttemptResult [attempt_id=%s", attemptID)

	// Get attempt with quiz info
	result := &model.AttemptResult{}
	qry := `SELECT a.id, a.quiz_id, a.user_id, a.status, a.started_at, a.submitted_at,
			a.time_limit_seconds, a.expires_at, a.score, a.max_score, a.percentage, a.passed,
			a.meta, a.created_at, a.updated_at,
			q.title, q.description
			FROM quiz_attempts a
			JOIN quizzes q ON a.quiz_id = q.id
			WHERE a.id = $1`

	var submittedAtSQL sql.NullTime
	var timeLimitSQL sql.NullInt32
	var expiresAtSQL sql.NullTime
	var scoreSQL sql.NullFloat64
	var maxScoreSQL sql.NullInt32
	var percentageSQL sql.NullFloat64
	var passedSQL sql.NullBool
	var metaJSON []byte
	var descSQL sql.NullString

	err := db.QueryRowContext(ctx, qry, attemptID).Scan(
		&result.ID,
		&result.QuizID,
		&result.UserID,
		&result.Status,
		&result.StartedAt,
		&submittedAtSQL,
		&timeLimitSQL,
		&expiresAtSQL,
		&scoreSQL,
		&maxScoreSQL,
		&percentageSQL,
		&passedSQL,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.QuizTitle,
		&descSQL,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("attempt-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.GetAttemptResult::QueryErr [err=%v", err)
		return nil, err
	}

	// Handle nullable fields
	if submittedAtSQL.Valid {
		result.SubmittedAt = &submittedAtSQL.Time
	}
	if timeLimitSQL.Valid {
		timeLimit := int(timeLimitSQL.Int32)
		result.TimeLimitSeconds = &timeLimit
	}
	if expiresAtSQL.Valid {
		result.ExpiresAt = &expiresAtSQL.Time
	}
	if scoreSQL.Valid {
		result.Score = &scoreSQL.Float64
	}
	if maxScoreSQL.Valid {
		maxScore := int(maxScoreSQL.Int32)
		result.MaxScore = &maxScore
	}
	if percentageSQL.Valid {
		result.Percentage = &percentageSQL.Float64
	}
	if passedSQL.Valid {
		result.Passed = &passedSQL.Bool
	}
	if descSQL.Valid {
		result.QuizDescription = descSQL.String
	}

	// Parse meta JSON
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &result.Meta)
	}

	// Get user answers with details
	answersQry := `SELECT
		ua.id, ua.attempt_id, ua.question_id, ua.answer_id, ua.answer_text,
		ua.is_correct, ua.points_earned, ua.answered_at, ua.meta, ua.created_at, ua.updated_at,
		q.question_text, q.question_type, q.points, q.explanation,
		sa.answer_text as selected_answer_text,
		ca.answer_text as correct_answer_text
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		LEFT JOIN answers sa ON ua.answer_id = sa.id
		LEFT JOIN answers ca ON q.id = ca.question_id AND ca.is_correct = true
		WHERE ua.attempt_id = $1
		ORDER BY q.order_index ASC`

	rows, err := db.QueryContext(ctx, answersQry, attemptID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetAttemptResult::AnswersQueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	answers := make([]model.UserAnswerWithDetails, 0)
	for rows.Next() {
		answer := model.UserAnswerWithDetails{}
		var answerIDSQL sql.NullString
		var answerTextSQL sql.NullString
		var isCorrectSQL sql.NullBool
		var answerMetaJSON []byte
		var explanationSQL sql.NullString
		var selectedAnswerTextSQL sql.NullString
		var correctAnswerTextSQL sql.NullString

		err := rows.Scan(
			&answer.ID,
			&answer.AttemptID,
			&answer.QuestionID,
			&answerIDSQL,
			&answerTextSQL,
			&isCorrectSQL,
			&answer.PointsEarned,
			&answer.AnsweredAt,
			&answerMetaJSON,
			&answer.CreatedAt,
			&answer.UpdatedAt,
			&answer.QuestionText,
			&answer.QuestionType,
			&answer.QuestionPoints,
			&explanationSQL,
			&selectedAnswerTextSQL,
			&correctAnswerTextSQL,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetAttemptResult::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if answerIDSQL.Valid {
			answer.AnswerID = &answerIDSQL.String
		}
		if answerTextSQL.Valid {
			answer.AnswerText = &answerTextSQL.String
		}
		if isCorrectSQL.Valid {
			answer.IsCorrect = &isCorrectSQL.Bool
		}
		if explanationSQL.Valid {
			answer.Explanation = &explanationSQL.String
		}
		if selectedAnswerTextSQL.Valid {
			answer.SelectedAnswerText = &selectedAnswerTextSQL.String
		}
		if correctAnswerTextSQL.Valid {
			answer.CorrectAnswerText = &correctAnswerTextSQL.String
		}

		// Parse meta JSON
		if len(answerMetaJSON) > 0 {
			json.Unmarshal(answerMetaJSON, &answer.Meta)
		}

		answers = append(answers, answer)
	}

	result.Answers = answers
	result.TotalQuestions = len(answers)

	// Count answered questions
	answeredCount := 0
	for _, a := range answers {
		if a.AnswerID != nil || a.AnswerText != nil {
			answeredCount++
		}
	}
	result.AnsweredCount = answeredCount

	commons.Log(ctx, "[dao.GetAttemptResult::OK [answers=%d", len(answers))
	return result, nil
}

// GetExpiredAttempts gets all in-progress attempts that have expired
func GetExpiredAttempts(ctx context.Context, limit int) ([]*model.QuizAttempt, error) {
	commons.Log(ctx, "[dao.GetExpiredAttempts [limit=%d", limit)

	qry := `SELECT id, quiz_id, user_id, status, started_at, submitted_at,
			time_limit_seconds, expires_at, score, max_score, percentage, passed,
			meta, created_at, updated_at
			FROM quiz_attempts
			WHERE status = 'in_progress'
			  AND expires_at IS NOT NULL
			  AND expires_at < NOW()
			ORDER BY expires_at ASC
			LIMIT $1`

	rows, err := db.QueryContext(ctx, qry, limit)
	if err != nil {
		commons.LogE(ctx, "[dao.GetExpiredAttempts::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	attempts := make([]*model.QuizAttempt, 0)
	for rows.Next() {
		attempt := &model.QuizAttempt{}
		var submittedAtSQL sql.NullTime
		var timeLimitSQL sql.NullInt32
		var expiresAtSQL sql.NullTime
		var scoreSQL sql.NullFloat64
		var maxScoreSQL sql.NullInt32
		var percentageSQL sql.NullFloat64
		var passedSQL sql.NullBool
		var metaJSON []byte

		err := rows.Scan(
			&attempt.ID,
			&attempt.QuizID,
			&attempt.UserID,
			&attempt.Status,
			&attempt.StartedAt,
			&submittedAtSQL,
			&timeLimitSQL,
			&expiresAtSQL,
			&scoreSQL,
			&maxScoreSQL,
			&percentageSQL,
			&passedSQL,
			&metaJSON,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetExpiredAttempts::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if submittedAtSQL.Valid {
			attempt.SubmittedAt = &submittedAtSQL.Time
		}
		if timeLimitSQL.Valid {
			timeLimit := int(timeLimitSQL.Int32)
			attempt.TimeLimitSeconds = &timeLimit
		}
		if expiresAtSQL.Valid {
			attempt.ExpiresAt = &expiresAtSQL.Time
		}
		if scoreSQL.Valid {
			attempt.Score = &scoreSQL.Float64
		}
		if maxScoreSQL.Valid {
			maxScore := int(maxScoreSQL.Int32)
			attempt.MaxScore = &maxScore
		}
		if percentageSQL.Valid {
			attempt.Percentage = &percentageSQL.Float64
		}
		if passedSQL.Valid {
			attempt.Passed = &passedSQL.Bool
		}

		// Parse meta JSON
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &attempt.Meta)
		}

		attempts = append(attempts, attempt)
	}

	commons.Log(ctx, "[dao.GetExpiredAttempts::OK [count=%d", len(attempts))
	return attempts, nil
}

// GetLatestAttemptsByQuizIDs retrieves the latest attempt for each quiz for a specific user
func GetLatestAttemptsByQuizIDs(ctx context.Context, userID uuid.UUID, quizIDs []string) (map[string]*model.LatestAttemptInfo, error) {
	commons.Log(ctx, "[dao.GetLatestAttemptsByQuizIDs [user_id=%s [quiz_count=%d", userID, len(quizIDs))

	if len(quizIDs) == 0 {
		return make(map[string]*model.LatestAttemptInfo), nil
	}

	// Build placeholders for quiz IDs
	placeholders := ""
	args := []interface{}{userID}
	for i, quizID := range quizIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "$" + fmt.Sprintf("%d", i+2)
		args = append(args, quizID)
	}

	qry := `
		WITH ranked_attempts AS (
			SELECT
				id, quiz_id, user_id, status, started_at, submitted_at,
				score, max_score, percentage, passed,
				ROW_NUMBER() OVER (PARTITION BY quiz_id ORDER BY started_at DESC) as rn
			FROM quiz_attempts
			WHERE user_id = $1 AND quiz_id IN (` + placeholders + `)
		)
		SELECT id, quiz_id, status, started_at, submitted_at, score, max_score, percentage, passed
		FROM ranked_attempts
		WHERE rn = 1`

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.GetLatestAttemptsByQuizIDs::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*model.LatestAttemptInfo)
	for rows.Next() {
		var attemptID, quizID, status string
		var startedAt time.Time
		var submittedAt sql.NullTime
		var score sql.NullFloat64
		var maxScore sql.NullInt64
		var percentage sql.NullFloat64
		var passed sql.NullBool

		err := rows.Scan(&attemptID, &quizID, &status, &startedAt, &submittedAt,
			&score, &maxScore, &percentage, &passed)
		if err != nil {
			commons.LogE(ctx, "[dao.GetLatestAttemptsByQuizIDs::ScanErr [err=%v", err)
			return nil, err
		}

		info := &model.LatestAttemptInfo{
			AttemptID: attemptID,
			Status:    status,
			StartedAt: startedAt,
		}

		if submittedAt.Valid {
			info.SubmittedAt = &submittedAt.Time
		}
		if score.Valid {
			info.Score = &score.Float64
		}
		if maxScore.Valid {
			maxScoreInt := int(maxScore.Int64)
			info.MaxScore = &maxScoreInt
		}
		if percentage.Valid {
			info.Percentage = &percentage.Float64
		}
		if passed.Valid {
			info.Passed = &passed.Bool
		}

		result[quizID] = info
	}

	commons.Log(ctx, "[dao.GetLatestAttemptsByQuizIDs::OK [found=%d", len(result))
	return result, nil
}
