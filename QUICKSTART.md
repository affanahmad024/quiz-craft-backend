# Quick Start Guide - Phase 1

## ✅ What We've Built

Phase 1 is complete! You now have:
- Complete project structure following clean architecture
- Configuration management (supports .env, YAML, and environment variables)
- PostgreSQL connection with GORM
- Redis connection (optional)
- Database migrations system
- Middleware (CORS, Logging, Error Handling)
- Health check endpoints
- Docker Compose setup for local development

## 🚀 How to Run

### Option 1: With Docker (Recommended)

1. **Start Docker Desktop** (required for PostgreSQL and Redis)

2. **Start the database services:**
```bash
cd quiz-craft/backend
docker-compose -f deployments/docker/docker-compose.yml up -d
```

3. **Wait for services to be ready (about 10 seconds)**

4. **Run the application:**
```bash
go run cmd/server/main.go
```

The API will start on `http://localhost:8080`

### Option 2: Without Docker

If you have PostgreSQL and Redis installed locally:

1. **Update .env file** with your local database credentials:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_postgres_user
DB_PASSWORD=your_postgres_password
DB_NAME=quiz_craft

REDIS_HOST=localhost
REDIS_PORT=6379
```

2. **Create the database:**
```sql
CREATE DATABASE quiz_craft;
```

3. **Run the application:**
```bash
go run cmd/server/main.go
```

## 🧪 Testing the API

### 1. Ping Endpoint
```bash
curl http://localhost:8080/api/v1/ping
```

Expected response:
```json
{
  "message": "pong"
}
```

### 2. Health Check Endpoint
```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
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

## 📊 What Happens on Startup

When you run the application, it will:

1. ✅ Load configuration from `.env` and `config/config.yaml`
2. ✅ Initialize the logger
3. ✅ Connect to PostgreSQL
4. ✅ Run database migrations (creates `users` and `user_passwords` tables)
5. ✅ Connect to Redis (optional, warns if fails)
6. ✅ Setup routes and middleware
7. ✅ Start HTTP server on port 8080

## 🗄️ Database Tables Created

After first run, you'll have these tables:
- `users` - User accounts
- `user_passwords` - Authentication credentials
- `schema_migrations` - Migration tracking

You can verify:
```bash
# Connect to PostgreSQL
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft

# List tables
\dt

# View users table structure
\d users
```

## 🐛 Troubleshooting

### "Failed to connect to database"
- Make sure Docker Desktop is running
- Check if containers are up: `docker ps`
- View logs: `docker logs quiz-craft-postgres`

### "Port 8080 already in use"
- Change `SERVER_PORT` in `.env` file
- Or stop the process using port 8080

### "Failed to connect to Redis"
- This is optional, the app will continue without Redis
- To fix: Make sure Redis container is running
- Check: `docker logs quiz-craft-redis`

## 📁 Project Structure Overview

```
backend/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── config/config.go         # Configuration loader
│   ├── handlers/                # HTTP request handlers
│   │   └── health_handler.go
│   └── middleware/              # HTTP middleware
│       ├── logger.go            # Request logging
│       ├── cors.go              # CORS configuration
│       └── error_handler.go    # Error handling
├── pkg/
│   ├── database/                # Database connection & migrations
│   │   ├── postgres.go
│   │   ├── migrate.go
│   │   └── migrations/          # SQL migration files
│   ├── cache/redis.go           # Redis connection
│   └── logger/logger.go         # Logger setup
├── api/routes/router.go         # Route definitions
├── .env                         # Environment variables
└── config/config.yaml           # Configuration file
```

## 🎯 Next Steps (Phase 2)

Phase 1 is complete! Next, we'll implement:
- User registration and signup
- Email verification
- Login with JWT authentication
- Password reset
- User profile management

## 💡 Tips

1. **View logs**: All logs are printed to console with structured format
2. **Hot reload**: Install `air` for automatic reloading during development:
   ```bash
   go install github.com/air-verse/air@latest
   air
   ```
3. **Database GUI**: Use pgAdmin or DBeaver to visualize the database
4. **Redis GUI**: Use RedisInsight to view Redis data

## 🔧 Useful Commands

```bash
# Build the application
go build -o bin/server.exe cmd/server/main.go

# Run the application
go run cmd/server/main.go

# Start Docker services
docker-compose -f deployments/docker/docker-compose.yml up -d

# Stop Docker services
docker-compose -f deployments/docker/docker-compose.yml down

# View Docker logs
docker-compose -f deployments/docker/docker-compose.yml logs -f

# Clean up (remove volumes)
docker-compose -f deployments/docker/docker-compose.yml down -v

# Format code
go fmt ./...

# Install dependencies
go mod download
go mod tidy
```

## ✅ Verification Checklist

- [ ] Docker Desktop is running
- [ ] PostgreSQL container is running (`docker ps`)
- [ ] Redis container is running (`docker ps`)
- [ ] Application starts without errors
- [ ] `/api/v1/ping` returns "pong"
- [ ] `/api/v1/health` shows all services healthy
- [ ] Database tables are created (users, user_passwords)

If all checks pass, Phase 1 is successfully set up! 🎉
