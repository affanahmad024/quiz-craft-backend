package handlers

import (
	"net/http"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/pkg/cache"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API is running and all dependencies are healthy
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "healthy"
	statusCode := http.StatusOK
	checks := make(map[string]interface{})

	// Check database
	if err := database.HealthCheck(); err != nil {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else {
		checks["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check Redis
	if err := cache.HealthCheck(); err != nil {
		checks["redis"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		status = "degraded"
		// Don't fail completely if only Redis is down
		if statusCode == http.StatusOK {
			statusCode = http.StatusOK // Keep 200 for degraded state
		}
	} else {
		checks["redis"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"checks":    checks,
	})
}

// Ping godoc
// @Summary Simple ping endpoint
// @Description Returns pong if the API is running
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
