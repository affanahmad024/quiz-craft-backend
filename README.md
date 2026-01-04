# Quiz-Craft Backend

A production-ready Go backend for Quiz-Craft - an AI-powered quiz generation and management platform.

## 🚀 Features

- **Phase 1 (Completed)**: Foundation & Setup
  - ✅ Project structure with clean architecture
  - ✅ Configuration management (Viper + environment variables)
  - ✅ PostgreSQL database connection with GORM
  - ✅ Redis cache integration
  - ✅ Database migrations (golang-migrate)
  - ✅ Middleware (CORS, Logging, Error handling)
  - ✅ Health check endpoints
  - ✅ Docker Compose for local development

- **Phase 2 (Completed)**: Authentication & User Management
  - ✅ User registration with email verification
  - ✅ JWT-based authentication (access + refresh tokens)
  - ✅ Login/Logout functionality
  - ✅ Password management (forgot, reset, change)
  - ✅ User profile CRUD operations
  - ✅ Soft delete with restore functionality
  - ✅ Account locking after failed attempts
  - ✅ Password history and strength validation
  - ✅ Email service with HTML templates
  - ✅ 17 authentication & user endpoints

- **Phase 3 (Upcoming)**: Quiz Management (CRUD)
- **Phase 4 (Upcoming)**: AI Integration & File Upload
- **Phase 5 (Upcoming)**: Quiz Attempts & Scoring

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (or use Docker)
- Redis 7+ (or use Docker)

## 🛠️ Installation

### 1. Clone the repository
```bash
cd quiz-craft/backend
```

### 2. Install dependencies
```bash
go mod download
```

### 3. Set up environment variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Start PostgreSQL and Redis (using Docker)
```bash
make docker-up
# or
docker-compose -f deployments/docker/docker-compose.yml up -d
```

### 5. Run the application
```bash
make run
# or
go run cmd/server/main.go
```

The API will be available at `http://localhost:8080`

## 📦 Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration loader
│   ├── models/                    # Database models (Phase 2+)
│   ├── repositories/              # Data access layer (Phase 2+)
│   ├── services/                  # Business logic (Phase 2+)
│   ├── handlers/
│   │   └── health_handler.go     # HTTP handlers
│   ├── middleware/
│   │   ├── logger.go              # Request logging
│   │   ├── cors.go                # CORS configuration
│   │   └── error_handler.go      # Error handling
│   ├── utils/                     # Utility functions (Phase 2+)
│   └── workers/                   # Background workers (Phase 4+)
├── pkg/
│   ├── database/
│   │   ├── postgres.go            # PostgreSQL connection
│   │   ├── migrate.go             # Migration runner
│   │   └── migrations/            # SQL migration files
│   ├── cache/
│   │   └── redis.go               # Redis connection
│   └── logger/
│       └── logger.go              # Logger initialization
├── api/
│   └── routes/
│       └── router.go              # Route definitions
├── tests/                         # Tests (Phase 8)
├── deployments/
│   └── docker/
│       └── docker-compose.yml     # Docker services
├── config/
│   └── config.yaml                # Configuration file
├── .env.example                   # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🎯 Available Endpoints

### Health Check
- `GET /api/v1/health` - Check API and dependencies health
- `GET /api/v1/ping` - Simple ping endpoint

Example:
```bash
curl http://localhost:8080/api/v1/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-03T18:00:00Z",
  "checks": {
    "database": {
      "status": "healthy"
    },
    "redis": {
      "status": "healthy"
    }
  }
}
```

## 🔧 Configuration

Configuration can be set via:
1. `config/config.yaml` file
2. Environment variables (takes precedence)
3. `.env` file (loaded automatically)

Key configuration options:
- `SERVER_PORT`: API server port (default: 8080)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL connection
- `REDIS_HOST`, `REDIS_PORT`: Redis connection
- `JWT_SECRET`: JWT signing secret (Phase 2)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## 📝 Makefile Commands

```bash
make help          # Show available commands
make run           # Run the application
make build         # Build the binary
make test          # Run tests
make docker-up     # Start PostgreSQL and Redis
make docker-down   # Stop docker containers
make docker-logs   # View docker logs
make docker-clean  # Stop and remove volumes
make deps          # Install dependencies
make fmt           # Format code
make clean         # Clean build artifacts
make setup         # Setup development environment
```

## 🗄️ Database Migrations

Migrations are automatically run when the application starts. You can also run them manually:

```bash
# Run pending migrations
make migrate-up

# Rollback last migration
make migrate-down
```

Migration files are located in `pkg/database/migrations/`:
- `000001_create_users_table.up.sql` / `000001_create_users_table.down.sql`
- `000002_create_user_passwords_table.up.sql` / `000002_create_user_passwords_table.down.sql`

## 🐛 Debugging

### Check if services are running:
```bash
docker-compose -f deployments/docker/docker-compose.yml ps
```

### View logs:
```bash
# Application logs (when running with make run)
# Check terminal output

# Docker services logs
make docker-logs
```

### Connect to PostgreSQL:
```bash
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft
```

### Connect to Redis:
```bash
docker exec -it quiz-craft-redis redis-cli
```

## 🧪 Testing

Run tests:
```bash
make test
# or
go test -v ./...
```

## 🚀 Deployment

See [approach-document.md](approach-document.md) for production deployment guidelines.

## 📖 API Documentation

API documentation will be available via Swagger/OpenAPI in Phase 8.

## 🤝 Contributing

1. Follow the project structure
2. Write tests for new features
3. Run `make fmt` before committing
4. Update documentation as needed

## 📄 License

[Add your license here]

## 🆘 Troubleshooting

### Port already in use
```bash
# Change SERVER_PORT in .env file
SERVER_PORT=8081
```

### Database connection failed
```bash
# Make sure PostgreSQL is running
make docker-up

# Check PostgreSQL logs
docker logs quiz-craft-postgres
```

### Redis connection failed
```bash
# Redis is optional, the application will warn but continue
# To fix: make sure Redis is running
make docker-up
```

## 📞 Support

For issues and questions, check the [approach-document.md](approach-document.md) or create an issue in the repository.
