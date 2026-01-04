package handlers

import (
	"net/http"

	"github.com/affanahmad024/quiz-craft/backend/internal/middleware"
	"github.com/affanahmad024/quiz-craft/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user endpoints
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetMe retrieves current user profile
// @Summary Get current user profile
// @Description Get authenticated user's profile
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.userService.GetProfile(userID.(uuid.UUID))
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve profile", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Profile retrieved successfully", user.ToResponse())
}

// UpdateMe updates current user profile
// @Summary Update current user profile
// @Description Update authenticated user's profile
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.UpdateProfileRequest true "Update request"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/me [patch]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req services.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	user, err := h.userService.UpdateProfile(userID.(uuid.UUID), req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Failed to update profile", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Profile updated successfully", user.ToResponse())
}

// DeleteMe performs soft delete on current user account
// @Summary Delete current user account
// @Description Soft delete authenticated user's account
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/me [delete]
func (h *UserHandler) DeleteMe(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.userService.DeleteAccount(userID.(uuid.UUID)); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "Failed to delete account", err)
		return
	}

	// Clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	middleware.RespondWithSuccess(c, http.StatusOK, "Account deleted successfully. You can contact support within 30 days to recover your account.", nil)
}

// GetUserByID retrieves a user by ID (admin endpoint)
// @Summary Get user by ID
// @Description Get any user's profile by ID (admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "User not found", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "User retrieved successfully", user.ToResponse())
}

// ListUsers lists all users with pagination (admin endpoint)
// @Summary List users
// @Description List all users with pagination and filters (admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Get pagination parameters
	page := 1
	limit := 20

	if p, exists := c.GetQuery("page"); exists {
		if parsedPage, err := uuid.Parse(p); err == nil {
			page = int(parsedPage.ID())
		}
	}

	if l, exists := c.GetQuery("limit"); exists {
		if parsedLimit, err := uuid.Parse(l); err == nil {
			limit = int(parsedLimit.ID())
		}
	}

	// Get filters (can be extended)
	filters := make(map[string]interface{})

	users, total, err := h.userService.ListUsers(page, limit, filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	// Convert to response format
	userResponses := make([]interface{}, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Users retrieved successfully", gin.H{
		"users": userResponses,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetStats retrieves user statistics (admin endpoint)
// @Summary Get user statistics
// @Description Get user statistics including total, active, and deleted counts
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/stats [get]
func (h *UserHandler) GetStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve stats", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Statistics retrieved successfully", stats)
}

// RestoreAccount restores a soft-deleted account (admin endpoint)
// @Summary Restore deleted account
// @Description Restore a soft-deleted user account (admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/{id}/restore [post]
func (h *UserHandler) RestoreAccount(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.userService.RestoreAccount(userID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Failed to restore account", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Account restored successfully", nil)
}

// DisableAccount disables a user account (admin endpoint)
// @Summary Disable user account
// @Description Disable a user account (admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/{id}/disable [post]
func (h *UserHandler) DisableAccount(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.userService.DisableAccount(userID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Failed to disable account", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Account disabled successfully", nil)
}

// EnableAccount enables a user account (admin endpoint)
// @Summary Enable user account
// @Description Enable a user account (admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /users/{id}/enable [post]
func (h *UserHandler) EnableAccount(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.userService.EnableAccount(userID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "Failed to enable account", err)
		return
	}

	middleware.RespondWithSuccess(c, http.StatusOK, "Account enabled successfully", nil)
}
