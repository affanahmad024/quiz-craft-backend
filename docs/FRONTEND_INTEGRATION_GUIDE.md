# QuizCraft Backend API Integration Guide
## Complete API Reference for Frontend Integration

**Version**: 1.0
**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: Bearer Token (JWT)
**Last Updated**: January 10, 2026

---

## Table of Contents

1. [API Overview](#api-overview)
2. [Authentication](#authentication)
3. [Authentication APIs](#authentication-apis)
4. [User Management APIs](#user-management-apis)
5. [Quiz Management APIs](#quiz-management-apis)
6. [Question & Answer APIs](#question--answer-apis)
7. [Quiz Attempt APIs](#quiz-attempt-apis)
8. [Search & Discovery APIs](#search--discovery-apis)
9. [Statistics APIs](#statistics-apis)
10. [Template APIs](#template-apis)
11. [Sharing APIs](#sharing-apis)
12. [Admin APIs](#admin-apis)
13. [File Upload APIs](#file-upload-apis)
14. [AI Generation APIs](#ai-generation-apis)
15. [Error Handling](#error-handling)
16. [Rate Limiting](#rate-limiting)
17. [WebSocket Support](#websocket-support)

---

## API Overview

### Base URL
```
Development: http://localhost:8080/api/v1
Production: https://api.quizcraft.com/api/v1
```

### Authentication Header
```
Authorization: Bearer {access_token}
```

### Response Format
All successful responses return JSON:
```json
{
  "data": { ... }  // or array for lists
}
```

All error responses return:
```json
{
  "error": {
    "code": "error-code",
    "message": "Human readable message",
    "details": { ... }  // optional
  }
}
```

### HTTP Status Codes
- `200 OK` - Successful GET, PATCH, POST (non-creation)
- `201 Created` - Successful POST (creation)
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Validation error
- `401 Unauthorized` - Not authenticated
- `403 Forbidden` - Not authorized
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., email exists)
- `422 Unprocessable Entity` - Invalid data format
- `500 Internal Server Error` - Server error

---

## Authentication

### Overview
QuizCraft uses JWT (JSON Web Tokens) for authentication. After login, you receive:
- **Access Token**: Short-lived token (24 hours) for API requests
- **Refresh Token**: Long-lived token (30 days) to get new access tokens

### Token Storage
**Recommended**: Store in httpOnly cookies (if you control the frontend server)
**Alternative**: Store in localStorage (easier but less secure)

### Token Usage
Include in Authorization header for all protected endpoints:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token Refresh Flow
When access token expires (401 error):
1. Call `/auth/refresh` with refresh token
2. Get new access token
3. Retry original request with new token

---

## Authentication APIs

### 1. Signup

**Endpoint**: `POST /auth/signup`
**Auth Required**: No
**Description**: Register a new user account

**Request Body**:
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "SecurePass123!",
  "phone_number": "+1234567890"  // optional
}
```

**Validation Rules**:
- `email`: Valid email format, required
- `name`: 2-255 characters, required
- `password`: Minimum 8 characters, required
- `phone_number`: Optional

**Success Response** (201 Created):
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "email_verification_sent": true
}
```

**Error Responses**:
- `400` - Validation error (weak password, invalid email)
- `409` - Email already exists

**Example cURL**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123!"
  }'
```

---

### 2. Verify Email

**Endpoint**: `POST /auth/verify-email`
**Auth Required**: No
**Description**: Verify user email with token received via email

**Request Body**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "verification-token-from-email"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Email verified successfully"
}
```

**Error Responses**:
- `400` - Invalid or expired token

---

### 3. Resend Verification Email

**Endpoint**: `POST /auth/resend-verification`
**Auth Required**: No
**Description**: Resend verification email

**Request Body**:
```json
{
  "email": "user@example.com"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Verification email sent"
}
```

**Error Responses**:
- `400` - Email already verified

---

### 4. Login

**Endpoint**: `POST /auth/login`
**Auth Required**: No
**Description**: Login and receive access/refresh tokens

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Success Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "token_type": "Bearer",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "phone_number": "+1234567890",
    "role": "user",
    "email_verified": true,
    "enabled": true,
    "created_at": "2026-01-10T12:00:00Z",
    "updated_at": "2026-01-10T12:00:00Z"
  }
}
```

**Error Responses**:
- `401` - Invalid credentials
- `401` - Email not verified
- `401` - Account locked (too many failed attempts)
- `401` - Account disabled

---

### 5. Refresh Token

**Endpoint**: `POST /auth/refresh`
**Auth Required**: No (uses refresh token)
**Description**: Get new access token using refresh token

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response** (200 OK):
```json
{
  "access_token": "new-access-token",
  "refresh_token": "same-refresh-token",
  "expires_in": 86400,
  "token_type": "Bearer",
  "user": { ... }
}
```

**Error Responses**:
- `401` - Invalid or expired refresh token

---

### 6. Forgot Password

**Endpoint**: `POST /auth/forgot-password`
**Auth Required**: No
**Description**: Request password reset email

**Request Body**:
```json
{
  "email": "user@example.com"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password reset email sent if account exists"
}
```

**Note**: Always returns success to prevent email enumeration

---

### 7. Reset Password

**Endpoint**: `POST /auth/reset-password`
**Auth Required**: No
**Description**: Reset password with token from email

**Request Body**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "reset-token-from-email",
  "new_password": "NewSecurePass456!"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password reset successfully"
}
```

**Error Responses**:
- `400` - Invalid or expired token
- `400` - Password too weak
- `400` - Password recently used (password history check)

---

### 8. Change Password (Authenticated)

**Endpoint**: `POST /auth/change-password`
**Auth Required**: Yes
**Description**: Change password when logged in

**Request Body**:
```json
{
  "current_password": "SecurePass123!",
  "new_password": "NewSecurePass456!"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password changed successfully"
}
```

**Error Responses**:
- `401` - Invalid current password
- `400` - Password recently used

---

### 9. Logout

**Endpoint**: `POST /auth/logout`
**Auth Required**: Yes
**Description**: Logout and invalidate tokens

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

---

## User Management APIs

### 10. Get Current User Profile

**Endpoint**: `GET /users/me`
**Auth Required**: Yes
**Description**: Get current user's profile

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "phone_number": "+1234567890",
  "role": "user",
  "email_verified": true,
  "enabled": true,
  "created_at": "2026-01-10T12:00:00Z",
  "updated_at": "2026-01-10T12:00:00Z",
  "last_login_at": "2026-01-10T14:30:00Z",
  "meta": {}
}
```

---

### 11. Update Profile

**Endpoint**: `PATCH /users/me`
**Auth Required**: Yes
**Description**: Update current user's profile

**Request Body**:
```json
{
  "name": "John Updated Doe",
  "phone_number": "+9876543210",
  "meta": {
    "bio": "Quiz enthusiast",
    "preferences": {
      "theme": "dark"
    }
  }
}
```

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Updated Doe",
  "phone_number": "+9876543210",
  ...
}
```

---

### 12. Delete Account

**Endpoint**: `DELETE /users/me`
**Auth Required**: Yes
**Description**: Soft delete current user's account

**Success Response** (204 No Content)

---

### 13. Get User Statistics

**Endpoint**: `GET /users/me/statistics`
**Auth Required**: Yes
**Description**: Get current user's statistics

**Success Response** (200 OK):
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "user@example.com",
  "quizzes_created": 15,
  "published_quizzes": 12,
  "total_quiz_views": 450,
  "total_quiz_attempts": 230,
  "quizzes_taken": 50,
  "attempts_completed": 45,
  "average_score": 78.5,
  "total_points_earned": 3525.0,
  "quizzes_passed": 38,
  "pass_rate": 84.4,
  "current_streak": 5,
  "longest_streak": 12,
  "last_activity_at": "2026-01-10T14:30:00Z",
  "joined_at": "2025-12-01T10:00:00Z"
}
```

---

## Quiz Management APIs

### 14. Create Quiz

**Endpoint**: `POST /quizzes`
**Auth Required**: Yes
**Description**: Create a new quiz

**Request Body**:
```json
{
  "title": "JavaScript Fundamentals",
  "description": "Test your knowledge of JavaScript basics",
  "category": "Programming",
  "difficulty": "medium",
  "time_limit_minutes": 30,
  "passing_score": 70,
  "is_public": true,
  "tags": ["javascript", "programming", "web"],
  "meta": {}
}
```

**Validation Rules**:
- `title`: 3-255 characters, required
- `description`: Max 2000 characters
- `difficulty`: One of: "easy", "medium", "hard"
- `time_limit_minutes`: 1-300 minutes
- `passing_score`: 0-100
- `category`: Max 100 characters

**Success Response** (201 Created):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "title": "JavaScript Fundamentals",
  "description": "Test your knowledge of JavaScript basics",
  "creator_id": "550e8400-e29b-41d4-a716-446655440000",
  "category": "Programming",
  "difficulty": "medium",
  "time_limit_minutes": 30,
  "passing_score": 70,
  "is_published": false,
  "is_public": true,
  "thumbnail_url": "",
  "tags": ["javascript", "programming", "web"],
  "share_code": null,
  "share_count": 0,
  "view_count": 0,
  "is_template": false,
  "template_usage_count": 0,
  "based_on_template_id": null,
  "meta": {},
  "created_at": "2026-01-10T12:00:00Z",
  "updated_at": "2026-01-10T12:00:00Z"
}
```

---

### 15. Get Quiz by ID

**Endpoint**: `GET /quizzes/:id`
**Auth Required**: Yes
**Description**: Get quiz details (without questions)

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "title": "JavaScript Fundamentals",
  ...
}
```

**Error Responses**:
- `404` - Quiz not found
- `403` - Access denied (private quiz not owned by user)

---

### 16. Get Quiz with Questions

**Endpoint**: `GET /quizzes/:id/full`
**Auth Required**: Yes
**Description**: Get quiz with all questions and answers

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "title": "JavaScript Fundamentals",
  "description": "Test your knowledge of JavaScript basics",
  ...,
  "questions": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
      "question_text": "What is the output of: console.log(typeof null)?",
      "question_type": "multiple_choice",
      "points": 10,
      "order_index": 0,
      "explanation": "typeof null returns 'object' due to a legacy bug",
      "meta": {},
      "created_at": "2026-01-10T12:05:00Z",
      "updated_at": "2026-01-10T12:05:00Z",
      "answers": [
        {
          "id": "880e8400-e29b-41d4-a716-446655440000",
          "question_id": "770e8400-e29b-41d4-a716-446655440000",
          "answer_text": "null",
          "is_correct": false,
          "order_index": 0,
          "created_at": "2026-01-10T12:05:00Z",
          "updated_at": "2026-01-10T12:05:00Z"
        },
        {
          "id": "990e8400-e29b-41d4-a716-446655440000",
          "question_id": "770e8400-e29b-41d4-a716-446655440000",
          "answer_text": "object",
          "is_correct": true,
          "order_index": 1,
          "created_at": "2026-01-10T12:05:00Z",
          "updated_at": "2026-01-10T12:05:00Z"
        },
        ...
      ]
    },
    ...
  ]
}
```

**Note**: For quiz creators, `is_correct` is shown. For quiz takers, `is_correct` is hidden until attempt is finished.

---

### 17. List My Quizzes

**Endpoint**: `GET /quizzes`
**Auth Required**: Yes
**Description**: List current user's quizzes

**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

**Success Response** (200 OK):
```json
{
  "quizzes": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "title": "JavaScript Fundamentals",
      "description": "Test your knowledge of JavaScript basics",
      "creator_id": "550e8400-e29b-41d4-a716-446655440000",
      "category": "Programming",
      "difficulty": "medium",
      "is_published": true,
      "is_public": true,
      "thumbnail_url": "",
      "tags": ["javascript"],
      "created_at": "2026-01-10T12:00:00Z",
      "question_count": 10
    },
    ...
  ],
  "total": 15,
  "page": 1,
  "limit": 20,
  "total_pages": 1
}
```

---

### 18. List Public Quizzes

**Endpoint**: `GET /quizzes/public`
**Auth Required**: Yes
**Description**: List all published public quizzes

**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

**Success Response** (200 OK): Same format as "List My Quizzes"

---

### 19. Update Quiz

**Endpoint**: `PATCH /quizzes/:id`
**Auth Required**: Yes (Creator only)
**Description**: Update quiz details

**URL Parameters**:
- `id`: Quiz UUID

**Request Body** (all fields optional):
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "category": "Updated Category",
  "difficulty": "hard",
  "time_limit_minutes": 45,
  "passing_score": 80,
  "is_public": false,
  "tags": ["updated", "tags"]
}
```

**Success Response** (200 OK): Updated quiz object

**Error Responses**:
- `403` - Not the quiz creator
- `400` - Cannot update published quiz (certain fields)

---

### 20. Delete Quiz

**Endpoint**: `DELETE /quizzes/:id`
**Auth Required**: Yes (Creator only)
**Description**: Soft delete quiz

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (204 No Content)

**Error Responses**:
- `403` - Not the quiz creator

---

### 21. Publish Quiz

**Endpoint**: `POST /quizzes/:id/publish`
**Auth Required**: Yes (Creator only)
**Description**: Publish quiz (make it available for attempts)

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "title": "JavaScript Fundamentals",
  "is_published": true,
  ...
}
```

**Error Responses**:
- `403` - Not the quiz creator
- `400` - Quiz must have at least one question

---

### 22. Unpublish Quiz

**Endpoint**: `POST /quizzes/:id/unpublish`
**Auth Required**: Yes (Creator only)
**Description**: Unpublish quiz

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK): Quiz object with `is_published: false`

---

## Question & Answer APIs

### 23. Add Question

**Endpoint**: `POST /quizzes/:id/questions`
**Auth Required**: Yes (Creator only)
**Description**: Add question to quiz

**URL Parameters**:
- `id`: Quiz UUID

**Request Body**:
```json
{
  "question_text": "What is the output of: console.log(typeof null)?",
  "question_type": "multiple_choice",
  "points": 10,
  "order_index": 0,
  "explanation": "typeof null returns 'object' due to a legacy bug in JavaScript",
  "answers": [
    {
      "answer_text": "null",
      "is_correct": false,
      "order_index": 0
    },
    {
      "answer_text": "object",
      "is_correct": true,
      "order_index": 1
    },
    {
      "answer_text": "undefined",
      "is_correct": false,
      "order_index": 2
    },
    {
      "answer_text": "number",
      "is_correct": false,
      "order_index": 3
    }
  ]
}
```

**Validation Rules**:
- `question_text`: Required, max 1000 characters
- `question_type`: One of: "multiple_choice", "true_false", "short_answer"
- `points`: 1-100
- `order_index`: Non-negative integer
- `explanation`: Max 500 characters
- `answers`:
  - Multiple choice: Minimum 2 answers
  - True/False: Exactly 2 answers
  - Short answer: 0 answers (text answer)
  - At least one answer must be correct

**Success Response** (201 Created):
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "question_text": "What is the output of: console.log(typeof null)?",
  "question_type": "multiple_choice",
  "points": 10,
  "order_index": 0,
  "explanation": "typeof null returns 'object' due to a legacy bug",
  "meta": {},
  "created_at": "2026-01-10T12:05:00Z",
  "updated_at": "2026-01-10T12:05:00Z",
  "answers": [ ... ]
}
```

---

### 24. Get Question

**Endpoint**: `GET /quizzes/questions/:qid`
**Auth Required**: Yes
**Description**: Get question with answers

**URL Parameters**:
- `qid`: Question UUID

**Success Response** (200 OK): Question object with answers

---

### 25. Update Question

**Endpoint**: `PATCH /quizzes/questions/:qid`
**Auth Required**: Yes (Quiz creator only)
**Description**: Update question

**URL Parameters**:
- `qid`: Question UUID

**Request Body** (all fields optional):
```json
{
  "question_text": "Updated question text",
  "points": 15,
  "explanation": "Updated explanation"
}
```

**Success Response** (200 OK): Updated question object

---

### 26. Delete Question

**Endpoint**: `DELETE /quizzes/questions/:qid`
**Auth Required**: Yes (Quiz creator only)
**Description**: Delete question

**URL Parameters**:
- `qid`: Question UUID

**Success Response** (204 No Content)

---

### 27. Add Answer

**Endpoint**: `POST /quizzes/questions/:qid/answers`
**Auth Required**: Yes (Quiz creator only)
**Description**: Add answer to question

**URL Parameters**:
- `qid`: Question UUID

**Request Body**:
```json
{
  "answer_text": "New answer option",
  "is_correct": false,
  "order_index": 4
}
```

**Success Response** (201 Created): Answer object

---

### 28. Update Answer

**Endpoint**: `PATCH /quizzes/answers/:aid`
**Auth Required**: Yes (Quiz creator only)
**Description**: Update answer

**URL Parameters**:
- `aid`: Answer UUID

**Request Body**:
```json
{
  "answer_text": "Updated answer text",
  "is_correct": true
}
```

**Success Response** (200 OK): Updated answer object

---

### 29. Delete Answer

**Endpoint**: `DELETE /quizzes/answers/:aid`
**Auth Required**: Yes (Quiz creator only)
**Description**: Delete answer

**URL Parameters**:
- `aid`: Answer UUID

**Success Response** (204 No Content)

---

## Quiz Attempt APIs

### 30. Start Quiz Attempt

**Endpoint**: `POST /quizzes/:id/attempts`
**Auth Required**: Yes
**Description**: Start a new quiz attempt

**URL Parameters**:
- `id`: Quiz UUID

**Request Body**:
```json
{
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000"
}
```

**Success Response** (201 Created):
```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440000",
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "in_progress",
  "started_at": "2026-01-10T14:00:00Z",
  "submitted_at": null,
  "time_limit_seconds": 1800,
  "expires_at": "2026-01-10T14:30:00Z",
  "score": null,
  "percentage": null,
  "passed": null,
  "max_score": 100,
  "meta": {},
  "created_at": "2026-01-10T14:00:00Z",
  "updated_at": "2026-01-10T14:00:00Z"
}
```

**Note**:
- `time_limit_seconds` and `expires_at` only present if quiz has time limit
- `max_score` is sum of all question points
- User can have multiple attempts for same quiz

---

### 31. Get Attempt

**Endpoint**: `GET /attempts/:id`
**Auth Required**: Yes (Attempt owner only)
**Description**: Get attempt details

**URL Parameters**:
- `id`: Attempt UUID

**Success Response** (200 OK): Attempt object

---

### 32. Submit Answer

**Endpoint**: `POST /attempts/:id/answers`
**Auth Required**: Yes (Attempt owner only)
**Description**: Submit/update answer for a question

**URL Parameters**:
- `id`: Attempt UUID

**Request Body**:
```json
{
  "question_id": "770e8400-e29b-41d4-a716-446655440000",
  "answer_id": "990e8400-e29b-41d4-a716-446655440000"
}
```

For short answer questions:
```json
{
  "question_id": "770e8400-e29b-41d4-a716-446655440000",
  "answer_text": "User's text answer"
}
```

**Success Response** (201 Created):
```json
{
  "id": "bb0e8400-e29b-41d4-a716-446655440000",
  "attempt_id": "aa0e8400-e29b-41d4-a716-446655440000",
  "question_id": "770e8400-e29b-41d4-a716-446655440000",
  "answer_id": "990e8400-e29b-41d4-a716-446655440000",
  "answer_text": null,
  "is_correct": true,
  "points_earned": 10.0,
  "answered_at": "2026-01-10T14:05:00Z",
  "meta": {}
}
```

**Note**:
- Progressive submission: Can update answer by submitting again for same question (upsert)
- `is_correct` and `points_earned` calculated immediately for MC/TF
- Short answer questions require manual grading (Phase 5: not supported)

**Error Responses**:
- `400` - Attempt expired
- `400` - Attempt already finished

---

### 33. Finish Attempt

**Endpoint**: `POST /attempts/:id/finish`
**Auth Required**: Yes (Attempt owner only)
**Description**: Finish attempt and calculate final score

**URL Parameters**:
- `id`: Attempt UUID

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440000",
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "submitted",
  "started_at": "2026-01-10T14:00:00Z",
  "submitted_at": "2026-01-10T14:25:00Z",
  "time_limit_seconds": 1800,
  "expires_at": "2026-01-10T14:30:00Z",
  "score": 85.0,
  "percentage": 85.0,
  "passed": true,
  "max_score": 100,
  "meta": {},
  "created_at": "2026-01-10T14:00:00Z",
  "updated_at": "2026-01-10T14:25:00Z"
}
```

**Calculation**:
- `score` = sum of all `points_earned`
- `percentage` = (score / max_score) * 100
- `passed` = percentage >= passing_score

---

### 34. Get Attempt Results

**Endpoint**: `GET /attempts/:id/results`
**Auth Required**: Yes (Attempt owner or quiz creator)
**Description**: Get full attempt results with correct answers

**URL Parameters**:
- `id`: Attempt UUID

**Success Response** (200 OK):
```json
{
  "attempt": {
    "id": "aa0e8400-e29b-41d4-a716-446655440000",
    "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "submitted",
    "score": 85.0,
    "percentage": 85.0,
    "passed": true,
    ...
  },
  "quiz": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "title": "JavaScript Fundamentals",
    ...
  },
  "questions": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "question_text": "What is the output of: console.log(typeof null)?",
      "question_type": "multiple_choice",
      "points": 10,
      "explanation": "typeof null returns 'object' due to a legacy bug",
      "answers": [
        {
          "id": "880e8400-e29b-41d4-a716-446655440000",
          "answer_text": "null",
          "is_correct": false,
          ...
        },
        {
          "id": "990e8400-e29b-41d4-a716-446655440000",
          "answer_text": "object",
          "is_correct": true,
          ...
        },
        ...
      ]
    },
    ...
  ],
  "user_answers": [
    {
      "id": "bb0e8400-e29b-41d4-a716-446655440000",
      "attempt_id": "aa0e8400-e29b-41d4-a716-446655440000",
      "question_id": "770e8400-e29b-41d4-a716-446655440000",
      "answer_id": "990e8400-e29b-41d4-a716-446655440000",
      "is_correct": true,
      "points_earned": 10.0,
      "answered_at": "2026-01-10T14:05:00Z"
    },
    ...
  ]
}
```

**Note**: Only available after attempt is finished

---

### 35. List My Attempts

**Endpoint**: `GET /attempts`
**Auth Required**: Yes
**Description**: List current user's quiz attempts

**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)
- `quiz_id`: Filter by quiz (optional)
- `status`: Filter by status: "in_progress", "submitted", "expired" (optional)

**Success Response** (200 OK):
```json
{
  "attempts": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440000",
      "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
      "quiz_title": "JavaScript Fundamentals",
      "status": "submitted",
      "score": 85.0,
      "percentage": 85.0,
      "passed": true,
      "started_at": "2026-01-10T14:00:00Z",
      "submitted_at": "2026-01-10T14:25:00Z"
    },
    ...
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

---

## Search & Discovery APIs

### 36. Search Quizzes

**Endpoint**: `GET /quizzes/search`
**Auth Required**: No
**Description**: Search and filter public quizzes

**Query Parameters**:
- `q`: Search query (searches in title and description)
- `category`: Filter by category
- `difficulty`: Filter by difficulty ("easy", "medium", "hard")
- `tags`: Comma-separated tags
- `creator_id`: Filter by creator
- `is_template`: Filter templates (true/false)
- `sort_by`: Sort by "newest", "popular", "most_attempted"
- `sort_order`: "asc" or "desc"
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)

**Example**:
```
GET /quizzes/search?q=javascript&category=Programming&difficulty=medium&sort_by=popular&page=1&limit=20
```

**Success Response** (200 OK):
```json
{
  "quizzes": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "title": "JavaScript Fundamentals",
      "description": "Test your knowledge of JavaScript basics",
      "category": "Programming",
      "difficulty": "medium",
      "question_count": 10,
      "attempt_count": 450,
      "average_score": 75.5,
      "creator": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "John Doe"
      },
      "tags": ["javascript", "programming"],
      "thumbnail_url": "",
      "view_count": 1250,
      "is_template": false,
      "template_usage_count": 0,
      "created_at": "2026-01-10T12:00:00Z"
    },
    ...
  ],
  "total": 150,
  "page": 1,
  "limit": 20,
  "total_pages": 8,
  "facets": {
    "categories": [
      { "value": "Programming", "count": 50 },
      { "value": "Math", "count": 30 },
      ...
    ],
    "difficulties": [
      { "value": "easy", "count": 40 },
      { "value": "medium", "count": 60 },
      { "value": "hard", "count": 50 }
    ],
    "tags": [
      { "value": "javascript", "count": 25 },
      { "value": "python", "count": 20 },
      ...
    ]
  }
}
```

---

### 37. Get Categories

**Endpoint**: `GET /categories`
**Auth Required**: No
**Description**: Get list of all quiz categories

**Success Response** (200 OK):
```json
{
  "categories": [
    "Programming",
    "Mathematics",
    "Science",
    "History",
    "Geography",
    ...
  ]
}
```

---

### 38. Get Popular Tags

**Endpoint**: `GET /tags`
**Auth Required**: No
**Description**: Get popular tags with usage counts

**Success Response** (200 OK):
```json
{
  "tags": [
    { "tag": "javascript", "count": 125 },
    { "tag": "python", "count": 98 },
    { "tag": "algebra", "count": 87 },
    ...
  ]
}
```

---

### 39. Get Quiz by Share Code

**Endpoint**: `GET /share/:code`
**Auth Required**: No
**Description**: Get quiz by share code

**URL Parameters**:
- `code`: Share code (e.g., "ABC123XYZ")

**Success Response** (200 OK): Quiz object

---

## Statistics APIs

### 40. Get Quiz Stats (Creator)

**Endpoint**: `GET /quizzes/:id/stats`
**Auth Required**: Yes (Creator only)
**Description**: Get basic quiz statistics

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK):
```json
{
  "total_attempts": 230,
  "unique_users": 185,
  "average_score": 75.5,
  "pass_rate": 68.7
}
```

---

### 41. Get Quiz Statistics (Detailed)

**Endpoint**: `GET /quizzes/:id/statistics`
**Auth Required**: Yes (Creator only)
**Description**: Get detailed quiz statistics

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK):
```json
{
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "title": "JavaScript Fundamentals",
  "creator_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_attempts": 230,
  "unique_users": 185,
  "completed_attempts": 215,
  "in_progress_attempts": 10,
  "abandoned_attempts": 3,
  "expired_attempts": 2,
  "average_score": 75.5,
  "highest_score": 100.0,
  "lowest_score": 25.0,
  "median_score": 78.0,
  "pass_rate": 68.7,
  "average_time_seconds": 1245.5,
  "fastest_time_seconds": 580.0,
  "view_count": 1250,
  "share_count": 45,
  "question_stats": [
    {
      "question_id": "770e8400-e29b-41d4-a716-446655440000",
      "question_text": "What is the output of: console.log(typeof null)?",
      "times_answered": 215,
      "correct_answers": 178,
      "incorrect_answers": 37,
      "success_rate": 82.8
    },
    ...
  ]
}
```

---

## Template APIs

### 42. List Templates

**Endpoint**: `GET /templates`
**Auth Required**: Yes
**Description**: List available quiz templates

**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)

**Success Response** (200 OK): Same format as quiz list

---

### 43. Save Quiz as Template

**Endpoint**: `POST /quizzes/:id/save-as-template`
**Auth Required**: Yes (Creator only)
**Description**: Save quiz as reusable template

**URL Parameters**:
- `id`: Quiz UUID

**Request Body**:
```json
{
  "template_name": "JavaScript Basics Template",
  "template_description": "A comprehensive JavaScript fundamentals quiz",
  "make_public": true
}
```

**Success Response** (200 OK): Quiz object with `is_template: true`

---

### 44. Create Quiz from Template

**Endpoint**: `POST /templates/:id/use`
**Auth Required**: Yes
**Description**: Create new quiz from template

**URL Parameters**:
- `id`: Template UUID

**Request Body**:
```json
{
  "title": "My JavaScript Quiz"
}
```

**Success Response** (201 Created): New quiz object with `based_on_template_id` set

**Note**: Copies all questions and answers from template

---

## Sharing APIs

### 45. Generate Share Link

**Endpoint**: `POST /quizzes/:id/share`
**Auth Required**: Yes (Creator only)
**Description**: Generate shareable link for quiz

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK):
```json
{
  "share_code": "ABC123XYZ",
  "share_url": "https://quizcraft.com/quiz/ABC123XYZ"
}
```

**Note**:
- Share code is permanent (no expiration)
- Increments quiz `share_count`
- Anyone with link can access quiz

---

## Admin APIs

**Note**: All admin endpoints require `role: "admin"` and return `403` for non-admin users.

### 46. List All Users (Admin)

**Endpoint**: `GET /admin/users`
**Auth Required**: Yes (Admin only)
**Description**: List all users with filters

**Query Parameters**:
- `page`: Page number
- `limit`: Items per page
- `role`: Filter by role ("user", "admin")
- `enabled`: Filter by enabled status (true/false)
- `email_verified`: Filter by verification status (true/false)
- `search`: Search in name/email
- `sort_by`: Sort by field
- `sort_order`: "asc" or "desc"

**Success Response** (200 OK):
```json
{
  "users": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user",
      "email_verified": true,
      "enabled": true,
      "created_at": "2026-01-10T12:00:00Z",
      "last_login_at": "2026-01-10T14:30:00Z",
      "quizzes_created": 15,
      "attempts_made": 50,
      "last_activity": "2026-01-10T14:30:00Z"
    },
    ...
  ],
  "total": 1250,
  "page": 1,
  "limit": 20
}
```

---

### 47. Get User Details (Admin)

**Endpoint**: `GET /admin/users/:id`
**Auth Required**: Yes (Admin only)
**Description**: Get detailed user information

**URL Parameters**:
- `id`: User UUID

**Success Response** (200 OK): Detailed user object with stats

---

### 48. Change User Role (Admin)

**Endpoint**: `PATCH /admin/users/:id/role`
**Auth Required**: Yes (Admin only)
**Description**: Change user role

**URL Parameters**:
- `id`: User UUID

**Request Body**:
```json
{
  "role": "admin"
}
```

**Success Response** (200 OK): Updated user object

**Note**: Activity is logged

---

### 49. Disable User (Admin)

**Endpoint**: `POST /admin/users/:id/disable`
**Auth Required**: Yes (Admin only)
**Description**: Disable user account

**URL Parameters**:
- `id`: User UUID

**Request Body**:
```json
{
  "reason": "Violation of terms of service"
}
```

**Success Response** (200 OK): User object with `enabled: false`

**Note**: User cannot login when disabled

---

### 50. Enable User (Admin)

**Endpoint**: `POST /admin/users/:id/enable`
**Auth Required**: Yes (Admin only)
**Description**: Enable disabled user account

**URL Parameters**:
- `id`: User UUID

**Success Response** (200 OK): User object with `enabled: true`

---

### 51. Delete User (Admin)

**Endpoint**: `DELETE /admin/users/:id`
**Auth Required**: Yes (Admin only)
**Description**: Permanently delete user (hard delete)

**URL Parameters**:
- `id`: User UUID

**Request Body** (optional):
```json
{
  "reason": "User requested account deletion",
  "permanent": true
}
```

**Success Response** (204 No Content)

**Warning**: This is irreversible

---

### 52. List All Quizzes (Admin)

**Endpoint**: `GET /admin/quizzes`
**Auth Required**: Yes (Admin only)
**Description**: List all quizzes with filters

**Query Parameters**: Same as user quiz list plus:
- `is_published`: Filter by published status
- `creator_id`: Filter by creator

**Success Response** (200 OK): Quiz list with admin details

---

### 53. Get Quiz Details (Admin)

**Endpoint**: `GET /admin/quizzes/:id`
**Auth Required**: Yes (Admin only)
**Description**: Get detailed quiz information

**URL Parameters**:
- `id`: Quiz UUID

**Success Response** (200 OK): Full quiz details including creator info

---

### 54. Delete Quiz (Admin)

**Endpoint**: `DELETE /admin/quizzes/:id`
**Auth Required**: Yes (Admin only)
**Description**: Delete quiz (soft or hard delete)

**URL Parameters**:
- `id`: Quiz UUID

**Request Body**:
```json
{
  "reason": "Inappropriate content",
  "permanent": false
}
```

**Success Response** (204 No Content)

---

### 55. Get Platform Statistics (Admin)

**Endpoint**: `GET /admin/stats`
**Auth Required**: Yes (Admin only)
**Description**: Get platform-wide statistics

**Success Response** (200 OK):
```json
{
  "total_users": 1250,
  "active_users": 890,
  "new_users_today": 15,
  "total_quizzes": 3450,
  "published_quizzes": 2890,
  "total_attempts": 45670,
  "completed_attempts": 42100,
  "average_completion_rate": 92.2,
  "generated_at": "2026-01-10T15:00:00Z"
}
```

---

### 56. Get User Activity Timeline (Admin)

**Endpoint**: `GET /admin/stats/activity`
**Auth Required**: Yes (Admin only)
**Description**: Get daily user activity timeline

**Query Parameters**:
- `days`: Number of days (1-90, default: 30)

**Success Response** (200 OK):
```json
{
  "activity": [
    {
      "date": "2026-01-10",
      "signups": 15,
      "active_users": 450,
      "quizzes_created": 23,
      "attempts_made": 890
    },
    {
      "date": "2026-01-09",
      "signups": 12,
      "active_users": 435,
      "quizzes_created": 19,
      "attempts_made": 856
    },
    ...
  ]
}
```

---

### 57. Get Activity Logs (Admin)

**Endpoint**: `GET /admin/logs`
**Auth Required**: Yes (Admin only)
**Description**: Get activity logs (audit trail)

**Query Parameters**:
- `page`: Page number
- `limit`: Items per page
- `action`: Filter by action type
- `user_id`: Filter by user
- `resource_type`: Filter by resource type ("quiz", "user", etc.)
- `start_date`: Filter from date
- `end_date`: Filter to date

**Success Response** (200 OK):
```json
{
  "logs": [
    {
      "id": "cc0e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "action": "quiz_deleted",
      "resource_type": "quiz",
      "resource_id": "660e8400-e29b-41d4-a716-446655440000",
      "details": {
        "reason": "Inappropriate content",
        "permanent": false,
        "admin_name": "Admin User"
      },
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2026-01-10T15:00:00Z"
    },
    ...
  ],
  "total": 5430,
  "page": 1,
  "limit": 50
}
```

---

## File Upload APIs

### 58. Upload File

**Endpoint**: `POST /files/upload`
**Auth Required**: Yes
**Description**: Upload file for AI quiz generation

**Request**: multipart/form-data
```
Content-Type: multipart/form-data

file: <binary file data>
```

**Supported File Types**: PDF, DOC, DOCX, TXT
**Max File Size**: 10 MB

**Success Response** (201 Created):
```json
{
  "file_id": "dd0e8400-e29b-41d4-a716-446655440000",
  "file_name": "lecture-notes.pdf",
  "file_size": 245678,
  "file_type": "application/pdf",
  "upload_url": "/uploads/dd0e8400-e29b-41d4-a716-446655440000.pdf",
  "created_at": "2026-01-10T15:00:00Z"
}
```

---

### 59. List Files

**Endpoint**: `GET /files`
**Auth Required**: Yes
**Description**: List user's uploaded files

**Success Response** (200 OK):
```json
{
  "files": [
    {
      "id": "dd0e8400-e29b-41d4-a716-446655440000",
      "file_name": "lecture-notes.pdf",
      "file_size": 245678,
      "file_type": "application/pdf",
      "upload_url": "/uploads/dd0e8400-e29b-41d4-a716-446655440000.pdf",
      "created_at": "2026-01-10T15:00:00Z"
    },
    ...
  ]
}
```

---

### 60. Delete File

**Endpoint**: `DELETE /files/:id`
**Auth Required**: Yes
**Description**: Delete uploaded file

**URL Parameters**:
- `id`: File UUID

**Success Response** (204 No Content)

---

## AI Generation APIs

**Note**: Requires `OPENAI_API_KEY` configured in backend

### 61. Generate Quiz from Text

**Endpoint**: `POST /ai/generate-quiz-from-text`
**Auth Required**: Yes
**Description**: Generate quiz from text content using AI

**Request Body**:
```json
{
  "text": "JavaScript is a high-level, interpreted programming language. It is a language which is also characterized as dynamic, weakly typed, prototype-based and multi-paradigm...",
  "num_questions": 10,
  "difficulty": "medium",
  "category": "Programming"
}
```

**Success Response** (202 Accepted):
```json
{
  "job_id": "ee0e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "message": "Quiz generation started. Check status at /ai/generation-status/{job_id}"
}
```

**Note**: This is an async operation. Poll status endpoint for completion.

---

### 62. Generate Quiz from File

**Endpoint**: `POST /ai/generate-quiz-from-file`
**Auth Required**: Yes
**Description**: Generate quiz from uploaded file using AI

**Request Body**:
```json
{
  "file_id": "dd0e8400-e29b-41d4-a716-446655440000",
  "num_questions": 10,
  "difficulty": "medium",
  "category": "Programming"
}
```

**Success Response** (202 Accepted): Same as text generation

---

### 63. Get Generation Status

**Endpoint**: `GET /ai/generation-status/:id`
**Auth Required**: Yes
**Description**: Check AI quiz generation status

**URL Parameters**:
- `id`: Job UUID

**Success Response** (200 OK):

**Pending**:
```json
{
  "id": "ee0e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "source_type": "text",
  "created_at": "2026-01-10T15:00:00Z",
  "updated_at": "2026-01-10T15:00:00Z"
}
```

**Processing**:
```json
{
  "id": "ee0e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "progress": 50,
  ...
}
```

**Completed**:
```json
{
  "id": "ee0e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "completed_at": "2026-01-10T15:05:00Z",
  ...
}
```

**Failed**:
```json
{
  "id": "ee0e8400-e29b-41d4-a716-446655440000",
  "status": "failed",
  "error_message": "Failed to parse file content",
  ...
}
```

---

### 64. List Generation Jobs

**Endpoint**: `GET /ai/jobs`
**Auth Required**: Yes
**Description**: List user's AI generation jobs

**Query Parameters**:
- `page`: Page number
- `limit`: Items per page
- `status`: Filter by status

**Success Response** (200 OK):
```json
{
  "jobs": [
    {
      "id": "ee0e8400-e29b-41d4-a716-446655440000",
      "status": "completed",
      "source_type": "file",
      "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-01-10T15:00:00Z",
      "completed_at": "2026-01-10T15:05:00Z"
    },
    ...
  ],
  "total": 15,
  "page": 1,
  "limit": 20
}
```

---

## Error Handling

### Standard Error Response

All errors return this format:
```json
{
  "error": {
    "code": "error-code",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional error details"
    }
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid-credentials` | 401 | Wrong email or password |
| `email-not-verified` | 401 | Email not verified yet |
| `account-locked` | 401 | Too many failed login attempts |
| `account-disabled` | 401 | Account disabled by admin |
| `invalid-token` | 401 | Invalid or expired JWT token |
| `access-denied` | 403 | Not authorized for this resource |
| `admin-access-required` | 403 | Admin role required |
| `email-already-exists` | 409 | Email already registered |
| `password-too-weak` | 400 | Password doesn't meet requirements |
| `quiz-not-published` | 400 | Quiz must be published |
| `attempt-expired` | 400 | Quiz attempt has expired |
| `invalid-quiz-id` | 400 | Invalid UUID format |
| `resource-not-found` | 404 | Resource doesn't exist |

### Error Handling Best Practices

1. **Always check HTTP status code first**
2. **Use error code for programmatic handling**
3. **Display error message to user**
4. **Log full error for debugging**
5. **Handle 401 by refreshing token or redirecting to login**
6. **Handle 403 by showing "not authorized" message**
7. **Handle validation errors (400) by showing field-specific errors**

---

## Rate Limiting

### Current Limits

| Endpoint Type | Limit |
|---------------|-------|
| Authentication (signup, login, forgot-password) | 5 requests/minute |
| General API | 100 requests/minute |
| Quiz Creation | No limit currently |
| File Upload | 10 requests/minute |

### Rate Limit Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1641825600
```

### Rate Limit Exceeded Response

```json
{
  "error": {
    "code": "rate-limit-exceeded",
    "message": "Too many requests. Please try again later.",
    "details": {
      "retry_after": 60
    }
  }
}
```

---

## WebSocket Support

**Status**: Not currently supported

**Future**: WebSocket support planned for:
- Real-time quiz sessions (multiplayer)
- Live leaderboards
- Instant result updates

---

## Integration Checklist

### Frontend Integration Steps

1. ✅ **Setup API Client**
   - Configure base URL
   - Add Authorization header interceptor
   - Add token refresh logic
   - Add error handling

2. ✅ **Implement Authentication**
   - Signup flow with email verification
   - Login with remember me
   - Token storage (localStorage or cookies)
   - Automatic token refresh
   - Logout

3. ✅ **Implement Quiz Management**
   - Create quiz wizard
   - Edit quiz
   - Add/edit/delete questions
   - Add/edit/delete answers
   - Publish/unpublish

4. ✅ **Implement Quiz Taking**
   - Start attempt
   - Display questions one by one
   - Submit answers (progressive)
   - Handle timer if present
   - Finish and show results

5. ✅ **Implement Discovery**
   - Search quizzes
   - Filter by category, difficulty, tags
   - Sort by newest, popular, most attempted

6. ✅ **Implement Statistics**
   - User statistics dashboard
   - Quiz statistics (for creators)
   - Attempt history

7. ✅ **Implement Admin (if applicable)**
   - User management
   - Quiz moderation
   - Platform statistics
   - Activity logs

8. ✅ **Error Handling**
   - Show user-friendly error messages
   - Handle network errors
   - Handle authentication errors
   - Retry failed requests

9. ✅ **Testing**
   - Test all API endpoints
   - Test error scenarios
   - Test token refresh
   - Test rate limiting

---

## Support & Resources

- **API Documentation**: http://localhost:8080/swagger/index.html
- **GitHub**: [Your Repository URL]
- **Support Email**: support@quizcraft.com

---

**End of API Integration Guide**

This document provides complete information about all backend APIs. Use the interactive Swagger documentation for testing endpoints directly in the browser.
