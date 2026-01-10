# Quiz-Craft Backend - Implementation Plan

Last Updated: January 10, 2026

## 📋 Overview

This document outlines the complete implementation plan for the Quiz-Craft backend, from the current state through all remaining phases to production deployment.

## Current Status: Phase 2 Complete ✅

**What's Done:**
- ✅ Authentication & User Management (17 endpoints)
- ✅ PostgreSQL database with migrations
- ✅ Email service (verification, password reset)
- ✅ JWT authentication
- ✅ Clean architecture structure

**What's Next:**
- Phase 3: Quiz Management (CRUD)
- Phase 4: AI Integration & File Upload
- Phase 5: Quiz Attempts & Scoring
- Phase 6: Advanced Features
- Phase 7: Admin Dashboard
- Phase 8: Testing & Production Readiness

---

## Phase 3: Quiz Management (CRUD) 🎯

**Goal**: Implement complete quiz and question management system

### 3.1 Database Schema Design

#### Tables to Create

**quizzes table:**
```sql
CREATE TABLE quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(100),
    difficulty VARCHAR(20) CHECK (difficulty IN ('easy', 'medium', 'hard')),
    time_limit_minutes INT, -- NULL means no time limit
    passing_score INT DEFAULT 70,
    is_published BOOLEAN DEFAULT false,
    is_public BOOLEAN DEFAULT true,
    thumbnail_url TEXT,
    tags TEXT[], -- Array of tags
    meta JSONB DEFAULT '{}'::jsonb, -- Additional metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_quizzes_creator_id ON quizzes(creator_id);
CREATE INDEX idx_quizzes_category ON quizzes(category);
CREATE INDEX idx_quizzes_is_published ON quizzes(is_published) WHERE deleted_at IS NULL;
CREATE INDEX idx_quizzes_tags USING GIN(tags);
```

**questions table:**
```sql
CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) NOT NULL CHECK (question_type IN ('multiple_choice', 'true_false', 'short_answer')),
    points INT DEFAULT 1,
    order_index INT NOT NULL, -- Order within quiz
    explanation TEXT, -- Explanation shown after answer
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_questions_quiz_id ON questions(quiz_id);
CREATE INDEX idx_questions_order ON questions(quiz_id, order_index);
```

**answers table:**
```sql
CREATE TABLE answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT false,
    order_index INT NOT NULL, -- Order within question
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_answers_question_id ON answers(question_id);
```

#### Migration Files to Create
- `000003_create_quizzes_table.up.sql` / `.down.sql`
- `000004_create_questions_table.up.sql` / `.down.sql`
- `000005_create_answers_table.up.sql` / `.down.sql`

### 3.2 Models & DAOs

#### Files to Create

**Models:**
- `quiz/model/quiz.go`
  ```go
  type Quiz struct {
      ID                 string
      Title              string
      Description        string
      CreatorID          string
      Category           string
      Difficulty         string
      TimeLimitMinutes   int
      PassingScore       int
      IsPublished        bool
      IsPublic           bool
      ThumbnailURL       string
      Tags               []string
      Meta               map[string]interface{}
      CreatedAt          time.Time
      UpdatedAt          time.Time
      DeletedAt          *time.Time
  }
  ```

- `quiz/model/question.go`
  ```go
  type Question struct {
      ID           string
      QuizID       string
      QuestionText string
      QuestionType string
      Points       int
      OrderIndex   int
      Explanation  string
      Meta         map[string]interface{}
      Answers      []Answer // For convenience
      CreatedAt    time.Time
      UpdatedAt    time.Time
  }
  ```

- `quiz/model/answer.go`
  ```go
  type Answer struct {
      ID         string
      QuestionID string
      AnswerText string
      IsCorrect  bool
      OrderIndex int
      CreatedAt  time.Time
      UpdatedAt  time.Time
  }
  ```

**DAOs:**
- `quiz/dao/quiz.go`
  - `Create(quiz *Quiz) error`
  - `GetByID(id string) (*Quiz, error)`
  - `GetByCreatorID(creatorID string) ([]Quiz, error)`
  - `List(filters map[string]interface{}) ([]Quiz, error)`
  - `Update(quiz *Quiz) error`
  - `Delete(id string) error` (soft delete)
  - `Publish(id string) error`
  - `Unpublish(id string) error`

- `quiz/dao/question.go`
  - `Create(question *Question) error`
  - `GetByID(id string) (*Question, error)`
  - `GetByQuizID(quizID string) ([]Question, error)`
  - `Update(question *Question) error`
  - `Delete(id string) error`
  - `Reorder(quizID string, questionIDs []string) error`

- `quiz/dao/answer.go`
  - `Create(answer *Answer) error`
  - `GetByID(id string) (*Answer, error)`
  - `GetByQuestionID(questionID string) ([]Answer, error)`
  - `Update(answer *Answer) error`
  - `Delete(id string) error`

### 3.3 Business Logic

**Files to Create:**
- `quiz/core/quiz.go` - Quiz business logic
  - Validation (title length, passing score range, etc.)
  - Authorization (only creator can edit/delete)
  - Publishing rules (must have questions to publish)

- `quiz/core/question.go` - Question business logic
  - Validation (at least 2 answers for multiple choice)
  - Answer validation (at least one correct answer)
  - Order management

### 3.4 API Endpoints

**Quiz Endpoints:**
```
POST   /api/v1/quizzes              - Create new quiz
GET    /api/v1/quizzes              - List quizzes (public + user's own)
GET    /api/v1/quizzes/:id          - Get quiz details
PUT    /api/v1/quizzes/:id          - Update quiz
DELETE /api/v1/quizzes/:id          - Delete quiz (soft delete)
POST   /api/v1/quizzes/:id/publish  - Publish quiz
POST   /api/v1/quizzes/:id/unpublish - Unpublish quiz
GET    /api/v1/quizzes/:id/questions - Get all questions for quiz
```

**Question Endpoints:**
```
POST   /api/v1/quizzes/:quiz_id/questions          - Add question to quiz
GET    /api/v1/questions/:id                       - Get question details
PUT    /api/v1/questions/:id                       - Update question
DELETE /api/v1/questions/:id                       - Delete question
POST   /api/v1/quizzes/:quiz_id/questions/reorder  - Reorder questions
```

**Answer Endpoints:**
```
POST   /api/v1/questions/:question_id/answers  - Add answer to question
PUT    /api/v1/answers/:id                     - Update answer
DELETE /api/v1/answers/:id                     - Delete answer
```

### 3.5 Validation Rules

- **Quiz:**
  - Title: 3-255 characters, required
  - Description: max 2000 characters, optional
  - Difficulty: must be 'easy', 'medium', or 'hard'
  - Time limit: 1-300 minutes (if set)
  - Passing score: 0-100

- **Question:**
  - Question text: required, max 1000 characters
  - Type: must be valid enum value
  - Points: 1-10
  - Must have at least 2 answers (for multiple choice)
  - Must have at least 1 correct answer

- **Answer:**
  - Answer text: required, max 500 characters
  - At least one answer must be marked correct per question

### 3.6 Authorization Rules

- **Create Quiz**: Any authenticated user
- **Edit/Delete Quiz**: Only quiz creator
- **Publish Quiz**: Only quiz creator
- **View Published Quiz**: Any user (if public) or creator
- **View Unpublished Quiz**: Only creator

### 3.7 Testing Checklist (Manual for now)

- [ ] Create quiz with valid data
- [ ] Create quiz with invalid data (validation errors)
- [ ] Add questions to quiz
- [ ] Add answers to questions
- [ ] Update quiz/questions/answers
- [ ] Delete quiz/questions/answers
- [ ] Publish quiz (with and without questions)
- [ ] Authorization tests (try editing someone else's quiz)
- [ ] Soft delete verification

**Estimated Time:** 3-5 days

---

## Phase 4: AI Integration & File Upload 🤖

**Goal**: Enable quiz generation from documents using Google AI

### 4.1 File Upload System

#### Files to Create

**File upload service:**
- `internal/services/file_upload.go`
  - `UploadFile(file multipart.File, userID string) (string, error)`
  - `DeleteFile(fileURL string) error`
  - `ValidateFile(file multipart.File) error`

**File storage:**
- Store files in `/uploads` directory (configurable)
- File naming: `{userID}_{timestamp}_{randomID}.{ext}`
- Supported formats: PDF, DOCX, TXT, MD

**Database table:**
```sql
CREATE TABLE uploaded_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'uploaded', -- uploaded, processing, processed, failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_uploaded_files_user_id ON uploaded_files(user_id);
CREATE INDEX idx_uploaded_files_status ON uploaded_files(status);
```

Migration: `000006_create_uploaded_files_table.up.sql`

### 4.2 Google AI Integration

#### Configuration
Add to `.env`:
```bash
GOOGLE_AI_API_KEY=your-api-key-here
GOOGLE_AI_MODEL=gemini-pro  # or gemini-1.5-pro
```

#### AI Service
- `internal/services/ai_service.go`
  - `GenerateQuizFromText(text string, options QuizGenOptions) (*Quiz, error)`
  - `GenerateQuizFromFile(fileID string, options QuizGenOptions) (*Quiz, error)`

**QuizGenOptions struct:**
```go
type QuizGenOptions struct {
    NumQuestions   int    // Number of questions to generate
    Difficulty     string // easy, medium, hard
    QuestionType   string // multiple_choice, true_false, mixed
    Category       string // Topic/category
    Language       string // en, es, fr, etc.
}
```

#### File Parsing
- `internal/services/file_parser.go`
  - `ParsePDF(filePath string) (string, error)`
  - `ParseDOCX(filePath string) (string, error)`
  - `ParseText(filePath string) (string, error)`

**Libraries needed:**
- PDF: `github.com/ledongthuc/pdf`
- DOCX: `github.com/nguyenthenguyen/docx`

### 4.3 AI Quiz Generation Flow

```
User uploads file
  ↓
Validate file (type, size)
  ↓
Store file in /uploads
  ↓
Save metadata to uploaded_files table
  ↓
Parse file content (PDF/DOCX/TXT)
  ↓
Send to Google AI with prompt
  ↓
Parse AI response (JSON)
  ↓
Create quiz + questions + answers in database
  ↓
Return quiz ID
```

### 4.4 AI Prompt Template

```
You are a quiz generator. Generate {num_questions} {difficulty} quiz questions based on the following content:

Content:
{file_content}

Requirements:
- Generate {num_questions} questions
- Difficulty level: {difficulty}
- Question type: {question_type}
- Include 4 answer options for multiple choice
- Mark the correct answer
- Include explanations for correct answers

Return the quiz in this JSON format:
{
  "title": "Generated Quiz Title",
  "description": "Brief description",
  "questions": [
    {
      "question_text": "Question here?",
      "question_type": "multiple_choice",
      "explanation": "Why this answer is correct",
      "answers": [
        {"answer_text": "Option A", "is_correct": false},
        {"answer_text": "Option B", "is_correct": true},
        {"answer_text": "Option C", "is_correct": false},
        {"answer_text": "Option D", "is_correct": false}
      ]
    }
  ]
}
```

### 4.5 API Endpoints

```
POST   /api/v1/files/upload               - Upload file
GET    /api/v1/files                       - List user's uploaded files
DELETE /api/v1/files/:id                   - Delete uploaded file

POST   /api/v1/ai/generate-quiz-from-text - Generate quiz from text input
POST   /api/v1/ai/generate-quiz-from-file - Generate quiz from uploaded file
```

### 4.6 Validation & Limits

- **File upload:**
  - Max size: 10MB (configurable)
  - Allowed types: PDF, DOCX, TXT, MD
  - Max 50 files per user (configurable)

- **AI generation:**
  - Text content: 100-50,000 characters
  - Questions: 5-50 per quiz
  - Rate limit: 10 requests per hour per user

### 4.7 Error Handling

- File too large → 400 Bad Request
- Invalid file type → 400 Bad Request
- AI API failure → 503 Service Unavailable
- Parsing error → 422 Unprocessable Entity

**Estimated Time:** 4-6 days

---

## Phase 5: Quiz Attempts & Scoring 📝

**Goal**: Allow users to take quizzes and track their results

### 5.1 Database Schema

**quiz_attempts table:**
```sql
CREATE TABLE quiz_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    score INT, -- Calculated score (0-100)
    total_points INT, -- Total points earned
    max_points INT, -- Maximum possible points
    time_taken_seconds INT, -- Time taken to complete
    status VARCHAR(50) DEFAULT 'in_progress', -- in_progress, completed, abandoned
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_quiz_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_status ON quiz_attempts(status);
```

**user_answers table:**
```sql
CREATE TABLE user_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    selected_answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,
    answer_text TEXT, -- For short answer questions
    is_correct BOOLEAN,
    points_earned INT DEFAULT 0,
    time_taken_seconds INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_answers_attempt_id ON user_answers(attempt_id);
CREATE INDEX idx_user_answers_question_id ON user_answers(question_id);
```

Migrations: `000007_create_quiz_attempts_table.up.sql`, `000008_create_user_answers_table.up.sql`

### 5.2 Models & DAOs

**Models:**
- `attempt/model/attempt.go`
- `attempt/model/user_answer.go`

**DAOs:**
- `attempt/dao/attempt.go`
  - `Create(attempt *QuizAttempt) error`
  - `GetByID(id string) (*QuizAttempt, error)`
  - `GetByUserID(userID string) ([]QuizAttempt, error)`
  - `GetByQuizID(quizID string) ([]QuizAttempt, error)`
  - `Update(attempt *QuizAttempt) error`
  - `Complete(attemptID string, score int) error`

- `attempt/dao/user_answer.go`
  - `Create(answer *UserAnswer) error`
  - `GetByAttemptID(attemptID string) ([]UserAnswer, error)`
  - `BulkCreate(answers []UserAnswer) error`

### 5.3 Business Logic

**Files:**
- `attempt/core/attempt.go`
  - `StartAttempt(quizID, userID string) (string, error)`
  - `SubmitAnswer(attemptID, questionID, answerID string) error`
  - `SubmitQuiz(attemptID string) (*QuizResult, error)`
  - `CalculateScore(attemptID string) (int, int, error)`
  - `GetAttemptDetails(attemptID string) (*AttemptDetails, error)`

**Scoring Logic:**
```go
score = (total_points_earned / max_possible_points) * 100
passing = score >= quiz.PassingScore
```

### 5.4 API Endpoints

```
POST   /api/v1/quizzes/:quiz_id/start     - Start a new quiz attempt
POST   /api/v1/attempts/:id/answers       - Submit answer for question
POST   /api/v1/attempts/:id/submit        - Submit completed quiz
GET    /api/v1/attempts/:id                - Get attempt details
GET    /api/v1/attempts/:id/results       - Get attempt results
GET    /api/v1/users/me/attempts          - Get user's quiz attempts
GET    /api/v1/quizzes/:quiz_id/attempts  - Get all attempts for quiz (creator only)
```

### 5.5 Quiz Taking Flow

```
User starts quiz
  ↓
Create quiz_attempt (status: in_progress)
  ↓
Record start time
  ↓
User answers questions
  ↓
For each answer:
  - Save to user_answers
  - Check if correct
  - Calculate points
  ↓
User submits quiz
  ↓
Calculate total score
  ↓
Update attempt (status: completed)
  ↓
Return results with breakdown
```

### 5.6 Time Limit Handling

If quiz has `time_limit_minutes`:
- Track `started_at` timestamp
- Calculate deadline: `started_at + time_limit_minutes`
- On submit, check if within time limit
- Auto-submit if time expires (implement via background job)

### 5.7 Results Response

```json
{
  "attempt_id": "uuid",
  "quiz_title": "Quiz Title",
  "score": 85,
  "passing_score": 70,
  "passed": true,
  "total_points": 17,
  "max_points": 20,
  "time_taken_seconds": 180,
  "completed_at": "2026-01-10T10:30:00Z",
  "questions": [
    {
      "question_text": "What is 2+2?",
      "user_answer": "4",
      "correct_answer": "4",
      "is_correct": true,
      "points_earned": 1,
      "explanation": "2+2 equals 4"
    }
  ]
}
```

**Estimated Time:** 3-4 days

---

## Phase 6: Advanced Features ⭐

**Goal**: Enhance user experience with additional features

### 6.1 Quiz Categories & Tags

**Implementation:**
- Use existing `category` and `tags` fields in quizzes table
- Add category management endpoints
- Add tag search/filtering

**Endpoints:**
```
GET /api/v1/categories           - List all categories
GET /api/v1/tags                 - List popular tags
GET /api/v1/quizzes?category=X   - Filter by category
GET /api/v1/quizzes?tags=X,Y     - Filter by tags
```

### 6.2 Quiz Search & Filtering

**Features:**
- Search by title/description
- Filter by category, difficulty, tags
- Sort by: newest, popular, highest rated
- Pagination

**Endpoint:**
```
GET /api/v1/quizzes/search?q=math&category=education&difficulty=medium&sort=popular&page=1&limit=20
```

### 6.3 Quiz Statistics

**Per Quiz:**
- Total attempts
- Average score
- Completion rate
- Time spent average

**Per User:**
- Total quizzes taken
- Average score
- Total points earned
- Streak (days in a row)

**Endpoints:**
```
GET /api/v1/quizzes/:id/stats    - Quiz statistics
GET /api/v1/users/me/stats       - User statistics
```

### 6.4 Leaderboard

**Database:**
```sql
CREATE TABLE leaderboard_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    best_score INT NOT NULL,
    best_time_seconds INT NOT NULL,
    rank INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(quiz_id, user_id)
);

CREATE INDEX idx_leaderboard_quiz_score ON leaderboard_entries(quiz_id, best_score DESC);
```

**Endpoints:**
```
GET /api/v1/quizzes/:id/leaderboard?limit=100  - Top scores for quiz
GET /api/v1/leaderboard/global?limit=100       - Global leaderboard
```

### 6.5 Quiz Templates

**Feature:** Save quiz structure as template for reuse

**Database:**
```sql
CREATE TABLE quiz_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_data JSONB NOT NULL, -- Full quiz structure
    is_public BOOLEAN DEFAULT false,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Endpoints:**
```
POST /api/v1/templates                    - Save quiz as template
GET  /api/v1/templates                    - List templates
POST /api/v1/templates/:id/use            - Create quiz from template
```

### 6.6 Quiz Sharing

**Feature:** Share quiz via unique link

**Implementation:**
- Generate shareable link: `/share/{short_code}`
- Track shares and views
- Option to make quiz public/private

**Database:**
```sql
ALTER TABLE quizzes ADD COLUMN share_code VARCHAR(20) UNIQUE;
ALTER TABLE quizzes ADD COLUMN share_count INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN view_count INT DEFAULT 0;
```

**Endpoints:**
```
POST /api/v1/quizzes/:id/share           - Generate share link
GET  /share/:code                         - Access shared quiz
```

**Estimated Time:** 3-4 days

---

## Phase 7: Admin Dashboard 👨‍💼

**Goal**: Provide admin interface for managing platform

### 7.1 Admin User Role

**Database:**
```sql
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user';
ALTER TABLE users ADD CONSTRAINT chk_role CHECK (role IN ('user', 'admin', 'moderator'));
```

Migration: `000009_add_user_roles.up.sql`

### 7.2 Admin Middleware

**File:** `internal/middleware/admin.go`
```go
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if userRole != "admin" {
            c.JSON(403, gin.H{"error": "Admin access required"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 7.3 Admin Endpoints

**User Management:**
```
GET    /api/v1/admin/users              - List all users
GET    /api/v1/admin/users/:id          - Get user details
PATCH  /api/v1/admin/users/:id/role     - Change user role
POST   /api/v1/admin/users/:id/disable  - Disable user
POST   /api/v1/admin/users/:id/enable   - Enable user
DELETE /api/v1/admin/users/:id          - Delete user permanently
```

**Quiz Moderation:**
```
GET    /api/v1/admin/quizzes              - List all quizzes
GET    /api/v1/admin/quizzes/:id          - Get quiz details
POST   /api/v1/admin/quizzes/:id/approve  - Approve quiz
POST   /api/v1/admin/quizzes/:id/reject   - Reject quiz
DELETE /api/v1/admin/quizzes/:id          - Delete quiz
```

**Platform Statistics:**
```
GET /api/v1/admin/stats                   - Platform-wide stats
GET /api/v1/admin/stats/users             - User statistics
GET /api/v1/admin/stats/quizzes           - Quiz statistics
GET /api/v1/admin/stats/activity          - Activity logs
```

### 7.4 Activity Logging

**Database:**
```sql
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
```

**Log events:**
- User login/logout
- Quiz created/updated/deleted
- Attempt started/completed
- Admin actions

### 7.5 System Settings

**Database:**
```sql
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Settings:**
- Max file upload size
- AI generation limits
- Rate limits
- Email templates
- Feature flags

**Endpoints:**
```
GET   /api/v1/admin/settings         - Get all settings
PATCH /api/v1/admin/settings/:key    - Update setting
```

**Estimated Time:** 3-4 days

---

## Phase 8: Testing & Production Readiness 🧪

**Goal**: Comprehensive testing and deployment preparation

### 8.1 Unit Tests

**Test coverage targets:**
- Models: 90%+
- DAOs: 85%+
- Business logic: 90%+
- Utilities: 95%+

**Files to create:**
```
user/dao/user_test.go
user/dao/password_test.go
auth/core/auth_test.go
quiz/dao/quiz_test.go
attempt/dao/attempt_test.go
internal/utils/jwt_test.go
internal/utils/password_test.go
```

**Test utilities:**
- Mock database
- Test fixtures
- Helper functions

### 8.2 Integration Tests

**Test scenarios:**
- Complete auth flow (signup → verify → login)
- Quiz creation flow
- Quiz taking flow
- File upload and AI generation
- Error handling

**Files:**
```
tests/integration/auth_test.go
tests/integration/quiz_test.go
tests/integration/attempt_test.go
```

### 8.3 API Documentation

**Swagger/OpenAPI:**
- Install: `github.com/swaggo/gin-swagger`
- Add annotations to handlers
- Generate docs: `swag init`

**Endpoint:** `GET /swagger/index.html`

**Also create:**
- Postman collection export
- curl examples for each endpoint
- API reference markdown

### 8.4 Performance Testing

**Load testing:**
- Use: `github.com/tsenart/vegeta` or `k6`
- Test scenarios:
  - 100 concurrent users
  - 1000 requests/second
  - Database query optimization

**Profiling:**
- Add pprof endpoints
- Identify bottlenecks
- Optimize slow queries

### 8.5 Security Audit

**Checklist:**
- [ ] SQL injection prevention (use parameterized queries)
- [ ] XSS prevention (sanitize inputs)
- [ ] CSRF protection (add CSRF tokens)
- [ ] Rate limiting (implement per-user/IP)
- [ ] Input validation (all endpoints)
- [ ] JWT expiration handling
- [ ] Password strength enforcement
- [ ] Secure file upload validation
- [ ] HTTPS enforcement (production)
- [ ] Secrets management (env variables)

### 8.6 Deployment

**Dockerfile:**
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/pkg/database/migrations ./pkg/database/migrations
EXPOSE 8080
CMD ["./main"]
```

**Docker Compose (production):**
- PostgreSQL with persistent volume
- Backend service
- Nginx reverse proxy
- SSL/TLS certificates

**Cloud deployment options:**
1. **AWS:** ECS + RDS + S3
2. **Google Cloud:** Cloud Run + Cloud SQL + Cloud Storage
3. **DigitalOcean:** App Platform + Managed Database
4. **Heroku:** Simple deployment

### 8.7 CI/CD Pipeline

**GitHub Actions:**
```yaml
name: CI/CD
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test -v ./...

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - run: docker build -t quiz-craft:latest .

  deploy:
    needs: [test, build]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to production
        run: echo "Deploy steps here"
```

### 8.8 Monitoring & Logging

**Add:**
- Application metrics (Prometheus)
- Error tracking (Sentry)
- Log aggregation (ELK stack or CloudWatch)
- Uptime monitoring (UptimeRobot)
- Performance monitoring (New Relic or Datadog)

**Estimated Time:** 5-7 days

---

## Summary Timeline

| Phase | Description | Estimated Time |
|-------|-------------|----------------|
| ✅ Phase 1-2 | Foundation + Auth | COMPLETE |
| Phase 3 | Quiz Management | 3-5 days |
| Phase 4 | AI Integration | 4-6 days |
| Phase 5 | Quiz Attempts | 3-4 days |
| Phase 6 | Advanced Features | 3-4 days |
| Phase 7 | Admin Dashboard | 3-4 days |
| Phase 8 | Testing & Deploy | 5-7 days |
| **Total** | | **21-30 days** |

---

## Recommended Implementation Order

### Week 1: Quiz Management (Phase 3)
- Days 1-2: Database schema + migrations
- Days 3-4: DAOs + Business logic
- Day 5: API endpoints + testing

### Week 2: AI Integration (Phase 4)
- Days 1-2: File upload system
- Days 3-4: Google AI integration
- Day 5: End-to-end testing

### Week 3: Quiz Taking (Phase 5)
- Days 1-2: Database + DAOs
- Days 3-4: Attempt logic + scoring
- Day 5: API endpoints

### Week 4: Polish & Deploy (Phases 6-8)
- Days 1-2: Advanced features
- Days 3-4: Admin dashboard
- Day 5: Testing + deployment

---

## Key Decision Points

### Before Phase 3:
- [ ] Decide on quiz visibility model (all public vs. creator-only)
- [ ] Decide on question type support (start with multiple choice only?)
- [ ] Decide on image support in questions (later phase?)

### Before Phase 4:
- [ ] Choose file storage (local vs. S3 vs. Cloud Storage)
- [ ] Confirm Google AI API access and quotas
- [ ] Decide on AI generation pricing model (free tier limits?)

### Before Phase 5:
- [ ] Decide on quiz retake policy (unlimited vs. limited attempts)
- [ ] Decide on answer review policy (show correct answers immediately?)
- [ ] Decide on leaderboard visibility (opt-in vs. opt-out)

### Before Phase 8:
- [ ] Choose deployment platform
- [ ] Choose monitoring/logging solution
- [ ] Decide on backup strategy
- [ ] Plan domain name and SSL certificate

---

## Next Steps

1. **Review this plan** - Adjust based on priorities
2. **Choose starting point** - Usually Phase 3
3. **Set up task tracking** - Create GitHub issues or use project board
4. **Begin implementation** - Start with database schema
5. **Test continuously** - Don't wait until Phase 8

**Ready to start Phase 3?** Let me know and I'll help you implement it step by step!
