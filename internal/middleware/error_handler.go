package middleware

import (
	"net/http"

	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorHandler middleware handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				// Log the panic
				logger.GetLogger().WithFields(map[string]interface{}{
					"request_id": requestID,
					"panic":      err,
				}).Error("Panic recovered")

				// Return error response
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Internal server error",
					Error:   "An unexpected error occurred",
				})

				c.Abort()
			}
		}()

		c.Next()

		// Handle errors added by handlers
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID, _ := c.Get("request_id")

			logger.GetLogger().WithFields(map[string]interface{}{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Request failed")

			// If no response has been written yet
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Request failed",
					Error:   err.Error(),
				})
			}
		}
	}
}

// RespondWithError is a helper to send error responses
func RespondWithError(c *gin.Context, statusCode int, message string, err error) {
	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}

// RespondWithSuccess is a helper to send success responses
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}
