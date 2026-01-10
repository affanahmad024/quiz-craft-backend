package unit

import (
	"context"
	"testing"

	quizcore "github.com/affanahmad024/quiz-craft/backend/quiz/core"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/affanahmad024/quiz-craft/backend/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupScoringTests(t *testing.T) (uuid.UUID, uuid.UUID, uuid.UUID, []uuid.UUID) {
	tests.SetupTestDB(t)
	tests.CleanupTestData(t)

	// Create test user
	userID := tests.CreateTestUser(t, "test@example.com", "Test User", "hashedpass", "user")

	// Create test quiz with passing score 60%
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)

	// Create multiple choice questions (10 points each)
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "What is 2+2?", 10, 1)   // Correct: Option B
	q2ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "What is 3+3?", 10, 2)   // Correct: Option C
	q3ID := tests.CreateTestQuestion(t, quizID, "true_false", "The sky is blue.", 10, true) // Correct: True

	return userID, quizID, q1ID, []uuid.UUID{q1ID, q2ID, q3ID}
}

func teardownScoringTests(t *testing.T) {
	tests.CleanupTestData(t)
	tests.TeardownTestDB(t)
}

// TestSubmitAnswer_CorrectMultipleChoice tests scoring correct MC answer
func TestSubmitAnswer_CorrectMultipleChoice(t *testing.T) {
	userID, quizID, q1ID, _ := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Get answers for the question
	answers, err := quizdao.ListByQuestionID(ctx, q1ID)
	require.NoError(t, err)
	require.NotEmpty(t, answers)

	// Find correct answer (index 1 based on test data)
	var correctAnswerID string
	for _, ans := range answers {
		if ans.IsCorrect {
			correctAnswerID = ans.ID
			break
		}
	}
	require.NotEmpty(t, correctAnswerID)

	// Submit correct answer
	submitInput := model.SubmitAnswerInput{
		QuestionID: q1ID.String(),
		AnswerID:   &correctAnswerID,
	}
	userAnswer, err := quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)

	require.NoError(t, err)
	assert.NotNil(t, userAnswer.IsCorrect)
	assert.True(t, *userAnswer.IsCorrect)
	assert.Equal(t, float64(10), userAnswer.PointsEarned) // Full points
}

// TestSubmitAnswer_IncorrectMultipleChoice tests scoring incorrect MC answer
func TestSubmitAnswer_IncorrectMultipleChoice(t *testing.T) {
	userID, quizID, q1ID, _ := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Get answers for the question
	answers, err := quizdao.ListByQuestionID(ctx, q1ID)
	require.NoError(t, err)

	// Find incorrect answer
	var incorrectAnswerID string
	for _, ans := range answers {
		if !ans.IsCorrect {
			incorrectAnswerID = ans.ID
			break
		}
	}
	require.NotEmpty(t, incorrectAnswerID)

	// Submit incorrect answer
	submitInput := model.SubmitAnswerInput{
		QuestionID: q1ID.String(),
		AnswerID:   &incorrectAnswerID,
	}
	userAnswer, err := quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)

	require.NoError(t, err)
	assert.NotNil(t, userAnswer.IsCorrect)
	assert.False(t, *userAnswer.IsCorrect)
	assert.Equal(t, float64(0), userAnswer.PointsEarned) // Zero points
}

// TestFinishAttempt_CalculateFinalScore tests final score calculation
func TestFinishAttempt_CalculateFinalScore(t *testing.T) {
	userID, quizID, _, questionIDs := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Answer 2 questions correctly (20 points) and 1 incorrectly (0 points)
	// Q1: Correct
	answers1, _ := quizdao.ListByQuestionID(ctx, questionIDs[0])
	var correctAnswer1ID string
	for _, ans := range answers1 {
		if ans.IsCorrect {
			correctAnswer1ID = ans.ID
			break
		}
	}
	submitInput1 := model.SubmitAnswerInput{
		QuestionID: questionIDs[0].String(),
		AnswerID:   &correctAnswer1ID,
	}
	quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput1)

	// Q2: Correct
	answers2, _ := quizdao.ListByQuestionID(ctx, questionIDs[1])
	var correctAnswer2ID string
	for _, ans := range answers2 {
		if ans.IsCorrect {
			correctAnswer2ID = ans.ID
			break
		}
	}
	submitInput2 := model.SubmitAnswerInput{
		QuestionID: questionIDs[1].String(),
		AnswerID:   &correctAnswer2ID,
	}
	quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput2)

	// Q3: Incorrect
	answers3, _ := quizdao.ListByQuestionID(ctx, questionIDs[2])
	var incorrectAnswer3ID string
	for _, ans := range answers3 {
		if !ans.IsCorrect {
			incorrectAnswer3ID = ans.ID
			break
		}
	}
	submitInput3 := model.SubmitAnswerInput{
		QuestionID: questionIDs[2].String(),
		AnswerID:   &incorrectAnswer3ID,
	}
	quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput3)

	// Finish attempt
	finishedAttempt, err := quizcore.FinishAttempt(ctx, attempt.ID, userID.String())

	require.NoError(t, err)
	assert.Equal(t, "submitted", finishedAttempt.Status)
	assert.NotNil(t, finishedAttempt.Score)
	assert.Equal(t, float64(20), *finishedAttempt.Score)         // 20 out of 30 points
	assert.NotNil(t, finishedAttempt.Percentage)
	assert.InDelta(t, 66.67, *finishedAttempt.Percentage, 0.1)   // ~66.67%
	assert.NotNil(t, finishedAttempt.Passed)
	assert.True(t, *finishedAttempt.Passed) // 66.67% >= 60% passing score
	assert.NotNil(t, finishedAttempt.SubmittedAt)
}

// TestFinishAttempt_PassedStatus tests passing score logic
func TestFinishAttempt_PassedStatus(t *testing.T) {
	userID, quizID, _, questionIDs := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Answer all 3 questions correctly (30 points = 100%)
	for _, qID := range questionIDs {
		answers, _ := quizdao.ListByQuestionID(ctx, qID)
		var correctAnswerID string
		for _, ans := range answers {
			if ans.IsCorrect {
				correctAnswerID = ans.ID
				break
			}
		}
		submitInput := model.SubmitAnswerInput{
			QuestionID: qID.String(),
			AnswerID:   &correctAnswerID,
		}
		quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)
	}

	// Finish attempt
	finishedAttempt, err := quizcore.FinishAttempt(ctx, attempt.ID, userID.String())

	require.NoError(t, err)
	assert.NotNil(t, finishedAttempt.Score)
	assert.Equal(t, float64(30), *finishedAttempt.Score)
	assert.NotNil(t, finishedAttempt.Percentage)
	assert.Equal(t, float64(100), *finishedAttempt.Percentage)
	assert.NotNil(t, finishedAttempt.Passed)
	assert.True(t, *finishedAttempt.Passed)
}

// TestFinishAttempt_FailedStatus tests failing score logic
func TestFinishAttempt_FailedStatus(t *testing.T) {
	userID, quizID, _, questionIDs := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Answer only 1 question correctly (10 points = 33.33% < 60% passing)
	answers1, _ := quizdao.ListByQuestionID(ctx, questionIDs[0])
	var correctAnswerID string
	for _, ans := range answers1 {
		if ans.IsCorrect {
			correctAnswerID = ans.ID
			break
		}
	}
	submitInput := model.SubmitAnswerInput{
		QuestionID: questionIDs[0].String(),
		AnswerID:   &correctAnswerID,
	}
	quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)

	// Answer other questions incorrectly
	for i := 1; i < len(questionIDs); i++ {
		answers, _ := quizdao.ListByQuestionID(ctx, questionIDs[i])
		var incorrectAnswerID string
		for _, ans := range answers {
			if !ans.IsCorrect {
				incorrectAnswerID = ans.ID
				break
			}
		}
		submitInput := model.SubmitAnswerInput{
			QuestionID: questionIDs[i].String(),
			AnswerID:   &incorrectAnswerID,
		}
		quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)
	}

	// Finish attempt
	finishedAttempt, err := quizcore.FinishAttempt(ctx, attempt.ID, userID.String())

	require.NoError(t, err)
	assert.NotNil(t, finishedAttempt.Score)
	assert.Equal(t, float64(10), *finishedAttempt.Score)
	assert.NotNil(t, finishedAttempt.Percentage)
	assert.InDelta(t, 33.33, *finishedAttempt.Percentage, 0.1)
	assert.NotNil(t, finishedAttempt.Passed)
	assert.False(t, *finishedAttempt.Passed) // 33.33% < 60% passing score
}

// TestFinishAttempt_ZeroScore tests handling of zero score
func TestFinishAttempt_ZeroScore(t *testing.T) {
	userID, quizID, _, questionIDs := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Answer all questions incorrectly
	for _, qID := range questionIDs {
		answers, _ := quizdao.ListByQuestionID(ctx, qID)
		var incorrectAnswerID string
		for _, ans := range answers {
			if !ans.IsCorrect {
				incorrectAnswerID = ans.ID
				break
			}
		}
		submitInput := model.SubmitAnswerInput{
			QuestionID: qID.String(),
			AnswerID:   &incorrectAnswerID,
		}
		quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)
	}

	// Finish attempt
	finishedAttempt, err := quizcore.FinishAttempt(ctx, attempt.ID, userID.String())

	require.NoError(t, err)
	assert.NotNil(t, finishedAttempt.Score)
	assert.Equal(t, float64(0), *finishedAttempt.Score)
	assert.NotNil(t, finishedAttempt.Percentage)
	assert.Equal(t, float64(0), *finishedAttempt.Percentage)
	assert.NotNil(t, finishedAttempt.Passed)
	assert.False(t, *finishedAttempt.Passed)
}

// TestFinishAttempt_PartialAnswers tests scoring with unanswered questions
func TestFinishAttempt_PartialAnswers(t *testing.T) {
	userID, quizID, _, questionIDs := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Answer only 1 question correctly, leave others unanswered
	answers1, _ := quizdao.ListByQuestionID(ctx, questionIDs[0])
	var correctAnswerID string
	for _, ans := range answers1 {
		if ans.IsCorrect {
			correctAnswerID = ans.ID
			break
		}
	}
	submitInput := model.SubmitAnswerInput{
		QuestionID: questionIDs[0].String(),
		AnswerID:   &correctAnswerID,
	}
	quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput)

	// Finish attempt without answering other questions
	finishedAttempt, err := quizcore.FinishAttempt(ctx, attempt.ID, userID.String())

	require.NoError(t, err)
	assert.NotNil(t, finishedAttempt.Score)
	assert.Equal(t, float64(10), *finishedAttempt.Score) // Only 1 question answered
	assert.NotNil(t, finishedAttempt.Percentage)
	assert.InDelta(t, 33.33, *finishedAttempt.Percentage, 0.1) // 10/30 points
	assert.NotNil(t, finishedAttempt.Passed)
	assert.False(t, *finishedAttempt.Passed)
}

// TestSubmitAnswer_UpdateExistingAnswer tests answer update with upsert
func TestSubmitAnswer_UpdateExistingAnswer(t *testing.T) {
	userID, quizID, q1ID, _ := setupScoringTests(t)
	defer teardownScoringTests(t)

	ctx := context.Background()

	// Start attempt
	startInput := model.StartAttemptInput{
		QuizID: quizID.String(),
	}
	attempt, err := quizcore.StartAttempt(ctx, userID.String(), startInput)
	require.NoError(t, err)

	// Get answers for the question
	answers, err := quizdao.ListByQuestionID(ctx, q1ID)
	require.NoError(t, err)

	// Find incorrect and correct answers
	var incorrectAnswerID, correctAnswerID string
	for _, ans := range answers {
		if ans.IsCorrect {
			correctAnswerID = ans.ID
		} else if incorrectAnswerID == "" {
			incorrectAnswerID = ans.ID
		}
	}

	// First submit incorrect answer
	submitInput1 := model.SubmitAnswerInput{
		QuestionID: q1ID.String(),
		AnswerID:   &incorrectAnswerID,
	}
	userAnswer1, err := quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput1)
	require.NoError(t, err)
	assert.False(t, *userAnswer1.IsCorrect)
	assert.Equal(t, float64(0), userAnswer1.PointsEarned)

	// Then update to correct answer
	submitInput2 := model.SubmitAnswerInput{
		QuestionID: q1ID.String(),
		AnswerID:   &correctAnswerID,
	}
	userAnswer2, err := quizcore.SubmitAnswer(ctx, attempt.ID, userID.String(), submitInput2)
	require.NoError(t, err)
	assert.True(t, *userAnswer2.IsCorrect)
	assert.Equal(t, float64(10), userAnswer2.PointsEarned)

	// Verify only one answer exists in database
	attemptUUID, _ := uuid.Parse(attempt.ID)
	userAnswers, err := quizdao.GetUserAnswersByAttempt(ctx, attemptUUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(userAnswers)) // Only one answer should exist (upserted)
}
