package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var db *sql.DB

// Init initializes the DAO with database connection
func Init(database *sql.DB) {
	db = database
}

// CreateQuiz creates a new quiz
func CreateQuiz(ctx context.Context, quiz *model.Quiz) error {
	commons.Log(ctx, "[dao.CreateQuiz [creator_id=%s [title=%s", quiz.CreatorID, quiz.Title)

	metaJSON, _ := json.Marshal(quiz.Meta)
	qry := `INSERT INTO quizzes (
				id, title, description, creator_id, category, difficulty,
				time_limit_minutes, passing_score, is_published, is_public,
				tags, meta, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()
			) RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		quiz.ID,
		quiz.Title,
		quiz.Description,
		quiz.CreatorID,
		quiz.Category,
		quiz.Difficulty,
		quiz.TimeLimitMinutes,
		quiz.PassingScore,
		quiz.IsPublished,
		quiz.IsPublic,
		pq.Array(quiz.Tags),
		metaJSON,
	).Scan(&quiz.CreatedAt, &quiz.UpdatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateQuiz::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateQuiz::OK [quiz_id=%s", quiz.ID)
	return nil
}

// FindByID finds quiz by ID (excludes soft deleted)
func FindByID(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	commons.Log(ctx, "[dao.FindByID [id=%s", id)

	quiz := &model.Quiz{}
	qry := `SELECT id, title, description, creator_id, category, difficulty,
			time_limit_minutes, passing_score, is_published, is_public,
			tags, meta, created_at, updated_at, deleted_at
			FROM quizzes
			WHERE id = $1 AND deleted_at IS NULL`

	var descSQL sql.NullString
	var categorySQL sql.NullString
	var difficultySQL sql.NullString
	var timeLimitSQL sql.NullInt32
	var metaJSON []byte

	err := db.QueryRowContext(ctx, qry, id).Scan(
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
		&quiz.Tags,
		&metaJSON,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
		&quiz.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("quiz-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	// Handle nullable fields
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

	// Parse meta JSON
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &quiz.Meta)
	}

	return quiz, nil
}

// FindByIDIncludingDeleted finds quiz by ID (includes soft deleted)
func FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	commons.Log(ctx, "[dao.FindByIDIncludingDeleted [id=%s", id)

	quiz := &model.Quiz{}
	qry := `SELECT id, title, description, creator_id, category, difficulty,
			time_limit_minutes, passing_score, is_published, is_public,
			tags, meta, created_at, updated_at, deleted_at
			FROM quizzes
			WHERE id = $1`

	var descSQL sql.NullString
	var categorySQL sql.NullString
	var difficultySQL sql.NullString
	var timeLimitSQL sql.NullInt32
	var metaJSON []byte

	err := db.QueryRowContext(ctx, qry, id).Scan(
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
		&quiz.Tags,
		&metaJSON,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
		&quiz.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("quiz-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByIDIncludingDeleted::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	// Handle nullable fields
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

	// Parse meta JSON
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &quiz.Meta)
	}

	return quiz, nil
}

// ListByCreator lists quizzes by creator ID with filters
func ListByCreator(ctx context.Context, creatorID uuid.UUID, filter model.QuizListFilter) ([]*model.Quiz, int, error) {
	commons.Log(ctx, "[dao.ListByCreator [creator_id=%s [page=%d [limit=%d", creatorID, filter.Page, filter.Limit)

	quizzes := []*model.Quiz{}

	// Build WHERE clause
	where := "WHERE creator_id = $1 AND deleted_at IS NULL"
	args := []interface{}{creatorID}
	argCount := 2

	if filter.IsPublished != nil {
		where += fmt.Sprintf(" AND is_published = $%d", argCount)
		args = append(args, *filter.IsPublished)
		argCount++
	}

	if filter.Category != "" {
		where += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, filter.Category)
		argCount++
	}

	if filter.Difficulty != "" {
		where += fmt.Sprintf(" AND difficulty = $%d", argCount)
		args = append(args, filter.Difficulty)
		argCount++
	}

	// Get total count
	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM quizzes %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByCreator::CountErr [err=%v", err)
		return quizzes, 0, err
	}

	// Get paginated results
	offset := (filter.Page - 1) * filter.Limit
	qry := fmt.Sprintf(`SELECT id, title, description, creator_id, category, difficulty,
						time_limit_minutes, passing_score, is_published, is_public,
						tags, meta, created_at, updated_at, deleted_at
						FROM quizzes %s
						ORDER BY created_at DESC
						LIMIT $%d OFFSET $%d`, where, argCount, argCount+1)

	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByCreator::SQLErr [err=%v", err)
		return quizzes, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		quiz := &model.Quiz{}
		var descSQL sql.NullString
		var categorySQL sql.NullString
		var difficultySQL sql.NullString
		var timeLimitSQL sql.NullInt32
		var metaJSON []byte

		err := rows.Scan(
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
			&quiz.Tags,
			&metaJSON,
			&quiz.CreatedAt,
			&quiz.UpdatedAt,
			&quiz.DeletedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByCreator::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
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

		// Parse meta JSON
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &quiz.Meta)
		}

		quizzes = append(quizzes, quiz)
	}

	commons.Log(ctx, "[dao.ListByCreator::OK [total=%d [returned=%d", total, len(quizzes))
	return quizzes, total, nil
}

// ListPublic lists public published quizzes with filters
func ListPublic(ctx context.Context, filter model.QuizListFilter) ([]*model.Quiz, int, error) {
	commons.Log(ctx, "[dao.ListPublic [page=%d [limit=%d", filter.Page, filter.Limit)

	quizzes := []*model.Quiz{}

	// Build WHERE clause - only public and published quizzes
	where := "WHERE is_public = true AND is_published = true AND deleted_at IS NULL"
	args := []interface{}{}
	argCount := 1

	if filter.Category != "" {
		where += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, filter.Category)
		argCount++
	}

	if filter.Difficulty != "" {
		where += fmt.Sprintf(" AND difficulty = $%d", argCount)
		args = append(args, filter.Difficulty)
		argCount++
	}

	if filter.Search != "" {
		where += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern)
		argCount++
	}

	// Get total count
	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM quizzes %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListPublic::CountErr [err=%v", err)
		return quizzes, 0, err
	}

	// Get paginated results
	offset := (filter.Page - 1) * filter.Limit
	qry := fmt.Sprintf(`SELECT id, title, description, creator_id, category, difficulty,
						time_limit_minutes, passing_score, is_published, is_public,
						tags, meta, created_at, updated_at, deleted_at
						FROM quizzes %s
						ORDER BY created_at DESC
						LIMIT $%d OFFSET $%d`, where, argCount, argCount+1)

	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListPublic::SQLErr [err=%v", err)
		return quizzes, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		quiz := &model.Quiz{}
		var descSQL sql.NullString
		var categorySQL sql.NullString
		var difficultySQL sql.NullString
		var timeLimitSQL sql.NullInt32
		var metaJSON []byte

		err := rows.Scan(
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
			&quiz.Tags,
			&metaJSON,
			&quiz.CreatedAt,
			&quiz.UpdatedAt,
			&quiz.DeletedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListPublic::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
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

		// Parse meta JSON
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &quiz.Meta)
		}

		quizzes = append(quizzes, quiz)
	}

	commons.Log(ctx, "[dao.ListPublic::OK [total=%d [returned=%d", total, len(quizzes))
	return quizzes, total, nil
}

// UpdateQuiz updates quiz fields
func UpdateQuiz(ctx context.Context, quiz *model.Quiz) error {
	commons.Log(ctx, "[dao.UpdateQuiz [quiz_id=%s", quiz.ID)

	metaJSON, _ := json.Marshal(quiz.Meta)
	qry := `UPDATE quizzes
			SET title = $1, description = $2, category = $3, difficulty = $4,
				time_limit_minutes = $5, passing_score = $6, is_public = $7,
				tags = $8, meta = $9, updated_at = NOW()
			WHERE id = $10 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry,
		quiz.Title,
		quiz.Description,
		quiz.Category,
		quiz.Difficulty,
		quiz.TimeLimitMinutes,
		quiz.PassingScore,
		quiz.IsPublic,
		pq.Array(quiz.Tags),
		metaJSON,
		quiz.ID,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.UpdateQuiz::SQLErr [quiz_id=%s [err=%v", quiz.ID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("quiz-not-found")
	}

	commons.Log(ctx, "[dao.UpdateQuiz::OK [quiz_id=%s", quiz.ID)
	return nil
}

// UpdatePublishStatus updates quiz publish status
func UpdatePublishStatus(ctx context.Context, id uuid.UUID, isPublished bool) error {
	commons.Log(ctx, "[dao.UpdatePublishStatus [id=%s [is_published=%v", id, isPublished)

	qry := `UPDATE quizzes
			SET is_published = $1, updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry, isPublished, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdatePublishStatus::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("quiz-not-found")
	}

	return nil
}

// SoftDelete soft deletes a quiz
func SoftDelete(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.SoftDelete [id=%s", id)

	qry := `UPDATE quizzes
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.SoftDelete::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("quiz-not-found")
	}

	commons.Log(ctx, "[dao.SoftDelete::OK [id=%s", id)
	return nil
}

// Restore restores a soft deleted quiz
func Restore(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.Restore [id=%s", id)

	qry := `UPDATE quizzes
			SET deleted_at = NULL, updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NOT NULL`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.Restore::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("quiz-not-found-or-not-deleted")
	}

	commons.Log(ctx, "[dao.Restore::OK [id=%s", id)
	return nil
}

// HardDelete permanently deletes a quiz
func HardDelete(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.HardDelete [level=CRITICAL [id=%s", id)

	qry := `DELETE FROM quizzes WHERE id = $1`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.HardDelete::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("quiz-not-found")
	}

	commons.Log(ctx, "[dao.HardDelete::OK [id=%s", id)
	return nil
}

// CountByCreator counts quizzes by creator
func CountByCreator(ctx context.Context, creatorID uuid.UUID) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM quizzes WHERE creator_id = $1 AND deleted_at IS NULL`
	err := db.QueryRowContext(ctx, qry, creatorID).Scan(&count)
	return count, err
}

// CountPublic counts public published quizzes
func CountPublic(ctx context.Context) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM quizzes WHERE is_public = true AND is_published = true AND deleted_at IS NULL`
	err := db.QueryRowContext(ctx, qry).Scan(&count)
	return count, err
}
