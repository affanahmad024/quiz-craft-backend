package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	aiagentdao "github.com/affanahmad024/quiz-craft/backend/ai-agent/dao"
	aiworker "github.com/affanahmad024/quiz-craft/backend/ai-agent/worker"
	"github.com/affanahmad024/quiz-craft/backend/api"
	authcore "github.com/affanahmad024/quiz-craft/backend/auth/core"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	_ "github.com/affanahmad024/quiz-craft/backend/docs" // Swagger docs
	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/internal/middleware"
	"github.com/affanahmad024/quiz-craft/backend/internal/services"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	quizworker "github.com/affanahmad024/quiz-craft/backend/quiz/worker"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
)

// @title QuizCraft API
// @version 1.0
// @description A comprehensive quiz management and taking platform with AI-powered quiz generation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@quizcraft.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var (
	cfg        *config.Config
	log        *logrus.Logger
	jwtManager *utils.JWTManager
	sqlDB      *sql.DB
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.App.LogLevel)
	log = logger.GetLogger()
	commons.SetLogLevel(cfg.App.LogLevel)

	log.Info("Starting QuizCraft Backend API")
	log.Infof("Environment: %s", cfg.App.Env)

	// Initialize PostgreSQL
	log.Info("Connecting to PostgreSQL...")
	if err := database.InitPostgres(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	defer database.Close()

	// Get SQL DB instance for DAO layer
	sqlDB = database.GetSQLDB()
	if sqlDB == nil {
		log.Fatal("Failed to get SQL database instance")
	}

	// Run migrations
	log.Info("Running database migrations...")
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT Manager
	jwtManager = utils.NewJWTManager(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessTokenExpire)*time.Hour,
		time.Duration(cfg.JWT.RefreshTokenExpire)*time.Hour,
	)

	// Initialize DAOs
	userdao.Init(sqlDB)
	quizdao.Init(sqlDB)
	aiagentdao.InitFileDAO(sqlDB)
	aiagentdao.InitJobDAO(sqlDB)

	// Initialize Auth Core with dependencies
	emailService := services.NewEmailService(&cfg.Email)
	authcore.Init(jwtManager, emailService)

	log.Info("Starting servers...")
	// Start background worker for AI quiz generation
	aiWorker := aiworker.NewQuizGeneratorWorker(sqlDB)
	go aiWorker.Start(context.Background())
	log.Info("Background worker started for quiz generation")

	// Start background worker for attempt auto-submit
	attemptWorker := quizworker.NewAttemptWorker()
	go attemptWorker.Start(context.Background())
	log.Info("Background worker started for attempt auto-submit")

	// Start internal and external servers
	go initInt()
	initExt() // Blocks until shutdown
}

// initInt initializes internal API server (service-to-service, no auth)
func initInt() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(TraceIDMiddleware())

	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", HealthCheck)

		// Internal-only endpoints
		// Add internal endpoints here as needed
	}

	intAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, "8081")
	intServer := &http.Server{
		Addr:           intAddr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Infof("Internal server starting on %s", intAddr)
	if err := intServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorf("Internal server error: %v", err)
	}
}

// initExt initializes external API server (public/authenticated endpoints)
func initExt() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(TraceIDMiddleware())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(&cfg.CORS)) // Enable CORS for frontend

	v1 := r.Group("/api/v1")

	// Health check (public)
	v1.GET("/health", HealthCheck)

	// Swagger documentation (public)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public auth endpoints
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", api.Signup)
		auth.POST("/verify-email", api.VerifyEmail)
		auth.POST("/resend-verification", api.ResendVerification)
		auth.POST("/login", api.Login)
		auth.POST("/refresh", api.RefreshToken)
		auth.POST("/forgot-password", api.ForgotPassword)
		auth.POST("/reset-password", api.ResetPassword)
	}

	// Protected auth endpoints
	protectedAuth := v1.Group("/auth")
	protectedAuth.Use(middleware.Auth(jwtManager))
	{
		protectedAuth.POST("/change-password", api.ChangePassword)
		protectedAuth.POST("/logout", api.Logout)
	}

	// Protected user endpoints
	users := v1.Group("/users")
	users.Use(middleware.Auth(jwtManager))
	{
		users.GET("/me", api.GetUserProfile)
		users.PATCH("/me", api.UpdateProfile)
		users.DELETE("/me", api.DeleteAccount)
		users.GET("", api.AdminListUsers)
		users.GET("/stats", api.GetUserStats)
		users.GET("/:id", api.GetUserByID)
	}

	// Quiz endpoints (protected)
	quizzes := v1.Group("/quizzes")
	quizzes.Use(middleware.Auth(jwtManager))
	{
		// Quiz CRUD
		quizzes.POST("", api.CreateQuiz)                   // Create quiz
		quizzes.GET("", api.ListMyQuizzes)                 // List my quizzes
		quizzes.GET("/public", api.ListPublicQuizzes)      // List public quizzes
		quizzes.GET("/:id", api.GetQuiz)                   // Get quiz by ID
		quizzes.GET("/:id/full", api.GetQuizWithQuestions) // Get quiz with questions
		quizzes.PATCH("/:id", api.UpdateQuiz)              // Update quiz
		quizzes.DELETE("/:id", api.DeleteQuiz)             // Delete quiz
		quizzes.POST("/:id/publish", api.PublishQuiz)      // Publish quiz
		quizzes.POST("/:id/unpublish", api.UnpublishQuiz)  // Unpublish quiz

		// Question CRUD
		quizzes.POST("/:id/questions", api.AddQuestion)       // Add question to quiz
		quizzes.GET("/questions/:qid", api.GetQuestion)       // Get question by ID
		quizzes.PATCH("/questions/:qid", api.UpdateQuestion)  // Update question
		quizzes.DELETE("/questions/:qid", api.DeleteQuestion) // Delete question

		// Answer CRUD
		quizzes.POST("/questions/:qid/answers", api.AddAnswer) // Add answer to question
		quizzes.PATCH("/answers/:aid", api.UpdateAnswer)       // Update answer
		quizzes.DELETE("/answers/:aid", api.DeleteAnswer)      // Delete answer
	}
	// File upload endpoints (protected)
	files := v1.Group("/files")
	files.Use(middleware.Auth(jwtManager))
	{
		files.POST("/upload", api.UploadFile) // Upload file
		files.GET("", api.ListFiles)          // List user files
		files.DELETE("/:id", api.DeleteFile)  // Delete file
	}

	// AI generation endpoints (protected)
	aiGroup := v1.Group("/ai")
	aiGroup.Use(middleware.Auth(jwtManager))
	{
		aiGroup.POST("/generate-quiz-from-text", api.GenerateQuizFromText) // Generate from text
		aiGroup.POST("/generate-quiz-from-file", api.GenerateQuizFromFile) // Generate from file
		aiGroup.GET("/generation-status/:id", api.GetGenerationStatus)     // Check status
		aiGroup.GET("/jobs", api.ListJobs)                                 // List jobs
	}

	// Attempt endpoints (protected)
	attempts := v1.Group("/attempts")
	attempts.Use(middleware.Auth(jwtManager))
	{
		attempts.GET("", api.ListMyAttempts)               // List my attempts
		attempts.GET("/:id", api.GetAttempt)               // Get attempt by ID
		attempts.GET("/:id/results", api.GetAttemptResult) // Get attempt result
		attempts.POST("/:id/answers", api.SubmitAnswer)    // Submit answer
		attempts.POST("/:id/finish", api.FinishAttempt)    // Finish attempt
	}

	// Add attempt routes to quizzes group
	quizzes.POST("/:id/attempts", api.StartAttempt) // Start quiz attempt
	quizzes.GET("/:id/stats", api.GetQuizStats)     // Get quiz stats (creator only)

	// Phase 6: Search and Discovery
	v1.GET("/quizzes/search", api.SearchQuizzes)   // Search quizzes (public)
	v1.GET("/categories", api.GetCategories)       // Get categories (public)
	v1.GET("/tags", api.GetPopularTags)            // Get popular tags (public)
	v1.GET("/share/:code", api.GetQuizByShareCode) // Get quiz by share code (public)

	// Phase 6: Quiz Statistics and Sharing (protected)
	quizzes.GET("/:id/statistics", api.GetQuizStatistics) // Get quiz statistics (creator only)
	quizzes.POST("/:id/share", api.GenerateShareLink)     // Generate share link (creator only)

	// Phase 6: User Statistics (protected)
	users.GET("/me/statistics", api.GetUserStatistics) // Get my statistics

	// Phase 6: Templates (protected)
	templates := v1.Group("/templates")
	templates.Use(middleware.Auth(jwtManager))
	{
		templates.GET("", api.ListTemplates)                   // List templates
		templates.POST("/:id/use", api.CreateQuizFromTemplate) // Create quiz from template
	}

	// Template creation from quiz (protected)
	quizzes.POST("/:id/save-as-template", api.SaveQuizAsTemplate) // Save quiz as template

	// Phase 7: Admin Dashboard (protected - admin only)
	admin := v1.Group("/admin")
	admin.Use(middleware.Auth(jwtManager))
	admin.Use(middleware.RequireAdmin())
	{
		// User Management
		admin.GET("/users", api.AdminListUsers)                // List all users
		admin.GET("/users/:id", api.GetUserDetails)            // Get user details
		admin.PATCH("/users/:id/role", api.ChangeUserRole)     // Change user role
		admin.POST("/users/:id/disable", api.AdminDisableUser) // Disable user
		admin.POST("/users/:id/enable", api.AdminEnableUser)   // Enable user
		admin.POST("/users/:id/restore", api.RestoreUser)      // Restore deleted user
		admin.DELETE("/users/:id", api.DeleteUser)             // Delete user (permanent)

		// Quiz Moderation
		admin.GET("/quizzes", api.ListQuizzesAdmin)        // List all quizzes
		admin.GET("/quizzes/:id", api.GetQuizDetailsAdmin) // Get quiz details
		admin.DELETE("/quizzes/:id", api.DeleteQuizAdmin)  // Delete quiz

		// Statistics
		admin.GET("/stats", api.GetPlatformStats)         // Get platform statistics
		admin.GET("/stats/activity", api.GetUserActivity) // Get activity timeline
		admin.GET("/logs", api.GetActivityLogs)           // Get activity logs
	}

	// Create HTTP server
	extAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	extServer := &http.Server{
		Addr:           extAddr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("External server starting on %s", extAddr)
		if err := extServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start external server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := extServer.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped gracefully")
}

// TraceIDMiddleware adds trace ID to each request
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := commons.Ctx(c)
		tid := commons.Tid(ctx)
		c.Set(commons.TidRespHdr, tid)
		c.Header(commons.TidRespHdr, tid)
		c.Next()
	}
}

// HealthCheck endpoint
func HealthCheck(c *gin.Context) {
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)

	// Check database
	dbHealthy := true
	if err := sqlDB.Ping(); err != nil {
		dbHealthy = false
	}

	status := "healthy"
	statusCode := http.StatusOK
	if !dbHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": gin.H{
			"database": gin.H{
				"status": map[bool]string{true: "healthy", false: "unhealthy"}[dbHealthy],
			},
		},
	})
}
