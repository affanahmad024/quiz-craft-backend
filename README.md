# QuizCraft Backend

A production-ready Go backend for QuizCraft - an AI-powered quiz generation and management platform with comprehensive testing and deployment infrastructure.

## 🚀 Project Status: **PRODUCTION READY** ✅

All 8 phases completed and production-ready!

### ✅ Phase 1: Foundation & Infrastructure (COMPLETE)
- Clean architecture project structure
- Configuration management (Viper + environment variables)
- PostgreSQL database with optimized queries
- Database migrations (golang-migrate)
- Middleware (CORS, Logging, Error handling, Authentication)
- Health check endpoints
- Docker Compose for local development

### ✅ Phase 2: Authentication & User Management (COMPLETE)
- User registration with email verification
- JWT-based authentication (access + refresh tokens)
- Login/Logout functionality
- Password management (forgot, reset, change)
- User profile CRUD operations
- Soft delete with restore functionality
- Account locking after failed attempts
- Password history and strength validation
- Email service with HTML templates
- Role-based access control (user, admin)

### ✅ Phase 3: Quiz Management (COMPLETE)
- Quiz CRUD operations with validation
- Question & Answer management
- Multiple question types (multiple choice, true/false, short answer, essay)
- Quiz publishing & unpublishing
- Public/private quiz visibility
- Time limits & passing scores
- Quiz categories & tags
- Quiz templates
- Share links with unique codes

### ✅ Phase 4: AI Integration & File Upload (COMPLETE)
- Google AI (Gemini) integration for quiz generation
- Generate quizzes from text content
- Generate quizzes from uploaded files (PDF, DOCX, TXT)
- Background job processing for AI generation
- File upload with validation (5MB limit, type checking)
- File storage and management
- Job status tracking

### ✅ Phase 5: Quiz Attempts & Scoring (COMPLETE)
- Start quiz attempts with time tracking
- Progressive answer submission
- Auto-submit on time expiry (background worker)
- Real-time scoring with partial credit
- Attempt history and statistics
- Answer review and feedback
- Retake attempts support

### ✅ Phase 6: Search & Discovery (COMPLETE)
- Advanced quiz search with filters (category, difficulty, tags)
- Pagination and sorting
- Popular quizzes discovery
- Recently created quizzes
- Quiz statistics (attempts, pass rate, average score)
- User statistics (total attempts, quizzes created, streak)
- Tag cloud and category listing
- Share link access

### ✅ Phase 7: Admin Dashboard (COMPLETE)
- User management (view, disable, enable, delete, role change)
- Quiz moderation (view all, delete)
- Platform statistics (total users, quizzes, attempts)
- Activity logs tracking
- Activity timeline
- User search and filtering

### ✅ Phase 8: Testing & Production Readiness (COMPLETE)
- **69 unit tests** covering auth, scoring, validation, and statistics
- **70+ manual test scenarios** for comprehensive coverage
- **Swagger/OpenAPI documentation** at `/swagger/index.html`
- **Docker production deployment** with multi-stage builds
- **Nginx reverse proxy** with SSL/TLS support
- **Rate limiting** (100 req/min API, 5 req/min auth)
- **Security headers** (HSTS, CSP, X-Frame-Options)
- **Health checks** and monitoring
- **Production environment templates**
- **Deployment guide** with backup strategy

## 📊 Implementation Statistics

- **67 REST API Endpoints** fully implemented
- **14 Database Tables** with optimized schemas
- **69 Unit Tests** (auth, scoring, validation, statistics)
- **70+ Manual Test Scenarios**
- **Comprehensive API documentation** covering all 105 endpoints
- **100% Phase Completion** (8/8 phases)

## 🎯 Recent Enhancements

### AI Quiz Extraction from Images (Jan 2026) 🆕
- **Upload quiz photos or PDFs** and automatically extract structure
- Google Gemini Vision API extracts questions, answers, and correct options
- **Automatic image detection** - extracts images embedded in questions
- Image compression (75% quality) optimizes storage
- Support for JPG, PNG, PDF formats (max 10MB)
- Background processing with progress tracking
- Images stored in `meta.image_url` field for easy access
- Perfect for digitizing paper quizzes, textbooks, or exam papers

### Bulk Question Creation (Jan 2026)
- Add multiple questions with answers in a single API call
- Atomic transaction ensures all-or-nothing creation
- Automatic order indexing for questions
- Full validation before database insertion

### Enhanced Quiz Listings (Jan 2026)
- Quiz listings now include **latest attempt information**
- See your score, status, and timestamps on each quiz
- Works for both "My Quizzes" and "Public Quizzes"
- Helps users track progress and decide which quizzes to retake

### Admin Security Fixes (Jan 2026)
- Fixed security vulnerability in user management endpoints
- User enable/disable/restore now **admin-only**
- Removed insecure routes from general user endpoints
- Added proper authorization checks and activity logging

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (or use Docker)
- Redis 7+ (optional, for caching)
- SMTP server (for email verification)
- Google AI API key (for AI quiz generation)

## 🛠️ Quick Start

### 1. Clone and navigate
```bash
cd quiz-craft/backend
```

### 2. Install dependencies
```bash
go mod download
```

### 3. Set up environment
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Start services (PostgreSQL)
```bash
docker-compose -f deployments/docker/docker-compose.yml up -d
```

### 5. Run the application
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

### 6. View API Documentation
Open your browser to `http://localhost:8080/swagger/index.html`

## 📦 Project Structure

```
backend/
├── main.go                        # Application entry point with Swagger
├── api/                           # HTTP handlers
│   ├── auth.go                    # Auth endpoints (9)
│   ├── user.go                    # User endpoints (4)
│   ├── quiz.go                    # Quiz endpoints (9)
│   ├── question.go                # Question/Answer endpoints (7)
│   ├── attempt.go                 # Attempt endpoints (6)
│   ├── search.go                  # Search endpoints (4)
│   ├── statistics.go              # Statistics endpoints (2)
│   ├── template.go                # Template endpoints (3)
│   ├── share.go                   # Share endpoints (1)
│   ├── admin.go                   # Admin endpoints (12)
│   ├── file.go                    # File endpoints (3)
│   └── ai.go                      # AI endpoints (4)
├── auth/core/                     # Authentication business logic
├── user/                          # User module
│   ├── dao/                       # Data Access Objects
│   └── model/                     # Data models
├── quiz/                          # Quiz module
│   ├── dao/                       # Data Access Objects
│   └── model/                     # Data models
├── internal/
│   ├── config/                    # Configuration loader
│   ├── middleware/                # HTTP middleware
│   ├── services/                  # Services (email, AI)
│   └── utils/                     # Utilities (JWT, password)
├── pkg/
│   ├── database/
│   │   ├── postgres.go
│   │   ├── migrate.go
│   │   └── migrations/            # 14 SQL migration files
│   └── logger/
├── worker/                        # Background workers
│   ├── worker.go                  # Worker orchestrator
│   ├── attempt_expiry.go          # Auto-submit expired attempts
│   └── ai_generation.go           # AI quiz generation
├── tests/                         # Test suite
│   ├── test_utils.go              # Test utilities
│   └── unit/                      # Unit tests
│       ├── auth_test.go           # 17 auth tests
│       ├── scoring_test.go        # 12 scoring tests
│       ├── validation_test.go     # 28 validation tests
│       └── statistics_test.go     # 12 statistics tests
├── docs/                          # Documentation
│   ├── README.md                  # Documentation index
│   ├── COMPLETE_API_REFERENCE.md  # Complete API reference (105 endpoints) ⭐
│   ├── FRONTEND_INTEGRATION.md    # Integration guide with code examples
│   ├── QUICKSTART.md              # Getting started guide
│   └── EDUCATOR_ROLE_IMPLEMENTATION.md  # Architecture reference
├── nginx/                         # Nginx configuration
│   ├── nginx.conf                 # Main config
│   └── conf.d/
│       └── quizcraft.conf         # Server config with SSL
├── Dockerfile                     # Multi-stage production build
├── docker-compose.prod.yml        # Production orchestration
├── .env.production.example        # Production environment template
└── README.md
```

## 🎯 API Endpoints Overview

### Authentication (9 endpoints)
- `POST /api/v1/auth/signup` - Register new user
- `POST /api/v1/auth/verify-email` - Verify email
- `POST /api/v1/auth/resend-verification` - Resend verification
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password
- `POST /api/v1/auth/change-password` - Change password
- `POST /api/v1/auth/logout` - Logout

### User Management (4 endpoints)
- `GET /api/v1/users/me` - Get current user
- `PATCH /api/v1/users/me` - Update profile
- `DELETE /api/v1/users/me` - Delete account
- `GET /api/v1/users/me/statistics` - User statistics

### Quiz Management (9 endpoints)
- `POST /api/v1/quizzes` - Create quiz
- `GET /api/v1/quizzes/:id` - Get quiz
- `GET /api/v1/quizzes/:id/full` - Get quiz with questions
- `GET /api/v1/quizzes` - List my quizzes
- `GET /api/v1/quizzes/public` - List public quizzes
- `PATCH /api/v1/quizzes/:id` - Update quiz
- `DELETE /api/v1/quizzes/:id` - Delete quiz
- `POST /api/v1/quizzes/:id/publish` - Publish quiz
- `POST /api/v1/quizzes/:id/unpublish` - Unpublish quiz

### Questions & Answers (11 endpoints)
- `POST /api/v1/quizzes/:id/questions` - Add question (with optional answers)
- `POST /api/v1/quizzes/:id/questions/bulk` - Add multiple questions at once
- `GET /api/v1/quizzes/questions/:qid` - Get question
- `PATCH /api/v1/quizzes/questions/:qid` - Update question
- `DELETE /api/v1/quizzes/questions/:qid` - Delete question
- `POST /api/v1/quizzes/questions/:qid/answers` - Add answer
- `PATCH /api/v1/quizzes/answers/:aid` - Update answer
- `DELETE /api/v1/quizzes/answers/:aid` - Delete answer
- `POST /api/v1/quizzes/:id/add-questions-from-text` - Add AI-generated questions from text 🆕
- `POST /api/v1/quizzes/:id/add-questions-from-file` - Add AI-generated questions from file 🆕
- `POST /api/v1/quizzes/:id/extract-questions-from-file` - Extract and add questions from file 🆕

### Quiz Attempts (6 endpoints)
- `POST /api/v1/quizzes/:id/attempts` - Start attempt
- `GET /api/v1/attempts/:id` - Get attempt
- `POST /api/v1/attempts/:id/answers` - Submit answer
- `POST /api/v1/attempts/:id/finish` - Finish attempt
- `GET /api/v1/attempts/:id/results` - Get results
- `GET /api/v1/attempts` - List my attempts

### Search & Discovery (4 endpoints)
- `GET /api/v1/quizzes/search` - Search quizzes
- `GET /api/v1/categories` - Get categories
- `GET /api/v1/tags` - Get popular tags
- `GET /api/v1/share/:code` - Get quiz by share code

### Statistics (2 endpoints)
- `GET /api/v1/quizzes/:id/stats` - Quiz statistics (creator)
- `GET /api/v1/quizzes/:id/statistics` - Detailed statistics

### Templates (3 endpoints)
- `GET /api/v1/templates` - List templates
- `POST /api/v1/quizzes/:id/save-as-template` - Save as template
- `POST /api/v1/templates/:id/use` - Create from template

### Sharing (1 endpoint)
- `POST /api/v1/quizzes/:id/share` - Generate share link

### Admin (13 endpoints)
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/admin/users/:id` - Get user details
- `PATCH /api/v1/admin/users/:id/role` - Change user role
- `POST /api/v1/admin/users/:id/disable` - Disable user
- `POST /api/v1/admin/users/:id/enable` - Enable user
- `POST /api/v1/admin/users/:id/restore` - Restore deleted user 🆕
- `DELETE /api/v1/admin/users/:id` - Delete user permanently
- `GET /api/v1/admin/quizzes` - List all quizzes
- `GET /api/v1/admin/quizzes/:id` - Get quiz details
- `DELETE /api/v1/admin/quizzes/:id` - Delete quiz
- `GET /api/v1/admin/stats` - Platform statistics
- `GET /api/v1/admin/stats/activity` - Activity timeline
- `GET /api/v1/admin/logs` - Activity logs

### File Upload (3 endpoints)
- `POST /api/v1/files/upload` - Upload file
- `GET /api/v1/files` - List files
- `DELETE /api/v1/files/:id` - Delete file

### AI Generation (5 endpoints)
- `POST /api/v1/ai/generate-quiz-from-text` - Generate new questions from text
- `POST /api/v1/ai/generate-quiz-from-file` - Generate new questions from file
- `POST /api/v1/ai/extract-quiz-from-file` - Extract existing questions from file (image/PDF) 🆕
- `GET /api/v1/ai/generation-status/:id` - Check generation status
- `GET /api/v1/ai/jobs` - List generation jobs

### Health Check (2 endpoints)
- `GET /api/v1/health` - Health check with dependencies
- `GET /api/v1/ping` - Simple ping

**See [docs/COMPLETE_API_REFERENCE.md](docs/COMPLETE_API_REFERENCE.md) for complete API documentation with all 105 endpoints and request/response examples**

## 🔧 Configuration

Key environment variables:

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=quiz_craft

# JWT
JWT_SECRET=your-super-secret-key
JWT_ACCESS_TOKEN_EXPIRE=24h
JWT_REFRESH_TOKEN_EXPIRE=168h

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@quizcraft.com

# Google AI
GOOGLE_AI_API_KEY=your-google-ai-key

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
AUTH_RATE_LIMIT_REQUESTS_PER_MINUTE=5
```

## 📝 Makefile Commands

```bash
make help          # Show available commands
make run           # Run the application
make build         # Build the binary
make test          # Run unit tests
make docker-up     # Start PostgreSQL and Redis
make docker-down   # Stop docker containers
make docker-logs   # View docker logs
make docker-clean  # Stop and remove volumes
make deps          # Install dependencies
make fmt           # Format code
make clean         # Clean build artifacts
make setup         # Setup development environment
```

## 🧪 Testing

### Run Unit Tests
```bash
# Run all tests
make test

# Run specific test file
go test -v ./tests/unit/auth_test.go

# Run with coverage
go test -cover ./tests/unit/...
```

### Test Coverage
- **69 unit tests** across 4 test files
- Auth tests: 17 tests
- Scoring tests: 12 tests
- Validation tests: 28 tests
- Statistics tests: 12 tests

### Manual Testing
See comprehensive manual testing guides:
- Test all API endpoints using Swagger UI at `http://localhost:8080/swagger/index.html`

## 🚀 Production Deployment

### Using Docker Compose (Recommended)

```bash
# 1. Copy production environment
cp .env.production.example .env

# 2. Edit .env with your production values
nano .env

# 3. Build and start services
docker-compose -f docker-compose.prod.yml up -d

# 4. Check health
curl http://your-domain.com/api/v1/health
```

### Services Included
- **API**: Go backend with auto-restart
- **PostgreSQL**: Database with health checks
- **Nginx**: Reverse proxy with SSL/TLS
- **Redis**: Optional caching layer

### SSL/TLS Configuration
```bash
# Using Let's Encrypt (Certbot)
certbot certonly --standalone -d your-domain.com
```

**See [.env.production.example](.env.production.example) for production environment configuration**

## 📖 Documentation

### 🚀 Quick Start
- **[QUICKSTART.md](docs/QUICKSTART.md)** - Get started in 5 minutes
- **Swagger UI** - Interactive API explorer at `http://localhost:8080/swagger/index.html`

### 📚 API Documentation
- **[COMPLETE_API_REFERENCE.md](docs/COMPLETE_API_REFERENCE.md)** ⭐ - **All 105 endpoints** with complete request/response formats
  - Authentication, Users, Quizzes, Questions, Attempts
  - AI Generation, Notifications, Live Sessions, Educators
  - Search, Templates, Admin - Error codes and examples

### 💻 Frontend Integration
- **[FRONTEND_INTEGRATION.md](docs/FRONTEND_INTEGRATION.md)** - Practical integration guide
  - React hooks and components
  - WebSocket connections (Notifications, Live Sessions)
  - Authentication flow with auto-refresh
  - Service modules and ready-to-use examples

### 🏗️ Architecture Reference
- **[EDUCATOR_ROLE_IMPLEMENTATION.md](docs/EDUCATOR_ROLE_IMPLEMENTATION.md)** - System architecture
  - Database schema design
  - WebSocket infrastructure
  - Live sessions architecture

### 📋 Documentation Index
- **[docs/README.md](docs/README.md)** - Complete documentation guide with all links

### ⚙️ Deployment
- **[.env.production.example](.env.production.example)** - Production environment template

## 🔒 Security Features

- ✅ JWT authentication with refresh tokens
- ✅ Password hashing with bcrypt
- ✅ Email verification required
- ✅ Account lockout after failed attempts
- ✅ Password strength validation
- ✅ Password history tracking
- ✅ Role-based access control (RBAC)
- ✅ Rate limiting per endpoint
- ✅ CORS configuration
- ✅ Security headers (HSTS, CSP, X-Frame-Options)
- ✅ SQL injection protection (parameterized queries)
- ✅ XSS prevention
- ✅ HTTPS enforcement in production

## ⚡ Performance Features

- ✅ Database connection pooling
- ✅ Optimized SQL queries with indexes
- ✅ Pagination for large datasets
- ✅ Gzip compression (Nginx)
- ✅ Background workers for long tasks
- ✅ Redis caching support
- ✅ Multi-stage Docker builds (small images)
- ✅ Health checks and monitoring

## 🐛 Troubleshooting

### Tests failing with database connection error
```bash
# Start PostgreSQL
docker-compose -f deployments/docker/docker-compose.yml up -d

# Create test database
docker exec -it quiz-craft-postgres psql -U postgres -c "CREATE DATABASE quiz_craft_test;"

# Update test credentials in tests/test_utils.go if needed
```

### Port already in use
```bash
# Change SERVER_PORT in .env
SERVER_PORT=8081
```

### Database connection failed
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check PostgreSQL logs
docker logs quiz-craft-postgres

# Verify connection
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft
```

### Migration errors
```bash
# Check migration status
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft -c "SELECT * FROM schema_migrations;"

# Reset database (WARNING: deletes all data)
docker-compose -f deployments/docker/docker-compose.yml down -v
docker-compose -f deployments/docker/docker-compose.yml up -d
```

## 🎯 Key Features

✨ **Complete Authentication System**
- Email verification workflow
- Password reset via email
- JWT with refresh tokens
- Account security (lockout, password history)

✨ **Comprehensive Quiz Management**
- Multiple question types
- AI-powered quiz generation
- Template system
- Share links
- Public/private visibility

✨ **Advanced Quiz Taking**
- Progressive answer submission
- Time limits with auto-submit
- Real-time scoring
- Attempt history
- Retake support

✨ **Discovery & Search**
- Advanced search with filters
- Category and tag browsing
- Popular quiz discovery
- Statistics and analytics

✨ **Admin Dashboard**
- User management
- Quiz moderation
- Platform statistics
- Activity tracking

✨ **Production Ready**
- 69 unit tests
- Swagger documentation
- Docker deployment
- Nginx reverse proxy
- Rate limiting
- Security hardened

## 📞 Support

For questions or issues:
1. Check the [Complete API Reference](docs/COMPLETE_API_REFERENCE.md) - All 105 endpoints
2. Review [Frontend Integration Guide](docs/FRONTEND_INTEGRATION.md) - Code examples
3. Check [Quick Start Guide](docs/QUICKSTART.md) - Get running in 5 minutes
4. View Swagger UI at `/swagger/index.html`

## 📄 License

[Add your license here]

---

**🎉 QuizCraft Backend - Production Ready & Feature Complete!**

Built with ❤️ using Go, PostgreSQL, Google AI, Docker, and Nginx.
