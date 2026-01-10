package core

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// StartAttempt starts a new quiz attempt for a user
func StartAttempt(ctx context.Context, userID string, input model.StartAttemptInput) (*model.QuizAttempt, error) {
	commons.Log(ctx, "[core.StartAttempt [user_id=%s [quiz_id=%s", userID, input.QuizID)

	// Validate input
	if err := input.Validate(); err != nil {
		commons.LogW(ctx, "[core.StartAttempt::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	quizUUID, err := uuid.Parse(input.QuizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.StartAttempt::FindQuizErr [quiz_id=%s [err=%v", input.QuizID, err)
		return nil, err
	}

	// Check if quiz is published
	if !quiz.IsPublished {
		commons.LogW(ctx, "[core.StartAttempt::QuizNotPublished [quiz_id=%s", input.QuizID)
		return nil, commons.BadInput("quiz-not-published")
	}

	// Check access control
	if !quiz.IsAccessibleBy(userID) {
		commons.LogW(ctx, "[core.StartAttempt::AccessDenied [quiz_id=%s [user_id=%s", input.QuizID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	// Get all questions for the quiz to calculate max score
	questions, err := quizdao.ListByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.StartAttempt::GetQuestionsErr [quiz_id=%s [err=%v", input.QuizID, err)
		return nil, err
	}

	if len(questions) == 0 {
		commons.LogW(ctx, "[core.StartAttempt::NoQuestions [quiz_id=%s", input.QuizID)
		return nil, commons.BadInput("quiz-has-no-questions")
	}

	// Calculate max score
	maxScore := 0
	for _, q := range questions {
		maxScore += q.Points
	}

	// Calculate time limit and expiry
	var timeLimitSeconds *int
	var expiresAt *time.Time
	if quiz.TimeLimitMinutes != nil {
		seconds := *quiz.TimeLimitMinutes * 60
		timeLimitSeconds = &seconds
		expiry := time.Now().Add(time.Duration(seconds) * time.Second)
		expiresAt = &expiry
	}

	// Create attempt
	attempt := &model.QuizAttempt{
		ID:               uuid.New().String(),
		QuizID:           input.QuizID,
		UserID:           userID,
		Status:           "in_progress",
		StartedAt:        time.Now(),
		TimeLimitSeconds: timeLimitSeconds,
		ExpiresAt:        expiresAt,
		MaxScore:         &maxScore,
		Meta:             make(model.JSONMap),
	}

	// Save to database
	if err := quizdao.CreateAttempt(ctx, attempt); err != nil {
		commons.LogE(ctx, "[core.StartAttempt::CreateErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.StartAttempt::OK [attempt_id=%s [max_score=%d", attempt.ID, maxScore)
	return attempt, nil
}

// GetAttempt retrieves an attempt with access control
func GetAttempt(ctx context.Context, attemptID string, userID string) (*model.QuizAttempt, error) {
	commons.Log(ctx, "[core.GetAttempt [attempt_id=%s [user_id=%s", attemptID, userID)

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, commons.BadInput("invalid-attempt-id")
	}

	attempt, err := quizdao.FindAttemptByID(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetAttempt::FindErr [attempt_id=%s [err=%v", attemptID, err)
		return nil, err
	}

	// Check ownership
	if attempt.UserID != userID {
		commons.LogW(ctx, "[core.GetAttempt::AccessDenied [attempt_id=%s [user_id=%s", attemptID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	return attempt, nil
}

// SubmitAnswer submits an answer for a question in an attempt
func SubmitAnswer(ctx context.Context, attemptID string, userID string, input model.SubmitAnswerInput) (*model.UserAnswer, error) {
	commons.Log(ctx, "[core.SubmitAnswer [attempt_id=%s [question_id=%s", attemptID, input.QuestionID)

	// Validate input
	if err := input.Validate(); err != nil {
		commons.LogW(ctx, "[core.SubmitAnswer::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, commons.BadInput("invalid-attempt-id")
	}

	questionUUID, err := uuid.Parse(input.QuestionID)
	if err != nil {
		return nil, commons.BadInput("invalid-question-id")
	}

	// Get attempt
	attempt, err := quizdao.FindAttemptByID(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[core.SubmitAnswer::FindAttemptErr [attempt_id=%s [err=%v", attemptID, err)
		return nil, err
	}

	// Check ownership
	if attempt.UserID != userID {
		commons.LogW(ctx, "[core.SubmitAnswer::AccessDenied [attempt_id=%s [user_id=%s", attemptID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	// Check if attempt is still in progress and not expired
	if !attempt.CanSubmitAnswer() {
		if attempt.IsExpired() {
			commons.LogW(ctx, "[core.SubmitAnswer::AttemptExpired [attempt_id=%s", attemptID)
			// Auto-expire the attempt
			if err := quizdao.UpdateAttemptStatus(ctx, attemptUUID, "expired"); err != nil {
				commons.LogE(ctx, "[core.SubmitAnswer::UpdateStatusErr [err=%v", err)
			}
			return nil, commons.BadInput("attempt-expired")
		}
		commons.LogW(ctx, "[core.SubmitAnswer::AttemptNotInProgress [attempt_id=%s [status=%s", attemptID, attempt.Status)
		return nil, commons.BadInput("attempt-not-in-progress")
	}

	// Get question to validate and score
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.SubmitAnswer::FindQuestionErr [question_id=%s [err=%v", input.QuestionID, err)
		return nil, err
	}

	// Verify question belongs to the quiz
	if question.QuizID != attempt.QuizID {
		commons.LogW(ctx, "[core.SubmitAnswer::QuestionNotInQuiz [question_id=%s [quiz_id=%s", input.QuestionID, attempt.QuizID)
		return nil, commons.BadInput("question-not-in-quiz")
	}

	// Phase 5: Only support MC and TF questions
	if question.QuestionType != "multiple_choice" && question.QuestionType != "true_false" {
		commons.LogW(ctx, "[core.SubmitAnswer::UnsupportedQuestionType [type=%s", question.QuestionType)
		return nil, commons.BadInput("question-type-not-supported")
	}

	// Validate answer_id is provided for MC/TF
	if input.AnswerID == nil {
		return nil, commons.BadInput("answer-id-required-for-mc-tf")
	}

	// Get the selected answer to verify it exists and belongs to this question
	answerUUID, err := uuid.Parse(*input.AnswerID)
	if err != nil {
		return nil, commons.BadInput("invalid-answer-id")
	}

	// Get all answers for the question
	answers, err := quizdao.ListByQuestionID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.SubmitAnswer::GetAnswersErr [question_id=%s [err=%v", input.QuestionID, err)
		return nil, err
	}

	// Find selected answer and check correctness
	var selectedAnswer *model.Answer
	for _, ans := range answers {
		if ans.ID == answerUUID.String() {
			selectedAnswer = ans
			break
		}
	}

	if selectedAnswer == nil {
		commons.LogW(ctx, "[core.SubmitAnswer::AnswerNotFound [answer_id=%s", *input.AnswerID)
		return nil, commons.BadInput("answer-not-found-for-question")
	}

	// Score the answer
	isCorrect := selectedAnswer.IsCorrect
	var pointsEarned float64
	if isCorrect {
		pointsEarned = float64(question.Points)
	} else {
		pointsEarned = 0
	}

	// Create user answer
	userAnswer := &model.UserAnswer{
		ID:           uuid.New().String(),
		AttemptID:    attemptID,
		QuestionID:   input.QuestionID,
		AnswerID:     input.AnswerID,
		IsCorrect:    &isCorrect,
		PointsEarned: pointsEarned,
		AnsweredAt:   time.Now(),
		Meta:         make(model.JSONMap),
	}

	// Save user answer (upsert)
	if err := quizdao.CreateOrUpdateUserAnswer(ctx, userAnswer); err != nil {
		commons.LogE(ctx, "[core.SubmitAnswer::SaveErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.SubmitAnswer::OK [user_answer_id=%s [is_correct=%v [points=%f",
		userAnswer.ID, isCorrect, pointsEarned)
	return userAnswer, nil
}

// FinishAttempt finishes an attempt and calculates final score
func FinishAttempt(ctx context.Context, attemptID string, userID string) (*model.QuizAttempt, error) {
	commons.Log(ctx, "[core.FinishAttempt [attempt_id=%s [user_id=%s", attemptID, userID)

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, commons.BadInput("invalid-attempt-id")
	}

	// Get attempt
	attempt, err := quizdao.FindAttemptByID(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[core.FinishAttempt::FindAttemptErr [attempt_id=%s [err=%v", attemptID, err)
		return nil, err
	}

	// Check ownership
	if attempt.UserID != userID {
		commons.LogW(ctx, "[core.FinishAttempt::AccessDenied [attempt_id=%s [user_id=%s", attemptID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	// Check if attempt can be finished
	if !attempt.CanFinish() {
		commons.LogW(ctx, "[core.FinishAttempt::CannotFinish [attempt_id=%s [status=%s", attemptID, attempt.Status)
		return nil, commons.BadInput("attempt-already-finished")
	}

	// Get all user answers
	userAnswers, err := quizdao.GetUserAnswersByAttempt(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[core.FinishAttempt::GetAnswersErr [attempt_id=%s [err=%v", attemptID, err)
		return nil, err
	}

	// Calculate total score
	var totalScore float64
	for _, ans := range userAnswers {
		totalScore += ans.PointsEarned
	}

	// Get quiz to check passing score
	quizUUID, _ := uuid.Parse(attempt.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.FinishAttempt::FindQuizErr [quiz_id=%s [err=%v", attempt.QuizID, err)
		return nil, err
	}

	// Calculate percentage and determine if passed
	percentage := attempt.CalculatePercentage()
	if attempt.MaxScore != nil && *attempt.MaxScore > 0 {
		percentage = (totalScore / float64(*attempt.MaxScore)) * 100
	}
	passed := percentage >= float64(quiz.PassingScore)

	// Determine status
	status := "submitted"
	if attempt.IsExpired() {
		status = "expired"
	}

	// Update attempt with final score
	now := time.Now()
	attempt.Status = status
	attempt.SubmittedAt = &now
	attempt.Score = &totalScore
	attempt.Percentage = &percentage
	attempt.Passed = &passed

	if err := quizdao.FinishAttempt(ctx, attempt); err != nil {
		commons.LogE(ctx, "[core.FinishAttempt::UpdateErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.FinishAttempt::OK [attempt_id=%s [score=%f [percentage=%f [passed=%v",
		attemptID, totalScore, percentage, passed)
	return attempt, nil
}

// GetAttemptResult gets the full result of an attempt with answers
func GetAttemptResult(ctx context.Context, attemptID string, userID string) (*model.AttemptResult, error) {
	commons.Log(ctx, "[core.GetAttemptResult [attempt_id=%s [user_id=%s", attemptID, userID)

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, commons.BadInput("invalid-attempt-id")
	}

	// Get attempt first to check ownership
	attempt, err := quizdao.FindAttemptByID(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetAttemptResult::FindAttemptErr [attempt_id=%s [err=%v", attemptID, err)
		return nil, err
	}

	// Check ownership - user can view their own attempts
	// OR quiz creator can view attempts on their quizzes
	if attempt.UserID != userID {
		// Check if user is the quiz creator
		quizUUID, _ := uuid.Parse(attempt.QuizID)
		quiz, err := quizdao.FindByID(ctx, quizUUID)
		if err != nil || quiz.CreatorID != userID {
			commons.LogW(ctx, "[core.GetAttemptResult::AccessDenied [attempt_id=%s [user_id=%s", attemptID, userID)
			return nil, commons.Forbidden("access-denied")
		}
	}

	// Get full result
	result, err := quizdao.GetAttemptResult(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetAttemptResult::GetResultErr [attempt_id=%s [err=%v", attemptID, err)
		return nil, err
	}

	commons.Log(ctx, "[core.GetAttemptResult::OK [attempt_id=%s", attemptID)
	return result, nil
}

// ListUserAttempts lists all attempts for a user
func ListUserAttempts(ctx context.Context, userID string, filter model.AttemptListFilter) ([]*model.AttemptListItem, int, error) {
	commons.Log(ctx, "[core.ListUserAttempts [user_id=%s [page=%d", userID, filter.Page)

	// Set defaults for pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, commons.BadInput("invalid-user-id")
	}

	attempts, total, err := quizdao.ListUserAttempts(ctx, userUUID, filter)
	if err != nil {
		commons.LogE(ctx, "[core.ListUserAttempts::ListErr [user_id=%s [err=%v", userID, err)
		return nil, 0, err
	}

	commons.Log(ctx, "[core.ListUserAttempts::OK [count=%d [total=%d", len(attempts), total)
	return attempts, total, nil
}

// GetQuizStats gets statistics for a quiz (creator only)
func GetQuizStats(ctx context.Context, quizID string, userID string) (*model.QuizStats, error) {
	commons.Log(ctx, "[core.GetQuizStats [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz to check ownership
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizStats::FindQuizErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	// Only creator can view stats
	if quiz.CreatorID != userID {
		commons.LogW(ctx, "[core.GetQuizStats::AccessDenied [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	stats, err := quizdao.GetQuizStats(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizStats::GetStatsErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	commons.Log(ctx, "[core.GetQuizStats::OK [quiz_id=%s [total_attempts=%d", quizID, stats.TotalAttempts)
	return stats, nil
}
