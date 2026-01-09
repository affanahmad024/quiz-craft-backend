# API Testing Guide - Phase 2

Complete guide to test all authentication and user management endpoints.

## Prerequisites

1. **Ensure PostgreSQL and Redis are running:**
   - PostgreSQL on port 5432
   - Redis on port 6379

2. **Run the application:**
   ```bash
   cd backend
   go run main.go
   ```

3. **Server should be running on:** `http://localhost:8080`

---

## 1. Health Check

### Check API Health
```bash
curl http://localhost:8080/api/v1/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-03T18:00:00Z",
  "checks": {
    "database": {"status": "healthy"},
    "redis": {"status": "healthy"}
  }
}
```

---

## 2. User Signup

### Register a New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "name": "John Doe"
  }'
```

**Note:** `phone_number` is optional and can be omitted.

**Expected Response:**
```json
{
  "success": true,
  "message": "Signup successful. Please check your email to verify your account.",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "email_verification_sent": true
  }
}
```

**Save the `user_id` for next steps!**

### Test Validation Errors

**Missing field:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Weak password:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "weak",
    "name": "Test User"
  }'
```

**Duplicate email:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "name": "Another John"
  }'
```

---

## 3. Email Verification

**Note:** In development, check your terminal/logs for the verification token (emails might not send if SMTP is not configured).

### Verify Email
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "YOUR_USER_ID_FROM_SIGNUP",
    "token": "YOUR_VERIFICATION_TOKEN_FROM_EMAIL"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Email verified successfully",
  "data": null
}
```

### Resend Verification Email
```bash
curl -X POST http://localhost:8080/api/v1/auth/resend-verification \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com"
  }'
```

---

## 4. Login

### Login with Valid Credentials
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'
```

**Expected Response:**
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
      "email": "john.doe@example.com",
      "name": "John Doe",
      "email_verified": true,
      "enabled": true
    }
  }
}
```

**Save the `access_token` - you'll need it for protected endpoints!**

### Test Login Failures

**Wrong password:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "WrongPassword123!"
  }'
```

**Unverified email (create a new user without verifying):**
```bash
# First create user
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "unverified@example.com",
    "password": "SecurePass123!",
    "name": "Unverified User"
  }'

# Try to login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "unverified@example.com",
    "password": "SecurePass123!"
  }'
```

---

## 4.5. Brute-Force Protection Testing 🔒

**The application includes built-in brute-force attack protection:**
- After **5 failed login attempts**, the account is **locked for 30 minutes**
- Failed attempts are tracked in the database using JSONB metadata
- Account lockout is automatic and temporary

### Test Account Lockout

**Step 1:** Create and verify a test user:
```bash
# Create user
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bruteforce.test@example.com",
    "password": "CorrectPass123!",
    "name": "Brute Force Test"
  }'

# Verify email (check logs for token)
# Then login successfully once to verify the account
```

**Step 2:** Make 5 failed login attempts (wrong password):
```bash
# Attempt 1
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bruteforce.test@example.com","password":"WrongPass1!"}'

# Attempt 2
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bruteforce.test@example.com","password":"WrongPass2!"}'

# Attempt 3
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bruteforce.test@example.com","password":"WrongPass3!"}'

# Attempt 4
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bruteforce.test@example.com","password":"WrongPass4!"}'

# Attempt 5 - This will trigger the lockout
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bruteforce.test@example.com","password":"WrongPass5!"}'
```

**Expected Response (after 5th attempt):**
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "account-locked-due-to-failed-attempts"
}
```

**Step 3:** Try logging in with the CORRECT password (should still be locked):
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bruteforce.test@example.com","password":"CorrectPass123!"}'
```

**Expected Response:**
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "account-locked"
}
```

**Step 4:** Check the database to see the lockout details:
```bash
psql -U postgres -d quiz_craft -c "SELECT user_id, meta->>'failed_login_attempts' as failed_attempts, meta->>'account_locked_until' as locked_until FROM user_passwords WHERE user_id IN (SELECT id FROM users WHERE email = 'bruteforce.test@example.com');"
```

**What you should see:**
- `failed_attempts`: 5
- `locked_until`: A timestamp 30 minutes from the 5th failed attempt

### Security Features Demonstrated

✅ **Failed Login Tracking** - Each wrong password increments a counter
✅ **Automatic Lockout** - After 5 attempts, account locks for 30 minutes
✅ **Temporal Lock** - Lock automatically expires after 30 minutes
✅ **JSONB Storage** - Security metadata stored efficiently in `meta` field
✅ **Lock Bypassing Prevention** - Even correct passwords won't work during lockout

---

## 5. Protected Endpoints (Requires Authentication)

**Set your access token as a variable:**
```bash
export ACCESS_TOKEN="your_access_token_from_login"
```

### Get Current User Profile
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "phone_number": "+1234567890",
    "email_verified": true,
    "enabled": true,
    "created_at": "2025-01-03T18:00:00Z",
    "updated_at": "2025-01-03T18:00:00Z",
    "meta": {}
  }
}
```

### Update Profile
```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated Doe",
    "phone_number": "+9876543210",
    "meta": {
      "timezone": "America/New_York",
      "theme": "dark"
    }
  }'
```

### Test Without Authentication
```bash
curl -X GET http://localhost:8080/api/v1/users/me
```

**Expected:** 401 Unauthorized

---

## 6. Token Refresh

### Refresh Access Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN_FROM_LOGIN"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "new_access_token...",
    "refresh_token": "same_refresh_token...",
    "expires_in": 86400,
    "token_type": "Bearer",
    "user": {...}
  }
}
```

---

## 7. Password Management

### Change Password (Authenticated)
```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

### Forgot Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com"
  }'
```

### Reset Password (with token from email)
```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "YOUR_USER_ID",
    "token": "RESET_TOKEN_FROM_EMAIL",
    "new_password": "ResetPass123!"
  }'
```

---

## 8. Soft Delete Operations

### Delete Your Account (Soft Delete)
```bash
curl -X DELETE http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Account deleted successfully. You can contact support within 30 days to recover your account.",
  "data": null
}
```

### Try to Login After Soft Delete
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'
```

**Expected:** User not found (soft deleted users can't login)

### Restore Account (Admin endpoint - requires admin token)
```bash
curl -X POST http://localhost:8080/api/v1/users/USER_ID/restore \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

---

## 9. Admin Endpoints

### List All Users
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Get User Statistics
```bash
curl -X GET http://localhost:8080/api/v1/users/stats \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Statistics retrieved successfully",
  "data": {
    "total_users": 10,
    "active_users": 10,
    "deleted_users": 2
  }
}
```

### Get User by ID
```bash
curl -X GET http://localhost:8080/api/v1/users/USER_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Disable User Account
```bash
curl -X POST http://localhost:8080/api/v1/users/USER_ID/disable \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Enable User Account
```bash
curl -X POST http://localhost:8080/api/v1/users/USER_ID/enable \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## 10. Logout

### Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## 🧪 Complete Test Scenario

Run these commands in sequence for a full test:

```bash
# 1. Health check
curl http://localhost:8080/api/v1/health

# 2. Signup
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"TestPass123!","name":"Test User"}'

# 3. Verify email (check logs for token)
# curl -X POST http://localhost:8080/api/v1/auth/verify-email ...

# 4. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"TestPass123!"}'

# 5. Get profile (use token from step 4)
export TOKEN="paste_your_access_token_here"
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"

# 6. Update profile
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name"}'

# 7. Change password
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"current_password":"TestPass123!","new_password":"NewTestPass123!"}'

# 8. Soft delete account
curl -X DELETE http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"

# 9. Try to login (should fail)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"NewTestPass123!"}'
```

---

## 💡 Tips

1. **Save tokens:** Store access and refresh tokens for reuse
2. **Check logs:** Application logs show detailed error messages
3. **Database:** Use pgAdmin or `psql` to inspect database state
4. **Postman:** Import these as Postman collection for easier testing
5. **Email Configuration:** Update `.env` with valid SMTP credentials to receive emails:
   ```env
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   SMTP_FROM_EMAIL=noreply@yourapp.com
   SMTP_FROM_NAME=YourApp
   ```
   **Note:** For Gmail, you need to use an [App Password](https://support.google.com/accounts/answer/185833), not your regular password.
6. **Check Verification Tokens:** If emails aren't configured, verification tokens are logged in the console

---

## 🐛 Troubleshooting

### "Authorization token not found"
- Make sure you're including the `Authorization` header
- Format: `Authorization: Bearer YOUR_TOKEN`

### "Invalid or expired token"
- Token might have expired (default: 24h for access tokens)
- Use refresh token to get a new access token

### "Email not verified"
- Check your email for verification link
- Or check application logs for the verification token
- Use the verify-email endpoint

### "User not found" after soft delete
- This is expected! Soft deleted users can't login
- Use restore endpoint to recover the account

---

## ✅ Success Checklist

### Core Authentication
- [ ] Health check returns healthy
- [ ] User signup works
- [ ] Email verification works (or tokens logged to console)
- [ ] Login returns tokens
- [ ] Protected endpoints require authentication
- [ ] Token refresh works
- [ ] Logout clears cookies

### User Management
- [ ] Profile retrieval works
- [ ] Profile update works
- [ ] Password change works
- [ ] Forgot password flow works
- [ ] Password reset with token works
- [ ] Soft delete works
- [ ] Deleted users can't login

### Security Features 🔒
- [ ] Brute-force protection: 5 failed attempts locks account
- [ ] Account lockout lasts 30 minutes
- [ ] Correct password doesn't work during lockout
- [ ] Failed attempts are tracked in database `meta` JSONB field
- [ ] Lockout automatically expires after timeout

### Database & Architecture
- [ ] JSONB `meta` field stores security metadata
- [ ] Password history tracked (prevents reuse)
- [ ] Migrations create correct schema
- [ ] All queries use proper column names

**All checked? Phase 2 with brute-force protection is working perfectly!** 🎉🔒
