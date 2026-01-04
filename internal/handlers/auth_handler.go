package handlers

import (
	"net/http"

	"github.com/affanahmad024/quiz-craft/backend/internal/middleware"
	"github.com/affanahmad024/quiz-craft/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Signup handles user registration
// @Summary User signup
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body services.SignupRequest true "Signup request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req services.SignupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	user, err := h.authService.Signup(req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Signup failed", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusCreated, "Signup successful. Please check your email to verify your account.", gin.H{
		"user_id":                user.ID,
		"email":                  user.Email,
		"email_verification_sent": true,
	})
}

// VerifyEmail handles email verification
// @Summary Verify email
// @Description Verify user email with token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body map[string]string true "Verification request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.authService.VerifyEmail(userID, req.Token); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Email verification failed", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Email verified successfully", nil)
}

// ResendVerification resends verification email
// @Summary Resend verification email
// @Description Resend email verification link
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if err := h.authService.ResendVerificationEmail(req.Email); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Failed to resend verification email", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Verification email sent successfully", nil)
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "Login request"
// @Success 200 {object} services.TokenResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	tokenResponse, err := h.authService.Login(req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	// Set secure cookies
	c.SetCookie(
		"access_token",
		tokenResponse.AccessToken,
		int(tokenResponse.ExpiresIn),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HttpOnly
	)

	c.SetCookie(
		"refresh_token",
		tokenResponse.RefreshToken,
		7*24*60*60, // 7 days
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HttpOnly
	)

	middleware.RespondWithSuccess(c, http.StatusOK, "Login successful", tokenResponse)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} services.TokenResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Try to get from cookie
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			middleware.RespondWithError(c, http.StatusBadRequest, "Refresh token is required", err)
			return
		}
		req.RefreshToken = refreshToken
	}

	tokenResponse, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		middleware.RespondWithError(c, http.StatusUnauthorized, "Token refresh failed", err)
		return
	}

	// Update access token cookie
	c.SetCookie(
		"access_token",
		tokenResponse.AccessToken,
		int(tokenResponse.ExpiresIn),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HttpOnly
	)

	middleware.RespondWithSuccess(c, http.StatusOK, "Token refreshed successfully", tokenResponse)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user and clear cookies
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	middleware.RespondWithSuccess(c, http.StatusOK, "Logged out successfully", nil)
}

// ForgotPassword handles forgot password request
// @Summary Forgot password
// @Description Request password reset link
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if err := h.authService.ForgotPassword(req.Email); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "Failed to process request", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "If an account exists with this email, a password reset link has been sent", nil)
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset password using reset token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body map[string]string true "Reset request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.authService.ResetPassword(userID, req.Token, req.NewPassword); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Password reset failed", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Password reset successfully", nil)
}

// ChangePassword handles password change (requires authentication)
// @Summary Change password
// @Description Change user password (requires old password)
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Change password request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if err := h.authService.ChangePassword(userID.(uuid.UUID), req.CurrentPassword, req.NewPassword); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Password change failed", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Password changed successfully", nil)
}
