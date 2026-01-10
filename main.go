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
	
	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/internal/middleware"
	"github.com/affanahmad024/quiz-craft/backend/internal/services"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"github.com/affanahmad024/quiz-craft/backend/api"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	authcore "github.com/affanahmad024/quiz-craft/backend/auth/core"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
)

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

	// Initialize Auth Core with dependencies
	emailService := services.NewEmailService(&cfg.Email)
	authcore.Init(jwtManager, emailService)

	log.Info("Starting servers...")

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

	v1 := r.Group("/api/v1")

	// Health check (public)
	v1.GET("/health", HealthCheck)

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
		users.GET("", api.ListUsers)
		users.GET("/stats", api.GetUserStats)
		users.GET("/:id", api.GetUserByID)
		users.POST("/:id/restore", api.RestoreUser)
		users.POST("/:id/disable", api.DisableUser)
		users.POST("/:id/enable", api.EnableUser)
	}

	// Quiz endpoints (protected)
	quizzes := v1.Group("/quizzes")
	quizzes.Use(middleware.Auth(jwtManager))
	{
		// Quiz CRUD
		quizzes.POST("", api.CreateQuiz)                    // Create quiz
		quizzes.GET("", api.ListMyQuizzes)                  // List my quizzes
		quizzes.GET("/public", api.ListPublicQuizzes)       // List public quizzes
		quizzes.GET("/:id", api.GetQuiz)                    // Get quiz by ID
		quizzes.GET("/:id/full", api.GetQuizWithQuestions)  // Get quiz with questions
		quizzes.PATCH("/:id", api.UpdateQuiz)               // Update quiz
		quizzes.DELETE("/:id", api.DeleteQuiz)              // Delete quiz
		quizzes.POST("/:id/publish", api.PublishQuiz)       // Publish quiz
		quizzes.POST("/:id/unpublish", api.UnpublishQuiz)   // Unpublish quiz

		// Question CRUD
		quizzes.POST("/:id/questions", api.AddQuestion)           // Add question to quiz
		quizzes.GET("/questions/:qid", api.GetQuestion)           // Get question by ID
		quizzes.PATCH("/questions/:qid", api.UpdateQuestion)      // Update question
		quizzes.DELETE("/questions/:qid", api.DeleteQuestion)     // Delete question

		// Answer CRUD
		quizzes.POST("/questions/:qid/answers", api.AddAnswer)    // Add answer to question
		quizzes.PATCH("/answers/:aid", api.UpdateAnswer)          // Update answer
		quizzes.DELETE("/answers/:aid", api.DeleteAnswer)         // Delete answer
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
