package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
)

// CreatePassword creates password record for user
func CreatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	commons.Log(ctx, "[dao.CreatePassword [user_id=%s", userID)

	qry := `INSERT INTO user_passwords (user_id, password_hash, password_history, failed_login_attempts,
			must_change_password, two_factor_enabled, last_password_change, meta, created_at, updated_at)
			VALUES ($1, $2, '[]'::jsonb, 0, false, false, NOW(), '{}'::jsonb, NOW(), NOW())`

	_, err := db.ExecContext(ctx, qry, userID, passwordHash)
	if err != nil {
		commons.LogE(ctx, "[dao.CreatePassword::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// FindPasswordByUserID finds password by user ID
func FindPasswordByUserID(ctx context.Context, userID uuid.UUID) (model.UserPasswordDAO, error) {
	commons.Log(ctx, "[dao.FindPasswordByUserID [user_id=%s", userID)

	pwd := model.UserPasswordDAO{}
	qry := `SELECT id, user_id, password_hash, password_history, verification_token, verification_token_expiry,
			reset_token, reset_token_expiry, failed_login_attempts, last_failed_login, account_locked_until,
			last_password_change, must_change_password, two_factor_enabled, two_factor_secret, meta, created_at, updated_at
			FROM user_passwords
			WHERE user_id = $1`

	err := db.QueryRowContext(ctx, qry, userID).Scan(
		&pwd.ID,
		&pwd.UserID,
		&pwd.PasswordHash,
		&pwd.PasswordHistory,
		&pwd.VerificationToken,
		&pwd.VerificationTokenExpiry,
		&pwd.ResetToken,
		&pwd.ResetTokenExpiry,
		&pwd.FailedLoginAttempts,
		&pwd.LastFailedLogin,
		&pwd.AccountLockedUntil,
		&pwd.LastPasswordChange,
		&pwd.MustChangePassword,
		&pwd.TwoFactorEnabled,
		&pwd.TwoFactorSecret,
		&pwd.Meta,
		&pwd.CreatedAt,
		&pwd.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return pwd, commons.NotFound("password-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindPasswordByUserID::SQLErr [user_id=%s [err=%v", userID, err)
		return pwd, err
	}

	return pwd, nil
}

// UpdatePasswordHash updates password hash and adds to history
func UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string, history []model.PasswordHistory) error {
	commons.Log(ctx, "[dao.UpdatePasswordHash [user_id=%s", userID)

	historyJSON, _ := json.Marshal(history)
	qry := `UPDATE user_passwords
			SET password_hash = $1, password_history = $2, last_password_change = NOW(), updated_at = NOW()
			WHERE user_id = $3`

	result, err := db.ExecContext(ctx, qry, newHash, historyJSON, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdatePasswordHash::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("password-not-found")
	}

	return nil
}

// SetVerificationToken sets email verification token
func SetVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) error {
	commons.Log(ctx, "[dao.SetVerificationToken [user_id=%s", userID)

	qry := `UPDATE user_passwords
			SET verification_token = $1, verification_token_expiry = $2, updated_at = NOW()
			WHERE user_id = $3`

	_, err := db.ExecContext(ctx, qry, token, expiry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.SetVerificationToken::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// ClearVerificationToken clears email verification token
func ClearVerificationToken(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.ClearVerificationToken [user_id=%s", userID)

	qry := `UPDATE user_passwords
			SET verification_token = NULL, verification_token_expiry = NULL, updated_at = NOW()
			WHERE user_id = $1`

	_, err := db.ExecContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.ClearVerificationToken::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// FindByVerificationToken finds password by verification token
func FindByVerificationToken(ctx context.Context, token string) (model.UserPasswordDAO, error) {
	commons.Log(ctx, "[dao.FindByVerificationToken")

	pwd := model.UserPasswordDAO{}
	qry := `SELECT id, user_id, password_hash, password_history, verification_token, verification_token_expiry,
			reset_token, reset_token_expiry, failed_login_attempts, last_failed_login, account_locked_until,
			last_password_change, must_change_password, two_factor_enabled, two_factor_secret, meta, created_at, updated_at
			FROM user_passwords
			WHERE verification_token = $1`

	err := db.QueryRowContext(ctx, qry, token).Scan(
		&pwd.ID,
		&pwd.UserID,
		&pwd.PasswordHash,
		&pwd.PasswordHistory,
		&pwd.VerificationToken,
		&pwd.VerificationTokenExpiry,
		&pwd.ResetToken,
		&pwd.ResetTokenExpiry,
		&pwd.FailedLoginAttempts,
		&pwd.LastFailedLogin,
		&pwd.AccountLockedUntil,
		&pwd.LastPasswordChange,
		&pwd.MustChangePassword,
		&pwd.TwoFactorEnabled,
		&pwd.TwoFactorSecret,
		&pwd.Meta,
		&pwd.CreatedAt,
		&pwd.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return pwd, commons.NotFound("token-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByVerificationToken::SQLErr [err=%v", err)
		return pwd, err
	}

	return pwd, nil
}

// SetResetToken sets password reset token
func SetResetToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) error {
	commons.Log(ctx, "[dao.SetResetToken [user_id=%s", userID)

	qry := `UPDATE user_passwords
			SET reset_token = $1, reset_token_expiry = $2, updated_at = NOW()
			WHERE user_id = $3`

	_, err := db.ExecContext(ctx, qry, token, expiry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.SetResetToken::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// ClearResetToken clears password reset token
func ClearResetToken(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.ClearResetToken [user_id=%s", userID)

	qry := `UPDATE user_passwords
			SET reset_token = NULL, reset_token_expiry = NULL, updated_at = NOW()
			WHERE user_id = $1`

	_, err := db.ExecContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.ClearResetToken::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// FindByResetToken finds password by reset token
func FindByResetToken(ctx context.Context, token string) (model.UserPasswordDAO, error) {
	commons.Log(ctx, "[dao.FindByResetToken")

	pwd := model.UserPasswordDAO{}
	qry := `SELECT id, user_id, password_hash, password_history, verification_token, verification_token_expiry,
			reset_token, reset_token_expiry, failed_login_attempts, last_failed_login, account_locked_until,
			last_password_change, must_change_password, two_factor_enabled, two_factor_secret, meta, created_at, updated_at
			FROM user_passwords
			WHERE reset_token = $1`

	err := db.QueryRowContext(ctx, qry, token).Scan(
		&pwd.ID,
		&pwd.UserID,
		&pwd.PasswordHash,
		&pwd.PasswordHistory,
		&pwd.VerificationToken,
		&pwd.VerificationTokenExpiry,
		&pwd.ResetToken,
		&pwd.ResetTokenExpiry,
		&pwd.FailedLoginAttempts,
		&pwd.LastFailedLogin,
		&pwd.AccountLockedUntil,
		&pwd.LastPasswordChange,
		&pwd.MustChangePassword,
		&pwd.TwoFactorEnabled,
		&pwd.TwoFactorSecret,
		&pwd.Meta,
		&pwd.CreatedAt,
		&pwd.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return pwd, commons.NotFound("token-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByResetToken::SQLErr [err=%v", err)
		return pwd, err
	}

	return pwd, nil
}

// IncrementFailedAttempts increments failed login attempts
func IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.IncrementFailedAttempts [user_id=%s", userID)

	qry := `UPDATE user_passwords
			SET failed_login_attempts = failed_login_attempts + 1, last_failed_login = NOW(), updated_at = NOW()
			WHERE user_id = $1`

	_, err := db.ExecContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.IncrementFailedAttempts::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// ResetFailedAttempts resets failed login attempts to 0
func ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.ResetFailedAttempts [user_id=%s", userID)

	qry := `UPDATE user_passwords
			SET failed_login_attempts = 0, account_locked_until = NULL, updated_at = NOW()
			WHERE user_id = $1`

	_, err := db.ExecContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.ResetFailedAttempts::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}

// LockAccount locks user account until specified time
func LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error {
	commons.Log(ctx, "[dao.LockAccount [user_id=%s [until=%v", userID, until)

	qry := `UPDATE user_passwords
			SET account_locked_until = $1, updated_at = NOW()
			WHERE user_id = $2`

	_, err := db.ExecContext(ctx, qry, until, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.LockAccount::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	return nil
}
