# Quiz-Craft Backend - Current State (Post-Refactoring)

Last Updated: January 10, 2026

## 📊 Project Status

### Completed Phases

#### ✅ Phase 1: Foundation & Infrastructure (COMPLETE)
- Clean architecture project structure
- Configuration management (Viper + environment variables)
- PostgreSQL database connection with GORM
- ~~Redis cache integration~~ **REMOVED** (Simplified architecture)
- Database migrations (golang-migrate)
- Middleware (CORS, Logging, Error handling)
- Health check endpoints
- Docker Compose for local development

#### ✅ Phase 2: Authentication & User Management (COMPLETE)
- User registration with email verification
- JWT-based authentication (access + refresh tokens)
- Login/Logout functionality
- Password management (forgot, reset, change)
- User profile CRUD operations
- Soft delete with restore functionality
- Account locking after failed attempts
- Password history and strength validation
- Email service with HTML templates
- 17 authentication & user endpoints

## 🏗️ Current Architecture

### Technology Stack
- **Language**: Go 1.25.3
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+ (GORM)
- **Authentication**: JWT (golang-jwt/jwt)
- **Email**: SMTP via gomail
- **Migrations**: golang-migrate
- **Configuration**: Viper + godotenv
- **Logging**: Logrus

### ~~Removed Components~~
- ❌ **Redis** - Removed for simplification (no caching layer currently)
  - Sessions now stored in PostgreSQL
  - No rate limiting implementation
  - Simplified health check

## 📁 Project Structure

```
backend/
├── main.go                          # Application entry point
├── api/                             # API handlers
│   ├── auth.go                      # Auth endpoints
│   └── user.go                      # User endpoints
├── auth/
│   └── core/                        # Auth business logic
│       ├── auth.go                  # Core auth functions
│       ├── login.go
│       ├── signup.go
│       ├── email_verification.go
│       ├── password_reset.go
│       └── password_change.go
├── user/
│   ├── dao/                         # Data Access Objects
│   │   ├── user.go                  # User CRUD operations
│   │   └── password.go              # Password/auth data operations
│   └── model/                       # Data models
│       ├── user.go
│       └── password.go
├── internal/
│   ├── config/
│   │   └── config.go                # Configuration loader
│   ├── middleware/
│   │   ├── auth.go                  # JWT authentication middleware
│   │   ├── cors.go                  # CORS configuration
│   │   ├── logger.go                # Request logging
│   │   └── error_handler.go        # Error handling
│   ├── services/
│   │   └── email.go                 # Email service
│   └── utils/
│       ├── jwt.go                   # JWT manager
│       ├── password.go              # Password hashing
│       └── validator.go             # Input validation
├── pkg/
│   ├── database/
│   │   ├── postgres.go              # PostgreSQL connection
│   │   ├── migrate.go               # Migration runner
│   │   └── migrations/              # SQL migration files
│   │       ├── 000001_create_users_table.up.sql
│   │       ├── 000001_create_users_table.down.sql
│   │       ├── 000002_create_user_passwords_table.up.sql
│   │       └── 000002_create_user_passwords_table.down.sql
│   └── logger/
│       └── logger.go                # Logger initialization
├── commons/
│   ├── context.go                   # Context utilities
│   ├── errors.go                    # Common errors
│   └── response.go                  # Response helpers
├── deployments/
│   └── docker/
│       └── docker-compose.yml       # PostgreSQL only (Redis removed)
├── config/
│   └── config.yaml                  # Configuration file
├── .env                             # Environment variables
├── .env.example                     # Environment template
├── go.mod
└── go.sum
```

## 🗄️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    email_verified BOOLEAN DEFAULT false,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    meta JSONB DEFAULT '{}'::jsonb
);

-- Indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_meta USING GIN(meta);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

### User Passwords Table
```sql
CREATE TABLE user_passwords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    password_history JSONB DEFAULT '[]'::jsonb,
    verification_token VARCHAR(255),
    verification_token_expiry TIMESTAMP WITH TIME ZONE,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP WITH TIME ZONE,
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret VARCHAR(255),
    meta JSONB DEFAULT '{}'::jsonb,  -- stores failed_login_attempts, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_user_passwords_user_id ON user_passwords(user_id);
CREATE INDEX idx_user_passwords_verification_token ON user_passwords(verification_token) WHERE verification_token IS NOT NULL;
CREATE INDEX idx_user_passwords_reset_token ON user_passwords(reset_token) WHERE reset_token IS NOT NULL;
CREATE INDEX idx_user_passwords_meta_gin USING GIN(meta);
```

## 🌐 API Endpoints

### Public Endpoints (No Authentication)

#### Health Check
- `GET /api/v1/health` - System health check
  - Checks: PostgreSQL connection
  - Returns: HTTP 200 (healthy) or 503 (unhealthy)

#### Authentication
- `POST /api/v1/auth/signup` - Register new user
- `POST /api/v1/auth/verify-email` - Verify email with token
- `POST /api/v1/auth/resend-verification` - Resend verification email
- `POST /api/v1/auth/login` - Login with email/password
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token

### Protected Endpoints (Requires JWT Token)

#### Authentication
- `POST /api/v1/auth/change-password` - Change password (requires old password)
- `POST /api/v1/auth/logout` - Logout current session

#### User Management
- `GET /api/v1/users/me` - Get current user profile
- `PATCH /api/v1/users/me` - Update current user profile
- `DELETE /api/v1/users/me` - Soft delete current user account
- `GET /api/v1/users` - List all users (admin)
- `GET /api/v1/users/stats` - Get user statistics (admin)
- `GET /api/v1/users/:id` - Get user by ID (admin)
- `POST /api/v1/users/:id/restore` - Restore deleted user (admin)
- `POST /api/v1/users/:id/disable` - Disable user account (admin)
- `POST /api/v1/users/:id/enable` - Enable user account (admin)

## 🔐 Authentication Flow

### 1. Registration Flow
```
User → POST /auth/signup
  ↓
Create user in database (email_verified = false)
  ↓
Generate verification token
  ↓
Send verification email
  ↓
Return success message
```

### 2. Email Verification Flow
```
User clicks link → POST /auth/verify-email
  ↓
Validate token (check expiry)
  ↓
Mark email_verified = true
  ↓
Clear verification token
  ↓
Return success message
```

### 3. Login Flow
```
User → POST /auth/login
  ↓
Validate credentials
  ↓
Check email_verified
  ↓
Check account enabled
  ↓
Generate JWT access + refresh tokens
  ↓
Return tokens
```

### 4. Protected Request Flow
```
User → GET /users/me (with Authorization header)
  ↓
Extract JWT from header
  ↓
Validate JWT signature
  ↓
Check token expiry
  ↓
Extract user_id from token
  ↓
Process request
  ↓
Return response
```

## 📦 Configuration

### Environment Variables (.env)
```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=debug

# Database (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=2024
DB_NAME=quiz_craft
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESS_TOKEN_EXPIRE=24h
JWT_REFRESH_TOKEN_EXPIRE=168h

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@quizcraft.com
SMTP_FROM_NAME=QuizCraft

# Application
APP_NAME=QuizCraft
APP_ENV=development
LOG_LEVEL=debug
```

## 🚀 How to Run

### Prerequisites
- Go 1.25.3+
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker)

### Quick Start
```bash
# 1. Install dependencies
go mod download

# 2. Copy environment file
cp .env.example .env
# Edit .env with your settings

# 3. Start PostgreSQL (Docker)
docker-compose -f deployments/docker/docker-compose.yml up -d

# 4. Run application
go run main.go

# Server starts on http://localhost:8080
```

### Verify Installation
```bash
# Check health
curl http://localhost:8080/api/v1/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2026-01-10T00:00:00+05:30",
  "checks": {
    "database": {
      "status": "healthy"
    }
  }
}
```

## 🔄 Startup Sequence

When the application starts (main.go), this happens in order:

1. **Load Configuration** (`config.Load()`)
   - Loads .env file
   - Loads config.yaml
   - Binds environment variables
   - Validates required fields

2. **Initialize Logger** (`logger.Init()`)
   - Sets up logrus logger
   - Configures log level

3. **Connect to PostgreSQL** (`database.InitPostgres()`)
   - Establishes database connection
   - Configures connection pool

4. **Run Migrations** (`database.RunMigrations()`)
   - Creates users table
   - Creates user_passwords table
   - Tracks migration versions

5. **Initialize JWT Manager** (`utils.NewJWTManager()`)
   - Sets up JWT signing/validation

6. **Initialize DAOs** (`userdao.Init()`)
   - Sets up UserDAO
   - Sets up PasswordDAO

7. **Initialize Services**
   - Email service
   - Auth core logic

8. **Start HTTP Servers**
   - Internal server (port 8081)
   - External server (port 8080)

## 🧪 Testing

### Manual Testing
See [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md) for detailed API testing examples.

### Automated Tests
*(To be implemented in Phase 8)*

## 🐛 Known Issues / Limitations

1. **No Rate Limiting** - Redis was removed, so no built-in rate limiting
2. **No Session Caching** - Sessions tracked in PostgreSQL (slower than Redis)
3. **No Real-time Features** - No pub/sub without Redis
4. **Email Verification Tokens** - Stored in database (manual expiry checking)
5. **No Automated Testing** - Test suite not yet implemented

## 🎯 What's Missing (To Be Implemented)

### Phase 3: Quiz Management
- Quiz CRUD operations
- Question management
- Answer management
- Quiz categorization

### Phase 4: AI Integration & File Upload
- Google AI API integration
- Quiz generation from files
- File upload handling
- PDF/Document parsing

### Phase 5: Quiz Attempts & Scoring
- Quiz attempt tracking
- Answer submission
- Automatic scoring
- Results & analytics

### Phase 6: Advanced Features
- Leaderboards
- Quiz sharing
- Quiz templates
- Batch operations

### Phase 7: Admin Dashboard
- User management
- Quiz moderation
- Analytics dashboard
- System settings

### Phase 8: Testing & Documentation
- Unit tests
- Integration tests
- API documentation (Swagger)
- Deployment guides

## 📚 Additional Documentation

- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md) - API testing examples
- [REFACTORING_COMPLETE.md](REFACTORING_COMPLETE.md) - Refactoring history
- [approach-document.md](approach-document.md) - Original design document

## 🔧 Maintenance Notes

### Recent Changes
- **2026-01-10**: Removed Redis completely to simplify architecture
  - Removed pkg/cache/redis.go
  - Updated config.go to remove RedisConfig
  - Updated docker-compose.yml to remove Redis service
  - Updated .env files to remove Redis variables
  - Simplified health check to only check PostgreSQL

### Future Improvements
- Add Redis back if rate limiting/caching becomes necessary
- Implement automated test suite
- Add API documentation with Swagger
- Implement proper logging middleware
- Add request tracing
- Implement graceful shutdown
