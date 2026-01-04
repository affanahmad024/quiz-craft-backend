package middleware

import (
	"log"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns CORS middleware with configuration
func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	// fallback for dev / misconfiguration safety
	if cfg == nil || len(cfg.AllowedOrigins) == 0 {
		log.Println("⚠️  CORS_ALLOWED_ORIGINS empty — falling back to AllowAllOrigins (DEV ONLY)")

		return cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		})
	}

	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

