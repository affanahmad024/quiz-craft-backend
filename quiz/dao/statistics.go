package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SearchQuizzes searches quizzes with filters, sorting, and pagination
func SearchQuizzes(ctx context.Context, input model.SearchQuizzesInput, userID *string) ([]model.QuizSummary, int, error) {
	commons.Log(ctx, "[dao.SearchQuizzes [query=%s [page=%d", input.Query, input.Page)

	// Build WHERE clause
	where := []string{"q.deleted_at IS NULL", "q.is_published = true", "q.is_public = true"}
	args := []interface{}{}
	argCount := 0

	// If user is logged in, also show their own quizzes
	if userID != nil {
		where[len(where)-1] = "(q.is_published = true AND q.is_public = true) OR q.creator_id = $1"
		args = append(args, *userID)
		argCount++
	}

	// Full-text search
	if input.Query != "" {
		argCount++
		where = append(where, fmt.Sprintf("to_tsvector('english', q.title || ' ' || COALESCE(q.description, '')) @@ plainto_tsquery('english', $%d)", argCount))
		args = append(args, input.Query)
	}

	// Category filter
	if input.Category != "" {
		argCount++
		where = append(where, fmt.Sprintf("q.category = $%d", argCount))
		args = append(args, input.Category)
	}

	// Difficulty filter
	if input.Difficulty != "" {
		argCount++
		where = append(where, fmt.Sprintf("q.difficulty = $%d", argCount))
		args = append(args, input.Difficulty)
	}

	// Tags filter (comma-separated)
	if input.Tags != "" {
		argCount++
		tags := strings.Split(input.Tags, ",")
		where = append(where, fmt.Sprintf("q.tags && $%d", argCount))
		args = append(args, pq.Array(tags))
	}

	// Creator filter
	if input.CreatorID != "" {
		argCount++
		where = append(where, fmt.Sprintf("q.creator_id = $%d", argCount))
		args = append(args, input.CreatorID)
	}

	// Template filter
	if input.IsTemplate != nil {
		argCount++
		where = append(where, fmt.Sprintf("q.is_template = $%d", argCount))
		args = append(args, *input.IsTemplate)
	}

	whereClause := strings.Join(where, " AND ")

	// Count total results
	countQry := fmt.Sprintf(`SELECT COUNT(DISTINCT q.id) FROM quizzes q WHERE %s`, whereClause)
	var total int
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.SearchQuizzes::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "q.created_at DESC"
	switch input.SortBy {
	case "newest":
		orderBy = "q.created_at DESC"
	case "popular":
		orderBy = "q.view_count DESC"
	case "most_attempted":
		orderBy = "attempt_count DESC"
	}

	if input.SortOrder == "asc" {
		orderBy = strings.Replace(orderBy, "DESC", "ASC", 1)
	}

	// Calculate offset
	offset := (input.Page - 1) * input.Limit

	// Build main query
	qry := fmt.Sprintf(`
		SELECT
			q.id, q.title, q.description, q.category, q.difficulty,
			q.thumbnail_url, q.tags, q.view_count, q.is_template, q.template_usage_count,
			q.created_at,
			u.id as creator_id, u.name as creator_name,
			COUNT(DISTINCT ques.id) as question_count,
			COUNT(DISTINCT qa.id) FILTER (WHERE qa.status = 'submitted') as attempt_count,
			COALESCE(AVG(qa.percentage) FILTER (WHERE qa.status = 'submitted'), 0) as average_score
		FROM quizzes q
		LEFT JOIN users u ON q.creator_id = u.id
		LEFT JOIN questions ques ON q.id = ques.quiz_id AND ques.deleted_at IS NULL
		LEFT JOIN quiz_attempts qa ON q.id = qa.quiz_id
		WHERE %s
		GROUP BY q.id, u.id, u.name
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argCount+1, argCount+2)

	args = append(args, input.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.SearchQuizzes::QueryErr [err=%v", err)
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]model.QuizSummary, 0)
	for rows.Next() {
		var quiz model.QuizSummary
		var descSQL sql.NullString
		var categorySQL sql.NullString
		var difficultySQL sql.NullString
		var thumbnailSQL sql.NullString

		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&descSQL,
			&categorySQL,
			&difficultySQL,
			&thumbnailSQL,
			pq.Array(&quiz.Tags),
			&quiz.ViewCount,
			&quiz.IsTemplate,
			&quiz.TemplateUsageCount,
			&quiz.CreatedAt,
			&quiz.Creator.ID,
			&quiz.Creator.Name,
			&quiz.QuestionCount,
			&quiz.AttemptCount,
			&quiz.AverageScore,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.SearchQuizzes::ScanErr [err=%v", err)
			continue
		}

		if descSQL.Valid {
			quiz.Description = &descSQL.String
		}
		if categorySQL.Valid {
			quiz.Category = &categorySQL.String
		}
		if difficultySQL.Valid {
			quiz.Difficulty = &difficultySQL.String
		}
		if thumbnailSQL.Valid {
			quiz.ThumbnailURL = &thumbnailSQL.String
		}

		results = append(results, quiz)
	}

	commons.Log(ctx, "[dao.SearchQuizzes::OK [count=%d [total=%d", len(results), total)
	return results, total, nil
}

// GetSearchFacets gets facet counts for filters
func GetSearchFacets(ctx context.Context, userID *string) (model.SearchFacets, error) {
	commons.Log(ctx, "[dao.GetSearchFacets")

	facets := model.SearchFacets{
		Categories:   make([]model.FacetCount, 0),
		Difficulties: make([]model.FacetCount, 0),
		Tags:         make([]model.FacetCount, 0),
	}

	// Base WHERE clause for visible quizzes
	whereClause := "deleted_at IS NULL AND is_published = true AND is_public = true"
	args := []interface{}{}
	if userID != nil {
		whereClause = "deleted_at IS NULL AND ((is_published = true AND is_public = true) OR creator_id = $1)"
		args = append(args, *userID)
	}

	// Get category counts
	catQry := fmt.Sprintf(`
		SELECT category, COUNT(*) as count
		FROM quizzes
		WHERE %s AND category IS NOT NULL
		GROUP BY category
		ORDER BY count DESC
		LIMIT 20
	`, whereClause)

	rows, err := db.QueryContext(ctx, catQry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.GetSearchFacets::CategoryErr [err=%v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var fc model.FacetCount
			if err := rows.Scan(&fc.Value, &fc.Count); err == nil {
				facets.Categories = append(facets.Categories, fc)
			}
		}
	}

	// Get difficulty counts
	diffQry := fmt.Sprintf(`
		SELECT difficulty, COUNT(*) as count
		FROM quizzes
		WHERE %s AND difficulty IS NOT NULL
		GROUP BY difficulty
		ORDER BY count DESC
	`, whereClause)

	rows, err = db.QueryContext(ctx, diffQry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.GetSearchFacets::DifficultyErr [err=%v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var fc model.FacetCount
			if err := rows.Scan(&fc.Value, &fc.Count); err == nil {
				facets.Difficulties = append(facets.Difficulties, fc)
			}
		}
	}

	// Get top tags
	tagQry := fmt.Sprintf(`
		SELECT unnest(tags) as tag, COUNT(*) as count
		FROM quizzes
		WHERE %s AND tags IS NOT NULL
		GROUP BY tag
		ORDER BY count DESC
		LIMIT 30
	`, whereClause)

	rows, err = db.QueryContext(ctx, tagQry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.GetSearchFacets::TagsErr [err=%v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var fc model.FacetCount
			if err := rows.Scan(&fc.Value, &fc.Count); err == nil {
				facets.Tags = append(facets.Tags, fc)
			}
		}
	}

	commons.Log(ctx, "[dao.GetSearchFacets::OK")
	return facets, nil
}

// GetQuizAttemptStats gets aggregate statistics for a quiz
func GetQuizAttemptStats(ctx context.Context, quizID uuid.UUID) (*model.QuizStatistics, error) {
	commons.Log(ctx, "[dao.GetQuizAttemptStats [quiz_id=%s", quizID)

	stats := &model.QuizStatistics{
		QuizID: quizID.String(),
	}

	qry := `
		SELECT
			COUNT(*) as total_attempts,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(*) FILTER (WHERE status = 'submitted') as completed_attempts,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress_attempts,
			COUNT(*) FILTER (WHERE status = 'abandoned') as abandoned_attempts,
			COUNT(*) FILTER (WHERE status = 'expired') as expired_attempts,
			COALESCE(AVG(percentage) FILTER (WHERE status = 'submitted'), 0) as average_score,
			COALESCE(MAX(percentage) FILTER (WHERE status = 'submitted'), 0) as highest_score,
			COALESCE(MIN(percentage) FILTER (WHERE status = 'submitted'), 0) as lowest_score,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY percentage) FILTER (WHERE status = 'submitted') as median_score,
			COUNT(*) FILTER (WHERE passed = true) as passed_count,
			AVG(EXTRACT(EPOCH FROM (submitted_at - started_at))) FILTER (WHERE status = 'submitted') as average_time_seconds,
			MIN(EXTRACT(EPOCH FROM (submitted_at - started_at))) FILTER (WHERE status = 'submitted') as fastest_time_seconds
		FROM quiz_attempts
		WHERE quiz_id = $1
	`

	var medianSQL sql.NullFloat64
	var avgTimeSQL sql.NullFloat64
	var fastestTimeSQL sql.NullFloat64
	var passedCount int

	err := db.QueryRowContext(ctx, qry, quizID).Scan(
		&stats.TotalAttempts,
		&stats.UniqueUsers,
		&stats.CompletedAttempts,
		&stats.InProgressAttempts,
		&stats.AbandonedAttempts,
		&stats.ExpiredAttempts,
		&stats.AverageScore,
		&stats.HighestScore,
		&stats.LowestScore,
		&medianSQL,
		&passedCount,
		&avgTimeSQL,
		&fastestTimeSQL,
	)

	if err != nil && err != sql.ErrNoRows {
		commons.LogE(ctx, "[dao.GetQuizAttemptStats::QueryErr [err=%v", err)
		return nil, err
	}

	if medianSQL.Valid {
		stats.MedianScore = &medianSQL.Float64
	}
	if avgTimeSQL.Valid {
		stats.AverageTimeSeconds = &avgTimeSQL.Float64
	}
	if fastestTimeSQL.Valid {
		stats.FastestTimeSeconds = &fastestTimeSQL.Float64
	}

	// Calculate pass rate
	if stats.CompletedAttempts > 0 {
		stats.PassRate = (float64(passedCount) / float64(stats.CompletedAttempts)) * 100
	}

	commons.Log(ctx, "[dao.GetQuizAttemptStats::OK [total=%d", stats.TotalAttempts)
	return stats, nil
}

// GetQuestionStats gets statistics for all questions in a quiz
func GetQuestionStats(ctx context.Context, quizID uuid.UUID) ([]model.QuestionStat, error) {
	commons.Log(ctx, "[dao.GetQuestionStats [quiz_id=%s", quizID)

	qry := `
		SELECT
			q.id as question_id,
			q.question_text,
			COUNT(ua.id) as times_answered,
			COUNT(*) FILTER (WHERE ua.is_correct = true) as correct_answers,
			COUNT(*) FILTER (WHERE ua.is_correct = false) as incorrect_answers,
			COALESCE(AVG(CASE WHEN ua.is_correct THEN 100.0 ELSE 0.0 END), 0) as success_rate
		FROM questions q
		LEFT JOIN user_answers ua ON q.id = ua.question_id
		WHERE q.quiz_id = $1 AND q.deleted_at IS NULL
		GROUP BY q.id, q.question_text
		ORDER BY q.order_index
	`

	rows, err := db.QueryContext(ctx, qry, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetQuestionStats::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	stats := make([]model.QuestionStat, 0)
	for rows.Next() {
		var stat model.QuestionStat
		err := rows.Scan(
			&stat.QuestionID,
			&stat.QuestionText,
			&stat.TimesAnswered,
			&stat.CorrectAnswers,
			&stat.IncorrectAnswers,
			&stat.SuccessRate,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetQuestionStats::ScanErr [err=%v", err)
			continue
		}
		stats = append(stats, stat)
	}

	commons.Log(ctx, "[dao.GetQuestionStats::OK [count=%d", len(stats))
	return stats, nil
}

// GetUserCreatorStats gets quiz creation statistics for a user
func GetUserCreatorStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	commons.Log(ctx, "[dao.GetUserCreatorStats [user_id=%s", userID)

	qry := `
		SELECT
			COUNT(*) as total_quizzes,
			COUNT(*) FILTER (WHERE is_published = true) as published_quizzes,
			COALESCE(SUM(view_count), 0) as total_views
		FROM quizzes
		WHERE creator_id = $1 AND deleted_at IS NULL
	`

	var totalQuizzes, publishedQuizzes, totalViews int
	err := db.QueryRowContext(ctx, qry, userID).Scan(&totalQuizzes, &publishedQuizzes, &totalViews)
	if err != nil && err != sql.ErrNoRows {
		commons.LogE(ctx, "[dao.GetUserCreatorStats::QueryErr [err=%v", err)
		return nil, err
	}

	// Get total attempts on user's quizzes
	attemptQry := `
		SELECT COUNT(*)
		FROM quiz_attempts qa
		JOIN quizzes q ON qa.quiz_id = q.id
		WHERE q.creator_id = $1 AND q.deleted_at IS NULL
	`
	var totalAttempts int
	db.QueryRowContext(ctx, attemptQry, userID).Scan(&totalAttempts)

	return map[string]interface{}{
		"total_quizzes":      totalQuizzes,
		"published_quizzes":  publishedQuizzes,
		"total_views":        totalViews,
		"total_attempts":     totalAttempts,
	}, nil
}

// GetUserAttemptStats gets quiz attempt statistics for a user
func GetUserAttemptStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	commons.Log(ctx, "[dao.GetUserAttemptStats [user_id=%s", userID)

	qry := `
		SELECT
			COUNT(DISTINCT quiz_id) as unique_quizzes,
			COUNT(*) FILTER (WHERE status = 'submitted') as completed_attempts,
			COALESCE(AVG(percentage) FILTER (WHERE status = 'submitted'), 0) as average_score,
			COALESCE(SUM(score) FILTER (WHERE status = 'submitted'), 0) as total_points_earned,
			COUNT(*) FILTER (WHERE passed = true) as quizzes_passed
		FROM quiz_attempts
		WHERE user_id = $1
	`

	var uniqueQuizzes, completedAttempts, quizzesPassed int
	var averageScore, totalPointsEarned float64

	err := db.QueryRowContext(ctx, qry, userID).Scan(
		&uniqueQuizzes,
		&completedAttempts,
		&averageScore,
		&totalPointsEarned,
		&quizzesPassed,
	)
	if err != nil && err != sql.ErrNoRows {
		commons.LogE(ctx, "[dao.GetUserAttemptStats::QueryErr [err=%v", err)
		return nil, err
	}

	// Calculate pass rate
	passRate := 0.0
	if completedAttempts > 0 {
		passRate = (float64(quizzesPassed) / float64(completedAttempts)) * 100
	}

	return map[string]interface{}{
		"unique_quizzes":      uniqueQuizzes,
		"completed_attempts":  completedAttempts,
		"average_score":       averageScore,
		"total_points_earned": totalPointsEarned,
		"quizzes_passed":      quizzesPassed,
		"pass_rate":           passRate,
	}, nil
}

// GetCategoryStats gets statistics for all categories
func GetCategoryStats(ctx context.Context) ([]model.CategoryStats, error) {
	commons.Log(ctx, "[dao.GetCategoryStats")

	qry := `
		SELECT
			q.category,
			COUNT(DISTINCT q.id) as quiz_count,
			COUNT(DISTINCT qa.id) as total_attempts,
			COALESCE(AVG(qa.percentage) FILTER (WHERE qa.status = 'submitted'), 0) as average_score
		FROM quizzes q
		LEFT JOIN quiz_attempts qa ON q.id = qa.quiz_id
		WHERE q.deleted_at IS NULL AND q.is_published = true AND q.category IS NOT NULL
		GROUP BY q.category
		ORDER BY quiz_count DESC
	`

	rows, err := db.QueryContext(ctx, qry)
	if err != nil {
		commons.LogE(ctx, "[dao.GetCategoryStats::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	stats := make([]model.CategoryStats, 0)
	for rows.Next() {
		var stat model.CategoryStats
		err := rows.Scan(&stat.Category, &stat.QuizCount, &stat.TotalAttempts, &stat.AverageScore)
		if err != nil {
			commons.LogE(ctx, "[dao.GetCategoryStats::ScanErr [err=%v", err)
			continue
		}
		stats = append(stats, stat)
	}

	commons.Log(ctx, "[dao.GetCategoryStats::OK [count=%d", len(stats))
	return stats, nil
}

// GetPopularTags gets most used tags
func GetPopularTags(ctx context.Context, limit int) ([]model.TagCount, error) {
	commons.Log(ctx, "[dao.GetPopularTags [limit=%d", limit)

	qry := `
		SELECT unnest(tags) as tag, COUNT(*) as count
		FROM quizzes
		WHERE deleted_at IS NULL AND is_published = true AND tags IS NOT NULL
		GROUP BY tag
		ORDER BY count DESC
		LIMIT $1
	`

	rows, err := db.QueryContext(ctx, qry, limit)
	if err != nil {
		commons.LogE(ctx, "[dao.GetPopularTags::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	tags := make([]model.TagCount, 0)
	for rows.Next() {
		var tag model.TagCount
		err := rows.Scan(&tag.Tag, &tag.Count)
		if err != nil {
			commons.LogE(ctx, "[dao.GetPopularTags::ScanErr [err=%v", err)
			continue
		}
		tags = append(tags, tag)
	}

	commons.Log(ctx, "[dao.GetPopularTags::OK [count=%d", len(tags))
	return tags, nil
}

// GetQuizByShareCode gets a quiz by its share code
func GetQuizByShareCode(ctx context.Context, shareCode string) (*model.Quiz, error) {
	commons.Log(ctx, "[dao.GetQuizByShareCode [share_code=%s", shareCode)

	qry := `SELECT id, title, description, creator_id, category, difficulty,
			time_limit_minutes, passing_score, is_published, is_public,
			tags, share_code, view_count, meta, created_at, updated_at, deleted_at
			FROM quizzes
			WHERE share_code = $1 AND deleted_at IS NULL`

	quiz := &model.Quiz{}
	var descSQL sql.NullString
	var categorySQL sql.NullString
	var difficultySQL sql.NullString
	var timeLimitSQL sql.NullInt32
	var shareCodeSQL sql.NullString
	var metaJSON []byte

	err := db.QueryRowContext(ctx, qry, shareCode).Scan(
		&quiz.ID,
		&quiz.Title,
		&descSQL,
		&quiz.CreatorID,
		&categorySQL,
		&difficultySQL,
		&timeLimitSQL,
		&quiz.PassingScore,
		&quiz.IsPublished,
		&quiz.IsPublic,
		pq.Array(&quiz.Tags),
		&shareCodeSQL,
		&quiz.ViewCount,
		&metaJSON,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
		&quiz.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("quiz-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.GetQuizByShareCode::QueryErr [err=%v", err)
		return nil, err
	}

	if descSQL.Valid {
		quiz.Description = descSQL.String
	}
	if categorySQL.Valid {
		quiz.Category = categorySQL.String
	}
	if difficultySQL.Valid {
		quiz.Difficulty = difficultySQL.String
	}
	if timeLimitSQL.Valid {
		timeLimit := int(timeLimitSQL.Int32)
		quiz.TimeLimitMinutes = &timeLimit
	}
	if shareCodeSQL.Valid {
		quiz.ShareCode = &shareCodeSQL.String
	}

	commons.Log(ctx, "[dao.GetQuizByShareCode::OK [quiz_id=%s", quiz.ID)
	return quiz, nil
}

// IncrementQuizViewCount increments the view count for a quiz
func IncrementQuizViewCount(ctx context.Context, quizID uuid.UUID) error {
	commons.Log(ctx, "[dao.IncrementQuizViewCount [quiz_id=%s", quizID)

	qry := `UPDATE quizzes SET view_count = view_count + 1 WHERE id = $1`
	_, err := db.ExecContext(ctx, qry, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.IncrementQuizViewCount::UpdateErr [err=%v", err)
		return err
	}

	return nil
}

// UpdateQuizShareCode updates the share code for a quiz
func UpdateQuizShareCode(ctx context.Context, quizID uuid.UUID, shareCode string) error {
	commons.Log(ctx, "[dao.UpdateQuizShareCode [quiz_id=%s [code=%s", quizID, shareCode)

	qry := `UPDATE quizzes SET share_code = $1, share_count = share_count + 1 WHERE id = $2`
	_, err := db.ExecContext(ctx, qry, shareCode, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateQuizShareCode::UpdateErr [err=%v", err)
		return err
	}

	return nil
}

// IncrementTemplateUsageCount increments usage count for a template
func IncrementTemplateUsageCount(ctx context.Context, templateID uuid.UUID) error {
	commons.Log(ctx, "[dao.IncrementTemplateUsageCount [template_id=%s", templateID)

	qry := `UPDATE quizzes SET template_usage_count = template_usage_count + 1 WHERE id = $1`
	_, err := db.ExecContext(ctx, qry, templateID)
	if err != nil {
		commons.LogE(ctx, "[dao.IncrementTemplateUsageCount::UpdateErr [err=%v", err)
		return err
	}

	return nil
}

// GetUserActivityDates gets all distinct activity dates for a user (quiz creation + attempts)
func GetUserActivityDates(ctx context.Context, userID uuid.UUID) ([]time.Time, error) {
	commons.Log(ctx, "[dao.GetUserActivityDates [user_id=%s", userID)

	qry := `
		SELECT DISTINCT DATE(created_at) as activity_date
		FROM (
			SELECT created_at FROM quizzes WHERE creator_id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT started_at FROM quiz_attempts WHERE user_id = $1
		) activities
		ORDER BY activity_date DESC
	`

	rows, err := db.QueryContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetUserActivityDates::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err == nil {
			dates = append(dates, date)
		}
	}

	commons.Log(ctx, "[dao.GetUserActivityDates::OK [count=%d", len(dates))
	return dates, nil
}
