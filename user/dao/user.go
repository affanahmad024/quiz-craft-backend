package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
)

var db *sql.DB

// Init initializes the DAO with database connection
func Init(database *sql.DB) {
	db = database
}

// CreateUser creates a new user
func CreateUser(ctx context.Context, email, name, phoneNumber string) (model.UserDAO, error) {
	commons.Log(ctx, "[dao.CreateUser [email=%s [name=%s", email, name)

	user := model.UserDAO{}
	qry := `INSERT INTO users (email, name, phone_number, email_verified, enabled, created_at, updated_at, meta)
			VALUES ($1, $2, $3, false, true, NOW(), NOW(), '{}'::jsonb)
			RETURNING id, email, name, phone_number, email_verified, enabled, created_at, updated_at, deleted_at, meta`

	var phoneSQL sql.NullString
	if phoneNumber != "" {
		phoneSQL = sql.NullString{String: phoneNumber, Valid: true}
	}

	err := db.QueryRowContext(ctx, qry, email, name, phoneSQL).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PhoneNumber,
		&user.EmailVerified,
		&user.Enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Meta,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateUser::SQLErr [email=%s [err=%v", email, err)
		return user, err
	}

	commons.Log(ctx, "[dao.CreateUser::OK [user_id=%s", user.ID)
	return user, nil
}

// FindByID finds user by ID (excludes soft deleted)
func FindByID(ctx context.Context, id uuid.UUID) (model.UserDAO, error) {
	commons.Log(ctx, "[dao.FindByID [id=%s", id)

	user := model.UserDAO{}
	qry := `SELECT id, email, name, phone_number, role, email_verified, enabled, created_at, updated_at, deleted_at, meta
			FROM users
			WHERE id = $1 AND deleted_at IS NULL`

	err := db.QueryRowContext(ctx, qry, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PhoneNumber,
		&user.Role,
		&user.EmailVerified,
		&user.Enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Meta,
	)

	if err == sql.ErrNoRows {
		return user, commons.NotFound("user-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByID::SQLErr [id=%s [err=%v", id, err)
		return user, err
	}

	return user, nil
}

// FindByIDIncludingDeleted finds user by ID (includes soft deleted)
func FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (model.UserDAO, error) {
	commons.Log(ctx, "[dao.FindByIDIncludingDeleted [id=%s", id)

	user := model.UserDAO{}
	qry := `SELECT id, email, name, phone_number, email_verified, enabled, created_at, updated_at, deleted_at, meta
			FROM users
			WHERE id = $1`

	err := db.QueryRowContext(ctx, qry, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PhoneNumber,
		&user.EmailVerified,
		&user.Enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Meta,
	)

	if err == sql.ErrNoRows {
		return user, commons.NotFound("user-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByIDIncludingDeleted::SQLErr [id=%s [err=%v", id, err)
		return user, err
	}

	return user, nil
}

// FindByEmail finds user by email (excludes soft deleted)
func FindByEmail(ctx context.Context, email string) (model.UserDAO, error) {
	commons.Log(ctx, "[dao.FindByEmail [email=%s", email)

	user := model.UserDAO{}
	qry := `SELECT id, email, name, phone_number, email_verified, enabled, created_at, updated_at, deleted_at, meta
			FROM users
			WHERE email = $1 AND deleted_at IS NULL`

	err := db.QueryRowContext(ctx, qry, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PhoneNumber,
		&user.EmailVerified,
		&user.Enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Meta,
	)

	if err == sql.ErrNoRows {
		return user, commons.NotFound("user-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByEmail::SQLErr [email=%s [err=%v", email, err)
		return user, err
	}

	return user, nil
}

// EmailExists checks if email exists (excludes soft deleted)
func EmailExists(ctx context.Context, email string) (bool, error) {
	commons.Log(ctx, "[dao.EmailExists [email=%s", email)

	var exists bool
	qry := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	err := db.QueryRowContext(ctx, qry, email).Scan(&exists)
	if err != nil {
		commons.LogE(ctx, "[dao.EmailExists::SQLErr [email=%s [err=%v", email, err)
		return false, err
	}

	return exists, nil
}

// EmailExistsIncludingDeleted checks if email exists (includes soft deleted)
func EmailExistsIncludingDeleted(ctx context.Context, email string) (bool, error) {
	var exists bool
	qry := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := db.QueryRowContext(ctx, qry, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateUser updates user fields
func UpdateUser(ctx context.Context, id uuid.UUID, name, phoneNumber string, meta map[string]interface{}) error {
	commons.Log(ctx, "[dao.UpdateUser [id=%s", id)

	metaJSON, _ := json.Marshal(meta)
	qry := `UPDATE users
			SET name = $1, phone_number = $2, meta = $3, updated_at = NOW()
			WHERE id = $4 AND deleted_at IS NULL`

	var phoneSQL sql.NullString
	if phoneNumber != "" {
		phoneSQL = sql.NullString{String: phoneNumber, Valid: true}
	}

	result, err := db.ExecContext(ctx, qry, name, phoneSQL, metaJSON, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateUser::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found")
	}

	return nil
}

// UpdateEmailVerified updates email verified status
func UpdateEmailVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	commons.Log(ctx, "[dao.UpdateEmailVerified [id=%s [verified=%v", id, verified)

	qry := `UPDATE users SET email_verified = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry, verified, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateEmailVerified::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found")
	}

	return nil
}

// UpdateEnabled updates enabled status
func UpdateEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	commons.Log(ctx, "[dao.UpdateEnabled [id=%s [enabled=%v", id, enabled)

	qry := `UPDATE users SET enabled = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry, enabled, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateEnabled::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found")
	}

	return nil
}

// UpdateRole updates user role (Phase 7: Admin)
func UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	commons.Log(ctx, "[dao.UpdateRole [id=%s [role=%s", id, role)

	qry := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry, role, id)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateRole::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found")
	}

	return nil
}

// SoftDelete soft deletes a user
func SoftDelete(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.SoftDelete [id=%s", id)

	qry := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.SoftDelete::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found")
	}

	commons.Log(ctx, "[dao.SoftDelete::OK [id=%s", id)
	return nil
}

// Restore restores a soft deleted user
func Restore(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.Restore [id=%s", id)

	qry := `UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NOT NULL`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.Restore::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found-or-not-deleted")
	}

	commons.Log(ctx, "[dao.Restore::OK [id=%s", id)
	return nil
}

// HardDelete permanently deletes a user
func HardDelete(ctx context.Context, id uuid.UUID) error {
	commons.Log(ctx, "[dao.HardDelete [level=CRITICAL [id=%s", id)

	qry := `DELETE FROM users WHERE id = $1`

	result, err := db.ExecContext(ctx, qry, id)
	if err != nil {
		commons.LogE(ctx, "[dao.HardDelete::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("user-not-found")
	}

	commons.Log(ctx, "[dao.HardDelete::OK [id=%s", id)
	return nil
}

// ListUsers lists users with pagination and filters
func ListUsers(ctx context.Context, filter model.UserListFilter) ([]model.UserDAO, int, error) {
	commons.Log(ctx, "[dao.ListUsers [page=%d [limit=%d", filter.Page, filter.Limit)

	users := []model.UserDAO{}

	// Build WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if !filter.IncludeDeleted {
		where += " AND deleted_at IS NULL"
	}

	if filter.EmailVerified != nil {
		where += fmt.Sprintf(" AND email_verified = $%d", argCount)
		args = append(args, *filter.EmailVerified)
		argCount++
	}

	if filter.Enabled != nil {
		where += fmt.Sprintf(" AND enabled = $%d", argCount)
		args = append(args, *filter.Enabled)
		argCount++
	}

	// Get total count
	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUsers::CountErr [err=%v", err)
		return users, 0, err
	}

	// Get paginated results
	offset := (filter.Page - 1) * filter.Limit
	qry := fmt.Sprintf(`SELECT id, email, name, phone_number, email_verified, enabled, created_at, updated_at, deleted_at, meta
						FROM users %s
						ORDER BY created_at DESC
						LIMIT $%d OFFSET $%d`, where, argCount, argCount+1)

	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUsers::SQLErr [err=%v", err)
		return users, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.UserDAO
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.PhoneNumber,
			&user.EmailVerified,
			&user.Enabled,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
			&user.Meta,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListUsers::ScanErr [err=%v", err)
			continue
		}
		users = append(users, user)
	}

	commons.Log(ctx, "[dao.ListUsers::OK [total=%d [returned=%d", total, len(users))
	return users, total, nil
}

// CountUsers counts active users
func CountUsers(ctx context.Context) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := db.QueryRowContext(ctx, qry).Scan(&count)
	return count, err
}

// CountDeletedUsers counts deleted users
func CountDeletedUsers(ctx context.Context) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM users WHERE deleted_at IS NOT NULL`
	err := db.QueryRowContext(ctx, qry).Scan(&count)
	return count, err
}
