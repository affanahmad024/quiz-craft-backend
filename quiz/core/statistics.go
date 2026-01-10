package core

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// GetQuizStatistics gets detailed statistics for a quiz (creator only)
func GetQuizStatistics(ctx context.Context, quizID string, userID string, includeQuestions bool) (*model.QuizStatistics, error) {
	commons.Log(ctx, "[core.GetQuizStatistics [quiz_id=%s [user_id=%s [include_questions=%v", quizID, userID, includeQuestions)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz and validate ownership
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizStatistics::FindErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	if quiz.CreatorID != userID {
		commons.LogW(ctx, "[core.GetQuizStatistics::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("not-quiz-creator")
	}

	// Get attempt statistics
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizStatistics::StatsErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	// Set quiz metadata
	stats.QuizID = quizID
	stats.Title = quiz.Title
	stats.CreatorID = quiz.CreatorID
	stats.ViewCount = quiz.ViewCount
	stats.ShareCount = quiz.ShareCount

	// Get question-level stats if requested
	if includeQuestions {
		questionStats, err := quizdao.GetQuestionStats(ctx, quizUUID)
		if err != nil {
			commons.LogW(ctx, "[core.GetQuizStatistics::QuestionStatsErr [err=%v", err)
			// Non-critical, continue without question stats
		} else {
			stats.QuestionStats = questionStats
		}
	}

	commons.Log(ctx, "[core.GetQuizStatistics::OK [quiz_id=%s [total_attempts=%d", quizID, stats.TotalAttempts)
	return stats, nil
}

// GetUserStatistics gets aggregate statistics for a user
func GetUserStatistics(ctx context.Context, userID string) (*model.UserStatistics, error) {
	commons.Log(ctx, "[core.GetUserStatistics [user_id=%s", userID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, commons.BadInput("invalid-user-id")
	}

	// Get user
	user, err := userdao.FindByID(ctx, userUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserStatistics::FindUserErr [user_id=%s [err=%v", userID, err)
		return nil, err
	}

	// Get quiz creation stats
	creatorStats, err := quizdao.GetUserCreatorStats(ctx, userUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserStatistics::CreatorStatsErr [err=%v", err)
		return nil, err
	}

	// Get quiz attempt stats
	attemptStats, err := quizdao.GetUserAttemptStats(ctx, userUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserStatistics::AttemptStatsErr [err=%v", err)
		return nil, err
	}

	// Calculate streak (both quiz creation AND quiz attempts)
	currentStreak, longestStreak, lastActivity := calculateUserStreak(ctx, userUUID)

	stats := &model.UserStatistics{
		UserID: userID,
		Name:   user.Name,
		Email:  user.Email,

		// Creator stats
		QuizzesCreated:    creatorStats["total_quizzes"].(int),
		PublishedQuizzes:  creatorStats["published_quizzes"].(int),
		TotalQuizViews:    creatorStats["total_views"].(int),
		TotalQuizAttempts: creatorStats["total_attempts"].(int),

		// Attempt stats
		QuizzesTaken:      attemptStats["unique_quizzes"].(int),
		AttemptsCompleted: attemptStats["completed_attempts"].(int),
		AverageScore:      attemptStats["average_score"].(float64),
		TotalPointsEarned: attemptStats["total_points_earned"].(float64),
		QuizzesPassed:     attemptStats["quizzes_passed"].(int),
		PassRate:          attemptStats["pass_rate"].(float64),

		// Streak
		CurrentStreak:  currentStreak,
		LongestStreak:  longestStreak,
		LastActivityAt: lastActivity,
		JoinedAt:       user.CreatedAt,
	}

	commons.Log(ctx, "[core.GetUserStatistics::OK [user_id=%s", userID)
	return stats, nil
}

// calculateUserStreak calculates current and longest streaks for a user
// Tracks BOTH quiz creation and quiz attempts as activity
func calculateUserStreak(ctx context.Context, userID uuid.UUID) (int, int, time.Time) {
	commons.Log(ctx, "[core.calculateUserStreak [user_id=%s", userID)

	// Get all activity dates from DAO
	activityDates, err := quizdao.GetUserActivityDates(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.calculateUserStreak::GetDatesErr [err=%v", err)
		return 0, 0, time.Now()
	}

	if len(activityDates) == 0 {
		return 0, 0, time.Now()
	}

	lastActivity := activityDates[0]

	// Calculate current streak
	currentStreak := 0
	today := time.Now().Truncate(24 * time.Hour)

	for i, date := range activityDates {
		expectedDate := today.AddDate(0, 0, -i)
		if date.Truncate(24*time.Hour).Equal(expectedDate) {
			currentStreak++
		} else {
			break
		}
	}

	// Calculate longest streak
	longestStreak := 0
	tempStreak := 1

	for i := 1; i < len(activityDates); i++ {
		prevDate := activityDates[i-1]
		currDate := activityDates[i]

		daysDiff := int(prevDate.Sub(currDate).Hours() / 24)
		if daysDiff == 1 {
			tempStreak++
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
		} else {
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
			tempStreak = 1
		}
	}

	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	commons.Log(ctx, "[core.calculateUserStreak::OK [current=%d [longest=%d", currentStreak, longestStreak)
	return currentStreak, longestStreak, lastActivity
}

// GetCategories gets statistics for all categories
func GetCategories(ctx context.Context) ([]model.CategoryStats, error) {
	commons.Log(ctx, "[core.GetCategories")

	stats, err := quizdao.GetCategoryStats(ctx)
	if err != nil {
		commons.LogE(ctx, "[core.GetCategories::GetErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.GetCategories::OK [count=%d", len(stats))
	return stats, nil
}

// GetPopularTags gets most used tags
func GetPopularTags(ctx context.Context, limit int) ([]model.TagCount, error) {
	commons.Log(ctx, "[core.GetPopularTags [limit=%d", limit)

	if limit < 1 || limit > 100 {
		limit = 50
	}

	tags, err := quizdao.GetPopularTags(ctx, limit)
	if err != nil {
		commons.LogE(ctx, "[core.GetPopularTags::GetErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.GetPopularTags::OK [count=%d", len(tags))
	return tags, nil
}
