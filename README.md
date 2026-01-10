# Quiz-Craft Backend

A production-ready Go backend for Quiz-Craft - an AI-powered quiz generation and management platform.

## 🚀 Current Status

### ✅ Phase 1: Foundation & Infrastructure (COMPLETE)
- Clean architecture project structure
- Configuration management (Viper + environment variables)
- PostgreSQL database with GORM
- Database migrations (golang-migrate)
- Middleware (CORS, Logging, Error handling)
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
- **17 authentication & user endpoints**

### 📋 Next Phases
- **Phase 3**: Quiz Management (CRUD)
- **Phase 4**: AI Integration & File Upload
- **Phase 5**: Quiz Attempts & Scoring
- **Phase 6**: Advanced Features
- **Phase 7**: Admin Dashboard
- **Phase 8**: Testing & Production

**See [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for complete roadmap**

## 📋 Prerequisites

- Go 1.25.3 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (or use Docker)

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

### 4. Start PostgreSQL (using Docker)
```bash
docker-compose -f deployments/docker/docker-compose.yml up -d
```

### 5. Run the application
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

## 📦 Project Structure

```
backend/
├── main.go                        # Application entry point
├── api/
│   ├── auth.go                    # Auth endpoints
│   └── user.go                    # User endpoints
├── auth/
│   └── core/                      # Auth business logic
├── user/
│   ├── dao/                       # Data Access Objects
│   └── model/                     # Data models
├── internal/
│   ├── config/                    # Configuration loader
│   ├── middleware/                # HTTP middleware
│   ├── services/                  # Services (email, etc.)
│   └── utils/                     # Utilities (JWT, password, etc.)
├── pkg/
│   ├── database/
│   │   ├── postgres.go
│   │   ├── migrate.go
│   │   └── migrations/            # SQL migration files
│   └── logger/
│       └── logger.go
├── commons/                       # Common utilities
├── deployments/docker/
│   └── docker-compose.yml         # PostgreSQL service
├── config/config.yaml
├── docs/                          # Documentation
├── .env.example
├── go.mod
└── README.md
```

**See [docs/CURRENT_STATE.md](docs/CURRENT_STATE.md) for detailed architecture**

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
  "timestamp": "2026-01-10T00:00:00+05:30",
  "checks": {
    "database": {
      "status": "healthy"
    }
  }
}
```

**See [docs/API_TESTING_GUIDE.md](docs/API_TESTING_GUIDE.md) for complete API documentation**

## 🔧 Configuration

Configuration can be set via:
1. `config/config.yaml` file
2. Environment variables (takes precedence)
3. `.env` file (loaded automatically)

Key configuration options:
- `SERVER_PORT`: API server port (default: 8080)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL connection
- `JWT_SECRET`: JWT signing secret
- `JWT_ACCESS_TOKEN_EXPIRE`: Access token expiry (default: 24h)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`: Email configuration
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

# List tables
\dt

# View users table
SELECT * FROM users;
```

## 🧪 Testing

Run tests:
```bash
make test
# or
go test -v ./...
```

## 🚀 Deployment

Production deployment guide will be available in Phase 8.

## 📖 Documentation

- **[CURRENT_STATE.md](docs/CURRENT_STATE.md)** - Current architecture and features
- **[IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md)** - Complete roadmap for all phases
- **[API_TESTING_GUIDE.md](docs/API_TESTING_GUIDE.md)** - API endpoint examples
- **[QUICKSTART.md](docs/QUICKSTART.md)** - Quick start guide
- **[approach-document.md](docs/approach-document.md)** - Original design document

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
docker-compose -f deployments/docker/docker-compose.yml up -d

# Check PostgreSQL logs
docker logs quiz-craft-postgres

# Verify connection
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft
```

### Migration errors
```bash
# Check schema_migrations table
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft -c "SELECT * FROM schema_migrations;"

# Reset database (WARNING: deletes all data)
docker-compose -f deployments/docker/docker-compose.yml down -v
docker-compose -f deployments/docker/docker-compose.yml up -d
# Wait 10 seconds, then restart app
```

## 📞 Support

- Check [docs/CURRENT_STATE.md](docs/CURRENT_STATE.md) for architecture details
- Check [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for roadmap
- Check [docs/API_TESTING_GUIDE.md](docs/API_TESTING_GUIDE.md) for API examples

## 🎯 What's Next?

We're currently at **Phase 2 Complete**. Ready to implement Phase 3 (Quiz Management)?

See [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for the complete implementation roadmap!
