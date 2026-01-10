package unit

import (
	"context"
	"testing"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/auth/core"
	"github.com/affanahmad024/quiz-craft/backend/constants"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/tests"
	"github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmailService for testing
type MockEmailService struct {
	SentEmails []string
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, email, name, token string, userID uuid.UUID) error {
	m.SentEmails = append(m.SentEmails, "verification:"+email)
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, email, name, token string, userID uuid.UUID) error {
	m.SentEmails = append(m.SentEmails, "reset:"+email)
	return nil
}

func (m *MockEmailService) SendPasswordChangedEmail(ctx context.Context, email, name string) error {
	m.SentEmails = append(m.SentEmails, "changed:"+email)
	return nil
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	m.SentEmails = append(m.SentEmails, "welcome:"+email)
	return nil
}

func setupAuthTests(t *testing.T) *MockEmailService {
	tests.SetupTestDB(t)
	tests.CleanupTestData(t)

	// Initialize JWT manager
	jwtSecret := "test-secret-key-for-testing-only"
	jwtManager := utils.NewJWTManager(jwtSecret, 24*time.Hour, 7*24*time.Hour)

	// Initialize email service
	emailService := &MockEmailService{SentEmails: []string{}}

	// Initialize auth core
	core.Init(jwtManager, emailService)

	return emailService
}

func teardownAuthTests(t *testing.T) {
	tests.CleanupTestData(t)
	tests.TeardownTestDB(t)
}

// TestSignup_ValidInput tests successful signup with valid input
func TestSignup_ValidInput(t *testing.T) {
	emailService := setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()
	req := model.SignupRequest{
		Email:       "test@example.com",
		Name:        "Test User",
		Password:    "StrongPass123!",
		PhoneNumber: "+1234567890",
	}

	resp, err := core.Signup(ctx, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.UserID)
	assert.Equal(t, req.Email, resp.Email)
	assert.True(t, resp.EmailVerificationSent)
	assert.Contains(t, emailService.SentEmails, "verification:test@example.com")
}

// TestSignup_WeakPassword tests signup fails with weak password
func TestSignup_WeakPassword(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()
	req := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "123", // Too weak
	}

	_, err := core.Signup(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

// TestSignup_DuplicateEmail tests signup fails with existing email
func TestSignup_DuplicateEmail(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create first user
	req1 := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User 1",
		Password: "StrongPass123!",
	}
	_, err := core.Signup(ctx, req1)
	require.NoError(t, err)

	// Try to create user with same email
	req2 := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User 2",
		Password: "StrongPass456!",
	}
	_, err = core.Signup(ctx, req2)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

// TestVerifyEmail_ValidToken tests successful email verification
func TestVerifyEmail_ValidToken(t *testing.T) {
	emailService := setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create user
	req := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, req)
	require.NoError(t, err)

	// Get verification token from database
	pwdDAO, err := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	require.NoError(t, err)
	require.True(t, pwdDAO.VerificationToken.Valid)

	// Verify email
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	err = core.VerifyEmail(ctx, verifyReq)

	require.NoError(t, err)
	assert.Contains(t, emailService.SentEmails, "welcome:test@example.com")

	// Check user is verified
	userDAO, err := dao.FindByID(ctx, signupResp.UserID)
	require.NoError(t, err)
	assert.True(t, userDAO.EmailVerified)
}

// TestVerifyEmail_InvalidToken tests verification fails with invalid token
func TestVerifyEmail_InvalidToken(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create user
	req := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, req)
	require.NoError(t, err)

	// Try to verify with invalid token
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  "invalid-token",
	}
	err = core.VerifyEmail(ctx, verifyReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

// TestVerifyEmail_ExpiredToken tests verification fails with expired token
func TestVerifyEmail_ExpiredToken(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create user
	req := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, req)
	require.NoError(t, err)

	// Get verification token
	pwdDAO, err := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	require.NoError(t, err)

	// Set token to expired
	expiredTime := time.Now().Add(-1 * time.Hour)
	err = dao.SetVerificationToken(ctx, signupResp.UserID, pwdDAO.VerificationToken.String, expiredTime)
	require.NoError(t, err)

	// Try to verify with expired token
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	err = core.VerifyEmail(ctx, verifyReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestLogin_ValidCredentials tests successful login
func TestLogin_ValidCredentials(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	// Verify email
	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	err = core.VerifyEmail(ctx, verifyReq)
	require.NoError(t, err)

	// Login
	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "StrongPass123!",
	}
	loginResp, err := core.Login(ctx, loginReq)

	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.NotEmpty(t, loginResp.User.ID)
	assert.Equal(t, "test@example.com", loginResp.User.Email)
}

// TestLogin_InvalidPassword tests login fails with wrong password
func TestLogin_InvalidPassword(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	// Try login with wrong password
	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	}
	_, err = core.Login(ctx, loginReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials")
}

// TestLogin_EmailNotVerified tests login fails with unverified email
func TestLogin_EmailNotVerified(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create user but don't verify email
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	_, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	// Try login without verification
	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "StrongPass123!",
	}
	_, err = core.Login(ctx, loginReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "verified")
}

// TestLogin_AccountLocked tests login fails with locked account
func TestLogin_AccountLocked(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	// Fail login attempts to lock account
	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword",
	}

	for i := 0; i < constants.MaxFailedAttempts; i++ {
		core.Login(ctx, loginReq)
	}

	// Try login with correct password (should fail due to lock)
	loginReq.Password = "StrongPass123!"
	_, err = core.Login(ctx, loginReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "locked")
}

// TestLogin_DisabledAccount tests login fails with disabled account
func TestLogin_DisabledAccount(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	// Disable user account
	err = dao.UpdateEnabled(ctx, signupResp.UserID, false)
	require.NoError(t, err)

	// Try login
	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "StrongPass123!",
	}
	_, err = core.Login(ctx, loginReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

// TestChangePassword_ValidInput tests successful password change
func TestChangePassword_ValidInput(t *testing.T) {
	emailService := setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	// Change password
	changeReq := model.ChangePasswordRequest{
		CurrentPassword: "StrongPass123!",
		NewPassword:     "NewStrongPass456!",
	}
	err = core.ChangePassword(ctx, signupResp.UserID, changeReq)

	require.NoError(t, err)
	assert.Contains(t, emailService.SentEmails, "changed:test@example.com")

	// Verify can login with new password
	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "NewStrongPass456!",
	}
	loginResp, err := core.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
}

// TestChangePassword_WrongCurrentPassword tests change fails with wrong current password
func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	// Try to change password with wrong current password
	changeReq := model.ChangePasswordRequest{
		CurrentPassword: "WrongPassword123!",
		NewPassword:     "NewStrongPass456!",
	}
	err = core.ChangePassword(ctx, signupResp.UserID, changeReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

// TestChangePassword_ReusedPassword tests change fails when reusing recent password
func TestChangePassword_ReusedPassword(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create and verify user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	// Try to change password to same password
	changeReq := model.ChangePasswordRequest{
		CurrentPassword: "StrongPass123!",
		NewPassword:     "StrongPass123!",
	}
	err = core.ChangePassword(ctx, signupResp.UserID, changeReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

// TestRefreshToken_ValidToken tests successful token refresh
func TestRefreshToken_ValidToken(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Create, verify and login user
	signupReq := model.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "StrongPass123!",
	}
	signupResp, err := core.Signup(ctx, signupReq)
	require.NoError(t, err)

	pwdDAO, _ := dao.FindPasswordByUserID(ctx, signupResp.UserID)
	verifyReq := model.VerifyEmailRequest{
		UserID: signupResp.UserID,
		Token:  pwdDAO.VerificationToken.String,
	}
	core.VerifyEmail(ctx, verifyReq)

	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "StrongPass123!",
	}
	loginResp, err := core.Login(ctx, loginReq)
	require.NoError(t, err)

	// Refresh token
	refreshResp, err := core.RefreshToken(ctx, loginResp.RefreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken) // Should be different
	assert.Equal(t, loginResp.RefreshToken, refreshResp.RefreshToken) // Should be same
}

// TestRefreshToken_InvalidToken tests refresh fails with invalid token
func TestRefreshToken_InvalidToken(t *testing.T) {
	setupAuthTests(t)
	defer teardownAuthTests(t)

	ctx := context.Background()

	// Try to refresh with invalid token
	_, err := core.RefreshToken(ctx, "invalid-token")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}
