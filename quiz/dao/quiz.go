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

	if quiz.PublishStatus == "" {
		quiz.PublishStatus = "draft"
	}

	metaJSON, _ := json.Marshal(quiz.Meta)
	qry := `INSERT INTO quizzes (
				id, title, description, creator_id, category, difficulty,
				time_limit_minutes, passing_score, is_published, publish_status, is_public,
				tags, meta, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW()
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
		quiz.PublishStatus,
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

func scanQuiz(row interface {
	Scan(dest ...interface{}) error
}) (*model.Quiz, error) {
	quiz := &model.Quiz{}
	var descSQL sql.NullString
	var categorySQL sql.NullString
	var difficultySQL sql.NullString
	var timeLimitSQL sql.NullInt32
	var publishStatusSQL sql.NullString
	var enrichmentJobSQL sql.NullString
	var metaJSON []byte

	err := row.Scan(
		&quiz.ID,
		&quiz.Title,
		&descSQL,
		&quiz.CreatorID,
		&categorySQL,
		&difficultySQL,
		&timeLimitSQL,
		&quiz.PassingScore,
		&quiz.IsPublished,
		&publishStatusSQL,
		&quiz.IsPublic,
		&quiz.Tags,
		&enrichmentJobSQL,
		&metaJSON,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
		&quiz.DeletedAt,
	)
	if err != nil {
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
	if publishStatusSQL.Valid {
		quiz.PublishStatus = publishStatusSQL.String
	}
	if enrichmentJobSQL.Valid {
		quiz.AIEnrichmentJobID = &enrichmentJobSQL.String
	}
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &quiz.Meta)
	}
	return quiz, nil
}

const quizSelectColumns = `id, title, description, creator_id, category, difficulty,
	time_limit_minutes, passing_score, is_published, publish_status, is_public,
	tags, ai_enrichment_job_id, meta, created_at, updated_at, deleted_at`

// FindByID finds quiz by ID (excludes soft deleted)
func FindByID(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	commons.Log(ctx, "[dao.FindByID [id=%s", id)

	qry := `SELECT ` + quizSelectColumns + ` FROM quizzes WHERE id = $1 AND deleted_at IS NULL`
	quiz, err := scanQuiz(db.QueryRowContext(ctx, qry, id))
	if err == sql.ErrNoRows {
		return nil, commons.NotFound("quiz-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}
	return quiz, nil
}

// FindByIDIncludingDeleted finds quiz by ID (includes soft deleted)
func FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	commons.Log(ctx, "[dao.FindByIDIncludingDeleted [id=%s", id)

	qry := `SELECT ` + quizSelectColumns + ` FROM quizzes WHERE id = $1`
	quiz, err := scanQuiz(db.QueryRowContext(ctx, qry, id))
	if err == sql.ErrNoRows {
		return nil, commons.NotFound("quiz-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByIDIncludingDeleted::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}
	return quiz, nil
}

// ListByCreator lists quizzes by creator ID with filters
func ListByCreator(ctx context.Context, creatorID uuid.UUID, filter model.QuizListFilter) ([]*model.Quiz, int, error) {
	commons.Log(ctx, "[dao.ListByCreator [creator_id=%s [page=%d [limit=%d", creatorID, filter.Page, filter.Limit)

	quizzes := []*model.Quiz{}
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

	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM quizzes %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByCreator::CountErr [err=%v", err)
		return quizzes, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	qry := fmt.Sprintf(`SELECT %s FROM quizzes %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, quizSelectColumns, where, argCount, argCount+1)
	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByCreator::SQLErr [err=%v", err)
		return quizzes, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		quiz, err := scanQuiz(rows)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByCreator::ScanErr [err=%v", err)
			continue
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
	where := "WHERE is_public = true AND is_published = true AND publish_status = 'published' AND deleted_at IS NULL"
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
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM quizzes %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListPublic::CountErr [err=%v", err)
		return quizzes, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	qry := fmt.Sprintf(`SELECT %s FROM quizzes %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, quizSelectColumns, where, argCount, argCount+1)
	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListPublic::SQLErr [err=%v", err)
		return quizzes, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		quiz, err := scanQuiz(rows)
		if err != nil {
			commons.LogE(ctx, "[dao.ListPublic::ScanErr [err=%v", err)
			continue
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
	status := "unpublished"
	if isPublished {
		status = "published"
	}
	return UpdatePublishWorkflowStatus(ctx, id, isPublished, status, nil)
}

// UpdatePublishWorkflowStatus updates public publish flag plus workflow status.
func UpdatePublishWorkflowStatus(ctx context.Context, id uuid.UUID, isPublished bool, status string, jobID *uuid.UUID) error {
	commons.Log(ctx, "[dao.UpdatePublishWorkflowStatus [id=%s [is_published=%v [status=%s", id, isPublished, status)

	qry := `UPDATE quizzes
			SET is_published = $1, publish_status = $2, ai_enrichment_job_id = COALESCE($3, ai_enrichment_job_id), updated_at = NOW()
			WHERE id = $4 AND deleted_at IS NULL`
	result, err := db.ExecContext(ctx, qry, isPublished, status, jobID, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdatePublishWorkflowStatus::SQLErr [id=%s [err=%v", id, err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("quiz-not-found")
	}
	return nil
}

// ApplyQuizEnrichment fills only missing quiz-level fields.
func ApplyQuizEnrichment(ctx context.Context, id uuid.UUID, difficulty, description string, tags []string) error {
	qry := `UPDATE quizzes
			SET difficulty = CASE WHEN COALESCE(difficulty, '') = '' THEN NULLIF($1, '') ELSE difficulty END,
				description = CASE WHEN COALESCE(description, '') = '' THEN NULLIF($2, '') ELSE description END,
				tags = CASE WHEN tags IS NULL OR cardinality(tags) = 0 THEN $3 ELSE tags END,
				meta = jsonb_set(COALESCE(meta, '{}'::jsonb), '{ai_enriched_at}', to_jsonb(NOW()::text), true),
				updated_at = NOW()
			WHERE id = $4 AND deleted_at IS NULL`
	_, err := db.ExecContext(ctx, qry, difficulty, description, pq.Array(tags), id)
	return err
}

// SoftDelete soft deletes a quiz
func SoftDelete(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.SoftDelete [id=%s", id)

	qry := `UPDATE quizzes SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
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

	qry := `UPDATE quizzes SET deleted_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NOT NULL`
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
	qry := `SELECT COUNT(*) FROM quizzes WHERE is_public = true AND is_published = true AND publish_status = 'published' AND deleted_at IS NULL`
	err := db.QueryRowContext(ctx, qry).Scan(&count)
	return count, err
}
