package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/google/uuid"
)

var jobDB *sql.DB

// InitJobDAO initializes the job DAO with database connection
func InitJobDAO(database *sql.DB) {
	jobDB = database
}

// CreateJob creates a new generation job
func CreateJob(ctx context.Context, job *model.GenerationJob) error {
	commons.Log(ctx, "[dao.CreateJob [user_id=%s [status=%s", job.UserID, job.Status)

	optionsJSON, err := json.Marshal(job.Options)
	if err != nil {
		commons.LogE(ctx, "[dao.CreateJob::JSONErr [err=%v", err)
		return err
	}

	qry := `INSERT INTO generation_jobs (
				id, user_id, file_id, text_content, options,
				status, progress, created_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8
			)`

	_, err = jobDB.ExecContext(ctx, qry,
		job.ID,
		job.UserID,
		job.FileID,
		job.TextContent,
		optionsJSON,
		job.Status,
		job.Progress,
		job.CreatedAt,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateJob::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateJob::OK [job_id=%s", job.ID)
	return nil
}

// GetJobByID retrieves a job by ID
func GetJobByID(ctx context.Context, id uuid.UUID) (*model.GenerationJob, error) {
	commons.Log(ctx, "[dao.GetJobByID [id=%s", id)

	job := &model.GenerationJob{}
	qry := `SELECT id, user_id, file_id, text_content, options, status,
			progress, quiz_id, error_message, started_at, completed_at, created_at
			FROM generation_jobs
			WHERE id = $1`

	var fileIDSQL sql.NullString
	var textContentSQL sql.NullString
	var optionsJSON []byte
	var quizIDSQL sql.NullString
	var errMsgSQL sql.NullString
	var startedAtSQL sql.NullTime
	var completedAtSQL sql.NullTime

	err := jobDB.QueryRowContext(ctx, qry, id).Scan(
		&job.ID,
		&job.UserID,
		&fileIDSQL,
		&textContentSQL,
		&optionsJSON,
		&job.Status,
		&job.Progress,
		&quizIDSQL,
		&errMsgSQL,
		&startedAtSQL,
		&completedAtSQL,
		&job.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("job-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.GetJobByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	// Handle nullable fields
	if fileIDSQL.Valid {
		job.FileID = &fileIDSQL.String
	}
	if textContentSQL.Valid {
		job.TextContent = &textContentSQL.String
	}
	if quizIDSQL.Valid {
		job.QuizID = &quizIDSQL.String
	}
	if errMsgSQL.Valid {
		job.ErrorMessage = &errMsgSQL.String
	}
	if startedAtSQL.Valid {
		job.StartedAt = &startedAtSQL.Time
	}
	if completedAtSQL.Valid {
		job.CompletedAt = &completedAtSQL.Time
	}

	// Parse options JSON
	if len(optionsJSON) > 0 {
		err = json.Unmarshal(optionsJSON, &job.Options)
		if err != nil {
			commons.LogE(ctx, "[dao.GetJobByID::JSONParseErr [err=%v", err)
			return nil, err
		}
	}

	return job, nil
}

// ListJobsByUserID lists jobs for a specific user with pagination
func ListJobsByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.GenerationJob, int, error) {
	commons.Log(ctx, "[dao.ListJobsByUserID [user_id=%s [page=%d [limit=%d", userID, page, limit)

	jobs := []*model.GenerationJob{}

	// Get total count
	var total int
	countQry := `SELECT COUNT(*) FROM generation_jobs WHERE user_id = $1`
	err := jobDB.QueryRowContext(ctx, countQry, userID).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListJobsByUserID::CountErr [err=%v", err)
		return jobs, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	qry := `SELECT id, user_id, file_id, text_content, options, status,
			progress, quiz_id, error_message, started_at, completed_at, created_at
			FROM generation_jobs
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`

	rows, err := jobDB.QueryContext(ctx, qry, userID, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.ListJobsByUserID::SQLErr [err=%v", err)
		return jobs, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		job := &model.GenerationJob{}
		var fileIDSQL sql.NullString
		var textContentSQL sql.NullString
		var optionsJSON []byte
		var quizIDSQL sql.NullString
		var errMsgSQL sql.NullString
		var startedAtSQL sql.NullTime
		var completedAtSQL sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&fileIDSQL,
			&textContentSQL,
			&optionsJSON,
			&job.Status,
			&job.Progress,
			&quizIDSQL,
			&errMsgSQL,
			&startedAtSQL,
			&completedAtSQL,
			&job.CreatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListJobsByUserID::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if fileIDSQL.Valid {
			job.FileID = &fileIDSQL.String
		}
		if textContentSQL.Valid {
			job.TextContent = &textContentSQL.String
		}
		if quizIDSQL.Valid {
			job.QuizID = &quizIDSQL.String
		}
		if errMsgSQL.Valid {
			job.ErrorMessage = &errMsgSQL.String
		}
		if startedAtSQL.Valid {
			job.StartedAt = &startedAtSQL.Time
		}
		if completedAtSQL.Valid {
			job.CompletedAt = &completedAtSQL.Time
		}

		// Parse options JSON
		if len(optionsJSON) > 0 {
			json.Unmarshal(optionsJSON, &job.Options)
		}

		jobs = append(jobs, job)
	}

	commons.Log(ctx, "[dao.ListJobsByUserID::OK [total=%d [returned=%d", total, len(jobs))
	return jobs, total, nil
}

// GetPendingJobs retrieves pending jobs for worker processing
func GetPendingJobs(ctx context.Context, limit int) ([]*model.GenerationJob, error) {
	commons.Log(ctx, "[dao.GetPendingJobs [limit=%d", limit)

	jobs := []*model.GenerationJob{}
	qry := `SELECT id, user_id, file_id, text_content, options, status,
			progress, quiz_id, error_message, started_at, completed_at, created_at
			FROM generation_jobs
			WHERE status = 'pending'
			ORDER BY created_at ASC
			LIMIT $1`

	rows, err := jobDB.QueryContext(ctx, qry, limit)
	if err != nil {
		commons.LogE(ctx, "[dao.GetPendingJobs::SQLErr [err=%v", err)
		return jobs, err
	}
	defer rows.Close()

	for rows.Next() {
		job := &model.GenerationJob{}
		var fileIDSQL sql.NullString
		var textContentSQL sql.NullString
		var optionsJSON []byte
		var quizIDSQL sql.NullString
		var errMsgSQL sql.NullString
		var startedAtSQL sql.NullTime
		var completedAtSQL sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&fileIDSQL,
			&textContentSQL,
			&optionsJSON,
			&job.Status,
			&job.Progress,
			&quizIDSQL,
			&errMsgSQL,
			&startedAtSQL,
			&completedAtSQL,
			&job.CreatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetPendingJobs::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if fileIDSQL.Valid {
			job.FileID = &fileIDSQL.String
		}
		if textContentSQL.Valid {
			job.TextContent = &textContentSQL.String
		}
		if quizIDSQL.Valid {
			job.QuizID = &quizIDSQL.String
		}
		if errMsgSQL.Valid {
			job.ErrorMessage = &errMsgSQL.String
		}
		if startedAtSQL.Valid {
			job.StartedAt = &startedAtSQL.Time
		}
		if completedAtSQL.Valid {
			job.CompletedAt = &completedAtSQL.Time
		}

		// Parse options JSON
		if len(optionsJSON) > 0 {
			json.Unmarshal(optionsJSON, &job.Options)
		}

		jobs = append(jobs, job)
	}

	commons.Log(ctx, "[dao.GetPendingJobs::OK [count=%d", len(jobs))
	return jobs, nil
}

// UpdateJobProgress updates job progress percentage
func UpdateJobProgress(ctx context.Context, id uuid.UUID, progress int) error {
	commons.Log(ctx, "[dao.UpdateJobProgress [id=%s [progress=%d", id, progress)

	qry := `UPDATE generation_jobs
			SET progress = $1
			WHERE id = $2`

	result, err := jobDB.ExecContext(ctx, qry, progress, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateJobProgress::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("job-not-found")
	}

	return nil
}

// UpdateJobStatus updates job status and optionally sets started_at
func UpdateJobStatus(ctx context.Context, id uuid.UUID, status string) error {
	commons.Log(ctx, "[dao.UpdateJobStatus [id=%s [status=%s", id, status)

	var qry string
	if status == "processing" {
		qry = `UPDATE generation_jobs
				SET status = $1, started_at = NOW()
				WHERE id = $2`
	} else {
		qry = `UPDATE generation_jobs
				SET status = $1
				WHERE id = $2`
	}

	result, err := jobDB.ExecContext(ctx, qry, status, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateJobStatus::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("job-not-found")
	}

	commons.Log(ctx, "[dao.UpdateJobStatus::OK [id=%s", id)
	return nil
}

// CompleteJob marks job as completed with quiz ID
func CompleteJob(ctx context.Context, id uuid.UUID, quizID string) error {
	commons.Log(ctx, "[dao.CompleteJob [id=%s [quiz_id=%s", id, quizID)

	qry := `UPDATE generation_jobs
			SET status = 'completed', progress = 100, quiz_id = $1, completed_at = NOW()
			WHERE id = $2`

	result, err := jobDB.ExecContext(ctx, qry, quizID, id)
	if err != nil {
		commons.LogE(ctx, "[dao.CompleteJob::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("job-not-found")
	}

	commons.Log(ctx, "[dao.CompleteJob::OK [id=%s", id)
	return nil
}

// FailJob marks job as failed with error message
func FailJob(ctx context.Context, id uuid.UUID, errorMessage string) error {
	commons.Log(ctx, "[dao.FailJob [id=%s [error=%s", id, errorMessage)

	qry := `UPDATE generation_jobs
			SET status = 'failed', error_message = $1, completed_at = NOW()
			WHERE id = $2`

	result, err := jobDB.ExecContext(ctx, qry, errorMessage, id)
	if err != nil {
		commons.LogE(ctx, "[dao.FailJob::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("job-not-found")
	}

	commons.Log(ctx, "[dao.FailJob::OK [id=%s", id)
	return nil
}

// DeleteOldJobs deletes completed/failed jobs older than specified duration
func DeleteOldJobs(ctx context.Context, olderThan time.Duration) (int, error) {
	commons.Log(ctx, "[dao.DeleteOldJobs [older_than=%v", olderThan)

	cutoffTime := time.Now().Add(-olderThan)
	qry := `DELETE FROM generation_jobs
			WHERE (status = 'completed' OR status = 'failed')
			AND completed_at < $1`

	result, err := jobDB.ExecContext(ctx, qry, cutoffTime)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteOldJobs::SQLErr [err=%v", err)
		return 0, err
	}

	rows, _ := result.RowsAffected()
	commons.Log(ctx, "[dao.DeleteOldJobs::OK [deleted=%d", rows)
	return int(rows), nil
}
