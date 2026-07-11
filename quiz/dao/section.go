package dao

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

func scanSection(row interface {
	Scan(dest ...interface{}) error
}) (*model.QuizSection, error) {
	section := &model.QuizSection{}
	var descriptionSQL sql.NullString
	var instructionsSQL sql.NullString
	var timeLimitSQL sql.NullInt32
	var metaJSON []byte

	err := row.Scan(
		&section.ID,
		&section.QuizID,
		&section.Title,
		&descriptionSQL,
		&instructionsSQL,
		&section.OrderIndex,
		&timeLimitSQL,
		&section.IsDefault,
		&metaJSON,
		&section.CreatedAt,
		&section.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if descriptionSQL.Valid {
		section.Description = descriptionSQL.String
	}
	if instructionsSQL.Valid {
		section.Instructions = instructionsSQL.String
	}
	if timeLimitSQL.Valid {
		limit := int(timeLimitSQL.Int32)
		section.TimeLimitMinutes = &limit
	}
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &section.Meta)
	}

	return section, nil
}

const sectionSelectColumns = `id, quiz_id, title, description, instructions, order_index,
	time_limit_minutes, is_default, meta, created_at, updated_at`

// CreateDefaultSection ensures a quiz has its default General section.
func CreateDefaultSection(ctx context.Context, quizID uuid.UUID) (*model.QuizSection, error) {
	qry := `INSERT INTO quiz_sections (id, quiz_id, title, order_index, is_default, meta, created_at, updated_at)
			SELECT $1, $2, 'General', 0, true, '{}'::jsonb, NOW(), NOW()
			WHERE NOT EXISTS (
				SELECT 1 FROM quiz_sections WHERE quiz_id = $2 AND is_default = true
			)`

	sectionID := uuid.New()
	_, err := db.ExecContext(ctx, qry, sectionID, quizID)
	if err != nil {
		commons.LogE(ctx, "[dao.CreateDefaultSection::SQLErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	return GetDefaultSection(ctx, quizID)
}

// EnsureDefaultSection returns the default section, creating it when missing.
func EnsureDefaultSection(ctx context.Context, quizID uuid.UUID) (*model.QuizSection, error) {
	section, err := GetDefaultSection(ctx, quizID)
	if err == nil {
		return section, nil
	}
	if !commons.IsNotFound(err) {
		return nil, err
	}
	return CreateDefaultSection(ctx, quizID)
}

// GetDefaultSection returns a quiz's default section.
func GetDefaultSection(ctx context.Context, quizID uuid.UUID) (*model.QuizSection, error) {
	qry := `SELECT ` + sectionSelectColumns + ` FROM quiz_sections WHERE quiz_id = $1 AND is_default = true LIMIT 1`
	section, err := scanSection(db.QueryRowContext(ctx, qry, quizID))
	if err == sql.ErrNoRows {
		return nil, commons.NotFound("default-section-not-found")
	}
	return section, err
}

// FindSectionByID finds a section by ID.
func FindSectionByID(ctx context.Context, id uuid.UUID) (*model.QuizSection, error) {
	qry := `SELECT ` + sectionSelectColumns + ` FROM quiz_sections WHERE id = $1`
	section, err := scanSection(db.QueryRowContext(ctx, qry, id))
	if err == sql.ErrNoRows {
		return nil, commons.NotFound("section-not-found")
	}
	return section, err
}

// ListSectionsByQuizID lists sections in display order.
func ListSectionsByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.QuizSection, error) {
	sections := []*model.QuizSection{}
	qry := `SELECT ` + sectionSelectColumns + ` FROM quiz_sections WHERE quiz_id = $1 ORDER BY order_index ASC, created_at ASC`
	rows, err := db.QueryContext(ctx, qry, quizID)
	if err != nil {
		return sections, err
	}
	defer rows.Close()

	for rows.Next() {
		section, err := scanSection(rows)
		if err != nil {
			commons.LogE(ctx, "[dao.ListSectionsByQuizID::ScanErr [err=%v", err)
			continue
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

// CreateSection creates a non-default section.
func CreateSection(ctx context.Context, section *model.QuizSection) error {
	metaJSON, _ := json.Marshal(section.Meta)
	qry := `INSERT INTO quiz_sections (
				id, quiz_id, title, description, instructions, order_index, time_limit_minutes, is_default, meta, created_at, updated_at
			) VALUES (
				$1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, $8, $9, NOW(), NOW()
			) RETURNING created_at, updated_at`

	return db.QueryRowContext(ctx, qry,
		section.ID,
		section.QuizID,
		section.Title,
		section.Description,
		section.Instructions,
		section.OrderIndex,
		section.TimeLimitMinutes,
		section.IsDefault,
		metaJSON,
	).Scan(&section.CreatedAt, &section.UpdatedAt)
}

// UpdateSection updates section fields.
func UpdateSection(ctx context.Context, section *model.QuizSection) error {
	metaJSON, _ := json.Marshal(section.Meta)
	qry := `UPDATE quiz_sections
			SET title = $1, description = NULLIF($2, ''), instructions = NULLIF($3, ''),
				time_limit_minutes = $4, meta = $5, updated_at = NOW()
			WHERE id = $6`
	result, err := db.ExecContext(ctx, qry,
		section.Title,
		section.Description,
		section.Instructions,
		section.TimeLimitMinutes,
		metaJSON,
		section.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("section-not-found")
	}
	return nil
}

// NextSectionOrderIndex returns the next section order for a quiz.
func NextSectionOrderIndex(ctx context.Context, quizID uuid.UUID) (int, error) {
	var next sql.NullInt32
	qry := `SELECT COALESCE(MAX(order_index) + 1, 0) FROM quiz_sections WHERE quiz_id = $1`
	err := db.QueryRowContext(ctx, qry, quizID).Scan(&next)
	if err != nil {
		return 0, err
	}
	if !next.Valid {
		return 0, nil
	}
	return int(next.Int32), nil
}

// NextSectionQuestionOrderIndex returns the next question order within a section.
func NextSectionQuestionOrderIndex(ctx context.Context, sectionID uuid.UUID) (int, error) {
	var next sql.NullInt32
	qry := `SELECT COALESCE(MAX(section_order_index) + 1, 0) FROM questions WHERE section_id = $1`
	err := db.QueryRowContext(ctx, qry, sectionID).Scan(&next)
	if err != nil {
		return 0, err
	}
	if !next.Valid {
		return 0, nil
	}
	return int(next.Int32), nil
}

// DeleteSectionMoveQuestionsToDefault deletes a non-default section and moves its questions to General.
func DeleteSectionMoveQuestionsToDefault(ctx context.Context, quizID, sectionID uuid.UUID) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var isDefault bool
	err = tx.QueryRowContext(ctx, `SELECT is_default FROM quiz_sections WHERE id = $1 AND quiz_id = $2`, sectionID, quizID).Scan(&isDefault)
	if err == sql.ErrNoRows {
		return commons.NotFound("section-not-found")
	}
	if err != nil {
		return err
	}
	if isDefault {
		return commons.BadInput("default-section-cannot-be-deleted")
	}

	var defaultSectionID uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT id FROM quiz_sections WHERE quiz_id = $1 AND is_default = true`, quizID).Scan(&defaultSectionID)
	if err == sql.ErrNoRows {
		return commons.NotFound("default-section-not-found")
	}
	if err != nil {
		return err
	}

	var startIndex int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(section_order_index) + 1, 0) FROM questions WHERE section_id = $1`, defaultSectionID).Scan(&startIndex)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `WITH moved AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY section_order_index ASC, order_index ASC, created_at ASC, id ASC) - 1 AS rn
			FROM questions
			WHERE quiz_id = $1 AND section_id = $2
		)
		UPDATE questions q
		SET section_id = $3, section_order_index = $4 + moved.rn, updated_at = NOW()
		FROM moved
		WHERE q.id = moved.id`, quizID, sectionID, defaultSectionID, startIndex)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM quiz_sections WHERE id = $1 AND quiz_id = $2`, sectionID, quizID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("section-not-found")
	}

	return tx.Commit()
}

// ReorderSections updates section ordering for a quiz.
func ReorderSections(ctx context.Context, quizID uuid.UUID, updates []model.SectionReorderItem) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE quiz_sections SET order_index = order_index + 100000 WHERE quiz_id = $1`, quizID)
	if err != nil {
		return err
	}

	for _, update := range updates {
		result, err := tx.ExecContext(ctx, `UPDATE quiz_sections SET order_index = $1, updated_at = NOW() WHERE id = $2 AND quiz_id = $3`, update.OrderIndex, update.SectionID, quizID)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return commons.NotFound("section-not-found")
		}
	}

	return tx.Commit()
}

// ReorderQuestionsInSections updates global and per-section question ordering.
func ReorderQuestionsInSections(ctx context.Context, quizID uuid.UUID, updates []model.QuestionSectionReorderItem) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE questions SET section_order_index = section_order_index + 100000 WHERE quiz_id = $1`, quizID)
	if err != nil {
		return err
	}

	for _, update := range updates {
		var sectionQuizID string
		err = tx.QueryRowContext(ctx, `SELECT quiz_id FROM quiz_sections WHERE id = $1`, update.SectionID).Scan(&sectionQuizID)
		if err == sql.ErrNoRows {
			return commons.NotFound("section-not-found")
		}
		if err != nil {
			return err
		}
		if sectionQuizID != quizID.String() {
			return commons.BadInput("section-does-not-belong-to-quiz")
		}

		result, err := tx.ExecContext(ctx, `UPDATE questions
			SET section_id = $1, section_order_index = $2, order_index = $3, updated_at = NOW()
			WHERE id = $4 AND quiz_id = $5`, update.SectionID, update.SectionOrderIndex, update.OrderIndex, update.QuestionID, quizID)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return commons.NotFound("question-not-found")
		}
	}

	return tx.Commit()
}
