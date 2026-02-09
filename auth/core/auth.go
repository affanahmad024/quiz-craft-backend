package core

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/constants"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
)

var (
	jwtManager    *utils.JWTManager
	emailService  EmailService // Interface for email sending
)

// EmailService interface for sending emails
type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, name, token string, userID uuid.UUID) error
	SendPasswordResetEmail(ctx context.Context, email, name, token string, userID uuid.UUID) error
	SendPasswordChangedEmail(ctx context.Context, email, name string) error
	SendWelcomeEmail(ctx context.Context, email, name string) error
}

// Init initializes the auth core with dependencies
func Init(jwt *utils.JWTManager, email EmailService) {
	jwtManager = jwt
	emailService = email
}

// Signup handles user registration
func Signup(ctx context.Context, req model.SignupRequest) (model.SignupResponse, error) {
	commons.Log(ctx, "[core.Signup [email=%s [name=%s", req.Email, req.Name)

	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		commons.LogW(ctx, "[core.Signup::WeakPassword [email=%s [err=%v", req.Email, err)
		return model.SignupResponse{}, commons.BadInput("password-too-weak")
	}

	// Check if email already exists
	exists, err := dao.EmailExists(ctx, req.Email)
	if err != nil {
		commons.LogE(ctx, "[core.Signup::EmailExistsErr [email=%s [err=%v", req.Email, err)
		return model.SignupResponse{}, err
	}
	if exists {
		commons.LogW(ctx, "[core.Signup::EmailExists [email=%s", req.Email)
		return model.SignupResponse{}, commons.Conflict("email-already-exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		commons.LogE(ctx, "[core.Signup::HashErr [email=%s [err=%v", req.Email, err)
		return model.SignupResponse{}, err
	}

	// Create user
	userDAO, err := dao.CreateUser(ctx, req.Email, req.Name, req.PhoneNumber)
	if err != nil {
		commons.LogE(ctx, "[core.Signup::CreateUserErr [email=%s [err=%v", req.Email, err)
		return model.SignupResponse{}, err
	}

	// Create password record
	err = dao.CreatePassword(ctx, userDAO.ID, hashedPassword)
	if err != nil {
		commons.LogE(ctx, "[core.Signup::CreatePasswordErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.SignupResponse{}, err
	}

	// Generate verification token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		commons.LogE(ctx, "[core.Signup::TokenGenErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.SignupResponse{}, err
	}

	expiry := time.Now().Add(time.Duration(constants.EmailVerificationExpiry) * time.Hour)
	err = dao.SetVerificationToken(ctx, userDAO.ID, token, expiry)
	if err != nil {
		commons.LogE(ctx, "[core.Signup::SetTokenErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.SignupResponse{}, err
	}

	// Send verification email
	if emailService != nil {
		go emailService.SendVerificationEmail(ctx, req.Email, req.Name, token, userDAO.ID)
	} else {
		commons.LogW(ctx, "[core.Signup::EmailServiceNotConfigured [token=%s", token)
	}

	resp := model.SignupResponse{
		UserID:                userDAO.ID,
		Email:                 req.Email,
		EmailVerificationSent: true,
	}

	commons.Log(ctx, "[core.Signup::OK [user_id=%s [email=%s", userDAO.ID, req.Email)
	return resp, nil
}

// VerifyEmail verifies user's email with token
func VerifyEmail(ctx context.Context, req model.VerifyEmailRequest) error {
	commons.Log(ctx, "[core.VerifyEmail [user_id=%s", req.UserID)

	// Find password by user ID
	pwdDAO, err := dao.FindPasswordByUserID(ctx, req.UserID)
	if err != nil {
		commons.LogE(ctx, "[core.VerifyEmail::FindPasswordErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Validate token
	if !pwdDAO.VerificationToken.Valid || pwdDAO.VerificationToken.String != req.Token {
		commons.LogW(ctx, "[core.VerifyEmail::InvalidToken [user_id=%s", req.UserID)
		return commons.BadInput("invalid-verification-token")
	}

	// Check expiry
	if !pwdDAO.VerificationTokenExpiry.Valid || time.Now().After(pwdDAO.VerificationTokenExpiry.Time) {
		commons.LogW(ctx, "[core.VerifyEmail::TokenExpired [user_id=%s", req.UserID)
		return commons.BadInput("verification-token-expired")
	}

	// Update user as verified
	err = dao.UpdateEmailVerified(ctx, req.UserID, true)
	if err != nil {
		commons.LogE(ctx, "[core.VerifyEmail::UpdateErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Clear verification token
	err = dao.ClearVerificationToken(ctx, req.UserID)
	if err != nil {
		commons.LogE(ctx, "[core.VerifyEmail::ClearTokenErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Get user details
	userDAO, _ := dao.FindByID(ctx, req.UserID)

	// Send welcome email
	if emailService != nil {
		go emailService.SendWelcomeEmail(ctx, userDAO.Email, userDAO.Name)
	}

	commons.Log(ctx, "[core.VerifyEmail::OK [user_id=%s", req.UserID)
	return nil
}

// ResendVerification resends verification email
func ResendVerification(ctx context.Context, req model.ResendVerificationRequest) error {
	commons.Log(ctx, "[core.ResendVerification [email=%s", req.Email)

	// Find user by email
	userDAO, err := dao.FindByEmail(ctx, req.Email)
	if err != nil {
		commons.LogE(ctx, "[core.ResendVerification::FindErr [email=%s [err=%v", req.Email, err)
		return err
	}

	// Check if already verified
	if userDAO.EmailVerified {
		commons.LogW(ctx, "[core.ResendVerification::AlreadyVerified [user_id=%s", userDAO.ID)
		return commons.BadInput("email-already-verified")
	}

	// Generate new token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		commons.LogE(ctx, "[core.ResendVerification::TokenGenErr [user_id=%s [err=%v", userDAO.ID, err)
		return err
	}

	expiry := time.Now().Add(time.Duration(constants.EmailVerificationExpiry) * time.Hour)
	err = dao.SetVerificationToken(ctx, userDAO.ID, token, expiry)
	if err != nil {
		commons.LogE(ctx, "[core.ResendVerification::SetTokenErr [user_id=%s [err=%v", userDAO.ID, err)
		return err
	}

	// Send verification email
	if emailService != nil {
		go emailService.SendVerificationEmail(ctx, userDAO.Email, userDAO.Name, token, userDAO.ID)
	}

	commons.Log(ctx, "[core.ResendVerification::OK [user_id=%s", userDAO.ID)
	return nil
}

// Login handles user authentication
func Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error) {
	commons.Log(ctx, "[core.Login [email=%s", req.Email)

	// Find user by email
	userDAO, err := dao.FindByEmail(ctx, req.Email)
	if err != nil {
		commons.LogW(ctx, "[core.Login::UserNotFound [email=%s", req.Email)
		return model.LoginResponse{}, commons.Unauthorized("invalid-credentials")
	}

	// Check if user is enabled
	if !userDAO.Enabled {
		commons.LogW(ctx, "[core.Login::UserDisabled [user_id=%s", userDAO.ID)
		return model.LoginResponse{}, commons.Unauthorized("account-disabled")
	}

	// Find password
	pwdDAO, err := dao.FindPasswordByUserID(ctx, userDAO.ID)
	if err != nil {
		commons.LogE(ctx, "[core.Login::FindPasswordErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.LoginResponse{}, commons.Unauthorized("invalid-credentials")
	}

	// Check if account is locked (using helper method)
	if pwdDAO.IsAccountLocked() {
		meta := pwdDAO.GetMeta()
		commons.LogW(ctx, "[core.Login::AccountLocked [user_id=%s [until=%v", userDAO.ID, meta.AccountLockedUntil)
		return model.LoginResponse{}, commons.Unauthorized("account-locked")
	}

	// Verify password
	err = utils.ComparePassword(pwdDAO.PasswordHash, req.Password)
	if err != nil {
		commons.LogW(ctx, "[core.Login::InvalidPassword [user_id=%s", userDAO.ID)

		// Get current meta to check failed attempts
		meta := pwdDAO.GetMeta()

		// Increment failed attempts
		dao.IncrementFailedAttempts(ctx, userDAO.ID)

		// Check if we need to lock account (after incrementing)
		if meta.FailedLoginAttempts+1 >= constants.MaxFailedAttempts {
			lockUntil := time.Now().Add(time.Duration(constants.AccountLockoutMinutes) * time.Minute)
			dao.LockAccount(ctx, userDAO.ID, lockUntil)
			commons.LogW(ctx, "[core.Login::AccountLockedDueToFailedAttempts [user_id=%s", userDAO.ID)
			return model.LoginResponse{}, commons.Unauthorized("account-locked-due-to-failed-attempts")
		}

		return model.LoginResponse{}, commons.Unauthorized("invalid-credentials")
	}

	// Check if email is verified
	if !userDAO.EmailVerified {
		commons.LogW(ctx, "[core.Login::EmailNotVerified [user_id=%s", userDAO.ID)
		return model.LoginResponse{}, commons.Unauthorized("email-not-verified")
	}

	// Reset failed attempts on successful login
	dao.ResetFailedAttempts(ctx, userDAO.ID)

	// Update last login timestamp
	if err := dao.UpdateLastLogin(ctx, userDAO.ID); err != nil {
		commons.LogW(ctx, "[core.Login::UpdateLastLoginErr [user_id=%s [err=%v", userDAO.ID, err)
		// Don't fail login if this fails, just log the error
	}

	// Generate tokens
	accessToken, err := jwtManager.GenerateAccessToken(userDAO.ID, userDAO.Email)
	if err != nil {
		commons.LogE(ctx, "[core.Login::AccessTokenErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.LoginResponse{}, err
	}

	refreshToken, err := jwtManager.GenerateRefreshToken(userDAO.ID, userDAO.Email)
	if err != nil {
		commons.LogE(ctx, "[core.Login::RefreshTokenErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.LoginResponse{}, err
	}

	resp := model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(constants.AccessTokenExpiry * 3600), // Convert hours to seconds
		TokenType:    "Bearer",
		User:         userDAO.ToAPI(),
	}

	commons.Log(ctx, "[core.Login::OK [user_id=%s [email=%s", userDAO.ID, req.Email)
	return resp, nil
}

// ForgotPassword initiates password reset flow
func ForgotPassword(ctx context.Context, req model.ForgotPasswordRequest) error {
	commons.Log(ctx, "[core.ForgotPassword [email=%s", req.Email)

	// Find user by email
	userDAO, err := dao.FindByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists
		commons.LogW(ctx, "[core.ForgotPassword::UserNotFound [email=%s", req.Email)
		return nil
	}

	// Generate reset token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		commons.LogE(ctx, "[core.ForgotPassword::TokenGenErr [user_id=%s [err=%v", userDAO.ID, err)
		return err
	}

	expiry := time.Now().Add(time.Duration(constants.PasswordResetExpiry) * time.Hour)
	err = dao.SetResetToken(ctx, userDAO.ID, token, expiry)
	if err != nil {
		commons.LogE(ctx, "[core.ForgotPassword::SetTokenErr [user_id=%s [err=%v", userDAO.ID, err)
		return err
	}

	// Send reset email
	if emailService != nil {
		go emailService.SendPasswordResetEmail(ctx, userDAO.Email, userDAO.Name, token, userDAO.ID)
	}

	commons.Log(ctx, "[core.ForgotPassword::OK [user_id=%s", userDAO.ID)
	return nil
}

// ResetPassword resets password with token
func ResetPassword(ctx context.Context, req model.ResetPasswordRequest) error {
	commons.Log(ctx, "[core.ResetPassword [user_id=%s", req.UserID)

	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.NewPassword); err != nil {
		commons.LogW(ctx, "[core.ResetPassword::WeakPassword [user_id=%s", req.UserID)
		return commons.BadInput("password-too-weak")
	}

	// Find password by user ID
	pwdDAO, err := dao.FindPasswordByUserID(ctx, req.UserID)
	if err != nil {
		commons.LogE(ctx, "[core.ResetPassword::FindPasswordErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Validate token
	if !pwdDAO.ResetToken.Valid || pwdDAO.ResetToken.String != req.Token {
		commons.LogW(ctx, "[core.ResetPassword::InvalidToken [user_id=%s", req.UserID)
		return commons.BadInput("invalid-reset-token")
	}

	// Check expiry
	if !pwdDAO.ResetTokenExpiry.Valid || time.Now().After(pwdDAO.ResetTokenExpiry.Time) {
		commons.LogW(ctx, "[core.ResetPassword::TokenExpired [user_id=%s", req.UserID)
		return commons.BadInput("reset-token-expired")
	}

	// Check password history
	history := pwdDAO.GetPasswordHistory()
	for _, h := range history {
		if utils.ComparePassword(h.Hash, req.NewPassword) == nil {
			commons.LogW(ctx, "[core.ResetPassword::PasswordReused [user_id=%s", req.UserID)
			return commons.BadInput("password-recently-used")
		}
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		commons.LogE(ctx, "[core.ResetPassword::HashErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Add current password to history
	history = append([]model.PasswordHistory{{
		Hash:      pwdDAO.PasswordHash,
		ChangedAt: time.Now(),
	}}, history...)

	// Keep only last N passwords
	if len(history) > constants.PasswordHistoryCount {
		history = history[:constants.PasswordHistoryCount]
	}

	// Update password
	err = dao.UpdatePasswordHash(ctx, req.UserID, hashedPassword, history)
	if err != nil {
		commons.LogE(ctx, "[core.ResetPassword::UpdateErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Clear reset token
	err = dao.ClearResetToken(ctx, req.UserID)
	if err != nil {
		commons.LogE(ctx, "[core.ResetPassword::ClearTokenErr [user_id=%s [err=%v", req.UserID, err)
		return err
	}

	// Get user details
	userDAO, _ := dao.FindByID(ctx, req.UserID)

	// Send password changed email
	if emailService != nil {
		go emailService.SendPasswordChangedEmail(ctx, userDAO.Email, userDAO.Name)
	}

	commons.Log(ctx, "[core.ResetPassword::OK [user_id=%s", req.UserID)
	return nil
}

// ChangePassword changes user password (authenticated)
func ChangePassword(ctx context.Context, userID uuid.UUID, req model.ChangePasswordRequest) error {
	commons.Log(ctx, "[core.ChangePassword [user_id=%s", userID)

	// Validate new password strength
	if err := utils.ValidatePasswordStrength(req.NewPassword); err != nil {
		commons.LogW(ctx, "[core.ChangePassword::WeakPassword [user_id=%s", userID)
		return commons.BadInput("password-too-weak")
	}

	// Find password
	pwdDAO, err := dao.FindPasswordByUserID(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.ChangePassword::FindPasswordErr [user_id=%s [err=%v", userID, err)
		return err
	}

	// Verify current password
	err = utils.ComparePassword(pwdDAO.PasswordHash, req.CurrentPassword)
	if err != nil {
		commons.LogW(ctx, "[core.ChangePassword::InvalidCurrentPassword [user_id=%s", userID)
		return commons.Unauthorized("invalid-current-password")
	}

	// Check password history
	history := pwdDAO.GetPasswordHistory()
	for _, h := range history {
		if utils.ComparePassword(h.Hash, req.NewPassword) == nil {
			commons.LogW(ctx, "[core.ChangePassword::PasswordReused [user_id=%s", userID)
			return commons.BadInput("password-recently-used")
		}
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		commons.LogE(ctx, "[core.ChangePassword::HashErr [user_id=%s [err=%v", userID, err)
		return err
	}

	// Add current password to history
	history = append([]model.PasswordHistory{{
		Hash:      pwdDAO.PasswordHash,
		ChangedAt: time.Now(),
	}}, history...)

	// Keep only last N passwords
	if len(history) > constants.PasswordHistoryCount {
		history = history[:constants.PasswordHistoryCount]
	}

	// Update password
	err = dao.UpdatePasswordHash(ctx, userID, hashedPassword, history)
	if err != nil {
		commons.LogE(ctx, "[core.ChangePassword::UpdateErr [user_id=%s [err=%v", userID, err)
		return err
	}

	// Get user details
	userDAO, _ := dao.FindByID(ctx, userID)

	// Send password changed email
	if emailService != nil {
		go emailService.SendPasswordChangedEmail(ctx, userDAO.Email, userDAO.Name)
	}

	commons.Log(ctx, "[core.ChangePassword::OK [user_id=%s", userID)
	return nil
}

// RefreshToken refreshes access token using refresh token
func RefreshToken(ctx context.Context, refreshToken string) (model.LoginResponse, error) {
	commons.Log(ctx, "[core.RefreshToken")

	// Validate refresh token
	claims, err := jwtManager.ValidateToken(refreshToken)
	if err != nil {
		commons.LogW(ctx, "[core.RefreshToken::InvalidToken [err=%v", err)
		return model.LoginResponse{}, commons.Unauthorized("invalid-refresh-token")
	}

	// Check token type
	if claims.Type != constants.TokenTypeRefresh {
		commons.LogW(ctx, "[core.RefreshToken::WrongTokenType [type=%s", claims.Type)
		return model.LoginResponse{}, commons.Unauthorized("invalid-token-type")
	}

	// Get user
	userDAO, err := dao.FindByID(ctx, claims.UserID)
	if err != nil {
		commons.LogE(ctx, "[core.RefreshToken::FindUserErr [user_id=%s [err=%v", claims.UserID, err)
		return model.LoginResponse{}, commons.Unauthorized("user-not-found")
	}

	// Update last login timestamp (treating token refresh as activity)
	if err := dao.UpdateLastLogin(ctx, userDAO.ID); err != nil {
		commons.LogW(ctx, "[core.RefreshToken::UpdateLastLoginErr [user_id=%s [err=%v", userDAO.ID, err)
		// Don't fail refresh if this fails, just log the error
	}

	// Generate new access token
	accessToken, err := jwtManager.GenerateAccessToken(userDAO.ID, userDAO.Email)
	if err != nil {
		commons.LogE(ctx, "[core.RefreshToken::AccessTokenErr [user_id=%s [err=%v", userDAO.ID, err)
		return model.LoginResponse{}, err
	}

	resp := model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Return same refresh token
		ExpiresIn:    int64(constants.AccessTokenExpiry * 3600),
		TokenType:    "Bearer",
		User:         userDAO.ToAPI(),
	}

	commons.Log(ctx, "[core.RefreshToken::OK [user_id=%s", userDAO.ID)
	return resp, nil
}
