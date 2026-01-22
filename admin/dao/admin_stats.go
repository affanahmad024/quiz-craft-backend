package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/google/uuid"
)

// GetPlatformStats retrieves comprehensive platform statistics
func GetPlatformStats(ctx context.Context) (*model.PlatformStats, error) {
	commons.Log(ctx, "[dao.GetPlatformStats")

	stats := &model.PlatformStats{
		GeneratedAt: time.Now(),
	}

	// User statistics
	if err := getUserStats(ctx, stats); err != nil {
		return nil, err
	}

	// Quiz statistics
	if err := getQuizStats(ctx, stats); err != nil {
		return nil, err
	}

	// Attempt statistics
	if err := getAttemptStats(ctx, stats); err != nil {
		return nil, err
	}

	// System health
	stats.DatabaseStatus = checkDatabaseHealth(ctx)

	commons.Log(ctx, "[dao.GetPlatformStats::OK")
	return stats, nil
}

func getUserStats(ctx context.Context, stats *model.PlatformStats) error {
	db := database.GetSQLDB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	query := `
		SELECT
			COUNT(*) FILTER (WHERE deleted_at IS NULL) as total_users,
			COUNT(*) FILTER (WHERE deleted_at IS NULL AND enabled = false) as disabled_users,
			COUNT(*) FILTER (WHERE created_at >= $1) as new_users_today,
			COUNT(*) FILTER (WHERE created_at >= $2) as new_users_this_week,
			COUNT(*) FILTER (WHERE created_at >= $3) as new_users_this_month
		FROM users
	`

	err := db.QueryRowContext(ctx, query, today, weekAgo, monthAgo).Scan(
		&stats.TotalUsers,
		&stats.DisabledUsers,
		&stats.NewUsersToday,
		&stats.NewUsersThisWeek,
		&stats.NewUsersThisMonth,
	)
	if err != nil {
		commons.LogE(ctx, "[dao.getUserStats::QueryErr [err=%v", err)
		return err
	}

	// Active users (last 30 days with attempts or quiz creations)
	activeQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM (
			SELECT user_id FROM quiz_attempts WHERE created_at >= $1
			UNION
			SELECT creator_id as user_id FROM quizzes WHERE created_at >= $1
		) as active_users
	`
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	err = db.QueryRowContext(ctx, activeQuery, thirtyDaysAgo).Scan(&stats.ActiveUsers)
	if err != nil {
		stats.ActiveUsers = 0 // Non-critical
	}

	return nil
}

func getQuizStats(ctx context.Context, stats *model.PlatformStats) error {
	db := database.GetSQLDB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	today := time.Now().Truncate(24 * time.Hour)

	query := `
		SELECT
			COUNT(*) FILTER (WHERE deleted_at IS NULL) as total_quizzes,
			COUNT(*) FILTER (WHERE deleted_at IS NULL AND is_published = true) as published_quizzes,
			COUNT(*) FILTER (WHERE deleted_at IS NULL AND is_public = false) as private_quizzes,
			COUNT(*) FILTER (WHERE created_at >= $1) as quizzes_created_today
		FROM quizzes
	`

	err := db.QueryRowContext(ctx, query, today).Scan(
		&stats.TotalQuizzes,
		&stats.PublishedQuizzes,
		&stats.PrivateQuizzes,
		&stats.QuizzesCreatedToday,
	)
	if err != nil {
		commons.LogE(ctx, "[dao.getQuizStats::QueryErr [err=%v", err)
		return err
	}

	// Count total questions
	questionQuery := `
		SELECT COUNT(*)
		FROM questions q
		JOIN quizzes qz ON q.quiz_id = qz.id
		WHERE qz.deleted_at IS NULL
	`
	err = db.QueryRowContext(ctx, questionQuery).Scan(&stats.TotalQuestions)
	if err != nil {
		stats.TotalQuestions = 0 // Non-critical
	}

	return nil
}

func getAttemptStats(ctx context.Context, stats *model.PlatformStats) error {
	db := database.GetSQLDB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	today := time.Now().Truncate(24 * time.Hour)

	query := `
		SELECT
			COUNT(*) as total_attempts,
			COUNT(*) FILTER (WHERE status = 'submitted') as completed_attempts,
			COUNT(*) FILTER (WHERE created_at >= $1) as attempts_today,
			COALESCE(AVG(percentage_score) FILTER (WHERE status = 'submitted'), 0) as average_score
		FROM quiz_attempts
	`

	err := db.QueryRowContext(ctx, query, today).Scan(
		&stats.TotalAttempts,
		&stats.CompletedAttempts,
		&stats.AttemptsToday,
		&stats.AverageScore,
	)
	if err != nil {
		commons.LogE(ctx, "[dao.getAttemptStats::QueryErr [err=%v", err)
		return err
	}

	return nil
}

func checkDatabaseHealth(ctx context.Context) string {
	db := database.GetSQLDB()
	if db == nil {
		return "unhealthy"
	}

	err := db.PingContext(ctx)
	if err != nil {
		return "unhealthy"
	}
	return "healthy"
}

// GetDailyUserActivity retrieves daily activity timeline
func GetDailyUserActivity(ctx context.Context, startDate time.Time, days int) ([]model.UserActivity, error) {
	commons.Log(ctx, "[dao.GetDailyUserActivity [days=%d", days)

	db := database.GetSQLDB()
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		WITH dates AS (
			SELECT generate_series(
				DATE($1),
				DATE(NOW()),
				'1 day'::interval
			)::date as date
		),
		new_users AS (
			SELECT DATE(created_at) as date, COUNT(*) as count
			FROM users
			WHERE created_at >= $1
			GROUP BY DATE(created_at)
		),
		active_users AS (
			SELECT DATE(created_at) as date, COUNT(DISTINCT user_id) as count
			FROM quiz_attempts
			WHERE created_at >= $1
			GROUP BY DATE(created_at)
		),
		quizzes_created AS (
			SELECT DATE(created_at) as date, COUNT(*) as count
			FROM quizzes
			WHERE created_at >= $1
			GROUP BY DATE(created_at)
		),
		attempts AS (
			SELECT DATE(created_at) as date, COUNT(*) as count
			FROM quiz_attempts
			WHERE created_at >= $1
			GROUP BY DATE(created_at)
		)
		SELECT
			d.date,
			COALESCE(nu.count, 0) as new_users,
			COALESCE(au.count, 0) as active_users,
			COALESCE(qc.count, 0) as quizzes_created,
			COALESCE(a.count, 0) as attempts_count
		FROM dates d
		LEFT JOIN new_users nu ON d.date = nu.date
		LEFT JOIN active_users au ON d.date = au.date
		LEFT JOIN quizzes_created qc ON d.date = qc.date
		LEFT JOIN attempts a ON d.date = a.date
		ORDER BY d.date DESC
	`

	rows, err := db.QueryContext(ctx, query, startDate)
	if err != nil {
		commons.LogE(ctx, "[dao.GetDailyUserActivity::QueryErr [err=%v", err)
		return nil, err
	}
	defer rows.Close()

	activities := []model.UserActivity{}
	for rows.Next() {
		var activity model.UserActivity
		var date time.Time

		err := rows.Scan(
			&date,
			&activity.NewUsers,
			&activity.ActiveUsers,
			&activity.QuizzesCreated,
			&activity.AttemptsCount,
		)
		if err != nil {
			commons.LogW(ctx, "[dao.GetDailyUserActivity::ScanErr [err=%v", err)
			continue
		}

		activity.Date = date.Format("2006-01-02")
		activities = append(activities, activity)
	}

	commons.Log(ctx, "[dao.GetDailyUserActivity::OK [count=%d", len(activities))
	return activities, nil
}

// ListUsersAdmin retrieves all users with admin view (includes stats and filters)
func ListUsersAdmin(ctx context.Context, page, limit int, role *string, enabled *bool, emailVerified *bool, search *string, sortBy, sortOrder string) ([]model.AdminUserDetail, int, error) {
	commons.Log(ctx, "[dao.ListUsersAdmin [page=%d [limit=%d", page, limit)

	db := database.GetSQLDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database connection is nil")
	}

	// Build WHERE clause
	where := []string{"u.deleted_at IS NULL"}
	args := []interface{}{}
	argCount := 0

	if role != nil {
		argCount++
		where = append(where, fmt.Sprintf("u.role = $%d", argCount))
		args = append(args, *role)
	}

	if enabled != nil {
		argCount++
		where = append(where, fmt.Sprintf("u.enabled = $%d", argCount))
		args = append(args, *enabled)
	}

	if emailVerified != nil {
		argCount++
		where = append(where, fmt.Sprintf("u.email_verified = $%d", argCount))
		args = append(args, *emailVerified)
	}

	if search != nil && *search != "" {
		argCount++
		where = append(where, fmt.Sprintf("(u.email ILIKE $%d OR u.name ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+*search+"%")
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	// Count total
	countQuery := "SELECT COUNT(*) FROM users u " + whereClause
	var total int
	err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUsersAdmin::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Validate sort
	validSortBy := map[string]bool{"created_at": true, "updated_at": true, "name": true, "email": true}
	if !validSortBy[sortBy] {
		sortBy = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Get users
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := `
		SELECT
			u.id,
			u.email,
			u.name,
			u.phone_number,
			u.role,
			u.email_verified,
			u.enabled,
			u.created_at,
			u.updated_at,
			COUNT(DISTINCT q.id) as quizzes_created,
			COUNT(DISTINCT qa.id) as quizzes_taken,
			COALESCE(AVG(qa.percentage_score) FILTER (WHERE qa.status = 'submitted'), 0) as average_score
		FROM users u
		LEFT JOIN quizzes q ON u.id = q.creator_id AND q.deleted_at IS NULL
		LEFT JOIN quiz_attempts qa ON u.id = qa.user_id
		` + whereClause + `
		GROUP BY u.id
		ORDER BY u.` + sortBy + ` ` + sortOrder + `
		LIMIT $` + fmt.Sprintf("%d", limitArg) + ` OFFSET $` + fmt.Sprintf("%d", offsetArg)

	offset := (page - 1) * limit
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUsersAdmin::QueryErr [err=%v", err)
		return nil, 0, err
	}
	defer rows.Close()

	users := []model.AdminUserDetail{}
	for rows.Next() {
		var user model.AdminUserDetail
		var phoneNumber sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&phoneNumber,
			&user.Role,
			&user.EmailVerified,
			&user.Enabled,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.QuizzesCreated,
			&user.QuizzesTaken,
			&user.AverageScore,
		)
		if err != nil {
			commons.LogW(ctx, "[dao.ListUsersAdmin::ScanErr [err=%v", err)
			continue
		}

		if phoneNumber.Valid {
			user.PhoneNumber = &phoneNumber.String
		}

		users = append(users, user)
	}

	commons.Log(ctx, "[dao.ListUsersAdmin::OK [count=%d [total=%d", len(users), total)
	return users, total, nil
}

// GetUserDetailAdmin retrieves detailed user info for admin
func GetUserDetailAdmin(ctx context.Context, userID uuid.UUID) (*model.AdminUserDetail, error) {
	commons.Log(ctx, "[dao.GetUserDetailAdmin [user_id=%s", userID)

	db := database.GetSQLDB()
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT
			u.id,
			u.email,
			u.name,
			u.phone_number,
			u.role,
			u.email_verified,
			u.enabled,
			u.created_at,
			u.updated_at,
			u.deleted_at,
			COUNT(DISTINCT q.id) as quizzes_created,
			COUNT(DISTINCT qa.id) as quizzes_taken,
			COALESCE(AVG(qa.percentage_score) FILTER (WHERE qa.status = 'submitted'), 0) as average_score
		FROM users u
		LEFT JOIN quizzes q ON u.id = q.creator_id AND q.deleted_at IS NULL
		LEFT JOIN quiz_attempts qa ON u.id = qa.user_id
		WHERE u.id = $1
		GROUP BY u.id
	`

	var user model.AdminUserDetail
	var phoneNumber, deletedAt sql.NullString

	err := db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&phoneNumber,
		&user.Role,
		&user.EmailVerified,
		&user.Enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
		&user.QuizzesCreated,
		&user.QuizzesTaken,
		&user.AverageScore,
	)
	if err == sql.ErrNoRows {
		return nil, commons.NotFound("user-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.GetUserDetailAdmin::QueryErr [err=%v", err)
		return nil, err
	}

	if phoneNumber.Valid {
		user.PhoneNumber = &phoneNumber.String
	}
	if deletedAt.Valid {
		parsedTime, _ := time.Parse(time.RFC3339, deletedAt.String)
		user.DeletedAt = &parsedTime
	}

	return &user, nil
}
