package dao

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const questionSelectColumns = `q.id, q.quiz_id, q.section_id, q.question_text, q.question_type, q.difficulty,
	q.tags, q.points, q.order_index, q.section_order_index, q.explanation, q.meta, q.created_at, q.updated_at`

func scanQuestion(row interface {
	Scan(dest ...interface{}) error
}) (*model.Question, error) {
	question := &model.Question{}
	var difficultySQL sql.NullString
	var explanationSQL sql.NullString
	var metaJSON []byte

	err := row.Scan(
		&question.ID,
		&question.QuizID,
		&question.SectionID,
		&question.QuestionText,
		&question.QuestionType,
		&difficultySQL,
		&question.Tags,
		&question.Points,
		&question.OrderIndex,
		&question.SectionOrderIndex,
		&explanationSQL,
		&metaJSON,
		&question.CreatedAt,
		&question.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if difficultySQL.Valid {
		question.Difficulty = difficultySQL.String
	}
	if explanationSQL.Valid {
		question.Explanation = explanationSQL.String
	}
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &question.Meta)
	}
	return question, nil
}

// CreateQuestion creates a new question
func CreateQuestion(ctx context.Context, question *model.Question) error {
	commons.Log(ctx, "[dao.CreateQuestion [quiz_id=%s [question_text=%s", question.QuizID, question.QuestionText)

	metaJSON, _ := json.Marshal(question.Meta)
	qry := `INSERT INTO questions (
				id, quiz_id, section_id, question_text, question_type, difficulty, tags, points, order_index,
				section_order_index, explanation, meta, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10, $11, $12, NOW(), NOW()
			) RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		question.ID,
		question.QuizID,
		question.SectionID,
		question.QuestionText,
		question.QuestionType,
		question.Difficulty,
		pq.Array(question.Tags),
		question.Points,
		question.OrderIndex,
		question.SectionOrderIndex,
		question.Explanation,
		metaJSON,
	).Scan(&question.CreatedAt, &question.UpdatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateQuestion::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateQuestion::OK [question_id=%s", question.ID)
	return nil
}

// FindQuestionByID finds question by ID
func FindQuestionByID(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	commons.Log(ctx, "[dao.FindQuestionByID [id=%s", id)

	qry := `SELECT ` + questionSelectColumns + `
			FROM questions q
			WHERE q.id = $1`
	question, err := scanQuestion(db.QueryRowContext(ctx, qry, id))
	if err == sql.ErrNoRows {
		return nil, commons.NotFound("question-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindQuestionByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}
	return question, nil
}

// ListByQuizID lists all questions for a quiz
func ListByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error) {
	commons.Log(ctx, "[dao.ListByQuizID [quiz_id=%s", quizID)

	questions := []*model.Question{}
	qry := `SELECT ` + questionSelectColumns + `
			FROM questions q
			LEFT JOIN quiz_sections s ON s.id = q.section_id
			WHERE q.quiz_id = $1
			ORDER BY COALESCE(s.order_index, 0) ASC, q.section_order_index ASC, q.order_index ASC, q.created_at ASC`

	rows, err := db.QueryContext(ctx, qry, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByQuizID::SQLErr [err=%v", err)
		return questions, err
	}
	defer rows.Close()

	for rows.Next() {
		question, err := scanQuestion(rows)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByQuizID::ScanErr [err=%v", err)
			continue
		}
		questions = append(questions, question)
	}

	commons.Log(ctx, "[dao.ListByQuizID::OK [quiz_id=%s [count=%d", quizID, len(questions))
	return questions, rows.Err()
}

// UpdateQuestion updates question fields
func UpdateQuestion(ctx context.Context, question *model.Question) error {
	commons.Log(ctx, "[dao.UpdateQuestion [question_id=%s", question.ID)

	metaJSON, _ := json.Marshal(question.Meta)
	qry := `UPDATE questions
			SET section_id = $1, question_text = $2, question_type = $3, difficulty = NULLIF($4, ''), tags = $5,
				points = $6, order_index = $7, section_order_index = $8, explanation = $9, meta = $10, updated_at = NOW()
			WHERE id = $11`

	result, err := db.ExecContext(ctx, qry,
		question.SectionID,
		question.QuestionText,
		question.QuestionType,
		question.Difficulty,
		pq.Array(question.Tags),
		question.Points,
		question.OrderIndex,
		question.SectionOrderIndex,
		question.Explanation,
		metaJSON,
		question.ID,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.UpdateQuestion::SQLErr [question_id=%s [err=%v", question.ID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("question-not-found")
	}

	commons.Log(ctx, "[dao.UpdateQuestion::OK [question_id=%s", question.ID)
	return nil
}

// DeleteQuestion deletes a question (CASCADE will delete answers)
func DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.DeleteQuestion [id=%s", id)

	qry := `DELETE FROM questions WHERE id = $1`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteQuestion::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("question-not-found")
	}

	commons.Log(ctx, "[dao.DeleteQuestion::OK [id=%s", id)
	return nil
}

// DeleteByQuizID deletes all questions for a quiz
func DeleteByQuizID(ctx context.Context, quizID uuid.UUID) error {
	commons.Log(ctx, "[dao.DeleteByQuizID [quiz_id=%s", quizID)

	qry := `DELETE FROM questions WHERE quiz_id = $1`

	result, err := db.ExecContext(ctx, qry, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteByQuizID::SQLErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	commons.Log(ctx, "[dao.DeleteByQuizID::OK [quiz_id=%s [deleted=%d", quizID, rows)
	return nil
}

// CountByQuizID counts questions for a quiz
func CountByQuizID(ctx context.Context, quizID uuid.UUID) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM questions WHERE quiz_id = $1`
	err := db.QueryRowContext(ctx, qry, quizID).Scan(&count)
	return count, err
}

// ReorderQuestions updates order_index for multiple questions
func ReorderQuestions(ctx context.Context, updates []model.QuestionReorder) error {
	commons.Log(ctx, "[dao.ReorderQuestions [count=%d", len(updates))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		commons.LogE(ctx, "[dao.ReorderQuestions::TxErr [err=%v", err)
		return err
	}
	defer tx.Rollback()

	qry := `UPDATE questions SET order_index = $1, updated_at = NOW() WHERE id = $2`

	for _, update := range updates {
		_, err := tx.ExecContext(ctx, qry, update.OrderIndex, update.QuestionID)
		if err != nil {
			commons.LogE(ctx, "[dao.ReorderQuestions::UpdateErr [question_id=%s [err=%v", update.QuestionID, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		commons.LogE(ctx, "[dao.ReorderQuestions::CommitErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.ReorderQuestions::OK [count=%d", len(updates))
	return nil
}

// CreateQuestionsBulk creates multiple questions in a transaction
func CreateQuestionsBulk(ctx context.Context, questions []*model.Question) error {
	commons.Log(ctx, "[dao.CreateQuestionsBulk [count=%d", len(questions))

	if len(questions) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		commons.LogE(ctx, "[dao.CreateQuestionsBulk::TxErr [err=%v", err)
		return err
	}
	defer tx.Rollback()

	qry := `INSERT INTO questions (
				id, quiz_id, section_id, question_text, question_type, difficulty, tags, points, order_index,
				section_order_index, explanation, meta, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10, $11, $12, NOW(), NOW()
			) RETURNING created_at, updated_at`

	for _, question := range questions {
		metaJSON, _ := json.Marshal(question.Meta)
		err := tx.QueryRowContext(ctx, qry,
			question.ID,
			question.QuizID,
			question.SectionID,
			question.QuestionText,
			question.QuestionType,
			question.Difficulty,
			pq.Array(question.Tags),
			question.Points,
			question.OrderIndex,
			question.SectionOrderIndex,
			question.Explanation,
			metaJSON,
		).Scan(&question.CreatedAt, &question.UpdatedAt)

		if err != nil {
			commons.LogE(ctx, "[dao.CreateQuestionsBulk::InsertErr [question_id=%s [err=%v", question.ID, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		commons.LogE(ctx, "[dao.CreateQuestionsBulk::CommitErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateQuestionsBulk::OK [count=%d", len(questions))
	return nil
}

// ApplyQuestionEnrichment fills only missing question-level fields.
func ApplyQuestionEnrichment(ctx context.Context, id uuid.UUID, difficulty, explanation string, tags []string) error {
	qry := `UPDATE questions
			SET difficulty = CASE WHEN COALESCE(difficulty, '') = '' THEN NULLIF($1, '') ELSE difficulty END,
				explanation = CASE WHEN COALESCE(explanation, '') = '' THEN NULLIF($2, '') ELSE explanation END,
				tags = CASE WHEN tags IS NULL OR cardinality(tags) = 0 THEN $3 ELSE tags END,
				updated_at = NOW()
			WHERE id = $4`
	_, err := db.ExecContext(ctx, qry, difficulty, explanation, pq.Array(tags), id)
	return err
}
