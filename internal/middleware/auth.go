package middleware

import (
	"net/http"
	"strings"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Auth is the authentication middleware
func Auth(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := commons.Ctx(c)

		// Try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" {
			// Extract Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				commons.LogW(ctx, "[middleware.Auth::InvalidAuthHeader")
				RespondWithError(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			// Try to get from cookie
			var err error
			tokenString, err = c.Cookie("access_token")
			if err != nil {
				commons.LogW(ctx, "[middleware.Auth::TokenNotFound")
				RespondWithError(c, http.StatusUnauthorized, "Authorization token not found", nil)
				c.Abort()
				return
			}
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			logger.GetLogger().Warnf("Invalid token: %v", err)
			commons.LogW(ctx, "[middleware.Auth::InvalidToken [err=%v", err)
			RespondWithError(c, http.StatusUnauthorized, "Invalid or expired token", err)
			c.Abort()
			return
		}

		// Check token type
		if claims.Type != "access" {
			commons.LogW(ctx, "[middleware.Auth::WrongTokenType [type=%s", claims.Type)
			RespondWithError(c, http.StatusUnauthorized, "Invalid token type", nil)
			c.Abort()
			return
		}

		// Set user info in context using commons
		commons.SetUserID(c, claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

// OptionalAuth is an optional authentication middleware
// It sets user info if token is present, but doesn't block if missing
func OptionalAuth(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		} else {
			tokenString, _ = c.Cookie("access_token")
		}

		// If token exists, validate and set user info
		if tokenString != "" {
			claims, err := jwtManager.ValidateToken(tokenString)
			if err == nil && claims.Type == "access" {
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}
