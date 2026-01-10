package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/google/uuid"
)

var fileDB *sql.DB

// InitFileDAO initializes the file DAO with database connection
func InitFileDAO(database *sql.DB) {
	fileDB = database
}

// CreateFile creates a new uploaded file record
func CreateFile(ctx context.Context, file *model.FileUpload) error {
	commons.Log(ctx, "[dao.CreateFile [user_id=%s [filename=%s", file.UserID, file.Filename)

	qry := `INSERT INTO uploaded_files (
				id, user_id, filename, original_filename, file_path,
				file_size, mime_type, status, created_at, expires_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			)`

	_, err := fileDB.ExecContext(ctx, qry,
		file.ID,
		file.UserID,
		file.Filename,
		file.OriginalFilename,
		file.FilePath,
		file.FileSize,
		file.MimeType,
		file.Status,
		file.CreatedAt,
		file.ExpiresAt,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateFile::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateFile::OK [file_id=%s", file.ID)
	return nil
}

// GetFileByID retrieves a file by ID
func GetFileByID(ctx context.Context, id uuid.UUID) (*model.FileUpload, error) {
	commons.Log(ctx, "[dao.GetFileByID [id=%s", id)

	file := &model.FileUpload{}
	qry := `SELECT id, user_id, filename, original_filename, file_path,
			file_size, mime_type, status, error_message, created_at, expires_at
			FROM uploaded_files
			WHERE id = $1`

	var errMsgSQL sql.NullString
	var expiresAtSQL sql.NullTime

	err := fileDB.QueryRowContext(ctx, qry, id).Scan(
		&file.ID,
		&file.UserID,
		&file.Filename,
		&file.OriginalFilename,
		&file.FilePath,
		&file.FileSize,
		&file.MimeType,
		&file.Status,
		&errMsgSQL,
		&file.CreatedAt,
		&expiresAtSQL,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("file-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.GetFileByID::SQLErr [id=%s [err=%v", id, err)
		return nil, err
	}

	// Handle nullable fields
	if errMsgSQL.Valid {
		file.ErrorMessage = &errMsgSQL.String
	}
	if expiresAtSQL.Valid {
		file.ExpiresAt = &expiresAtSQL.Time
	}

	return file, nil
}

// ListFilesByUserID lists files for a specific user with pagination
func ListFilesByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.FileUpload, int, error) {
	commons.Log(ctx, "[dao.ListFilesByUserID [user_id=%s [page=%d [limit=%d", userID, page, limit)

	files := []*model.FileUpload{}

	// Get total count
	var total int
	countQry := `SELECT COUNT(*) FROM uploaded_files WHERE user_id = $1`
	err := fileDB.QueryRowContext(ctx, countQry, userID).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListFilesByUserID::CountErr [err=%v", err)
		return files, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	qry := `SELECT id, user_id, filename, original_filename, file_path,
			file_size, mime_type, status, error_message, created_at, expires_at
			FROM uploaded_files
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`

	rows, err := fileDB.QueryContext(ctx, qry, userID, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.ListFilesByUserID::SQLErr [err=%v", err)
		return files, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		file := &model.FileUpload{}
		var errMsgSQL sql.NullString
		var expiresAtSQL sql.NullTime

		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Filename,
			&file.OriginalFilename,
			&file.FilePath,
			&file.FileSize,
			&file.MimeType,
			&file.Status,
			&errMsgSQL,
			&file.CreatedAt,
			&expiresAtSQL,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListFilesByUserID::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if errMsgSQL.Valid {
			file.ErrorMessage = &errMsgSQL.String
		}
		if expiresAtSQL.Valid {
			file.ExpiresAt = &expiresAtSQL.Time
		}

		files = append(files, file)
	}

	commons.Log(ctx, "[dao.ListFilesByUserID::OK [total=%d [returned=%d", total, len(files))
	return files, total, nil
}

// UpdateFileStatus updates file status and error message
func UpdateFileStatus(ctx context.Context, id uuid.UUID, status string, errorMessage *string) error {
	commons.Log(ctx, "[dao.UpdateFileStatus [id=%s [status=%s", id, status)

	qry := `UPDATE uploaded_files
			SET status = $1, error_message = $2
			WHERE id = $3`

	result, err := fileDB.ExecContext(ctx, qry, status, errorMessage, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateFileStatus::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("file-not-found")
	}

	commons.Log(ctx, "[dao.UpdateFileStatus::OK [id=%s", id)
	return nil
}

// DeleteFile deletes a file record
func DeleteFile(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.DeleteFile [id=%s", id)

	qry := `DELETE FROM uploaded_files WHERE id = $1`

	result, err := fileDB.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteFile::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("file-not-found")
	}

	commons.Log(ctx, "[dao.DeleteFile::OK [id=%s", id)
	return nil
}

// GetExpiredFiles retrieves files that have passed their expiration date
func GetExpiredFiles(ctx context.Context) ([]*model.FileUpload, error) {
	commons.Log(ctx, "[dao.GetExpiredFiles")

	files := []*model.FileUpload{}
	qry := `SELECT id, user_id, filename, original_filename, file_path,
			file_size, mime_type, status, error_message, created_at, expires_at
			FROM uploaded_files
			WHERE expires_at IS NOT NULL AND expires_at < $1`

	rows, err := fileDB.QueryContext(ctx, qry, time.Now())
	if err != nil {
		commons.LogE(ctx, "[dao.GetExpiredFiles::SQLErr [err=%v", err)
		return files, err
	}
	defer rows.Close()

	for rows.Next() {
		file := &model.FileUpload{}
		var errMsgSQL sql.NullString
		var expiresAtSQL sql.NullTime

		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Filename,
			&file.OriginalFilename,
			&file.FilePath,
			&file.FileSize,
			&file.MimeType,
			&file.Status,
			&errMsgSQL,
			&file.CreatedAt,
			&expiresAtSQL,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetExpiredFiles::ScanErr [err=%v", err)
			continue
		}

		// Handle nullable fields
		if errMsgSQL.Valid {
			file.ErrorMessage = &errMsgSQL.String
		}
		if expiresAtSQL.Valid {
			file.ExpiresAt = &expiresAtSQL.Time
		}

		files = append(files, file)
	}

	commons.Log(ctx, "[dao.GetExpiredFiles::OK [count=%d", len(files))
	return files, nil
}
