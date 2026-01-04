package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/internal/models"
	"github.com/affanahmad024/quiz-craft/backend/internal/repositories"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo     *repositories.UserRepository
	emailService *EmailService
	jwtManager   *utils.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repositories.UserRepository,
	emailService *EmailService,
	jwtManager *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: emailService,
		jwtManager:   jwtManager,
	}
}

// SignupRequest represents signup request data
type SignupRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
}

// LoginRequest represents login request data
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	TokenType    string    `json:"token_type"`
	User         models.UserResponse `json:"user"`
}

// Signup creates a new user account
func (s *AuthService) Signup(req SignupRequest) (*models.User, error) {
	// Validate and normalize input
	req.Email = utils.NormalizeEmail(req.Email)
	req.Name = utils.SanitizeString(req.Name)

	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}

	if err := utils.ValidateName(req.Name); err != nil {
		return nil, err
	}

	if err := utils.ValidatePhoneNumber(req.PhoneNumber); err != nil {
		return nil, err
	}

	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Check if email already exists (including soft deleted)
	exists, err := s.userRepo.EmailExistsIncludingDeleted(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}

	if exists {
		// Check if it's a soft deleted user
		user, err := s.userRepo.FindByEmail(req.Email)
		if err != nil {
			// Email exists but user is soft deleted
			return nil, errors.New("email already exists. Please contact support to recover your account")
		}

		// User exists and is active
		if user != nil {
			return nil, errors.New("email already registered")
		}
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate salt
	salt, err := utils.GenerateSalt(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate email verification token
	verificationToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	tokenExpiry := time.Now().Add(24 * time.Hour)

	// Create user
	user := &models.User{
		Email:         req.Email,
		Name:          req.Name,
		PhoneNumber:   req.PhoneNumber,
		EmailVerified: false,
		Enabled:       true,
		Meta:          models.Meta{},
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create user password
	userPassword := &models.UserPassword{
		UserID:                 user.ID,
		PasswordHash:           passwordHash,
		Salt:                   salt,
		EmailVerificationToken: &verificationToken,
		EmailTokenExpiresAt:    &tokenExpiry,
		TOTPEnabled:            false,
		Meta:                   models.PasswordMeta{},
	}

	if err := s.userRepo.CreatePassword(userPassword); err != nil {
		// Rollback user creation
		_ = s.userRepo.HardDelete(user.ID)
		return nil, fmt.Errorf("failed to create user password: %w", err)
	}

	// Send verification email (async - don't fail if email fails)
	go func() {
		if err := s.emailService.SendVerificationEmail(
			user.Email,
			user.Name,
			verificationToken,
			user.ID.String(),
		); err != nil {
			logger.GetLogger().Errorf("Failed to send verification email to %s: %v", user.Email, err)
		}
	}()

	logger.GetLogger().Infof("User created successfully: %s", user.Email)
	return user, nil
}

// VerifyEmail verifies user's email with token
func (s *AuthService) VerifyEmail(userID uuid.UUID, token string) error {
	user, err := s.userRepo.FindByIDWithPassword(userID)
	if err != nil {
		return errors.New("invalid verification link")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	if user.Password == nil {
		return errors.New("user password not found")
	}

	if user.Password.EmailVerificationToken == nil || *user.Password.EmailVerificationToken != token {
		return errors.New("invalid verification token")
	}

	if !user.Password.IsEmailTokenValid() {
		return errors.New("verification token has expired")
	}

	// Update user
	user.EmailVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Clear verification token
	user.Password.EmailVerificationToken = nil
	user.Password.EmailTokenExpiresAt = nil
	if err := s.userRepo.UpdatePassword(user.Password); err != nil {
		logger.GetLogger().Errorf("Failed to clear verification token: %v", err)
	}

	// Send welcome email (async)
	go func() {
		if err := s.emailService.SendWelcomeEmail(user.Email, user.Name); err != nil {
			logger.GetLogger().Errorf("Failed to send welcome email to %s: %v", user.Email, err)
		}
	}()

	logger.GetLogger().Infof("Email verified successfully: %s", user.Email)
	return nil
}

// ResendVerificationEmail resends verification email
func (s *AuthService) ResendVerificationEmail(email string) error {
	email = utils.NormalizeEmail(email)

	user, err := s.userRepo.FindByEmailWithPassword(email)
	if err != nil {
		return errors.New("user not found")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	if user.Password == nil {
		return errors.New("user password not found")
	}

	// Generate new token
	verificationToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	tokenExpiry := time.Now().Add(24 * time.Hour)

	// Update token
	user.Password.EmailVerificationToken = &verificationToken
	user.Password.EmailTokenExpiresAt = &tokenExpiry

	if err := s.userRepo.UpdatePassword(user.Password); err != nil {
		return fmt.Errorf("failed to update verification token: %w", err)
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(
		user.Email,
		user.Name,
		verificationToken,
		user.ID.String(),
	); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	logger.GetLogger().Infof("Verification email resent to: %s", user.Email)
	return nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req LoginRequest) (*TokenResponse, error) {
	req.Email = utils.NormalizeEmail(req.Email)

	user, err := s.userRepo.FindByEmailWithPassword(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if user.Password == nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if account is locked
	if user.Password.IsAccountLocked() {
		lockUntil := user.Password.Meta.AccountLockedUntil
		return nil, fmt.Errorf("account is locked until %s due to multiple failed login attempts", lockUntil.Format("2006-01-02 15:04:05"))
	}

	// Verify password
	if err := utils.ComparePassword(user.Password.PasswordHash, req.Password); err != nil {
		// Increment failed login attempts
		user.Password.IncrementFailedLoginAttempts()
		_ = s.userRepo.UpdatePassword(user.Password)

		return nil, errors.New("invalid email or password")
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, errors.New("account is disabled")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return nil, errors.New("email not verified. Please check your email for verification link")
	}

	// Reset failed login attempts on successful login
	user.Password.ResetFailedLoginAttempts()
	_ = s.userRepo.UpdatePassword(user.Password)

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresIn := s.jwtManager.GetTokenExpiryTime("access").Seconds()

	logger.GetLogger().Infof("User logged in successfully: %s", user.Email)

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiresIn),
		TokenType:    "Bearer",
		User:         user.ToResponse(),
	}, nil
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Get user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.Enabled {
		return nil, errors.New("account is disabled")
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	expiresIn := s.jwtManager.GetTokenExpiryTime("access").Seconds()

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Keep the same refresh token
		ExpiresIn:    int64(expiresIn),
		TokenType:    "Bearer",
		User:         user.ToResponse(),
	}, nil
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(email string) error {
	email = utils.NormalizeEmail(email)

	user, err := s.userRepo.FindByEmailWithPassword(email)
	if err != nil {
		// Don't reveal if email exists or not (security best practice)
		logger.GetLogger().Warnf("Password reset requested for non-existent email: %s", email)
		return nil
	}

	if user.Password == nil {
		return errors.New("user password not found")
	}

	// Generate reset token
	resetToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	tokenExpiry := time.Now().Add(1 * time.Hour)

	// Update reset token
	user.Password.PasswordResetToken = &resetToken
	user.Password.ResetTokenExpiresAt = &tokenExpiry

	if err := s.userRepo.UpdatePassword(user.Password); err != nil {
		return fmt.Errorf("failed to update reset token: %w", err)
	}

	// Send reset email
	if err := s.emailService.SendPasswordResetEmail(
		user.Email,
		user.Name,
		resetToken,
		user.ID.String(),
	); err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	logger.GetLogger().Infof("Password reset email sent to: %s", user.Email)
	return nil
}

// ResetPassword resets user password with token
func (s *AuthService) ResetPassword(userID uuid.UUID, token, newPassword string) error {
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	user, err := s.userRepo.FindByIDWithPassword(userID)
	if err != nil {
		return errors.New("invalid reset link")
	}

	if user.Password == nil {
		return errors.New("user password not found")
	}

	if user.Password.PasswordResetToken == nil || *user.Password.PasswordResetToken != token {
		return errors.New("invalid reset token")
	}

	if !user.Password.IsResetTokenValid() {
		return errors.New("reset token has expired")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if password was used before
	if user.Password.IsPasswordInHistory(newPasswordHash) {
		return errors.New("password was used recently. Please choose a different password")
	}

	// Add current password to history
	user.Password.AddPasswordToHistory()

	// Update password
	user.Password.PasswordHash = newPasswordHash
	user.Password.PasswordResetToken = nil
	user.Password.ResetTokenExpiresAt = nil

	if err := s.userRepo.UpdatePassword(user.Password); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Send confirmation email (async)
	go func() {
		if err := s.emailService.SendPasswordChangedEmail(user.Email, user.Name); err != nil {
			logger.GetLogger().Errorf("Failed to send password changed email to %s: %v", user.Email, err)
		}
	}()

	logger.GetLogger().Infof("Password reset successfully for: %s", user.Email)
	return nil
}

// ChangePassword changes user password (requires old password)
func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	user, err := s.userRepo.FindByIDWithPassword(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Password == nil {
		return errors.New("user password not found")
	}

	// Verify old password
	if err := utils.ComparePassword(user.Password.PasswordHash, oldPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if password was used before
	if user.Password.IsPasswordInHistory(newPasswordHash) {
		return errors.New("password was used recently. Please choose a different password")
	}

	// Add current password to history
	user.Password.AddPasswordToHistory()

	// Update password
	user.Password.PasswordHash = newPasswordHash

	if err := s.userRepo.UpdatePassword(user.Password); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Send confirmation email (async)
	go func() {
		if err := s.emailService.SendPasswordChangedEmail(user.Email, user.Name); err != nil {
			logger.GetLogger().Errorf("Failed to send password changed email to %s: %v", user.Email, err)
		}
	}()

	logger.GetLogger().Infof("Password changed successfully for: %s", user.Email)
	return nil
}
