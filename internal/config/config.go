package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Email       EmailConfig
	App         AppConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	FileUpload  FileUploadConfig
	GoogleAI    GoogleAIConfig
}

type ServerConfig struct {
	Port string
	Host string
	Mode string // debug, release, test
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret              string
	AccessTokenExpire   time.Duration
	RefreshTokenExpire  time.Duration
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type AppConfig struct {
	Name     string
	Env      string // development, staging, production
	LogLevel string // debug, info, warn, error
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

type FileUploadConfig struct {
	MaxSize      int64
	UploadDir    string
	AllowedTypes []string
}

type GoogleAIConfig struct {
	APIKey string
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	// Set config file details
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Read config file (optional - will use env vars if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Enable environment variable reading
	viper.AutomaticEnv()

	// Map environment variables to config keys
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.mode", "GIN_MODE")

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.access_token_expire", "JWT_ACCESS_TOKEN_EXPIRE")
	viper.BindEnv("jwt.refresh_token_expire", "JWT_REFRESH_TOKEN_EXPIRE")

	viper.BindEnv("email.smtp_host", "SMTP_HOST")
	viper.BindEnv("email.smtp_port", "SMTP_PORT")
	viper.BindEnv("email.smtp_username", "SMTP_USERNAME")
	viper.BindEnv("email.smtp_password", "SMTP_PASSWORD")
	viper.BindEnv("email.from_email", "SMTP_FROM_EMAIL")
	viper.BindEnv("email.from_name", "SMTP_FROM_NAME")

	viper.BindEnv("app.name", "APP_NAME")
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("app.log_level", "LOG_LEVEL")

	viper.BindEnv("google_ai.api_key", "GOOGLE_AI_API_KEY")

	// Set defaults
	setDefaults()

	// Unmarshal config into struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Parse durations
	accessExpire, err := time.ParseDuration(viper.GetString("jwt.access_token_expire"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT access token expire duration: %w", err)
	}
	cfg.JWT.AccessTokenExpire = accessExpire

	refreshExpire, err := time.ParseDuration(viper.GetString("jwt.refresh_token_expire"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT refresh token expire duration: %w", err)
	}
	cfg.JWT.RefreshTokenExpire = refreshExpire

	connMaxLifetime, err := time.ParseDuration(viper.GetString("database.conn_max_lifetime"))
	if err != nil {
		return nil, fmt.Errorf("invalid database connection max lifetime: %w", err)
	}
	cfg.Database.ConnMaxLifetime = connMaxLifetime

	rateLimitWindow, err := time.ParseDuration(viper.GetString("rate_limit.window"))
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit window duration: %w", err)
	}
	cfg.RateLimit.Window = rateLimitWindow

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "debug")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// JWT defaults
	viper.SetDefault("jwt.access_token_expire", "24h")
	viper.SetDefault("jwt.refresh_token_expire", "168h")

	// App defaults
	viper.SetDefault("app.name", "QuizCraft")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.log_level", "debug")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})

	// Rate limit defaults
	viper.SetDefault("rate_limit.requests", 100)
	viper.SetDefault("rate_limit.window", "1m")

	// File upload defaults
	viper.SetDefault("file_upload.max_size", 10485760) // 10MB
	viper.SetDefault("file_upload.upload_dir", "./uploads")
	viper.SetDefault("file_upload.allowed_types", []string{"image/jpeg", "image/png", "application/pdf"})
}

func validateConfig(cfg *Config) error {
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
