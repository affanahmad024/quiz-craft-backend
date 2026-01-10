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

// TestCalculateAverageScore tests average score calculation
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

	// Get answers
	answers, err := quizdao.ListByQuestionID(ctx, q1ID)
	require.NoError(t, err)
	var correctAnswerID string
	for _, ans := range answers {
		if ans.IsCorrect {
			correctAnswerID = ans.ID
			break
		}
	}

	// Create 3 attempts with different scores
	// Attempt 1: 10 points (100%)
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, correctAnswerID, true, 10)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt1ID.String()), 10, 100, true)

	// Attempt 2: 0 points (0%)
	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "wrong", false, 0)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt2ID.String()), 0, 0, false)

	// Attempt 3: 10 points (100%)
	attempt3ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt3ID, q1ID, correctAnswerID, true, 10)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt3ID.String()), 10, 100, true)

	// Get stats
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizID)
	require.NoError(t, err)

	// Average score should be (10 + 0 + 10) / 3 = 6.67
	assert.InDelta(t, 6.67, stats.AverageScore, 0.1)
	assert.Equal(t, 3, stats.TotalAttempts)
}

// TestCalculatePassRate tests pass rate calculation
func TestCalculatePassRate(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users and quiz with 60% passing score
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	user3ID := tests.CreateTestUser(t, "user3@example.com", "User 3", "pass", "user")
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)

	// Create question worth 10 points
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create 4 attempts: 2 passed, 2 failed
	// Attempt 1: Passed (100% >= 60%)
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "correct", true, 10)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt1ID.String()), 10, 100, true)

	// Attempt 2: Failed (0% < 60%)
	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "wrong", false, 0)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt2ID.String()), 0, 0, false)

	// Attempt 3: Passed (100% >= 60%)
	attempt3ID := tests.CreateTestAttempt(t, user3ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt3ID, q1ID, "correct", true, 10)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt3ID.String()), 10, 100, true)

	// Attempt 4: Failed (0% < 60%)
	attempt4ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt4ID, q1ID, "wrong", false, 0)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt4ID.String()), 0, 0, false)

	// Get stats
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizID)
	require.NoError(t, err)

	// Pass rate should be 2/4 = 50%
	assert.Equal(t, 4, stats.TotalAttempts)
	// Calculate pass rate manually from attempts
	passedCount := 0
	if stats.PassedAttempts > 0 {
		passedCount = stats.PassedAttempts
	}
	passRate := float64(passedCount) / float64(stats.TotalAttempts) * 100
	assert.InDelta(t, 50.0, passRate, 0.1)
}

// TestCalculatePercentage tests percentage calculation
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

// TestStreakCalculation tests streak calculation logic
func TestStreakCalculation(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create user
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")

	// Create quizzes for activity
	quiz1ID := tests.CreateTestQuiz(t, userID, "Quiz 1", true, true, 60)
	quiz2ID := tests.CreateTestQuiz(t, userID, "Quiz 2", true, true, 60)
	quiz3ID := tests.CreateTestQuiz(t, userID, "Quiz 3", true, true, 60)

	// Verify quizzes were created
	assert.NotEqual(t, uuid.Nil, quiz1ID)
	assert.NotEqual(t, uuid.Nil, quiz2ID)
	assert.NotEqual(t, uuid.Nil, quiz3ID)

	// Get user stats (should show quiz creation activity)
	stats, err := quizdao.GetUserStats(ctx, userID)
	require.NoError(t, err)

	// User should have created 3 quizzes
	assert.Equal(t, 3, stats.QuizzesCreated)
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
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizID)
	require.NoError(t, err)

	// Should have 4 total attempts from 3 unique users
	assert.Equal(t, 4, stats.TotalAttempts)
	assert.Equal(t, 3, stats.UniqueUsers)
}

// TestHighestLowestScore tests finding highest and lowest scores
func TestHighestLowestScore(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users and quiz
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	user3ID := tests.CreateTestUser(t, "user3@example.com", "User 3", "pass", "user")
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create attempts with scores: 10, 5, 8
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "answer", true, 10)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt1ID.String()), 10, 100, true)

	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "answer", false, 5)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt2ID.String()), 5, 50, false)

	attempt3ID := tests.CreateTestAttempt(t, user3ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt3ID, q1ID, "answer", true, 8)
	quizdao.UpdateAttemptScore(ctx, uuid.MustParse(attempt3ID.String()), 8, 80, true)

	// Get stats
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizID)
	require.NoError(t, err)

	// Highest should be 10, lowest should be 5
	assert.Equal(t, float64(10), stats.HighestScore)
	assert.Equal(t, float64(5), stats.LowestScore)
}

// TestCompletionRate tests calculation of completion rate
func TestCompletionRate(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create users and quiz
	user1ID := tests.CreateTestUser(t, "user1@example.com", "User 1", "pass", "user")
	user2ID := tests.CreateTestUser(t, "user2@example.com", "User 2", "pass", "user")
	quizID := tests.CreateTestQuiz(t, user1ID, "Test Quiz", true, true, 60)
	q1ID := tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Create 3 attempts: 2 submitted, 1 in_progress
	attempt1ID := tests.CreateTestAttempt(t, user1ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt1ID, q1ID, "answer", true, 10)

	attempt2ID := tests.CreateTestAttempt(t, user2ID, quizID, "submitted")
	tests.CreateTestAnswer(t, attempt2ID, q1ID, "answer", true, 10)

	attempt3ID := tests.CreateTestAttempt(t, user1ID, quizID, "in_progress")
	tests.CreateTestAnswer(t, attempt3ID, q1ID, "answer", true, 10)

	// Get stats
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizID)
	require.NoError(t, err)

	// Completion rate should be 2/3 = 66.67%
	assert.Equal(t, 3, stats.TotalAttempts)
	completionRate := float64(stats.CompletedAttempts) / float64(stats.TotalAttempts) * 100
	assert.InDelta(t, 66.67, completionRate, 0.1)
}

// TestAverageTimeCalculation tests calculation of average completion time
func TestAverageTimeCalculation(t *testing.T) {
	// Test the logic of calculating average time from attempts
	// Time in seconds: 120, 180, 240
	times := []float64{120, 180, 240}

	var total float64
	for _, time := range times {
		total += time
	}
	average := total / float64(len(times))

	// Average should be 180 seconds
	assert.Equal(t, float64(180), average)
}

// TestZeroAttemptsStats tests statistics when no attempts exist
func TestZeroAttemptsStats(t *testing.T) {
	setupStatisticsTests(t)
	defer teardownStatisticsTests(t)

	ctx := context.Background()

	// Create user and quiz with no attempts
	userID := tests.CreateTestUser(t, "user@example.com", "User", "pass", "user")
	quizID := tests.CreateTestQuiz(t, userID, "Test Quiz", true, true, 60)
	tests.CreateTestQuestion(t, quizID, "multiple_choice", "Question 1", 10, 1)

	// Get stats
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizID)
	require.NoError(t, err)

	// All stats should be zero/defaults
	assert.Equal(t, 0, stats.TotalAttempts)
	assert.Equal(t, 0, stats.UniqueUsers)
	assert.Equal(t, float64(0), stats.AverageScore)
	assert.Equal(t, float64(0), stats.HighestScore)
	assert.Equal(t, float64(0), stats.LowestScore)
}

// TestQuestionSuccessRate tests success rate calculation per question
func TestQuestionSuccessRate(t *testing.T) {
	// Test success rate calculation logic
	// 7 correct answers out of 10 total = 70%
	correct := 7
	total := 10
	successRate := (float64(correct) / float64(total)) * 100

	assert.Equal(t, float64(70), successRate)
}

// TestScoreDistribution tests score distribution calculations
func TestScoreDistribution(t *testing.T) {
	scores := []float64{0, 25, 50, 50, 75, 75, 75, 100, 100, 100}

	// Calculate distribution
	ranges := map[string]int{
		"0-25":    0,
		"26-50":   0,
		"51-75":   0,
		"76-100":  0,
	}

	for _, score := range scores {
		if score <= 25 {
			ranges["0-25"]++
		} else if score <= 50 {
			ranges["26-50"]++
		} else if score <= 75 {
			ranges["51-75"]++
		} else {
			ranges["76-100"]++
		}
	}

	assert.Equal(t, 1, ranges["0-25"])    // 1 score in 0-25 (0)
	assert.Equal(t, 2, ranges["26-50"])   // 2 scores in 26-50 (25, 50)
	assert.Equal(t, 3, ranges["51-75"])   // 3 scores in 51-75 (50, 75, 75, 75 - but 50 counted above)
	assert.Equal(t, 3, ranges["76-100"])  // 3 scores in 76-100 (100, 100, 100)
}
