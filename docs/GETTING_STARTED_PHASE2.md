# Getting Started with Phase 2

Complete guide to run and test the authentication system with soft delete.

## 🎯 What You'll Learn

- How to start the application
- How to test all authentication endpoints
- How soft delete works
- How to restore deleted accounts

---

## 📋 Prerequisites

Make sure you completed Phase 1:
- ✅ Docker Desktop installed and running
- ✅ Go 1.21+ installed
- ✅ PostgreSQL and Redis running (via Docker)

---

## 🚀 Step 1: Start Docker Services

```bash
cd quiz-craft/backend

# Start PostgreSQL and Redis
docker-compose -f deployments/docker/docker-compose.yml up -d

# Verify they're running
docker ps
```

You should see:
- `quiz-craft-postgres`
- `quiz-craft-redis`

---

## 🔧 Step 2: Configure Environment

The `.env` file was created in Phase 1. For Phase 2, verify these settings:

```bash
# Check your .env file
cat .env | grep -E "JWT|SMTP"
```

**Minimum required:**
```env
JWT_SECRET=your-super-secret-jwt-key
JWT_ACCESS_TOKEN_EXPIRE=24h
JWT_REFRESH_TOKEN_EXPIRE=168h
```

**Optional (for actual emails):**
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@quizcraft.com
SMTP_FROM_NAME=QuizCraft
```

> **Note:** Without SMTP config, verification tokens will only appear in logs (which is fine for testing!)

---

## 🏃 Step 3: Run the Application

```bash
# Option 1: Run directly
go run cmd/server/main.go

# Option 2: Build and run
go build -o bin/server.exe cmd/server/main.go
./bin/server.exe
```

**You should see:**
```
INFO[0000] Starting QuizCraft Backend API
INFO[0000] Environment: development
INFO[0000] Connecting to PostgreSQL...
INFO[0000] PostgreSQL connected successfully
INFO[0000] Running database migrations...
INFO[0000] Migrations applied successfully
INFO[0000] Connecting to Redis...
INFO[0000] Redis connected successfully
INFO[0000] Setting up routes...
INFO[0000] Server starting on 0.0.0.0:8080
```

---

## 🧪 Step 4: Test the System

### 4.1 Health Check

```bash
curl http://localhost:8080/api/v1/health
```

**Expected:** All services healthy ✅

### 4.2 Create Your First User

```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!",
    "name": "John Doe",
    "phone_number": "+1234567890"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Signup successful. Please check your email to verify your account.",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john@example.com",
    "email_verification_sent": true
  }
}
```

**Save the `user_id` - you'll need it!**

### 4.3 Get Verification Token

**If SMTP not configured**, check your terminal/logs:

```
INFO Email verification token: abc123xyz789...
```

**If SMTP configured**, check your email inbox.

### 4.4 Verify Email

```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "token": "your-verification-token-here"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Email verified successfully",
  "data": null
}
```

### 4.5 Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400,
    "token_type": "Bearer",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john@example.com",
      "name": "John Doe",
      "email_verified": true,
      "enabled": true
    }
  }
}
```

**Copy the `access_token`!**

### 4.6 Access Protected Endpoint

```bash
# Set your token
export TOKEN="paste_your_access_token_here"

# Get your profile
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john@example.com",
    "name": "John Doe",
    "phone_number": "+1234567890",
    "email_verified": true,
    "enabled": true,
    "created_at": "2025-01-03T18:00:00Z",
    "updated_at": "2025-01-03T18:00:00Z"
  }
}
```

✅ **Authentication working!**

---

## 🗑️ Step 5: Test Soft Delete

### 5.1 Delete Your Account

```bash
curl -X DELETE http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Account deleted successfully. You can contact support within 30 days to recover your account.",
  "data": null
}
```

### 5.2 Verify You Can't Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "success": false,
  "message": "Login failed",
  "error": "user not found"
}
```

✅ **Soft delete working - user can't login!**

### 5.3 Check Database (Optional)

```bash
# Connect to PostgreSQL
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft

# Check users table
SELECT id, email, name, deleted_at FROM users;
```

You'll see `deleted_at` has a timestamp! 📅

```
                  id                  |       email        |    name    |         deleted_at
--------------------------------------+--------------------+------------+----------------------------
 550e8400-e29b-41d4-a716-446655440000 | john@example.com   | John Doe   | 2025-01-03 18:30:00.123456
```

### 5.4 Restore Account

To restore, you need admin access. For testing, you can call the endpoint with any valid token:

```bash
# Save the user ID from earlier
export USER_ID="550e8400-e29b-41d4-a716-446655440000"

# Restore (using same token for demo)
curl -X POST http://localhost:8080/api/v1/users/$USER_ID/restore \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Account restored successfully",
  "data": null
}
```

### 5.5 Login Again

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

✅ **Login works again - account restored!**

---

## 🎯 Step 6: Test Other Features

### Update Profile

```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "meta": {
      "timezone": "America/New_York",
      "theme": "dark"
    }
  }'
```

### Change Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

### Forgot Password Flow

```bash
# 1. Request reset
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com"}'

# 2. Check logs for reset token
# 3. Reset password
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "YOUR_USER_ID",
    "token": "RESET_TOKEN_FROM_EMAIL",
    "new_password": "AnotherSecure123!"
  }'
```

### Get User Statistics

```bash
curl -X GET http://localhost:8080/api/v1/users/stats \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📊 Understanding Soft Delete

### What Happens When You Delete?

1. **Soft Delete (DELETE /users/me)**:
   - Sets `deleted_at = NOW()`
   - Record stays in database
   - GORM automatically excludes from queries
   - User can't login
   - Email remains reserved

2. **Restore (POST /users/:id/restore)**:
   - Sets `deleted_at = NULL`
   - User visible again in queries
   - Can login again
   - All data intact

3. **Hard Delete (Admin only)**:
   - Permanently removes record
   - **Cannot be undone**
   - Use with extreme caution

### Querying in Code

```go
// Normal query - excludes deleted
db.Find(&users)  // deleted_at IS NULL

// Include deleted
db.Unscoped().Find(&users)  // All records

// Only deleted
db.Unscoped().Where("deleted_at IS NOT NULL").Find(&users)
```

---

## 🐛 Troubleshooting

### "Failed to connect to database"
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs quiz-craft-postgres

# Restart if needed
docker-compose -f deployments/docker/docker-compose.yml restart postgres
```

### "Authorization token not found"
- Make sure you're using: `Authorization: Bearer YOUR_TOKEN`
- Check token is not expired (24h default)
- Use refresh token to get new access token

### "Email not verified"
- Check application logs for verification token
- Or configure SMTP to receive real emails
- Use `/auth/resend-verification` to get new token

### "Password too weak"
Requirements:
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

---

## ✅ Success Checklist

After following this guide, you should have:

- [x] Application running successfully
- [x] Created a user account
- [x] Verified email
- [x] Logged in and received JWT tokens
- [x] Accessed protected endpoints
- [x] Updated profile
- [x] Soft deleted account
- [x] Verified deleted user can't login
- [x] Restored deleted account
- [x] Logged in again after restore

**All checked? Phase 2 is working perfectly!** 🎉

---

## 📚 Next Steps

1. **Test all endpoints**: See [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md)
2. **Review code**: Check out the implementation
3. **Configure SMTP**: Set up real email sending
4. **Ready for Phase 3**: Quiz management coming next!

---

## 🆘 Need More Help?

- **API Reference**: [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md)
- **Phase 2 Summary**: [PHASE2_COMPLETE.md](PHASE2_COMPLETE.md)
- **Quick Reference**: [PHASE2_SUMMARY.md](PHASE2_SUMMARY.md)
- **Main README**: [README.md](README.md)

---

## 🎓 What You Learned

- ✅ Running a Go backend application
- ✅ Testing REST APIs with curl
- ✅ JWT authentication flow
- ✅ Soft delete implementation
- ✅ Account restoration
- ✅ Password management
- ✅ Protected vs public endpoints

**Congratulations on completing Phase 2!** 🚀
