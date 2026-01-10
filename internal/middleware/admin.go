package middleware

import (
	"github.com/affanahmad024/quiz-craft/backend/commons"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireAdmin middleware ensures the user has admin role
// Must be used after Auth() middleware
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := commons.Ctx(c)

		// Get user ID from context (set by Auth middleware)
		userID, exists := commons.UsrID(ctx)
		if !exists {
			commons.LogW(ctx, "[middleware.RequireAdmin::NoUserID")
			commons.Resp401(c, "authentication-required")
			c.Abort()
			return
		}

		// Get user from database to check role
		user, err := userdao.FindByID(ctx, userID)
		if err != nil {
			commons.LogW(ctx, "[middleware.RequireAdmin::FindUserErr [user_id=%s [err=%v", userID, err)
			commons.Resp404(c, "user-not-found")
			c.Abort()
			return
		}

		// Check if user is admin
		if user.Role != "admin" {
			commons.LogW(ctx, "[middleware.RequireAdmin::NotAdmin [user_id=%s [role=%s", userID, user.Role)
			commons.Resp403(c, "admin-access-required")
			c.Abort()
			return
		}

		// Set role in context for later use
		c.Set("user_role", user.Role)
		c.Next()
	}
}

// IsAdmin checks if the given user ID belongs to an admin
func IsAdmin(c *gin.Context, userID uuid.UUID) bool {
	ctx := commons.Ctx(c)

	user, err := userdao.FindByID(ctx, userID)
	if err != nil {
		return false
	}

	return user.Role == "admin"
}
