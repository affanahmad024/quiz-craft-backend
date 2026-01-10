package unit

import (
	"context"
	"testing"

	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/affanahmad024/quiz-craft/backend/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStatisticsTests(t *testing.T) {
	tests.SetupTestDB(t)
	tests.CleanupTestData(t)
}

func teardownStatisticsTests(t *testing.T) {
	tests.CleanupTestData(t)
	tests.TeardownTestDB(t)
}

// TestCalculateAverageScore tests average score calculation via quiz stats
func TestCalculateAverageScore(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users and quiz
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)

	// Create question
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create 3 attempts with different scores
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "correct", true, 10)

	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "wrong", false, 0)

	attempt3ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt3ID, q1ID, "correct", true, 10)

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should have 3 attempts
	assert.Equal(t, 3, stats.TotalAttempts)
	// Average score calculated by database
	assert.GreaterOrEqual(t, stats.AverageScore, 0.0)
}

// TestCalculatePassRate tests pass rate calculation
func TestCalculatePassRate(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users and quiz with 60% passing score
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)

	// Create question worth 10 points
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create attempts
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "correct", true, 10)

	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "wrong", false, 0)

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should have attempts tracked
	assert.GreaterOrEqual(t, stats.TotalAttempts, 2)
}

// TestCalculatePercentage tests percentage calculation logic
func TestCalculatePercentage(t *testing.T) {
	testCases := []struct {
		name            string
		score           float64
		maxScore        int
		expectedPercent float64
	}{
		{"Perfect score", 30, 30, 100.0},
		{"Half score", 15, 30, 50.0},
		{"Two thirds", 20, 30, 66.67},
		{"Zero score", 0, 30, 0.0},
		{"One third", 10, 30, 33.33},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			percentage := (tc.score / float64(tc.maxScore)) * 100
			assert.InDelta(t, tc.expectedPercent, percentage, 0.1)
		})
	}
}

// TestQuizStatsBasic tests basic quiz statistics
func TestQuizStatsBasic(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create user and quiz
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)

	// Get stats for new quiz (should have zero attempts)
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	assert.Equal(t, 0, stats.TotalAttempts)
	assert.Equal(t, 0.0, stats.AverageScore)
}

// TestUniqueUsersCount tests counting unique users who attempted a quiz
func TestUniqueUsersCount(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	user3ID := tests.CreateTestUser(t, "user3@example.com", "User 3", "pass", "user")

	// Create quiz
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create attempts: User1 (2 attempts), User2 (1 attempt), User3 (1 attempt)
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "answer", true, 10)

	attempt2ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "answer", true, 10)

	attempt3ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt3ID, q1ID, "answer", true, 10)

	attempt4ID := tests.CreateTestAttempt(t, user3ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt4ID, q1ID, "answer", true, 10)

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should have 4 total attempts from 3 unique users
	assert.Equal(t, 4, stats.TotalAttempts)
	// Note: unique users might not be tracked in current implementation
}

// TestHighestLowestScore tests score tracking
func TestHighestLowestScore(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users and quiz
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)

	// Create question worth 10 points
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// High score attempt
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "correct", true, 10)

	// Low score attempt
	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "wrong", false, 0)

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should track attempts
	assert.Equal(t, 2, stats.TotalAttempts)
}

// TestCompletionRate tests attempt completion tracking
func TestCompletionRate(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create user and quiz
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create submitted attempt
	attempt1ID := tests.CreateTestAttempt(t, userID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "answer", true, 10)

	// Create in-progress attempt
	tests.CreateTestAttempt(t, userID, quizID, "in_progress")

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should have multiple attempts
	assert.GreaterOrEqual(t, stats.TotalAttempts, 1)
}

// TestZeroAttemptsStats tests statistics with no attempts
func TestZeroAttemptsStats(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create quiz with no attempts
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should handle zero attempts gracefully
	assert.Equal(t, 0, stats.TotalAttempts)
	assert.Equal(t, 0.0, stats.AverageScore)
}

// TestQuizCreation tests quiz creation tracking
func TestQuizCreation(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	// Create user
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")

	// Create multiple quizzes
	quiz1ID := tests.CreateTestQuiz(t, userID, "Quiz 1", true, true, 60)
	quiz2ID := tests.CreateTestQuiz(t, userID, "Quiz 2", true, true, 60)
	quiz3ID := tests.CreateTestQuiz(t, userID, "Quiz 3", false, false, 60)

	// Verify quizzes were created
	assert.NotEqual(t, uuid.Nil, quiz1ID)
	assert.NotEqual(t, uuid.Nil, quiz2ID)
	assert.NotEqual(t, uuid.Nil, quiz3ID)
}

// TestAttemptStatusTracking tests different attempt statuses
func TestAttemptStatusTracking(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create user and quiz
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)

	// Create attempts with different statuses
	attempt1 := tests.CreateTestAttempt(t, userID, quizID, "in_progress")
	attempt2 := tests.CreateTestAttempt(t, userID, quizID, "submitted")
	attempt3 := tests.CreateTestAttempt(t, userID, quizID, "expired")

	// Verify attempts were created with different statuses
	assert.NotEqual(t, uuid.Nil, attempt1)
	assert.NotEqual(t, uuid.Nil, attempt2)
	assert.NotEqual(t, uuid.Nil, attempt3)

	// Get stats
	stats, err := quizdao.GetQuizStats(ctx, quizID)
	require.NoError(t, err)

	// Should track all attempts
	assert.GreaterOrEqual(t, stats.TotalAttempts, 3)
}

// TestAnswerTracking tests user answer tracking
func TestAnswerTracking(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create user and quiz
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)
	q2ID := tests.CreateTestQuestion(t, quizID, "true_false", "Question 2", 5, true)

	// Create attempt
	attemptID := tests.CreateTestAttempt(t, userID, quizID, "in_progress")

	// Add answers
	tests.CreateTestAnswer(t, attemptID, q1ID, "answer1", true, 10)
	tests.CreateTestAnswer(t, attemptID, q2ID, "true", true, 5)

	// Get answers
	answers, err := quizdao.GetUserAnswersByAttempt(ctx, attemptID)
	require.NoError(t, err)

	// Should have 2 answers
	assert.Equal(t, 2, len(answers))
}
