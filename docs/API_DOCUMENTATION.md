# QuizCraft API Documentation

## Overview

The QuizCraft API is a comprehensive REST API for managing quizzes, user authentication, AI-powered quiz generation, and quiz-taking functionality.

**Base URL**: `http://localhost:8080/api/v1`

**Interactive Documentation**: `http://localhost:8080/swagger/index.html`

---

## Authentication

The API uses JWT (JSON Web Token) for authentication. Most endpoints require authentication.

### Getting Authenticated

1. **Signup**: `POST /auth/signup`
2. **Verify Email**: `POST /auth/verify-email`
3. **Login**: `POST /auth/login` - Returns access_token and refresh_token
4. **Use Token**: Include in Authorization header: `Bearer {access_token}`

### Token Refresh

Access tokens expire after a set period. Use refresh token to get new access token:

**Endpoint**: `POST /auth/refresh`

```json
{
  "refresh_token": "your_refresh_token"
}
```

---

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/signup` | Register new user | No |
| POST | `/auth/verify-email` | Verify email with token | No |
| POST | `/auth/resend-verification` | Resend verification email | No |
| POST | `/auth/login` | Login | No |
| POST | `/auth/refresh` | Refresh access token | No |
| POST | `/auth/forgot-password` | Request password reset | No |
| POST | `/auth/reset-password` | Reset password with token | No |
| POST | `/auth/change-password` | Change password (authenticated) | Yes |
| POST | `/auth/logout` | Logout | Yes |

### Users

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/users/me` | Get own profile | Yes |
| PATCH | `/users/me` | Update own profile | Yes |
| DELETE | `/users/me` | Delete own account | Yes |
| GET | `/users/me/statistics` | Get own statistics | Yes |
| GET | `/users/stats` | Get user stats | Yes |
| GET | `/users/:id` | Get user by ID | Yes |

### Quizzes

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/quizzes` | Create new quiz | Yes |
| GET | `/quizzes` | List own quizzes | Yes |
| GET | `/quizzes/public` | List public quizzes | Yes |
| GET | `/quizzes/:id` | Get quiz by ID | Yes |
| GET | `/quizzes/:id/full` | Get quiz with questions | Yes |
| PATCH | `/quizzes/:id` | Update quiz | Yes (Creator) |
| DELETE | `/quizzes/:id` | Delete quiz | Yes (Creator) |
| POST | `/quizzes/:id/publish` | Publish quiz | Yes (Creator) |
| POST | `/quizzes/:id/unpublish` | Unpublish quiz | Yes (Creator) |
| GET | `/quizzes/:id/stats` | Get quiz stats | Yes (Creator) |
| GET | `/quizzes/:id/statistics` | Get detailed statistics | Yes (Creator) |

### Questions & Answers

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/quizzes/:id/questions` | Add question to quiz (with optional answers) | Yes (Creator) |
| POST | `/quizzes/:id/questions/bulk` | Add multiple questions at once | Yes (Creator) |
| GET | `/quizzes/questions/:qid` | Get question by ID | Yes |
| PATCH | `/quizzes/questions/:qid` | Update question | Yes (Creator) |
| DELETE | `/quizzes/questions/:qid` | Delete question | Yes (Creator) |
| POST | `/quizzes/questions/:qid/answers` | Add answer to question | Yes (Creator) |
| PATCH | `/quizzes/answers/:aid` | Update answer | Yes (Creator) |
| DELETE | `/quizzes/answers/:aid` | Delete answer | Yes (Creator) |

### Quiz Attempts

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/quizzes/:id/attempts` | Start quiz attempt | Yes |
| GET | `/attempts` | List own attempts | Yes |
| GET | `/attempts/:id` | Get attempt details | Yes |
| POST | `/attempts/:id/answers` | Submit answer | Yes |
| POST | `/attempts/:id/finish` | Finish attempt | Yes |
| GET | `/attempts/:id/results` | Get attempt results | Yes |

### Search & Discovery

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/quizzes/search` | Search quizzes | No |
| GET | `/categories` | List categories | No |
| GET | `/tags` | Get popular tags | No |
| GET | `/share/:code` | Get quiz by share code | No |

### Templates

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/templates` | List quiz templates | Yes |
| POST | `/templates/:id/use` | Create quiz from template | Yes |
| POST | `/quizzes/:id/save-as-template` | Save quiz as template | Yes (Creator) |

### Sharing

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/quizzes/:id/share` | Generate share link | Yes (Creator) |

### AI Generation

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/ai/generate-quiz-from-text` | Generate new questions from text (with optional images) 🆕 | Yes |
| POST | `/ai/generate-quiz-from-file` | Generate new questions from file (with optional images) 🆕 | Yes |
| POST | `/ai/extract-quiz-from-file` | Extract existing questions from file (image/PDF) | Yes |
| GET | `/ai/generation-status/:id` | Check generation status | Yes |
| GET | `/ai/jobs` | List generation jobs | Yes |

### File Uploads

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/files/upload` | Upload file | Yes |
| GET | `/files` | List user files | Yes |
| DELETE | `/files/:id` | Delete file | Yes |

### Admin Dashboard

All admin endpoints require **admin role**.

#### User Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/users` | List all users | Yes (Admin) |
| GET | `/admin/users/:id` | Get user details | Yes (Admin) |
| PATCH | `/admin/users/:id/role` | Change user role | Yes (Admin) |
| POST | `/admin/users/:id/disable` | Disable user account | Yes (Admin) |
| POST | `/admin/users/:id/enable` | Enable user account | Yes (Admin) |
| POST | `/admin/users/:id/restore` | Restore deleted user | Yes (Admin) |
| DELETE | `/admin/users/:id` | Delete user permanently | Yes (Admin) |

**Security Notes:**
- User enable/disable/restore endpoints are **admin-only**
- Previously these routes existed in `/users/:id/disable` which was a security issue (fixed)
- Admins cannot disable/delete themselves
- All admin actions are logged in activity logs

#### Quiz Moderation

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/quizzes` | List all quizzes | Yes (Admin) |
| GET | `/admin/quizzes/:id` | Get quiz details | Yes (Admin) |
| DELETE | `/admin/quizzes/:id` | Delete quiz | Yes (Admin) |

#### Statistics & Logs

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/stats` | Get platform statistics | Yes (Admin) |
| GET | `/admin/stats/activity` | Get activity timeline | Yes (Admin) |
| GET | `/admin/logs` | Get activity logs | Yes (Admin) |

### Health Check

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |

---

## Request/Response Examples

### Signup

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123!",
    "phone_number": "+1234567890"
  }'
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "email_verification_sent": true
}
```

### Login

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "email_verified": true
  }
}
```

### Create Quiz

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/quizzes \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "JavaScript Basics",
    "description": "Test your JavaScript knowledge",
    "category": "Programming",
    "difficulty": "medium",
    "time_limit_minutes": 30,
    "passing_score": 70,
    "is_public": true
  }'
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "title": "JavaScript Basics",
  "description": "Test your JavaScript knowledge",
  "creator_id": "550e8400-e29b-41d4-a716-446655440000",
  "category": "Programming",
  "difficulty": "medium",
  "time_limit_minutes": 30,
  "passing_score": 70,
  "is_published": false,
  "is_public": true,
  "created_at": "2026-01-10T12:00:00Z"
}
```

### List My Quizzes

Lists all quizzes created by the authenticated user, including latest attempt information.

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/quizzes?page=1&limit=20" \
  -H "Authorization: Bearer {access_token}"
```

**Response:**
```json
{
  "message": "Quizzes retrieved successfully",
  "data": {
    "quizzes": [
      {
        "id": "quiz-123",
        "title": "JavaScript Basics",
        "description": "Test your JS knowledge",
        "category": "Programming",
        "difficulty": "medium",
        "is_published": true,
        "is_public": true,
        "created_at": "2026-01-10T12:00:00Z",
        "latest_attempt": {
          "attempt_id": "attempt-456",
          "status": "submitted",
          "started_at": "2026-01-20T10:00:00Z",
          "submitted_at": "2026-01-20T10:15:00Z",
          "score": 85.0,
          "max_score": 100,
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

**Features:**
- Shows your latest attempt on each quiz (score, status, timestamps)
- If no attempt exists, `latest_attempt` field is omitted
- Useful for tracking which quizzes you've taken and your scores

### List Public Quizzes

Lists all public published quizzes, including your latest attempts if you're authenticated.

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/quizzes/public?category=Programming&difficulty=medium" \
  -H "Authorization: Bearer {access_token}"
```

**Response:**
```json
{
  "message": "Public quizzes retrieved successfully",
  "data": {
    "quizzes": [
      {
        "id": "quiz-789",
        "title": "Python Advanced",
        "description": "Advanced Python concepts",
        "creator_id": "other-user-id",
        "category": "Programming",
        "difficulty": "hard",
        "is_published": true,
        "is_public": true,
        "created_at": "2026-01-15T14:00:00Z",
        "latest_attempt": {
          "attempt_id": "attempt-999",
          "status": "submitted",
          "started_at": "2026-01-19T09:00:00Z",
          "submitted_at": "2026-01-19T09:45:00Z",
          "score": 72.0,
          "max_score": 100,
          "percentage": 72.0,
          "passed": true
        }
      },
      {
        "id": "quiz-321",
        "title": "React Fundamentals",
        "description": "Learn React basics",
        "creator_id": "another-user-id",
        "is_published": true,
        "is_public": true
        // No latest_attempt = you haven't taken this quiz yet
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 20
  }
}
```

**Features:**
- Shows all public quizzes from all users
- Includes your attempt history on each quiz
- Helps you decide which quizzes to take or retake
- Filter by category, difficulty, search terms

### Add Question

Add a single question to a quiz, optionally with answers and an image attachment.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/questions \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "question_text": "What is the output of: console.log(typeof null)?",
    "question_type": "multiple_choice",
    "points": 10,
    "order_index": 0,
    "explanation": "typeof null returns object due to a legacy bug in JavaScript",
    "image_file_id": "abc123-def456",
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
  }'
```

**Request Fields:**
- `question_text` (string, required): The question text
- `question_type` (string, required): Type of question (`multiple_choice`, `true_false`, `short_answer`)
- `points` (int, required): Points awarded for correct answer
- `explanation` (string, optional): Explanation of the answer
- `image_file_id` (string, optional): File ID of an uploaded image to attach to the question
- `answers` (array, optional): Array of answer objects to create with the question

**Image Workflow:**
1. First upload an image: `POST /api/v1/files/upload`
2. Get the `file_id` from the response
3. Include the `file_id` in the `image_file_id` field when creating the question
4. The question will automatically have `meta.image_url` and `meta.image_file_id` populated

### Add Multiple Questions (Bulk)

Add multiple questions with answers in a single API call. All questions are created in an atomic transaction. Each question can optionally have an attached image.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/questions/bulk \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "questions": [
      {
        "question_text": "What is 2+2?",
        "question_type": "multiple_choice",
        "points": 10,
        "explanation": "Basic arithmetic",
        "image_file_id": "img-123",
        "answers": [
          {"answer_text": "3", "is_correct": false},
          {"answer_text": "4", "is_correct": true},
          {"answer_text": "5", "is_correct": false}
        ]
      },
      {
        "question_text": "What is the capital of France?",
        "question_type": "multiple_choice",
        "points": 10,
        "image_file_id": "img-456",
        "answers": [
          {"answer_text": "London", "is_correct": false},
          {"answer_text": "Paris", "is_correct": true},
          {"answer_text": "Berlin", "is_correct": false}
        ]
      }
    ]
  }'
```

**Response:**
```json
{
  "message": "Questions added successfully",
  "data": {
    "questions": [
      {
        "id": "q1-uuid",
        "quiz_id": "quiz-uuid",
        "question_text": "What is 2+2?",
        "question_type": "multiple_choice",
        "points": 10,
        "order_index": 0
      },
      {
        "id": "q2-uuid",
        "quiz_id": "quiz-uuid",
        "question_text": "What is the capital of France?",
        "question_type": "multiple_choice",
        "points": 10,
        "order_index": 1
      }
    ],
    "count": 2
  }
}
```

**Features:**
- All-or-nothing transaction: If one question fails, none are created
- Automatic order indexing
- Each question can include its answers
- Each question can optionally have an attached image via `image_file_id`
- Validates all questions, answers, and image files before creating
- Images must be uploaded first via `POST /files/upload`

### Update Question

Update an existing question's details, including adding, updating, or removing an attached image.

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/v1/quizzes/questions/{question_id} \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "question_text": "Updated question text",
    "points": 15,
    "explanation": "Updated explanation",
    "image_file_id": "new-image-id"
  }'
```

**Image Operations:**
- **Add image**: Provide `image_file_id` with a valid file ID
- **Update image**: Provide `image_file_id` with a different file ID
- **Remove image**: Provide `image_file_id` with an empty string `""`
- **No change**: Omit the `image_file_id` field entirely

**Example - Remove Image:**
```json
{
  "image_file_id": ""
}
```

### Start Quiz Attempt

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/attempts \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "quiz_id": "660e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response:**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "in_progress",
  "started_at": "2026-01-10T12:00:00Z",
  "time_limit_seconds": 1800,
  "expires_at": "2026-01-10T12:30:00Z",
  "max_score": 100
}
```

### Submit Answer

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/attempts/{attempt_id}/answers \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "question_id": "880e8400-e29b-41d4-a716-446655440000",
    "answer_id": "990e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response:**
```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440000",
  "attempt_id": "770e8400-e29b-41d4-a716-446655440000",
  "question_id": "880e8400-e29b-41d4-a716-446655440000",
  "answer_id": "990e8400-e29b-41d4-a716-446655440000",
  "is_correct": true,
  "points_earned": 10,
  "answered_at": "2026-01-10T12:05:00Z"
}
```

### Finish Attempt

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/attempts/{attempt_id}/finish \
  -H "Authorization: Bearer {access_token}"
```

**Response:**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "quiz_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "submitted",
  "started_at": "2026-01-10T12:00:00Z",
  "submitted_at": "2026-01-10T12:25:00Z",
  "score": 85,
  "percentage": 85.0,
  "passed": true,
  "max_score": 100
}
```

---

## Error Responses

### Standard Error Format

```json
{
  "error": {
    "code": "error-code",
    "message": "Human-readable error message",
    "details": {}
  }
}
```

### Common HTTP Status Codes

| Status Code | Meaning | Example |
|-------------|---------|---------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 204 | No Content | Delete successful |
| 400 | Bad Request | Validation error |
| 401 | Unauthorized | Not authenticated or invalid token |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Invalid data format |
| 500 | Internal Server Error | Server error |

### Common Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `invalid-credentials` | 401 | Invalid email or password |
| `email-not-verified` | 401 | Email not verified |
| `account-locked` | 401 | Too many failed login attempts |
| `account-disabled` | 401 | Account disabled by admin |
| `invalid-token` | 401 | Invalid or expired token |
| `access-denied` | 403 | Not authorized for this resource |
| `admin-access-required` | 403 | Admin role required |
| `email-already-exists` | 409 | Email already registered |
| `password-too-weak` | 400 | Password doesn't meet requirements |
| `quiz-not-published` | 400 | Quiz must be published |
| `attempt-expired` | 400 | Quiz attempt has expired |
| `invalid-quiz-id` | 400 | Invalid UUID format for quiz |

---

## Pagination

List endpoints support pagination via query parameters:

- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

**Example:**
```bash
GET /api/v1/quizzes?page=2&limit=50
```

**Response includes pagination metadata:**
```json
{
  "items": [...],
  "total": 250,
  "page": 2,
  "limit": 50,
  "total_pages": 5
}
```

---

## Filtering & Sorting

### Quiz List Filters

- `category`: Filter by category
- `difficulty`: Filter by difficulty (easy, medium, hard)
- `search`: Search in title and description
- `is_template`: Filter templates (true/false)
- `sort`: Sort order (newest, popular, most_attempted)

**Example:**
```bash
GET /api/v1/quizzes/public?category=Programming&difficulty=medium&sort=popular
```

### Attempt List Filters

- `quiz_id`: Filter by quiz
- `status`: Filter by status (in_progress, submitted, expired)

---

## Rate Limiting

**Current implementation**: No rate limiting

**Recommended for production**:
- Authentication endpoints: 5 requests per minute
- Quiz creation: 10 per hour
- Quiz attempts: 20 per hour
- General API: 100 requests per minute

---

## WebSocket Support

**Current implementation**: Not supported

**Future consideration**: Real-time quiz features (live quizzes, multiplayer)

---

## API Versioning

The API uses URI versioning: `/api/v1/...`

Breaking changes will result in a new version (v2, v3, etc.).

---

## Interactive Documentation

Visit **`http://localhost:8080/swagger/index.html`** to:

- View all endpoints with detailed documentation
- Test endpoints directly from the browser
- See request/response schemas
- Authenticate and make authenticated requests

---

## SDKs & Client Libraries

**Current implementation**: None

**Recommended**: Generate clients using OpenAPI spec:
- JavaScript/TypeScript: openapi-generator
- Python: openapi-generator
- Go: openapi-generator
- Mobile: Kotlin (Android), Swift (iOS)

---

## Best Practices

### Authentication

1. **Store tokens securely**: Never store in localStorage (XSS risk). Use httpOnly cookies or secure storage.
2. **Refresh tokens**: Implement automatic token refresh before expiry.
3. **Logout**: Always call logout endpoint to invalidate tokens.

### Error Handling

1. **Check HTTP status codes**: Don't rely solely on response body.
2. **Handle all error cases**: Network errors, timeouts, server errors.
3. **Display user-friendly messages**: Don't show technical error codes to users.

### Performance

1. **Use pagination**: Don't fetch all items at once.
2. **Cache responses**: Cache public quiz lists, categories, etc.
3. **Compress requests**: Use gzip compression for large payloads.

### Security

1. **Validate input**: Client-side validation for better UX.
2. **Use HTTPS**: Always in production.
3. **Sanitize user content**: Prevent XSS in quiz descriptions, questions.

---

## Support & Contact

- **Email**: support@quizcraft.com
- **Documentation**: http://localhost:8080/swagger/index.html
- **Issues**: GitHub Issues (if applicable)

---

## Changelog

### Version 1.0.0 (Current)

- Complete authentication system
- Quiz CRUD operations
- Quiz attempt and scoring
- AI-powered quiz generation
- AI quiz extraction from images
- Search and discovery
- Templates and sharing
- Admin dashboard
- Platform statistics

---

## AI Quiz Extraction from Image/PDF

The new quiz extraction feature allows users to upload a photo or PDF of an existing quiz, and the system will automatically extract the quiz structure, questions, answers, and even images embedded in the questions.

### Extract Quiz from File (Image/PDF)

**Endpoint:** `POST /api/v1/ai/extract-quiz-from-file`

**Description:** Upload an image or PDF containing existing quiz questions, and the system will extract the quiz structure using Google Gemini Vision AI. The extraction process automatically detects questions, answers, correct options, and images associated with questions. This endpoint is for extracting already-written questions from files, not generating new questions.

**Request (multipart/form-data):**
```bash
curl -X POST http://localhost:8080/api/v1/ai/extract-quiz-from-file \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@quiz_photo.jpg" \
  -F "extract_images=true"
```

**Parameters:**
- `file` (required): Image or PDF file (JPG, PNG, or PDF)
  - Max size: 10MB
  - Supported formats: JPEG, PNG, PDF
- `extract_images` (optional): Whether to extract and attach images from questions
  - Default: `true`
  - Values: `true` or `false`

**Response (202 Accepted):**
```json
{
  "success": true,
  "message": "Quiz extraction started",
  "data": {
    "generation_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "extract_images": true,
    "estimated_time_seconds": 45
  }
}
```

**Processing Flow:**
1. User uploads quiz image/PDF
2. System creates extraction job (async processing)
3. Gemini Vision API extracts quiz structure
4. System detects and extracts images in questions
5. Images are compressed (quality: 75%, max 1200x1200px)
6. Images saved to `/uploads/questions/` directory
7. Quiz created with questions containing `meta.image_url`

**Check Status:**
```bash
curl -X GET http://localhost:8080/api/v1/ai/generation-status/{generation_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response (Completed):**
```json
{
  "success": true,
  "data": {
    "status": "completed",
    "progress": 100,
    "quiz_id": "660e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Question with Image Example:**

When you retrieve the quiz, questions with images will have the image URL in their `meta` field:

```json
{
  "id": "question-1",
  "question_text": "What animal is shown in the image?",
  "question_type": "multiple_choice",
  "points": 1,
  "meta": {
    "image_url": "/api/v1/files/serve/abc123-image.jpg",
    "image_file_id": "abc123"
  },
  "answers": [
    {"answer_text": "Dog", "is_correct": false},
    {"answer_text": "Cat", "is_correct": true},
    {"answer_text": "Bird", "is_correct": false},
    {"answer_text": "Fish", "is_correct": false}
  ]
}
```

**Image Access:**

Images are served via the file serving endpoint:
```bash
curl http://localhost:8080/api/v1/files/serve/{image_file_id}
```

**Features:**
- ✅ Extracts quiz title and description
- ✅ Detects all questions and answer options
- ✅ Identifies correct answers automatically
- ✅ Extracts images embedded in questions
- ✅ Compresses images to save storage (75% quality)
- ✅ Supports multiple question types (multiple choice, true/false)
- ✅ Maintains question order
- ✅ Background processing (non-blocking)
- ✅ Progress tracking

**Use Cases:**
- Convert paper quizzes to digital format
- Import quizzes from textbooks (photo/scan)
- Digitize exam papers
- Bulk import existing quiz materials
- Preserve quiz formatting and images

**Limitations:**
- Max file size: 10MB
- Supported formats: JPG, PNG, PDF only
- AI may struggle with poor quality images
- Handwritten text may not extract accurately
- Complex layouts may need manual review

---

## AI Quiz Generation with Reference Images

The AI generation endpoints now support **reference images** to create more contextual and visual quizzes. Users can upload reference images (diagrams, charts, photos) that the AI will use to generate relevant questions.

### Generate Quiz from Text with Images

**Endpoint:** `POST /api/v1/ai/generate-quiz-from-text`

**Description:** Generate a quiz from text content, optionally providing reference images that the AI can use to create visual questions.

**Request (JSON):**
```bash
curl -X POST http://localhost:8080/api/v1/ai/generate-quiz-from-text \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Photosynthesis is the process by which plants convert light energy into chemical energy...",
    "reference_images": ["file-id-1", "file-id-2"],
    "options": {
      "num_questions": 10,
      "difficulty": "medium",
      "question_type": "multiple_choice",
      "category": "Biology",
      "language": "en"
    }
  }'
```

**Parameters:**
- `text` (required): Text content to generate quiz from (100-50,000 characters)
- `options` (required): Quiz generation options
- `reference_images` (optional): Array of file IDs for reference images
  - Upload images first using `/api/v1/files/upload`
  - Then pass the file IDs in this array
  - AI will use these images to generate visual questions

**Response (202 Accepted):**
```json
{
  "success": true,
  "message": "Quiz generation started",
  "data": {
    "generation_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "estimated_time_seconds": 30,
    "reference_images_count": 2
  }
}
```

### Generate Quiz from File with Images

**Endpoint:** `POST /api/v1/ai/generate-quiz-from-file`

**Description:** Generate a quiz from an uploaded PDF/DOCX file, with optional reference images.

**Request (JSON):**
```bash
curl -X POST http://localhost:8080/api/v1/ai/generate-quiz-from-file \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": "abc123-def456-...",
    "reference_images": ["image-id-1", "image-id-2"],
    "options": {
      "num_questions": 15,
      "difficulty": "hard",
      "question_type": "mixed",
      "category": "Science",
      "language": "en"
    }
  }'
```

**Parameters:**
- `file_id` (required): ID of uploaded PDF/DOCX file
- `options` (required): Quiz generation options
- `reference_images` (optional): Array of file IDs for reference images

### How Reference Images Work

1. **Upload Reference Images:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/files/upload \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -F "file=@diagram.jpg"
   ```
   Response contains `file_id`

2. **Generate Quiz with Images:**
   - Include file IDs in `reference_images` array
   - AI analyzes both text content and images
   - AI generates questions that reference the images

3. **Generated Questions:**
   - Some questions will have `image_data` field (base64)
   - Images are compressed and saved automatically
   - Questions include `meta.image_url` for accessing images

**Example Generated Question:**
```json
{
  "question_text": "What component is labeled 'A' in the cell diagram?",
  "question_type": "multiple_choice",
  "meta": {
    "image_url": "/api/v1/files/serve/xyz789",
    "image_file_id": "xyz789"
  },
  "answers": [
    {"answer_text": "Nucleus", "is_correct": true},
    {"answer_text": "Mitochondria", "is_correct": false},
    {"answer_text": "Cell wall", "is_correct": false},
    {"answer_text": "Chloroplast", "is_correct": false}
  ]
}
```

### Workflow

```
1. Upload reference images → Get file IDs
   POST /files/upload

2. Generate quiz with image references
   POST /ai/generate-quiz-from-text
   {
     "text": "...",
     "reference_images": ["file-id-1", "file-id-2"]
   }

3. Check status
   GET /ai/generation-status/{generation_id}

4. Retrieve generated quiz
   GET /quizzes/{quiz_id}/full

5. Questions with images have meta.image_url
```

### Use Cases

- **Educational Content:** Generate quizzes with diagrams from textbooks
- **Medical Training:** Create questions referencing anatomy images
- **Technical Documentation:** Quiz users on system architecture diagrams
- **Visual Learning:** Add context with charts, graphs, or infographics
- **Language Learning:** Include images for vocabulary questions

### Benefits

- ✅ **Contextual Questions:** AI creates questions based on visual content
- ✅ **Automatic Image Handling:** Images compressed and stored automatically
- ✅ **Flexible:** Use images only where relevant
- ✅ **Same API:** Works with existing generation workflow

---

## Manual Question Image Support

### Overview

Attach images to manually created questions for visual learning. Images are uploaded separately and then referenced by ID when creating or updating questions.

### Key Features

- 📷 **Attach Images to Questions:** Add images to single or multiple questions
- 🔄 **Update Images:** Change or remove question images anytime
- ✅ **Automatic Validation:** System verifies file ownership and existence
- 📦 **Bulk Support:** Attach images to multiple questions at once
- 🔗 **File ID Reference:** Clean API using pre-uploaded file IDs

### How It Works

**1. Upload Image First**
```bash
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@question-image.jpg"
```

Response:
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "file_id": "abc123-def456",
    "filename": "question-image.jpg",
    "size": 245678,
    "mime_type": "image/jpeg"
  }
}
```

**2. Create Question with Image**
```bash
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/questions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "question_text": "What animal is shown in the image?",
    "question_type": "multiple_choice",
    "points": 10,
    "image_file_id": "abc123-def456",
    "answers": [
      {"answer_text": "Dog", "is_correct": false},
      {"answer_text": "Cat", "is_correct": true},
      {"answer_text": "Bird", "is_correct": false}
    ]
  }'
```

**3. Question Response Includes Image URL**
```json
{
  "id": "question-uuid",
  "question_text": "What animal is shown in the image?",
  "meta": {
    "image_url": "/api/v1/files/serve/abc123-def456",
    "image_file_id": "abc123-def456"
  }
}
```

### Bulk Questions with Images

Create multiple questions with images in a single transaction:

```bash
curl -X POST http://localhost:8080/api/v1/quizzes/{quiz_id}/questions/bulk \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "questions": [
      {
        "question_text": "Question 1 with image",
        "question_type": "multiple_choice",
        "points": 10,
        "image_file_id": "image-1-id",
        "answers": [...]
      },
      {
        "question_text": "Question 2 with image",
        "question_type": "multiple_choice",
        "points": 10,
        "image_file_id": "image-2-id",
        "answers": [...]
      },
      {
        "question_text": "Question 3 without image",
        "question_type": "multiple_choice",
        "points": 10,
        "answers": [...]
      }
    ]
  }'
```

### Update Question Images

**Add Image to Existing Question:**
```bash
curl -X PATCH http://localhost:8080/api/v1/quizzes/questions/{question_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "image_file_id": "new-image-id"
  }'
```

**Change Question Image:**
```bash
curl -X PATCH http://localhost:8080/api/v1/quizzes/questions/{question_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "image_file_id": "different-image-id"
  }'
```

**Remove Image from Question:**
```bash
curl -X PATCH http://localhost:8080/api/v1/quizzes/questions/{question_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "image_file_id": ""
  }'
```

### Image Operations Summary

| Operation | `image_file_id` Value | Result |
|-----------|----------------------|--------|
| Add Image | Valid file ID | Image attached to question |
| Update Image | Different file ID | Image replaced |
| Remove Image | Empty string `""` | Image removed from question |
| No Change | Omit field | Image unchanged |

### Workflow Example

```
1. Upload multiple images
   POST /files/upload (image 1) → file_id_1
   POST /files/upload (image 2) → file_id_2
   POST /files/upload (image 3) → file_id_3

2. Create quiz
   POST /quizzes → quiz_id

3. Add questions with images
   POST /quizzes/{quiz_id}/questions/bulk
   {
     "questions": [
       {"question_text": "...", "image_file_id": "file_id_1", ...},
       {"question_text": "...", "image_file_id": "file_id_2", ...},
       {"question_text": "...", "image_file_id": "file_id_3", ...}
     ]
   }

4. Update a question's image
   PATCH /quizzes/questions/{question_id}
   {"image_file_id": "file_id_4"}

5. Remove image from another question
   PATCH /quizzes/questions/{question_id}
   {"image_file_id": ""}
```

### Validation Rules

- ✅ File must exist in the system
- ✅ File must belong to the requesting user
- ✅ File ID must be a valid UUID format
- ✅ Same validation applies to single and bulk operations
- ✅ Bulk operations are atomic: if any validation fails, no questions are created

### Use Cases

- **Visual Questions:** Math diagrams, geometry problems, charts
- **Identification:** Name the object, animal, landmark, person
- **Technical Diagrams:** Circuit diagrams, network topology, architecture
- **Medical Images:** Anatomy identification, X-rays, scans
- **Language Learning:** Picture vocabulary, scene descriptions
- **Art & History:** Identify paintings, artifacts, historical events

### Benefits

- ✅ **Clean API:** Upload once, reference by ID
- ✅ **Reusable Images:** Use same image across multiple questions
- ✅ **Atomic Bulk Operations:** All-or-nothing question creation
- ✅ **Flexible Updates:** Add, change, or remove images anytime
- ✅ **Validation:** Automatic file ownership and existence checks
- ✅ **Consistent Storage:** All images stored in `meta.image_url`

### Error Handling

**Invalid File ID:**
```json
{
  "success": false,
  "error": "invalid-image-file-id"
}
```

**File Not Found:**
```json
{
  "success": false,
  "error": "image-file-not-found"
}
```

**File Not Owned:**
```json
{
  "success": false,
  "error": "not-image-file-owner"
}
```

### Limits

- No specific limit on images per quiz
- File upload size: Depends on file upload endpoint limits
- All images must be uploaded before creating questions
