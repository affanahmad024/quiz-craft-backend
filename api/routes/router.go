package routes

import (
	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/internal/handlers"
	"github.com/affanahmad024/quiz-craft/backend/internal/middleware"
	"github.com/affanahmad024/quiz-craft/backend/internal/repositories"
	"github.com/affanahmad024/quiz-craft/backend/internal/services"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures all routes
func SetupRouter(cfg *config.Config) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())                     // Recover from panics
	router.Use(middleware.ErrorHandler())          // Custom error handler
	router.Use(middleware.Logger())                // Custom logger
	router.Use(middleware.CORS(&cfg.CORS))         // CORS

	// Initialize dependencies
	db := database.GetDB()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpire,
		cfg.JWT.RefreshTokenExpire,
	)

	// Initialize services
	emailService := services.NewEmailService(&cfg.Email)
	authService := services.NewAuthService(userRepo, emailService, jwtManager)
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		// Health check routes
		public.GET("/health", healthHandler.HealthCheck)
		public.GET("/ping", healthHandler.Ping)

		// Authentication routes
		auth := public.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/v1")
	protected.Use(middleware.Auth(jwtManager))
	{
		// Auth routes that require authentication
		auth := protected.Group("/auth")
		{
			auth.POST("/change-password", authHandler.ChangePassword)
		}

		// User routes
		users := protected.Group("/users")
		{
			// Current user endpoints
			users.GET("/me", userHandler.GetMe)
			users.PATCH("/me", userHandler.UpdateMe)
			users.DELETE("/me", userHandler.DeleteMe)

			// Admin endpoints (TODO: Add admin middleware in future)
			users.GET("", userHandler.ListUsers)
			users.GET("/stats", userHandler.GetStats)
			users.GET("/:id", userHandler.GetUserByID)
			users.POST("/:id/restore", userHandler.RestoreAccount)
			users.POST("/:id/disable", userHandler.DisableAccount)
			users.POST("/:id/enable", userHandler.EnableAccount)
		}

		// Quiz routes - TODO: Implement in Phase 3
		// quizzes := protected.Group("/quizzes")
		// {
		//     quizzes.POST("", quizHandler.Create)
		//     quizzes.GET("", quizHandler.List)
		//     quizzes.GET("/:id", quizHandler.GetByID)
		//     quizzes.PUT("/:id", quizHandler.Update)
		//     quizzes.DELETE("/:id", quizHandler.Delete)
		// }
	}

	return router
}
