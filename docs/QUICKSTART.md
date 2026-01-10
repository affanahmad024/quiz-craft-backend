# Quick Start Guide

Last Updated: January 10, 2026

## ✅ What's Complete

**Phase 1 & 2 Complete!** You have a fully functional authentication system:
- Clean architecture with PostgreSQL database
- User registration with email verification
- JWT-based authentication (access + refresh tokens)
- Login/Logout with session management
- Password reset and change functionality
- 17 authentication & user endpoints
- Automatic database migrations
- Docker Compose for development

## 🚀 How to Run

### Prerequisites
- Go 1.25.3+
- Docker Desktop (for PostgreSQL)
- Git

### Step-by-Step Setup

**1. Navigate to backend directory:**
```bash
cd quiz-craft/backend
```

**2. Install dependencies:**
```bash
go mod download
```

**3. Copy environment file:**
```bash
cp .env.example .env
```

**4. Edit .env file** with your settings:
```env
# Update these values
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
JWT_SECRET=change-this-to-a-random-secret-key
```

**5. Start PostgreSQL (Docker):**
```bash
docker-compose -f deployments/docker/docker-compose.yml up -d
```

**6. Wait 10 seconds** for PostgreSQL to be ready

**7. Run the application:**
```bash
go run main.go
```

You should see:
```
INFO[2026-01-10T...] Starting QuizCraft Backend API
INFO[2026-01-10T...] Environment: development
INFO[2026-01-10T...] Connecting to PostgreSQL...
INFO[2026-01-10T...] PostgreSQL connected successfully
INFO[2026-01-10T...] Running database migrations...
INFO[2026-01-10T...] Starting servers...
INFO[2026-01-10T...] Internal server starting on 0.0.0.0:8081
INFO[2026-01-10T...] External server starting on 0.0.0.0:8080
```

The API is now running at `http://localhost:8080`

## 🧪 Testing the API

### 1. Health Check
```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
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

### 2. Test User Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "name": "Test User"
  }'
```

Expected response:
```json
{
  "message": "User registered successfully. Please check your email to verify your account."
}
```

### 3. Check Email
You should receive an email verification link (check spam folder if using real email)

**For more API examples, see [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md)**

## 📊 What Happens on Startup

When you run `go run main.go`, this happens in order:

1. ✅ **Load Configuration** - Reads .env and config.yaml
2. ✅ **Initialize Logger** - Sets up structured logging
3. ✅ **Connect to PostgreSQL** - Establishes database connection
4. ✅ **Run Migrations** - Creates/updates database tables:
   - `users` table
   - `user_passwords` table
   - `schema_migrations` table (tracks migrations)
5. ✅ **Initialize JWT Manager** - Sets up token signing/validation
6. ✅ **Initialize DAOs** - Sets up data access layer
7. ✅ **Initialize Services** - Sets up email service
8. ✅ **Start HTTP Servers** - Internal (8081) and External (8080)

## 🗄️ Database Tables Created

After first run, check your database:

```bash
# Connect to PostgreSQL
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft

# List tables
\dt
```

You should see:
- `users` - User accounts (email, name, etc.)
- `user_passwords` - Authentication data (password hashes, tokens, 2FA)
- `schema_migrations` - Migration version tracking

## 🐛 Troubleshooting

### "Failed to connect to database"
```bash
# Check if PostgreSQL container is running
docker ps

# Should see: quiz-craft-postgres

# If not running, start it
docker-compose -f deployments/docker/docker-compose.yml up -d

# Check logs
docker logs quiz-craft-postgres
```

### "Port 8080 already in use"
```bash
# Option 1: Change port in .env
SERVER_PORT=8081

# Option 2: Find and kill process using port 8080
# Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac:
lsof -ti:8080 | xargs kill -9
```

### "Migration failed"
```bash
# Check current migration version
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft -c "SELECT * FROM schema_migrations;"

# If corrupted, reset database (WARNING: deletes all data)
docker-compose -f deployments/docker/docker-compose.yml down -v
docker-compose -f deployments/docker/docker-compose.yml up -d
# Wait 10 seconds, then restart app
```

### "SMTP connection failed"
Email sending will fail, but app will continue running. To fix:
1. Use Gmail with App Password (not regular password)
2. Enable 2-factor authentication on Gmail
3. Generate App Password: https://myaccount.google.com/apppasswords
4. Update `SMTP_PASSWORD` in .env with the app password

## 📁 Project Structure

```
backend/
├── main.go                    # Entry point
├── api/                       # HTTP handlers
│   ├── auth.go               # Auth endpoints
│   └── user.go               # User endpoints
├── auth/core/                # Auth business logic
├── user/
│   ├── dao/                  # Database operations
│   └── model/                # Data models
├── internal/
│   ├── config/               # Configuration
│   ├── middleware/           # HTTP middleware
│   ├── services/             # Services (email, etc.)
│   └── utils/                # Utilities (JWT, password)
├── pkg/
│   ├── database/
│   │   └── migrations/       # SQL migration files
│   └── logger/
├── .env                      # Your configuration
└── config/config.yaml        # Default config
```

## 📚 Available Endpoints

### Public (No Auth Required)
- `POST /api/v1/auth/signup` - Register new user
- `POST /api/v1/auth/login` - Login with credentials
- `POST /api/v1/auth/verify-email` - Verify email with token
- `POST /api/v1/auth/resend-verification` - Resend verification email
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token
- `POST /api/v1/auth/refresh` - Refresh access token

### Protected (Requires JWT Token)
- `POST /api/v1/auth/logout` - Logout current session
- `POST /api/v1/auth/change-password` - Change password
- `GET /api/v1/users/me` - Get current user profile
- `PATCH /api/v1/users/me` - Update profile
- `DELETE /api/v1/users/me` - Soft delete account
- `GET /api/v1/users` - List all users (admin)
- `GET /api/v1/users/:id` - Get user by ID (admin)
- `POST /api/v1/users/:id/restore` - Restore deleted user (admin)
- `POST /api/v1/users/:id/disable` - Disable user (admin)
- `POST /api/v1/users/:id/enable` - Enable user (admin)

**For detailed API examples, see [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md)**

## 💡 Tips

1. **Use a REST client** like Postman or Insomnia for easier API testing
2. **Install air** for hot reload during development:
   ```bash
   go install github.com/air-verse/air@latest
   air
   ```
3. **Use a database GUI** like pgAdmin or DBeaver to visualize data
4. **Check logs** - All application logs print to console

## ✅ Verification Checklist

Before moving to Phase 3, verify:
- [ ] Docker Desktop is running
- [ ] PostgreSQL container is running (`docker ps`)
- [ ] Application starts without errors
- [ ] Health check returns HTTP 200
- [ ] Can register a new user
- [ ] Database tables are created (users, user_passwords)
- [ ] Can login with correct credentials
- [ ] JWT tokens are returned on login

If all checks pass, you're ready for Phase 3! 🎉

## 🎯 What's Next?

Phase 1 & 2 are complete! Ready to build quiz management features?

**See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for Phase 3 roadmap**

Phase 3 will add:
- Quiz creation and management
- Question and answer management
- Quiz publishing and sharing
- Category and tag system

Estimated time: 3-5 days

## 📖 More Documentation

- **[CURRENT_STATE.md](CURRENT_STATE.md)** - Complete architecture overview
- **[IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)** - Full roadmap (Phases 3-8)
- **[API_TESTING_GUIDE.md](API_TESTING_GUIDE.md)** - API examples and testing
- **[../README.md](../README.md)** - Main project README

## 🆘 Need Help?

1. Check [CURRENT_STATE.md](CURRENT_STATE.md) for architecture details
2. Check [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md) for API examples
3. Check Docker logs: `docker logs quiz-craft-postgres`
4. Check application logs in your terminal
5. Verify .env configuration

Happy coding! 🚀
