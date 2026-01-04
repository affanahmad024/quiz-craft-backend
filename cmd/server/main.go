package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/api/routes"
	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/pkg/cache"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.App.LogLevel)
	log := logger.GetLogger()

	log.Info("Starting QuizCraft Backend API")
	log.Infof("Environment: %s", cfg.App.Env)

	// Initialize PostgreSQL
	log.Info("Connecting to PostgreSQL...")
	if err := database.InitPostgres(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Info("Running database migrations...")
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	log.Info("Connecting to Redis...")
	if err := cache.InitRedis(&cfg.Redis); err != nil {
		log.Warnf("Failed to initialize Redis: %v", err)
		log.Warn("Continuing without Redis (some features may not work)")
	} else {
		defer cache.Close()
	}

	// Setup router
	log.Info("Setting up routes...")
	router := routes.SetupRouter(cfg)

	// Create HTTP server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:           serverAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Server starting on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped gracefully")
}
