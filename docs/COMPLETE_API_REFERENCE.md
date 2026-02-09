# QuizCraft - Complete API Reference

**Version:** 2.0
**Last Updated:** February 2026
**Base URL:** `http://localhost:8080/api/v1`
**Authentication:** Bearer Token (JWT)

---

## 📋 Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Users](#users)
4. [Quizzes](#quizzes)
5. [Questions & Answers](#questions--answers)
6. [Quiz Attempts](#quiz-attempts)
7. [AI Quiz Generation](#ai-quiz-generation)
8. [File Management](#file-management)
9. [Notifications](#notifications)
10. [Live Quiz Sessions](#live-quiz-sessions)
11. [Educators](#educators)
12. [Search & Discovery](#search--discovery)
13. [Templates](#templates)
14. [Admin](#admin)
15. [Error Codes](#error-codes)

---

## Getting Started

### Base URL
```
Development: http://localhost:8080/api/v1
Production: https://api.quizcraft.com/api/v1
```

### Authentication
Most endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer {access_token}
```

### Response Format
All responses follow this format:
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

Error responses:
```json
{
  "success": false,
  "error": "error-code",
  "message": "Error description"
}
```

---

## Authentication

### 1. Signup
**POST** `/auth/signup`
**Auth:** None
**Description:** Register a new user

**Request:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "SecurePass123!",
  "phone_number": "+1234567890" // optional
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Signup successful. Please check your email to verify your account.",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com",
    "email_verification_sent": true
  }
}
```

### 2. Verify Email
**POST** `/auth/verify-email`
**Auth:** None

**Request:**
```json
{
  "user_id": "uuid",
  "token": "verification-token"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Email verified successfully"
}
```

### 3. Resend Verification
**POST** `/auth/resend-verification`
**Auth:** None

**Request:**
```json
{
  "email": "user@example.com"
}
```

### 4. Login
**POST** `/auth/login`
**Auth:** None

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400,
    "token_type": "Bearer",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user",
      "email_verified": true
    }
  }
}
```

### 5. Refresh Token
**POST** `/auth/refresh`
**Auth:** None

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "new-token",
    "refresh_token": "same-refresh-token",
    "expires_in": 86400
  }
}
```

### 6. Forgot Password
**POST** `/auth/forgot-password`
**Auth:** None

**Request:**
```json
{
  "email": "user@example.com"
}
```

### 7. Reset Password
**POST** `/auth/reset-password`
**Auth:** None

**Request:**
```json
{
  "user_id": "uuid",
  "token": "reset-token",
  "new_password": "NewSecurePass456!"
}
```

### 8. Change Password
**POST** `/auth/change-password`
**Auth:** Required

**Request:**
```json
{
  "current_password": "OldPass123!",
  "new_password": "NewPass456!"
}
```

### 9. Logout
**POST** `/auth/logout`
**Auth:** Required

**Request:** None

---

## Users

### 1. Get My Profile
**GET** `/users/me`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "phone_number": "+1234567890",
    "role": "user",
    "email_verified": true,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

### 2. Update Profile
**PATCH** `/users/me`
**Auth:** Required

**Request:**
```json
{
  "name": "John Updated",
  "phone_number": "+9876543210",
  "meta": {
    "bio": "Quiz enthusiast"
  }
}
```

### 3. Delete Account
**DELETE** `/users/me`
**Auth:** Required
Soft deletes the account

### 4. Get User Statistics
**GET** `/users/me/statistics`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "quizzes_created": 15,
    "quizzes_attempted": 42,
    "total_score": 3250,
    "average_score": 77.4,
    "passed_quizzes": 35,
    "current_streak": 7
  }
}
```

### 5. Get User Stats
**GET** `/users/stats`
**Auth:** Required

### 6. Get User by ID
**GET** `/users/:id`
**Auth:** Required

### 7. List All Users (Admin)
**GET** `/users`
**Auth:** Required (Admin)

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)
- `search`: Search by name/email

---

## Quizzes

### 1. Create Quiz
**POST** `/quizzes`
**Auth:** Required

**Request:**
```json
{
  "title": "JavaScript Fundamentals",
  "description": "Test your JavaScript knowledge",
  "category": "Programming",
  "difficulty": "medium",
  "time_limit_minutes": 30,
  "passing_score": 70,
  "is_public": true,
  "tags": ["javascript", "programming", "web"]
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "JavaScript Fundamentals",
    "creator_id": "uuid",
    "is_published": false,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

### 2. List My Quizzes
**GET** `/quizzes`
**Auth:** Required

**Query Parameters:**
- `page`: Page number
- `limit`: Items per page
- `category`: Filter by category
- `difficulty`: Filter by difficulty

**Response (200):**
```json
{
  "success": true,
  "data": {
    "quizzes": [
      {
        "id": "uuid",
        "title": "JavaScript Fundamentals",
        "category": "Programming",
        "difficulty": "medium",
        "is_published": true,
        "created_at": "2026-01-01T00:00:00Z",
        "latest_attempt": {
          "attempt_id": "uuid",
          "status": "submitted",
          "score": 85,
          "percentage": 85.0,
          "passed": true
        }
      }
    ],
    "total": 15,
    "page": 1,
    "limit": 20
  }
}
```

### 3. List Public Quizzes
**GET** `/quizzes/public`
**Auth:** Required

Same query parameters and response format as "List My Quizzes"

### 4. Get Quiz
**GET** `/quizzes/:id`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "JavaScript Fundamentals",
    "description": "Test your knowledge",
    "creator_id": "uuid",
    "category": "Programming",
    "difficulty": "medium",
    "time_limit_minutes": 30,
    "passing_score": 70,
    "is_published": true,
    "is_public": true,
    "total_questions": 20,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

### 5. Get Quiz with Questions
**GET** `/quizzes/:id/full`
**Auth:** Required

Returns quiz with all questions and answers

### 6. Update Quiz
**PATCH** `/quizzes/:id`
**Auth:** Required (Creator only)

**Request:**
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "time_limit_minutes": 45
}
```

### 7. Delete Quiz
**DELETE** `/quizzes/:id`
**Auth:** Required (Creator only)

### 8. Publish Quiz
**POST** `/quizzes/:id/publish`
**Auth:** Required (Creator only)

Makes quiz available to others

### 9. Unpublish Quiz
**POST** `/quizzes/:id/unpublish`
**Auth:** Required (Creator only)

### 10. Get Quiz Stats
**GET** `/quizzes/:id/stats`
**Auth:** Required (Creator only)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_attempts": 150,
    "unique_users": 85,
    "average_score": 76.5,
    "pass_rate": 68.5,
    "completion_rate": 92.3
  }
}
```

### 11. Get Quiz Statistics
**GET** `/quizzes/:id/statistics`
**Auth:** Required (Creator only)

Returns detailed statistics with per-question analytics

### 12. Generate Share Link
**POST** `/quizzes/:id/share`
**Auth:** Required (Creator only)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "share_code": "ABC123",
    "share_url": "http://localhost:8080/api/v1/share/ABC123"
  }
}
```

### 13. Save as Template
**POST** `/quizzes/:id/save-as-template`
**Auth:** Required (Creator only)

---

## Questions & Answers

### 1. Add Question
**POST** `/quizzes/:id/questions`
**Auth:** Required (Creator only)

**Request (JSON):**
```json
{
  "question_text": "What is 2+2?",
  "question_type": "multiple_choice",
  "points": 10,
  "explanation": "Basic arithmetic",
  "answers": [
    {
      "answer_text": "3",
      "is_correct": false,
      "order_index": 0
    },
    {
      "answer_text": "4",
      "is_correct": true,
      "order_index": 1
    }
  ]
}
```

**Or Multipart (with image):**
```bash
curl -X POST /api/v1/quizzes/{id}/questions \
  -F "image=@question_image.jpg" \
  -F 'question_data={"question_text":"What is shown?","question_type":"multiple_choice","points":10,"answers":[...]}'
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "quiz_id": "uuid",
    "question_text": "What is 2+2?",
    "question_type": "multiple_choice",
    "points": 10,
    "order_index": 0,
    "meta": {
      "image_url": "/api/v1/files/serve/file-id",
      "image_file_id": "file-id"
    }
  }
}
```

### 2. Add Questions Bulk
**POST** `/quizzes/:id/questions/bulk`
**Auth:** Required (Creator only)

**Request:**
```json
{
  "questions": [
    {
      "question_text": "Question 1",
      "question_type": "multiple_choice",
      "points": 10,
      "answers": [...]
    },
    {
      "question_text": "Question 2",
      "question_type": "true_false",
      "points": 5,
      "answers": [...]
    }
  ]
}
```

### 3. Get Question
**GET** `/quizzes/questions/:qid`
**Auth:** Required

### 4. Update Question
**PATCH** `/quizzes/questions/:qid`
**Auth:** Required (Creator only)

**Request:**
```json
{
  "question_text": "Updated question",
  "points": 15,
  "explanation": "Updated explanation"
}
```

### 5. Delete Question
**DELETE** `/quizzes/questions/:qid`
**Auth:** Required (Creator only)

### 6. Add Answer
**POST** `/quizzes/questions/:qid/answers`
**Auth:** Required (Creator only)

**Request:**
```json
{
  "answer_text": "Answer text",
  "is_correct": true,
  "order_index": 0
}
```

### 7. Update Answer
**PATCH** `/quizzes/answers/:aid`
**Auth:** Required (Creator only)

### 8. Delete Answer
**DELETE** `/quizzes/answers/:aid`
**Auth:** Required (Creator only)

---

## Quiz Attempts

### 1. Start Attempt
**POST** `/quizzes/:id/attempts`
**Auth:** Required

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "quiz_id": "uuid",
    "user_id": "uuid",
    "status": "in_progress",
    "started_at": "2026-01-01T10:00:00Z",
    "expires_at": "2026-01-01T10:30:00Z",
    "max_score": 100
  }
}
```

### 2. List My Attempts
**GET** `/attempts`
**Auth:** Required

**Query Parameters:**
- `quiz_id`: Filter by quiz
- `status`: Filter by status (in_progress, submitted, expired)

### 3. Get Attempt
**GET** `/attempts/:id`
**Auth:** Required

### 4. Submit Answer
**POST** `/attempts/:id/answers`
**Auth:** Required

**Request:**
```json
{
  "question_id": "uuid",
  "answer_id": "uuid",
  "answer_text": "For short answer questions"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "is_correct": true,
    "points_earned": 10,
    "answered_at": "2026-01-01T10:05:00Z"
  }
}
```

### 5. Finish Attempt
**POST** `/attempts/:id/finish`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "submitted",
    "score": 85,
    "percentage": 85.0,
    "passed": true,
    "submitted_at": "2026-01-01T10:25:00Z"
  }
}
```

### 6. Get Attempt Results
**GET** `/attempts/:id/results`
**Auth:** Required

Returns detailed results with all questions, answers, and feedback

---

## AI Quiz Generation

### 1. Generate from Text
**POST** `/ai/generate-quiz-from-text`
**Auth:** Required

**Request:**
```json
{
  "text": "Your content here (100-50000 chars)",
  "options": {
    "num_questions": 10,
    "difficulty": "medium",
    "question_type": "mixed",
    "category": "Science",
    "language": "en"
  },
  "reference_images": ["file-id-1", "file-id-2"]
}
```

**Response (202):**
```json
{
  "success": true,
  "message": "Quiz generation started",
  "data": {
    "generation_id": "uuid",
    "status": "pending",
    "estimated_time_seconds": 30
  }
}
```

### 2. Generate from File
**POST** `/ai/generate-quiz-from-file`
**Auth:** Required

**Request (Multipart):**
```bash
curl -X POST /api/v1/ai/generate-quiz-from-file \
  -F "file=@document.pdf" \
  -F "options[num_questions]=10" \
  -F "options[difficulty]=medium"
```

### 3. Extract from Image/PDF
**POST** `/ai/extract-quiz-from-file`
**Auth:** Required

**Request (Multipart):**
```bash
curl -X POST /api/v1/ai/extract-quiz-from-file \
  -F "file=@quiz_image.jpg" \
  -F "extract_images=true"
```

Extracts existing questions from images or PDFs

### 4. Check Generation Status
**GET** `/ai/generation-status/:id`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "status": "completed",
    "progress": 100,
    "quiz_id": "uuid"
  }
}
```

Status values: `pending`, `processing`, `completed`, `failed`

### 5. List Generation Jobs
**GET** `/ai/jobs`
**Auth:** Required

### 6. Add Questions to Existing Quiz (from text)
**POST** `/quizzes/:id/add-questions-from-text`
**Auth:** Required (Creator only)

Same request format as "Generate from Text" but adds to existing quiz

### 7. Add Questions to Existing Quiz (from file)
**POST** `/quizzes/:id/add-questions-from-file`
**Auth:** Required (Creator only)

### 8. Extract Questions to Existing Quiz
**POST** `/quizzes/:id/extract-questions-from-file`
**Auth:** Required (Creator only)

---

## File Management

### 1. Upload File
**POST** `/files/upload`
**Auth:** Required

**Request (Multipart):**
```bash
curl -X POST /api/v1/files/upload \
  -F "file=@document.pdf"
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "file_id": "uuid",
    "filename": "document.pdf",
    "size": 245678,
    "mime_type": "application/pdf"
  }
}
```

### 2. List Files
**GET** `/files`
**Auth:** Required

### 3. Serve File
**GET** `/files/serve/:id`
**Auth:** Required

Returns the actual file for viewing/downloading

### 4. Delete File
**DELETE** `/files/:id`
**Auth:** Required

---

## Notifications

### 1. Get Notifications
**GET** `/notifications`
**Auth:** Required

**Query Parameters:**
- `limit`: Items per page (default: 20, max: 100)
- `offset`: Offset for pagination
- `type`: Filter by notification type
- `unread_only`: true/false

**Response (200):**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "type": "new_quiz",
        "title": "New Quiz Published",
        "message": "John Doe published a new quiz",
        "data": {
          "quiz_id": "uuid",
          "educator_id": "uuid"
        },
        "is_read": false,
        "created_at": "2026-01-01T10:00:00Z"
      }
    ],
    "total": 45
  }
}
```

**Notification Types:**
- `new_quiz` - Educator published new quiz
- `live_session_scheduled` - Live session scheduled
- `live_session_starting` - Session starting soon
- `live_session_invitation` - Invited to session
- `new_follower` - Someone followed you
- `system` - System notification

### 2. Get Unread Count
**GET** `/notifications/unread-count`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "count": 5
  }
}
```

### 3. Mark as Read
**PATCH** `/notifications/:id/read`
**Auth:** Required

### 4. Mark All as Read
**PATCH** `/notifications/read-all`
**Auth:** Required

### 5. Delete Notification
**DELETE** `/notifications/:id`
**Auth:** Required

### 6. WebSocket Connection
**WS** `/notifications/ws?token={access_token}`
**Auth:** Required (token in query)

Real-time notification delivery

**Connection Example:**
```javascript
const ws = new WebSocket(`ws://localhost:8080/api/v1/notifications/ws?token=${token}`);

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('New notification:', notification);
};
```

---

## Live Quiz Sessions

### 1. Create Live Session
**POST** `/live-sessions`
**Auth:** Required

**Request:**
```json
{
  "quiz_id": "uuid",
  "title": "JavaScript Challenge",
  "description": "Real-time quiz competition",
  "session_type": "public",
  "progression_mode": "automatic",
  "question_timer_seconds": 30,
  "scheduled_at": "2026-01-15T15:00:00Z",
  "max_participants": 100,
  "is_discoverable": true
}
```

**Fields:**
- `session_type`: `public` (discoverable) or `private` (access code)
- `progression_mode`: `manual` (host controls) or `automatic` (timer-based)
- `scheduled_at`: Optional, session auto-starts at this time

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "access_code": "ABC123",
    "status": "waiting",
    "created_at": "2026-01-15T10:00:00Z"
  }
}
```

### 2. List My Sessions
**GET** `/live-sessions`
**Auth:** Required

Returns sessions created by the user

### 3. List Public Sessions
**GET** `/live-sessions/public`
**Auth:** Required

**Query Parameters:**
- `status`: `upcoming`, `live`, `completed`

### 4. Get My Invitations
**GET** `/live-sessions/invitations`
**Auth:** Required

Returns sessions you've been invited to

### 5. Get Live Session
**GET** `/live-sessions/:id`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "JavaScript Challenge",
    "creator": {
      "id": "uuid",
      "name": "John Doe"
    },
    "status": "live",
    "current_question_index": 2,
    "total_questions": 20,
    "participant_count": 45,
    "scheduled_at": "2026-01-15T15:00:00Z"
  }
}
```

### 6. Join Session
**POST** `/live-sessions/:id/join`
**Auth:** Required

**Request:**
```json
{
  "access_code": "ABC123"
}
```

Access code required for private sessions

### 7. Start Session (Creator only)
**POST** `/live-sessions/:id/start`
**Auth:** Required

### 8. Advance Question (Creator only)
**POST** `/live-sessions/:id/advance`
**Auth:** Required

For manual progression mode

### 9. End Session (Creator only)
**POST** `/live-sessions/:id/end`
**Auth:** Required

### 10. Submit Live Answer
**POST** `/live-sessions/:id/answer`
**Auth:** Required

**Request:**
```json
{
  "question_id": "uuid",
  "answer_id": "uuid",
  "response_time_ms": 1250
}
```

Response time affects scoring in speed-based sessions

### 11. Get Leaderboard
**GET** `/live-sessions/:id/leaderboard`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "leaderboard": [
      {
        "rank": 1,
        "user_id": "uuid",
        "username": "John Doe",
        "score": 950,
        "correct_answers": 18,
        "total_questions": 20
      }
    ]
  }
}
```

### 12. Get Session Results
**GET** `/live-sessions/:id/results`
**Auth:** Required

Returns detailed results after session ends

### 13. Invite Users (Creator only)
**POST** `/live-sessions/:id/invite`
**Auth:** Required

**Request:**
```json
{
  "user_ids": ["uuid1", "uuid2"]
}
```

Sends in-app notifications to invited users

### 14. WebSocket Connection
**WS** `/live-sessions/:id/ws?token={access_token}`
**Auth:** Required

Real-time session updates

---

## Educators

### Public Endpoints

### 1. List Educators
**GET** `/educators`
**Auth:** None

**Query Parameters:**
- `page`, `limit`: Pagination
- `sort`: `popular`, `newest`, `followers`

### 2. Search Educators
**GET** `/educators/search`
**Auth:** None

**Query Parameters:**
- `q`: Search query
- `category`: Filter by expertise

### 3. Get Featured Educators
**GET** `/educators/featured`
**Auth:** None

### 4. Get Educator Profile
**GET** `/educators/:id`
**Auth:** None

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "bio": "Experienced educator",
    "expertise": ["JavaScript", "Python"],
    "total_quizzes": 45,
    "total_followers": 1250,
    "joined_as_educator_at": "2025-12-01T00:00:00Z"
  }
}
```

### 5. Get Educator Quizzes
**GET** `/educators/:id/quizzes`
**Auth:** None

### Protected Endpoints

### 6. Promote to Educator
**POST** `/educator/promote`
**Auth:** Required

**Request:**
```json
{
  "bio": "I'm passionate about teaching",
  "expertise": ["JavaScript", "React", "Node.js"],
  "credentials": "10 years teaching experience"
}
```

**Requirements:**
- Email verified
- Phone number added
- No restrictions on account

### 7. Get My Educator Profile
**GET** `/educator/profile`
**Auth:** Required (Educator only)

### 8. Update My Educator Profile
**PATCH** `/educator/profile`
**Auth:** Required (Educator only)

**Request:**
```json
{
  "bio": "Updated bio",
  "expertise": ["JavaScript", "TypeScript"],
  "credentials": "Updated credentials"
}
```

### 9. Get My Followers
**GET** `/educator/followers`
**Auth:** Required (Educator only)

### 10. Get Follower Stats
**GET** `/educator/followers/stats`
**Auth:** Required (Educator only)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_followers": 1250,
    "new_this_week": 45,
    "new_this_month": 180
  }
}
```

### Follow System

### 11. Follow Educator
**POST** `/follow/educators/:id/follow`
**Auth:** Required

### 12. Unfollow Educator
**DELETE** `/follow/educators/:id/unfollow`
**Auth:** Required

### 13. Get My Following
**GET** `/follow/following`
**Auth:** Required

Returns list of educators you follow

---

## Search & Discovery

### 1. Search Quizzes
**GET** `/quizzes/search`
**Auth:** None

**Query Parameters:**
- `q`: Search query
- `category`: Filter by category
- `difficulty`: Filter by difficulty
- `tags`: Filter by tags (comma-separated)
- `sort`: `newest`, `popular`, `most_attempted`
- `page`, `limit`: Pagination

### 2. Get Categories
**GET** `/categories`
**Auth:** None

**Response (200):**
```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "name": "Programming",
        "count": 450
      },
      {
        "name": "Science",
        "count": 320
      }
    ]
  }
}
```

### 3. Get Popular Tags
**GET** `/tags`
**Auth:** None

### 4. Get Quiz by Share Code
**GET** `/share/:code`
**Auth:** None

Access quiz via share code

---

## Templates

### 1. List Templates
**GET** `/templates`
**Auth:** Required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "templates": [
      {
        "id": "uuid",
        "title": "General Knowledge Template",
        "description": "A starter template",
        "category": "General",
        "question_count": 10
      }
    ]
  }
}
```

### 2. Create Quiz from Template
**POST** `/templates/:id/use`
**Auth:** Required

**Request:**
```json
{
  "title": "My Quiz from Template",
  "description": "Custom description"
}
```

---

## Admin

**All admin endpoints require admin role**

### User Management

### 1. List All Users
**GET** `/admin/users`
**Auth:** Required (Admin)

**Query Parameters:**
- `page`, `limit`: Pagination
- `search`: Search by name/email
- `role`: Filter by role
- `status`: Filter by status (active/disabled)

### 2. Get User Details
**GET** `/admin/users/:id`
**Auth:** Required (Admin)

### 3. Change User Role
**PATCH** `/admin/users/:id/role`
**Auth:** Required (Admin)

**Request:**
```json
{
  "role": "admin"
}
```

Roles: `user`, `admin`

### 4. Disable User
**POST** `/admin/users/:id/disable`
**Auth:** Required (Admin)

### 5. Enable User
**POST** `/admin/users/:id/enable`
**Auth:** Required (Admin)

### 6. Restore Deleted User
**POST** `/admin/users/:id/restore`
**Auth:** Required (Admin)

### 7. Delete User (Permanent)
**DELETE** `/admin/users/:id`
**Auth:** Required (Admin)

### 8. Demote Educator
**POST** `/admin/users/:id/demote`
**Auth:** Required (Admin)

Removes educator status

### Quiz Management

### 9. List All Quizzes
**GET** `/admin/quizzes`
**Auth:** Required (Admin)

### 10. Get Quiz Details
**GET** `/admin/quizzes/:id`
**Auth:** Required (Admin)

### 11. Delete Quiz (Admin)
**DELETE** `/admin/quizzes/:id`
**Auth:** Required (Admin)

### Platform Statistics

### 12. Get Platform Stats
**GET** `/admin/stats`
**Auth:** Required (Admin)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_users": 5420,
    "total_educators": 245,
    "total_quizzes": 1850,
    "total_attempts": 45620,
    "active_users_today": 320,
    "quizzes_created_today": 12
  }
}
```

### 13. Get User Activity
**GET** `/admin/stats/activity`
**Auth:** Required (Admin)

### 14. Get Activity Logs
**GET** `/admin/logs`
**Auth:** Required (Admin)

**Query Parameters:**
- `action_type`: Filter by action
- `user_id`: Filter by user
- `from_date`, `to_date`: Date range

---

## Error Codes

### Authentication Errors (401)
- `invalid-credentials` - Wrong email/password
- `email-not-verified` - Email verification required
- `account-locked` - Too many failed login attempts
- `account-disabled` - Account disabled by admin
- `invalid-token` - Invalid or expired token

### Authorization Errors (403)
- `access-denied` - Not authorized for this resource
- `admin-access-required` - Admin role required
- `educator-access-required` - Educator role required
- `creator-access-required` - Must be creator of resource

### Validation Errors (400)
- `invalid-email` - Invalid email format
- `password-too-weak` - Password doesn't meet requirements
- `invalid-quiz-id` - Invalid UUID format
- `quiz-not-published` - Quiz must be published
- `attempt-expired` - Quiz attempt has expired
- `file-too-large` - File exceeds size limit
- `invalid-file-type` - File type not supported

### Resource Errors (404)
- `user-not-found` - User doesn't exist
- `quiz-not-found` - Quiz doesn't exist
- `question-not-found` - Question doesn't exist
- `attempt-not-found` - Attempt doesn't exist
- `session-not-found` - Live session doesn't exist

### Conflict Errors (409)
- `email-already-exists` - Email already registered
- `quiz-already-published` - Quiz is already published
- `already-following` - Already following this educator

### Server Errors (500)
- `internal-server-error` - Unexpected server error
- `database-error` - Database operation failed
- `ai-generation-failed` - AI generation failed

---

## Rate Limits

- **Authentication endpoints:** 5 requests per minute
- **General API:** 100 requests per minute
- **File uploads:** 10 per hour
- **AI generation:** 5 per hour

---

## Pagination

List endpoints support pagination:
```
?page=1&limit=20
```

Response includes:
```json
{
  "data": [...],
  "total": 250,
  "page": 1,
  "limit": 20,
  "total_pages": 13
}
```

---

## WebSocket Connections

### Notification WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/notifications/ws?token=' + token);
```

### Live Session WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/live-sessions/' + sessionId + '/ws?token=' + token);
```

**Message Types:**
- `notification` - New notification
- `question_start` - New question started
- `question_end` - Question ended
- `leaderboard_update` - Leaderboard updated
- `session_end` - Session ended

---

## Examples

### Complete Quiz Creation Flow

```bash
# 1. Create quiz
curl -X POST http://localhost:8080/api/v1/quizzes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"My Quiz","category":"General"}'

# 2. Add questions
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/questions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question_text":"What is 2+2?","question_type":"multiple_choice","points":10,"answers":[...]}'

# 3. Publish quiz
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/publish \
  -H "Authorization: Bearer $TOKEN"

# 4. Generate share link
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/share \
  -H "Authorization: Bearer $TOKEN"
```

### AI Quiz Generation Flow

```bash
# 1. Upload file (optional)
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@textbook.pdf"

# 2. Generate quiz from file
curl -X POST http://localhost:8080/api/v1/ai/generate-quiz-from-file \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@textbook.pdf" \
  -F "options[num_questions]=10"

# 3. Check status
curl http://localhost:8080/api/v1/ai/generation-status/{generation_id} \
  -H "Authorization: Bearer $TOKEN"

# 4. Get generated quiz
curl http://localhost:8080/api/v1/quizzes/{quiz_id}/full \
  -H "Authorization: Bearer $TOKEN"
```

---

## Support

For issues or questions, please refer to:
- [QUICKSTART.md](QUICKSTART.md) - Getting started guide
- [GitHub Issues](https://github.com/your-repo/issues)
- Email: support@quizcraft.com

---

**Last Updated:** February 7, 2026
**API Version:** 2.0
**Total Endpoints:** 105
