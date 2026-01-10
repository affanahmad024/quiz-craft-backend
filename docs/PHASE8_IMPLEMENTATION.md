# Phase 8 Implementation Document: Testing & Production Readiness

**Date:** January 10, 2026
**Status:** Planning
**Estimated Duration:** 5 working days (25-30 hours)
**Risk Level:** Low

---

## Executive Summary

Phase 8 focuses on **quality assurance, documentation, and production deployment preparation**. This phase ensures the quiz-craft platform is **production-ready, well-tested, documented, and deployable** with confidence. Based on user requirements, we'll focus on manual testing with key unit tests for critical business logic.

### Key Features:
- ✅ Unit tests for critical business logic (scoring, auth)
- ✅ Comprehensive manual testing guide
- ✅ API documentation (Swagger/OpenAPI)
- ✅ Deployment configuration (Docker, Docker Compose)
- ✅ Production environment setup guide
- ✅ Monitoring and logging setup
- ✅ Performance optimization
- ✅ Security hardening

### Out of Scope (Deferred):
- ❌ 90%+ code coverage (focusing on critical paths only)
- ❌ E2E automated tests
- ❌ Load testing with high concurrency
- ❌ Advanced monitoring dashboards

---

## User Requirements Summary

Based on user clarification:

| Requirement | Decision |
|-------------|----------|
| **Testing Approach** | Manual testing + key unit tests (critical business logic) |
| **Code Coverage** | Focus on critical paths (scoring, auth, payments if any) |
| **Deployment** | Docker-based, cloud-ready |
| **Documentation** | API docs (Swagger) + deployment guide |

---

## Testing Strategy

### 1. Unit Tests (Critical Business Logic Only)

Focus areas for unit testing:
- **Authentication Logic** (signup, login, token validation)
- **Scoring Logic** (quiz scoring, pass/fail calculation)
- **Quiz Validation** (question validation, answer validation)
- **Statistics Calculations** (averages, percentages, streaks)

**Not testing**:
- Basic CRUD operations (covered by manual testing)
- Simple getter/setter methods
- Trivial utility functions

---

### 2. Manual Testing Guide

Comprehensive manual test scenarios for all features:

#### Authentication Flow Tests
```bash
# Test 1: Complete signup flow
POST /api/v1/auth/signup
→ Check email sent
→ Verify with token
→ Login successfully

# Test 2: Email verification expiry
→ Wait for token to expire
→ Try to verify
→ Should fail with "token-expired"
→ Resend verification
→ Verify successfully

# Test 3: Password reset flow
→ Request password reset
→ Check email
→ Reset with token
→ Login with new password

# Test 4: Failed login attempts
→ Login with wrong password 5 times
→ Should lock account
→ Try to login again
→ Should fail with "account-locked"
```

#### Quiz Management Tests
```bash
# Test 1: Create quiz with questions
→ Create quiz (unpublished)
→ Add multiple choice questions
→ Add true/false questions
→ Add answers to questions
→ Mark correct answers
→ Publish quiz

# Test 2: Quiz validation
→ Try to publish quiz without questions
→ Should fail
→ Try to add question without answers
→ Should fail
→ Try to add question with no correct answer
→ Should fail

# Test 3: Quiz permissions
→ User A creates quiz
→ User B tries to edit
→ Should fail with 403
→ User B can view if public
```

#### Quiz Attempt Tests
```bash
# Test 1: Complete quiz attempt
→ Start quiz
→ Submit answers progressively
→ Finish quiz
→ Check score calculation
→ Verify pass/fail status
→ View results with correct answers

# Test 2: Time limit enforcement
→ Start timed quiz (2 minutes)
→ Wait for time to expire
→ Try to submit answer
→ Should fail with "time-limit-exceeded"
→ Check attempt auto-submitted by worker

# Test 3: Resume quiz
→ Start quiz
→ Submit some answers
→ Close browser
→ Reopen and get attempt details
→ Continue from where left off
→ Finish quiz
```

#### Search and Filter Tests
```bash
# Test 1: Full-text search
→ Search for "javascript"
→ Check results contain search term
→ Try various search terms

# Test 2: Multiple filters
→ Filter by category + difficulty + tags
→ Check results match all criteria
→ Try different combinations

# Test 3: Sorting
→ Sort by newest
→ Sort by most attempted
→ Sort by highest rated
→ Verify order is correct
```

#### Admin Features Tests
```bash
# Test 1: Admin authorization
→ Regular user tries to access /admin/users
→ Should fail with 403
→ Admin user accesses same endpoint
→ Should succeed

# Test 2: User management
→ Admin disables user
→ User tries to login
→ Should fail
→ Admin enables user
→ User can login again

# Test 3: Activity logging
→ Admin disables user with reason
→ Check activity logs
→ Should show action, reason, admin info
```

---

### 3. Test Files to Create

```
tests/
├── unit/
│   ├── auth_test.go           ← Authentication logic tests
│   ├── scoring_test.go        ← Quiz scoring tests
│   ├── validation_test.go     ← Input validation tests
│   └── statistics_test.go     ← Statistics calculation tests
├── manual/
│   ├── auth_test_cases.md     ← Manual auth test scenarios
│   ├── quiz_test_cases.md     ← Manual quiz test scenarios
│   ├── attempt_test_cases.md  ← Manual attempt test scenarios
│   └── admin_test_cases.md    ← Manual admin test scenarios
└── test_utils.go              ← Test helpers and fixtures
```

---

## Unit Test Examples

### Authentication Tests

```go
// tests/unit/auth_test.go
package unit

import (
    "testing"
    "time"
    "quiz-craft/auth/core"
    "quiz-craft/user/model"
    "github.com/stretchr/testify/assert"
)

func TestSignup_Success(t *testing.T) {
    // Arrange
    input := authcore.SignupInput{
        Email:    "test@example.com",
        Name:     "Test User",
        Password: "SecurePass123!",
    }

    // Act
    user, err := authcore.Signup(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, input.Email, user.Email)
    assert.False(t, user.EmailVerified)
}

func TestSignup_DuplicateEmail(t *testing.T) {
    // Arrange
    input := authcore.SignupInput{
        Email:    "existing@example.com",
        Name:     "Test User",
        Password: "SecurePass123!",
    }

    // Create first user
    authcore.Signup(context.Background(), input)

    // Act - try to create duplicate
    user, err := authcore.Signup(context.Background(), input)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Contains(t, err.Error(), "email-already-exists")
}

func TestVerifyEmail_ValidToken(t *testing.T) {
    // Arrange
    user := createTestUser()
    token := generateVerificationToken(user.ID)

    // Act
    err := authcore.VerifyEmail(context.Background(), token)

    // Assert
    assert.NoError(t, err)
    updatedUser := getUser(user.ID)
    assert.True(t, updatedUser.EmailVerified)
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
    // Arrange
    user := createTestUser()
    token := generateExpiredToken(user.ID)

    // Act
    err := authcore.VerifyEmail(context.Background(), token)

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "token-expired")
}
```

---

### Scoring Tests

```go
// tests/unit/scoring_test.go
package unit

import (
    "testing"
    "quiz-craft/quiz/core"
    "github.com/stretchr/testify/assert"
)

func TestScoreAnswer_CorrectMultipleChoice(t *testing.T) {
    // Arrange
    question := createTestQuestion("multiple_choice", 10)
    correctAnswer := createCorrectAnswer(question.ID)

    // Act
    isCorrect, pointsEarned := quizcore.ScoreAnswer(question, correctAnswer.ID)

    // Assert
    assert.True(t, isCorrect)
    assert.Equal(t, 10.0, pointsEarned)
}

func TestScoreAnswer_IncorrectMultipleChoice(t *testing.T) {
    // Arrange
    question := createTestQuestion("multiple_choice", 10)
    incorrectAnswer := createIncorrectAnswer(question.ID)

    // Act
    isCorrect, pointsEarned := quizcore.ScoreAnswer(question, incorrectAnswer.ID)

    // Assert
    assert.False(t, isCorrect)
    assert.Equal(t, 0.0, pointsEarned)
}

func TestCalculateFinalScore(t *testing.T) {
    // Arrange
    questions := []Question{
        {ID: "q1", Points: 10},
        {ID: "q2", Points: 10},
        {ID: "q3", Points: 10},
    }

    userAnswers := []UserAnswer{
        {QuestionID: "q1", IsCorrect: true, PointsEarned: 10},
        {QuestionID: "q2", IsCorrect: false, PointsEarned: 0},
        {QuestionID: "q3", IsCorrect: true, PointsEarned: 10},
    }

    // Act
    score, maxScore, percentage := quizcore.CalculateFinalScore(questions, userAnswers)

    // Assert
    assert.Equal(t, 20.0, score)
    assert.Equal(t, 30.0, maxScore)
    assert.Equal(t, 66.67, percentage) // 20/30 * 100
}

func TestCalculatePassFail(t *testing.T) {
    quiz := Quiz{PassingScore: 70}

    // Test pass
    passed := quizcore.CalculatePassFail(&quiz, 75.0)
    assert.True(t, passed)

    // Test fail
    failed := quizcore.CalculatePassFail(&quiz, 65.0)
    assert.False(t, failed)

    // Test exact passing score
    exactPass := quizcore.CalculatePassFail(&quiz, 70.0)
    assert.True(t, exactPass)
}
```

---

### Statistics Tests

```go
// tests/unit/statistics_test.go
package unit

import (
    "testing"
    "quiz-craft/quiz/core"
    "github.com/stretchr/testify/assert"
)

func TestCalculateStreak_Consecutive(t *testing.T) {
    // Arrange
    attempts := []time.Time{
        time.Now(),
        time.Now().AddDate(0, 0, -1),
        time.Now().AddDate(0, 0, -2),
    }

    // Act
    streak := quizcore.CalculateStreak(attempts)

    // Assert
    assert.Equal(t, 3, streak)
}

func TestCalculateStreak_Broken(t *testing.T) {
    // Arrange
    attempts := []time.Time{
        time.Now(),
        time.Now().AddDate(0, 0, -1),
        time.Now().AddDate(0, 0, -5), // Gap
    }

    // Act
    streak := quizcore.CalculateStreak(attempts)

    // Assert
    assert.Equal(t, 2, streak) // Only counts first 2
}

func TestCalculateAverageScore(t *testing.T) {
    // Arrange
    scores := []float64{80.0, 90.0, 70.0, 85.0}

    // Act
    average := quizcore.CalculateAverage(scores)

    // Assert
    assert.Equal(t, 81.25, average)
}
```

---

## API Documentation (Swagger)

### Setup Swagger

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Install gin-swagger
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

### Add Swagger Annotations

```go
// main.go
package main

import (
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "quiz-craft/docs" // Import generated docs
)

// @title QuizCraft API
// @version 1.0
// @description Quiz creation and management platform
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
// @description Type "Bearer" followed by a space and JWT token

func main() {
    router := gin.Default()

    // Swagger endpoint
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // ... rest of setup
}
```

### Annotate Endpoints

```go
// api/auth.go

// Signup godoc
// @Summary Register new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param input body authcore.SignupInput true "Signup input"
// @Success 201 {object} commons.SuccessResponse{data=user.User}
// @Failure 400 {object} commons.ErrorResponse
// @Failure 409 {object} commons.ErrorResponse
// @Router /auth/signup [post]
func Signup(c *gin.Context) {
    // Implementation
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param input body authcore.LoginInput true "Login credentials"
// @Success 200 {object} commons.SuccessResponse{data=authcore.LoginResponse}
// @Failure 400 {object} commons.ErrorResponse
// @Failure 401 {object} commons.ErrorResponse
// @Router /auth/login [post]
func Login(c *gin.Context) {
    // Implementation
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} commons.SuccessResponse{data=user.User}
// @Failure 401 {object} commons.ErrorResponse
// @Failure 404 {object} commons.ErrorResponse
// @Router /users/me [get]
func GetCurrentUser(c *gin.Context) {
    // Implementation
}
```

### Generate Swagger Docs

```bash
# Generate swagger documentation
swag init

# This creates:
# - docs/docs.go
# - docs/swagger.json
# - docs/swagger.yaml

# Access Swagger UI at:
# http://localhost:8080/swagger/index.html
```

---

## Deployment Configuration

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/config ./config
COPY --from=builder /app/pkg/database/migrations ./pkg/database/migrations

# Create uploads directory
RUN mkdir -p /root/uploads

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the application
CMD ["./main"]
```

---

### Docker Compose (Production)

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: quiz-craft-postgres
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: quiz-craft-backend
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - JWT_SECRET=${JWT_SECRET}
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT}
      - SMTP_USERNAME=${SMTP_USERNAME}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
      - GIN_MODE=release
      - LOG_LEVEL=info
    ports:
      - "8080:8080"
    volumes:
      - ./uploads:/root/uploads
      - ./logs:/root/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  nginx:
    image: nginx:alpine
    container_name: quiz-craft-nginx
    depends_on:
      - backend
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./static:/usr/share/nginx/html:ro
    restart: unless-stopped

volumes:
  postgres_data:
    driver: local
```

---

### Nginx Configuration

```nginx
# nginx/nginx.conf
upstream backend {
    server backend:8080;
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # API proxy
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Swagger UI
    location /swagger/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Health check
    location /health {
        proxy_pass http://backend/api/v1/health;
        access_log off;
    }

    # Static files (if any)
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ =404;
    }

    # File upload size limit
    client_max_body_size 10M;
}
```

---

## Production Environment Setup

### Environment Variables

```bash
# .env.production
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=release

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=quiz_craft_user
DB_PASSWORD=<strong-random-password>
DB_NAME=quiz_craft_prod
DB_SSLMODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT (generate strong secret)
JWT_SECRET=<64-char-random-string>
JWT_ACCESS_TOKEN_EXPIRE=24h
JWT_REFRESH_TOKEN_EXPIRE=168h

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=<app-specific-password>
SMTP_FROM_EMAIL=noreply@quizcraft.com
SMTP_FROM_NAME=QuizCraft

# Application
APP_NAME=QuizCraft
APP_ENV=production
LOG_LEVEL=info
```

---

### Deployment Steps

```bash
# 1. Clone repository on server
git clone https://github.com/your-repo/quiz-craft.git
cd quiz-craft/backend

# 2. Copy and configure environment file
cp .env.example .env.production
nano .env.production  # Edit with production values

# 3. Generate SSL certificates (Let's Encrypt)
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem

# 4. Build and start services
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d --build

# 5. Check logs
docker-compose -f docker-compose.prod.yml logs -f backend

# 6. Verify health
curl http://localhost:8080/api/v1/health

# 7. Access Swagger docs
# https://yourdomain.com/swagger/index.html
```

---

## Monitoring and Logging

### Application Logging

```go
// Use structured logging with logrus
import (
    log "github.com/sirupsen/logrus"
)

func init() {
    // Set log format
    log.SetFormatter(&log.JSONFormatter{})

    // Set log level from environment
    level, _ := log.ParseLevel(os.Getenv("LOG_LEVEL"))
    log.SetLevel(level)

    // Log to file in production
    if os.Getenv("APP_ENV") == "production" {
        file, err := os.OpenFile("/root/logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err == nil {
            log.SetOutput(file)
        }
    }
}

// Log with context
log.WithFields(log.Fields{
    "user_id": userID,
    "action": "quiz_created",
    "quiz_id": quizID,
}).Info("User created quiz")

// Log errors with stack trace
log.WithError(err).Error("Failed to create quiz")
```

---

### Health Check Endpoint Enhancement

```go
// pkg/database/postgres.go
func HealthCheck() error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    return DB.PingContext(ctx)
}

// api/health.go
func HealthCheck(c *gin.Context) {
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
        "checks": map[string]interface{}{},
    }

    // Check database
    if err := postgres.HealthCheck(); err != nil {
        health["status"] = "unhealthy"
        health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
            "status": "unhealthy",
            "error": err.Error(),
        }
        c.JSON(503, health)
        return
    }

    health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
        "status": "healthy",
    }

    c.JSON(200, health)
}
```

---

### Monitoring with Prometheus (Optional)

```go
// Add Prometheus metrics endpoint
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request latency",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

// Add metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## Performance Optimization

### Database Optimization

```sql
-- Add missing indexes (if not already added)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_quizzes_creator_published
ON quizzes(creator_id, is_published) WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_quiz_attempts_user_status
ON quiz_attempts(user_id, status);

-- Analyze tables for query planner
ANALYZE users;
ANALYZE quizzes;
ANALYZE questions;
ANALYZE quiz_attempts;
ANALYZE user_answers;

-- Check for slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

---

### Connection Pooling Configuration

```go
// pkg/database/postgres.go
func InitPostgres(cfg *config.Config) error {
    db.SetMaxOpenConns(cfg.Database.MaxOpenConns)       // 25
    db.SetMaxIdleConns(cfg.Database.MaxIdleConns)       // 5
    db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime) // 5m

    return nil
}
```

---

### Response Caching (Optional)

```go
// Cache frequently accessed data
import "github.com/patrickmn/go-cache"

var Cache = cache.New(5*time.Minute, 10*time.Minute)

func GetQuizWithCache(quizID string) (*Quiz, error) {
    // Try cache first
    if cached, found := Cache.Get("quiz:" + quizID); found {
        return cached.(*Quiz), nil
    }

    // Fetch from database
    quiz := dao.GetQuizByID(quizID)

    // Cache result
    Cache.Set("quiz:"+quizID, quiz, cache.DefaultExpiration)

    return quiz, nil
}
```

---

## Security Hardening

### Security Checklist

```markdown
## Pre-Deployment Security Checklist

### Authentication & Authorization
- [x] JWT secret is strong (64+ chars)
- [x] Passwords hashed with bcrypt
- [x] Email verification enforced
- [x] Account lockout after failed attempts
- [x] Admin role properly enforced

### Input Validation
- [x] All user inputs validated
- [x] SQL injection prevented (parameterized queries)
- [x] XSS prevention (sanitized outputs)
- [x] File upload validation (size, type)
- [x] Rate limiting on sensitive endpoints

### Data Protection
- [x] HTTPS enforced in production
- [x] Sensitive data not logged
- [x] Database credentials in env variables
- [x] JWT tokens in Authorization header only
- [x] CORS properly configured

### Infrastructure
- [x] Firewall configured (ports 80, 443 only)
- [x] Database not exposed publicly
- [x] Regular backups configured
- [x] SSL certificates valid
- [x] Docker containers run as non-root

### Monitoring
- [x] Error logging enabled
- [x] Activity logging for admin actions
- [x] Health checks configured
- [x] Alerts for failures
```

---

### CORS Configuration

```go
// internal/middleware/cors.go
func CORSMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"https://yourdomain.com", "https://www.yourdomain.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}

// In production, DO NOT use AllowOrigins: []string{"*"}
```

---

### Rate Limiting (Optional)

```go
// Add rate limiting middleware
import "github.com/ulule/limiter/v3"
import "github.com/ulule/limiter/v3/drivers/store/memory"

func RateLimitMiddleware() gin.HandlerFunc {
    rate := limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  60, // 60 requests per minute
    }

    store := memory.NewStore()
    instance := limiter.New(store, rate)

    return func(c *gin.Context) {
        limiterCtx, err := instance.Get(c, c.ClientIP())
        if err != nil {
            c.AbortWithStatusJSON(500, gin.H{"error": "rate-limiter-error"})
            return
        }

        if limiterCtx.Reached {
            c.AbortWithStatusJSON(429, gin.H{"error": "rate-limit-exceeded"})
            return
        }

        c.Next()
    }
}
```

---

## Implementation Timeline

### Day 1: Unit Tests Setup & Critical Tests (5-6 hours)
**Tasks**:
- [ ] Set up testing framework (testify)
- [ ] Create test utilities and helpers
- [ ] Create `tests/unit/auth_test.go`
- [ ] Write authentication tests
  - Signup validation
  - Email verification
  - Login logic
  - Token validation
- [ ] Create `tests/unit/scoring_test.go`
- [ ] Write scoring tests
  - Answer scoring
  - Final score calculation
  - Pass/fail logic
- [ ] Run tests and fix issues

**Acceptance Criteria**:
- ✅ Test framework configured
- ✅ Auth tests pass
- ✅ Scoring tests pass

---

### Day 2: Validation & Statistics Tests + Manual Test Guide (6-7 hours)
**Tasks**:
- [ ] Create `tests/unit/validation_test.go`
- [ ] Write validation tests
  - Quiz validation
  - Question validation
  - Answer validation
- [ ] Create `tests/unit/statistics_test.go`
- [ ] Write statistics tests
  - Streak calculation
  - Average calculations
  - Percentage calculations
- [ ] Create manual testing guides
  - `tests/manual/auth_test_cases.md`
  - `tests/manual/quiz_test_cases.md`
  - `tests/manual/attempt_test_cases.md`
  - `tests/manual/admin_test_cases.md`
- [ ] Perform manual testing

**Acceptance Criteria**:
- ✅ All unit tests pass
- ✅ Manual test guides complete
- ✅ Manual testing completed

---

### Day 3: API Documentation (Swagger) (5-6 hours)
**Tasks**:
- [ ] Install swag and gin-swagger
- [ ] Add Swagger annotations to main.go
- [ ] Annotate all authentication endpoints
- [ ] Annotate all quiz endpoints
- [ ] Annotate all attempt endpoints
- [ ] Annotate all admin endpoints
- [ ] Generate Swagger documentation
- [ ] Test Swagger UI
- [ ] Write API usage examples

**Acceptance Criteria**:
- ✅ Swagger docs generated
- ✅ All endpoints documented
- ✅ Swagger UI accessible
- ✅ Examples provided

---

### Day 4: Docker & Deployment Configuration (6-7 hours)
**Tasks**:
- [ ] Create production Dockerfile
- [ ] Create docker-compose.prod.yml
- [ ] Configure Nginx
  - Create nginx.conf
  - Configure SSL
  - Configure proxy settings
- [ ] Create .env.production template
- [ ] Write deployment guide
  - Server setup
  - SSL certificate setup
  - Docker deployment steps
  - Verification steps
- [ ] Test deployment locally
- [ ] Document rollback procedures

**Acceptance Criteria**:
- ✅ Docker images build successfully
- ✅ Services start with docker-compose
- ✅ Nginx proxies correctly
- ✅ Health checks pass
- ✅ Deployment guide complete

---

### Day 5: Monitoring, Security & Final Polish (6-7 hours)
**Tasks**:
- [ ] Enhance logging
  - Structured logging
  - Log rotation
  - Error tracking
- [ ] Enhance health check endpoint
- [ ] Add Prometheus metrics (optional)
- [ ] Security hardening
  - Review CORS configuration
  - Add rate limiting
  - Review input validation
  - Check for exposed secrets
- [ ] Performance optimization
  - Add database indexes
  - Configure connection pooling
  - Add caching (optional)
- [ ] Create monitoring guide
- [ ] Create troubleshooting guide
- [ ] Final security review
- [ ] Final testing

**Acceptance Criteria**:
- ✅ Logging configured
- ✅ Monitoring in place
- ✅ Security checklist complete
- ✅ Performance optimized
- ✅ All documentation complete

---

## Documentation Deliverables

### 1. API Documentation
- ✅ Swagger/OpenAPI specification
- ✅ Interactive Swagger UI
- ✅ Example requests/responses
- ✅ Authentication guide

### 2. Deployment Documentation
- ✅ Server requirements
- ✅ Installation guide
- ✅ Environment configuration
- ✅ Docker deployment guide
- ✅ SSL setup guide
- ✅ Troubleshooting guide

### 3. Testing Documentation
- ✅ Unit test coverage report
- ✅ Manual testing guide
- ✅ Test scenarios
- ✅ How to run tests

### 4. Operations Documentation
- ✅ Monitoring guide
- ✅ Log analysis guide
- ✅ Backup and restore procedures
- ✅ Incident response plan

---

## Success Metrics

Phase 8 is successful when:

✅ **Testing**:
- Critical business logic has unit tests
- All unit tests pass
- Manual testing guide complete
- All features manually tested

✅ **Documentation**:
- API fully documented in Swagger
- Deployment guide complete
- All README files updated
- Troubleshooting guide available

✅ **Deployment**:
- Docker images build successfully
- Services deploy with docker-compose
- Health checks pass
- SSL configured
- Application accessible

✅ **Security**:
- Security checklist complete
- No secrets in code
- CORS configured
- Rate limiting in place

✅ **Performance**:
- Database optimized
- Connection pooling configured
- Response times acceptable

✅ **Monitoring**:
- Logging configured
- Health checks working
- Error tracking in place

---

## Critical Files Summary

### New Files (15+ files):
1. `tests/unit/auth_test.go`
2. `tests/unit/scoring_test.go`
3. `tests/unit/validation_test.go`
4. `tests/unit/statistics_test.go`
5. `tests/test_utils.go`
6. `tests/manual/auth_test_cases.md`
7. `tests/manual/quiz_test_cases.md`
8. `tests/manual/attempt_test_cases.md`
9. `tests/manual/admin_test_cases.md`
10. `Dockerfile`
11. `docker-compose.prod.yml`
12. `nginx/nginx.conf`
13. `.env.production.example`
14. `docs/DEPLOYMENT_GUIDE.md`
15. `docs/MONITORING_GUIDE.md`

### Modified Files (3+ files):
1. `main.go` - Add Swagger annotations
2. `api/*.go` - Add Swagger annotations to all endpoints
3. `README.md` - Update with deployment info

---

## Dependencies

### New Dependencies:
```bash
# Testing
go get github.com/stretchr/testify

# Swagger
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files

# Rate limiting (optional)
go get github.com/ulule/limiter/v3
go get github.com/ulule/limiter/v3/drivers/store/memory

# Caching (optional)
go get github.com/patrickmn/go-cache

# Prometheus (optional)
go get github.com/prometheus/client_golang/prometheus
```

---

## Total Estimated Effort

**Implementation**: 25-30 hours (5 working days at 5-6 hours/day)

**Breakdown**:
- Day 1: Unit Tests (5-6 hours)
- Day 2: More Tests + Manual Guide (6-7 hours)
- Day 3: API Documentation (5-6 hours)
- Day 4: Docker & Deployment (6-7 hours)
- Day 5: Monitoring & Security (6-7 hours)

**Complexity**: Medium (multiple concerns, documentation-heavy)

**Risk Level**: Low (mostly configuration and documentation)

---

## Post-Deployment Checklist

```markdown
## After First Deployment

### Immediate (Day 1)
- [ ] Verify all services running
- [ ] Check health endpoint
- [ ] Test API endpoints
- [ ] Verify SSL certificate
- [ ] Check logs for errors
- [ ] Test admin access
- [ ] Verify email sending

### First Week
- [ ] Monitor error logs daily
- [ ] Check performance metrics
- [ ] Review user feedback
- [ ] Backup database
- [ ] Test restore from backup
- [ ] Review security logs

### First Month
- [ ] Analyze usage patterns
- [ ] Optimize slow queries
- [ ] Review and update documentation
- [ ] Plan feature enhancements
- [ ] Security audit
```

---

**Ready to deploy!** 🚀
