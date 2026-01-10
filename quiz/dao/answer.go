package dao

import (
	"context"
	"database/sql"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// CreateAnswer creates a new answer
func CreateAnswer(ctx context.Context, answer *model.Answer) error {
	commons.Log(ctx, "[dao.CreateAnswer [question_id=%s [answer_text=%s", answer.QuestionID, answer.AnswerText)

	qry := `INSERT INTO answers (
				id, question_id, answer_text, is_correct, order_index,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, NOW(), NOW()
			) RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		answer.ID,
		answer.QuestionID,
		answer.AnswerText,
		answer.IsCorrect,
		answer.OrderIndex,
	).Scan(&answer.CreatedAt, &answer.UpdatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateAnswer::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateAnswer::OK [answer_id=%s", answer.ID)
	return nil
}

// FindAnswerByID finds answer by ID
func FindAnswerByID(ctx context.Context, id uuid.UUID) (*model.Answer, error) {
	commons.Log(ctx, "[dao.FindAnswerByID [id=%s", id)

	answer := &model.Answer{}
	qry := `SELECT id, question_id, answer_text, is_correct, order_index,
			created_at, updated_at
			FROM answers
			WHERE id = $1`

	err := db.QueryRowContext(ctx, qry, id).Scan(
		&answer.ID,
		&answer.QuestionID,
		&answer.AnswerText,
		&answer.IsCorrect,
		&answer.OrderIndex,
		&answer.CreatedAt,
		&answer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("answer-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindAnswerByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	return answer, nil
}

// ListByQuestionID lists all answers for a question
func ListByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*model.Answer, error) {
	commons.Log(ctx, "[dao.ListByQuestionID [question_id=%s", questionID)

	answers := []*model.Answer{}
	qry := `SELECT id, question_id, answer_text, is_correct, order_index,
			created_at, updated_at
			FROM answers
			WHERE question_id = $1
			ORDER BY order_index ASC`

	rows, err := db.QueryContext(ctx, qry, questionID)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByQuestionID::SQLErr [err=%v", err)
		return answers, err
	}
	defer rows.Close()

	for rows.Next() {
		answer := &model.Answer{}
		err := rows.Scan(
			&answer.ID,
			&answer.QuestionID,
			&answer.AnswerText,
			&answer.IsCorrect,
			&answer.OrderIndex,
			&answer.CreatedAt,
			&answer.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByQuestionID::ScanErr [err=%v", err)
			continue
		}

		answers = append(answers, answer)
	}

	commons.Log(ctx, "[dao.ListByQuestionID::OK [question_id=%s [count=%d", questionID, len(answers))
	return answers, nil
}

// ListByQuestionIDHideCorrect lists answers with is_correct hidden (for quiz attempts)
func ListByQuestionIDHideCorrect(ctx context.Context, questionID uuid.UUID) ([]model.AnswerListItem, error) {
	commons.Log(ctx, "[dao.ListByQuestionIDHideCorrect [question_id=%s", questionID)

	answers := []model.AnswerListItem{}
	qry := `SELECT id, answer_text, order_index
			FROM answers
			WHERE question_id = $1
			ORDER BY order_index ASC`

	rows, err := db.QueryContext(ctx, qry, questionID)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByQuestionIDHideCorrect::SQLErr [err=%v", err)
		return answers, err
	}
	defer rows.Close()

	for rows.Next() {
		answer := model.AnswerListItem{}
		err := rows.Scan(
			&answer.ID,
			&answer.AnswerText,
			&answer.OrderIndex,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByQuestionIDHideCorrect::ScanErr [err=%v", err)
			continue
		}

		answers = append(answers, answer)
	}

	commons.Log(ctx, "[dao.ListByQuestionIDHideCorrect::OK [question_id=%s [count=%d", questionID, len(answers))
	return answers, nil
}

// UpdateAnswer updates answer fields
func UpdateAnswer(ctx context.Context, answer *model.Answer) error {
	commons.Log(ctx, "[dao.UpdateAnswer [answer_id=%s", answer.ID)

	qry := `UPDATE answers
			SET answer_text = $1, is_correct = $2, order_index = $3, updated_at = NOW()
			WHERE id = $4`

	result, err := db.ExecContext(ctx, qry,
		answer.AnswerText,
		answer.IsCorrect,
		answer.OrderIndex,
		answer.ID,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.UpdateAnswer::SQLErr [answer_id=%s [err=%v", answer.ID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("answer-not-found")
	}

	commons.Log(ctx, "[dao.UpdateAnswer::OK [answer_id=%s", answer.ID)
	return nil
}

// DeleteAnswer deletes an answer
func DeleteAnswer(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.DeleteAnswer [id=%s", id)

	qry := `DELETE FROM answers WHERE id = $1`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteAnswer::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("answer-not-found")
	}

	commons.Log(ctx, "[dao.DeleteAnswer::OK [id=%s", id)
	return nil
}

// DeleteByQuestionID deletes all answers for a question
func DeleteByQuestionID(ctx context.Context, questionID uuid.UUID) error {
	commons.Log(ctx, "[dao.DeleteByQuestionID [question_id=%s", questionID)

	qry := `DELETE FROM answers WHERE question_id = $1`

	result, err := db.ExecContext(ctx, qry, questionID)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteByQuestionID::SQLErr [question_id=%s [err=%v", questionID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	commons.Log(ctx, "[dao.DeleteByQuestionID::OK [question_id=%s [deleted=%d", questionID, rows)
	return nil
}

// CountByQuestionID counts answers for a question
func CountByQuestionID(ctx context.Context, questionID uuid.UUID) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM answers WHERE question_id = $1`
	err := db.QueryRowContext(ctx, qry, questionID).Scan(&count)
	return count, err
}

// CountCorrectByQuestionID counts correct answers for a question
func CountCorrectByQuestionID(ctx context.Context, questionID uuid.UUID) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM answers WHERE question_id = $1 AND is_correct = true`
	err := db.QueryRowContext(ctx, qry, questionID).Scan(&count)
	return count, err
}

// ReorderAnswers updates order_index for multiple answers
func ReorderAnswers(ctx context.Context, updates []model.AnswerReorder) error {
	commons.Log(ctx, "[dao.ReorderAnswers [count=%d", len(updates))

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		commons.LogE(ctx, "[dao.ReorderAnswers::TxErr [err=%v", err)
		return err
	}
	defer tx.Rollback()

	qry := `UPDATE answers SET order_index = $1, updated_at = NOW() WHERE id = $2`

	for _, update := range updates {
		_, err := tx.ExecContext(ctx, qry, update.OrderIndex, update.AnswerID)
		if err != nil {
			commons.LogE(ctx, "[dao.ReorderAnswers::UpdateErr [answer_id=%s [err=%v", update.AnswerID, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		commons.LogE(ctx, "[dao.ReorderAnswers::CommitErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.ReorderAnswers::OK [count=%d", len(updates))
	return nil
}

// CreateAnswersBulk creates multiple answers in a transaction
func CreateAnswersBulk(ctx context.Context, answers []*model.Answer) error {
	commons.Log(ctx, "[dao.CreateAnswersBulk [count=%d", len(answers))

	if len(answers) == 0 {
		return nil
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		commons.LogE(ctx, "[dao.CreateAnswersBulk::TxErr [err=%v", err)
		return err
	}
	defer tx.Rollback()

	qry := `INSERT INTO answers (id, question_id, answer_text, is_correct, order_index, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`

	for _, answer := range answers {
		_, err := tx.ExecContext(ctx, qry,
			answer.ID,
			answer.QuestionID,
			answer.AnswerText,
			answer.IsCorrect,
			answer.OrderIndex,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.CreateAnswersBulk::InsertErr [err=%v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		commons.LogE(ctx, "[dao.CreateAnswersBulk::CommitErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateAnswersBulk::OK [count=%d", len(answers))
	return nil
}
