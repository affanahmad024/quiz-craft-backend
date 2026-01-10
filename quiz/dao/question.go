package dao

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// CreateQuestion creates a new question
func CreateQuestion(ctx context.Context, question *model.Question) error {
	commons.Log(ctx, "[dao.CreateQuestion [quiz_id=%s [question_text=%s", question.QuizID, question.QuestionText)

	metaJSON, _ := json.Marshal(question.Meta)
	qry := `INSERT INTO questions (
				id, quiz_id, question_text, question_type, points, order_index,
				explanation, meta, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
			) RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		question.ID,
		question.QuizID,
		question.QuestionText,
		question.QuestionType,
		question.Points,
		question.OrderIndex,
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

	question := &model.Question{}
	qry := `SELECT id, quiz_id, question_text, question_type, points, order_index,
			explanation, meta, created_at, updated_at
			FROM questions
			WHERE id = $1`

	var explanationSQL sql.NullString
	var metaJSON []byte

	err := db.QueryRowContext(ctx, qry, id).Scan(
		&question.ID,
		&question.QuizID,
		&question.QuestionText,
		&question.QuestionType,
		&question.Points,
		&question.OrderIndex,
		&explanationSQL,
		&metaJSON,
		&question.CreatedAt,
		&question.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("question-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindQuestionByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	// Handle nullable fields
	if explanationSQL.Valid {
		question.Explanation = explanationSQL.String
	}

	// Parse meta JSON
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &question.Meta)
	}

	return question, nil
}

// ListByQuizID lists all questions for a quiz
func ListByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error) {
	commons.Log(ctx, "[dao.ListByQuizID [quiz_id=%s", quizID)

	questions := []*model.Question{}
	qry := `SELECT id, quiz_id, question_text, question_type, points, order_index,
			explanation, meta, created_at, updated_at
			FROM questions
			WHERE quiz_id = $1
			ORDER BY order_index ASC`

	rows, err := db.QueryContext(ctx, qry, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByQuizID::SQLErr [err=%v", err)
		return questions, err
	}
	defer rows.Close()

	for rows.Next() {
		question := &model.Question{}
		var explanationSQL sql.NullString
		var metaJSON []byte

		err := rows.Scan(
			&question.ID,
			&question.QuizID,
			&question.QuestionText,
			&question.QuestionType,
			&question.Points,
			&question.OrderIndex,
			&explanationSQL,
			&metaJSON,
			&question.CreatedAt,
			&question.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByQuizID::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if explanationSQL.Valid {
			question.Explanation = explanationSQL.String
		}

		// Parse meta JSON
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &question.Meta)
		}

		questions = append(questions, question)
	}

	commons.Log(ctx, "[dao.ListByQuizID::OK [quiz_id=%s [count=%d", quizID, len(questions))
	return questions, nil
}

// UpdateQuestion updates question fields
func UpdateQuestion(ctx context.Context, question *model.Question) error {
	commons.Log(ctx, "[dao.UpdateQuestion [question_id=%s", question.ID)

	metaJSON, _ := json.Marshal(question.Meta)
	qry := `UPDATE questions
			SET question_text = $1, question_type = $2, points = $3, order_index = $4,
				explanation = $5, meta = $6, updated_at = NOW()
			WHERE id = $7`

	result, err := db.ExecContext(ctx, qry,
		question.QuestionText,
		question.QuestionType,
		question.Points,
		question.OrderIndex,
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

	// Start transaction
	tx, err := db.Begin()
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
