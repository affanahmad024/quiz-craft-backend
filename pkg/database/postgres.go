package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitPostgres initializes PostgreSQL connection
func InitPostgres(cfg *config.DatabaseConfig) error {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	var logLevel gormlogger.LogLevel
	switch logger.GetLogger().Level {
	case 5: // Debug
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Warn
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.GetLogger().Info("PostgreSQL connected successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// GetSQLDB returns the underlying *sql.DB from GORM
// This is needed for the DAO layer which uses raw SQL
func GetSQLDB() *sql.DB {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.GetLogger().Errorf("Failed to get *sql.DB from GORM: %v", err)
		return nil
	}

	return sqlDB
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// HealthCheck checks if database connection is alive
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
